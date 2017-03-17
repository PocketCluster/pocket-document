# INSERT OR REPLACE only works when a colum has a PK, or UNIQUE property

INSERT OR REPLACE INTO Employee (id, role, name) 
VALUES (1, 'code monkey', (SELECT name FROM Employee WHERE id = 1) );


CREATE TABLE IF NOT EXISTS TEST (ttl BIGINT, key TEXT NOT NULL, path TEXT NOT NULL, value TEXT);
INSERT INTO TEST (ttl, key, path, value) values (1, 'key-0', 'path-0', 'value-0');
INSERT INTO TEST (ttl, key, path, value) values (1, 'key-1', 'path-1', 'value-1');
INSERT INTO TEST (ttl, key, path, value) values (1, 'key-2', 'path-2', 'value-2');

    INSERT INTO TEST (ttl, key, path, value) VALUES (0, (SELECT key FROM TEST WHERE key = 'key-0'), (SELECT path FROM TEST WHERE path = 'path-0'), 'value-000');


INSERT INTO EVENTTYPE (EventTypeName)
SELECT 'ANI Received'
WHERE NOT EXISTS (SELECT 1 FROM EVENTTYPE WHERE EventTypeName = 'ANI Received');