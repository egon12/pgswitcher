# Pgswitcher Prof of Concept

## How to run it

First you need to build the pgswitcher that can be run by

```sh
make build
```

Then you can run it by run

```sh
make
```

that will run
- docker-compose up
- run pgswitcher
- run the inserter
- switch the connection from old to new

and if you want to clean up (the docker)

you can run 

```sh
make dd
```

## Explanation

The are numbers of services that need to be run this POC
- postgres-main (in docker), is postgres database
  that act as old target. We want to migrate the data
  in this database to postgres-standby
- postgres-standby (in docker) is postgres database
  that act as new target. We want to use this database
  in future.
- pgswitcher is proxy database that can connect to either
  postgres-main and postgres-standby. For the first time,
  usually it will connect to postgres-main, then if we switch
  it through http-request, it will wait until no connection
  to postgres-main execut some script, then it will 
  give new connection to postgres-standby
- inserter is our dummy app that have a task to do insert
  query in 100 RPS

The main actor are pgswitcher.

In here pgswitcher will read config.json that have config like
- old postgres connection
- new postgres connection
- client connection (to set how client will connect to this pgswitcher)
- postgres port
- http port
- sql that should be execute befor we use new sql

For now, we only use one config connection per `connection 
type` (old, new and client)
