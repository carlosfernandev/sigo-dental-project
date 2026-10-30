-- Enable btree_gist extension for EXCLUDE constraints (used in scheduling)
-- This script runs after 001_databases.sql creates the sigo database
\c sigo
CREATE EXTENSION IF NOT EXISTS btree_gist;
