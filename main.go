package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/egon12/pgswitcher/config"
	"github.com/egon12/pgswitcher/ctrl"
	"github.com/egon12/pgswitcher/impl"
)

func main() {
	tp, err := impl.NewTargetPool(
		config.C.Old[0],
		config.C.New[0],
		false,
	)
	if err != nil {
		panic(err)
	}

	hijacked, release, err := tp.HijackOne()
	if err != nil {
		panic(err)
	}

	sourceFactory, err := impl.NewSourceFactory(config.C.Client[0], hijacked)
	if err != nil {
		panic(err)
	}
	release()

	hooks, err := impl.NewHooks(config.C.ExecuteBeforeUseNewSQL)
	if err != nil {
		panic(err)
	}

	controller := ctrl.NewController(sourceFactory, tp, impl.PiperFactory, hooks, impl.TargetReplyer)
	defer controller.Close()

	l, err := net.Listen("tcp", config.C.Listen)
	if err != nil {
		panic(err)
	}

	sigterm := make(chan os.Signal)
	signal.Notify(sigterm, os.Interrupt)

	go handle(l, controller)

	go serveHTTP(controller)

	<-sigterm
}

func handle(l net.Listener, controller *ctrl.Controller) {
	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		err = controller.Handle(c)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func serveHTTP(controller *ctrl.Controller) {
	var handlerFunc http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/switch" {
			err := controller.SwitchToNew()
			if err != nil {
				w.Write([]byte("hore\n" + err.Error()))

			}
			w.Write([]byte("hore\n1234"))
			return
		}

		w.Write([]byte(r.URL.Path))
		return
	}
	http.ListenAndServe(config.C.HTTPListen, handlerFunc)

}
