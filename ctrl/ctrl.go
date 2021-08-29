package ctrl

import (
	"io"
	"log"
	"net"
)

type (
	// Source is caller or user or client that
	Source interface {
		io.ReadWriter
		GetID() uint32
	}

	// SourceFactory will create Source from
	SourceFactory interface {
		New(net.Conn) (Source, error)
	}

	// Target is what we called, or service, or server
	Target interface {
		io.ReadWriter
		Release()
	}

	// TargetPool is the pool that will do cutover
	TargetPool interface {
		// Switch if called with new parameter will block UseNew from be
		// called and until all acquired target from Target Pool is
		// released
		Switch(new bool) error

		// UseNew will block process if in the process of switching
		UseNew() bool

		// Acquire is how we acquire the target. This function won't be
		// blocked
		Acquire() (Target, error)

		// AcquireNew is how we acquire the target from the new Pool.
		// This function won't be blocked
		AcquireNew() (Target, error)

		// Close is called when we want to shutdown all the cutover process
		// Only call this if the process already done
		Close() error
	}

	// target replyer is object that know how ussually the
	// service should be reply. it will be used when
	// the we cannot acquire the target, or the target is broken
	TargetReplyer interface {
		Cancel(Source)
	}

	// Piper is object that will understand the message that sent
	// it will know when to end chat
	Piper interface {
		// this function should be block the process
		// and will return when source starting the chat
		WaitForChat(Source) error
		Chat(Source, Target) error
	}

	// PiperFactory will produce Piper
	PiperFactory interface {
		New() Piper
	}

	// Hooks will be called in the event
	Hooks interface {
		BeforeUseNew(old, new Target) error
	}

	// Controller is the main
	Controller struct {
		sf SourceFactory
		p  TargetPool
		pf PiperFactory
		h  Hooks
		tr TargetReplyer
	}
)

func NewController(sf SourceFactory, p TargetPool, pf PiperFactory, h Hooks, tr TargetReplyer) *Controller {
	return &Controller{sf, p, pf, h, tr}
}

// SwitchToNew will
func (c *Controller) SwitchToNew() error {
	err := c.p.Switch(true) // wait until all acquired released
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = c.p.Switch(false) // rollback
		}
	}()

	newTarget, err := c.p.AcquireNew()
	if err != nil {
		return err
	}

	oldTarget, err := c.p.Acquire()
	if err != nil {
		return err
	}

	err = c.h.BeforeUseNew(oldTarget, newTarget)
	return err
}

func (c *Controller) SwitchToOld() error {
	return c.p.Switch(false)
}

// Handle will receive connection and process it
func (c *Controller) Handle(conn net.Conn) error {
	s, err := c.sf.New(conn)
	if err != nil {
		return err
	}

	go func() {
		err := c.stream(s)
		if err != nil {
			log.Printf("stream error: %v", err)
		}
	}()
	return nil
}

func (c *Controller) Close() error {
	return c.p.Close()
}

// stream have loop forever
func (c *Controller) stream(s Source) error {
	var err error
	p := c.pf.New()
	for {
		err = p.WaitForChat(s)
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		err = c.chat(p, s)
		if err != nil {
			return err
			// do we need error
		}
	}
}

// is when we send the data to target and send back
func (c *Controller) chat(p Piper, s Source) error {
	t, err := c.getTarget()
	if err != nil {
		c.tr.Cancel(s)
		return err
	}

	defer t.Release()

	return p.Chat(s, t)
}

// will get Target and when
func (c *Controller) getTarget() (Target, error) {
	if c.p.UseNew() {
		return c.p.AcquireNew()
	}
	return c.p.Acquire()
}
