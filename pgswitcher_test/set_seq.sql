SELECT setval('table01_id_seq', (SELECT max(id) FROM table01))
