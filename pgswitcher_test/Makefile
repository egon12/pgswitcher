default: du run

run:
	sleep 2
	./pgswitcher &
	sleep 2
	go run inserter/main.go &
	sleep 2
	curl localhost:8080/switch
	sleep 2
	killall pgswitcher
	

build:
	go build ..

du:
	docker-compose up -d

dd: 
	docker-compose down

check:
	psql postgres://system:123456@127.0.0.1:5441/trial01\?sslmode=disable -c "select * from table01 order by id;" > master.txt
	psql postgres://system:123456@127.0.0.1:5442/trial01\?sslmode=disable -c "select * from table01 order by id;" > standby.txt
	diff master.txt standby.txt
