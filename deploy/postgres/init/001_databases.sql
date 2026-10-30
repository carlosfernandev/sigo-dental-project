-- Monolithic application: single database with all domains in the public schema.
--
-- Idempotent on purpose: \gexec lets this file be re-run against a volume that
-- already has the database, instead of failing on the first duplicate.
-- The postgres entrypoint only executes it when the data volume is empty.

SELECT 'CREATE DATABASE sigo'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'sigo')\gexec
