# Pgswitcher

Postgresql Proxy for switch server

## Why do we build this?

When we migrate our database from your own 
Postgres server into Cloud Postgres Server and
the other way around, we ussually have a long 
downtime. The downtime is caused by manually
do these task:

- turn off the app (if app are cluster then need to
  turn of all of app in the cluster)
- waiting unitl there is no connection to database
- execute some script to set the sequence for the table
- change the config the url in app that point into new database
- turn up the database

*pgswitcher* will act as proxy so we can execute 
all of the task automatically, by providing proxy for 
the postgresql database.

## Getting Started

You can see the example of how it works by see 
(pgswitcher_test)[https://github.com/egon12/pgswitcher/tree/main/pgswitcher_test]

## When to used

This project is still in early development stage
So it still unstable aspect like the proxy doesn't support
cancel request with pid and secret key.

Use it carefully
