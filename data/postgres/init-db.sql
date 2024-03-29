-- Taken from golang database-access docs

DROP TABLE IF EXISTS album;

CREATE TABLE album (
  id         SERIAL PRIMARY KEY, -- SERIAL for auto-increment in PostgreSQL
  title      VARCHAR(128) NOT NULL,
  artist     VARCHAR(255) NOT NULL,
  price      NUMERIC(5,2) NOT NULL -- NUMERIC type for decimal values
);

INSERT INTO album (title, artist, price) VALUES
  ('Blue Train', 'John Coltrane', 56.99),
  ('Giant Steps', 'John Coltrane', 63.99),
  ('Jeru', 'Gerry Mulligan', 17.99),
  ('Sarah Vaughan', 'Sarah Vaughan', 34.98);

