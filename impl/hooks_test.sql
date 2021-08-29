SELECT setval('table_01_id_seq', (SELECT max(id) FROM table_01))
