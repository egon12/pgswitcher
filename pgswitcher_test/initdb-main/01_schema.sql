CREATE TABLE table01 (id serial primary key, name varchar);

INSERT INTO table01 (name) values('Egon'), ('Marsha');

CREATE PUBLICATION pub01 FOR ALL TABLES;
