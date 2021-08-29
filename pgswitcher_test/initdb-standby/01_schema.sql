CREATE TABLE table01 (id serial primary key, name varchar);

CREATE SUBSCRIPTION sub01
CONNECTION 'user=system password=123456 host=postgres-main dbname=trial01'
PUBLICATION pub01;

