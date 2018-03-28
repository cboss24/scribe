#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE USER scribe;
    CREATE DATABASE scribe;
    GRANT ALL PRIVILEGES ON DATABASE scribe TO scribe;
    \connect scribe;

	CREATE SEQUENCE job_id_sequence;
    CREATE TYPE job_status AS ENUM ('SUBMITTED', 'PENDING', 'RUNNABLE', 'STARTING', 'RUNNING', 'SUCCEEDED', 'FAILED');
	CREATE TABLE job (
		id 				bigint NOT NULL DEFAULT nextval('job_id_sequence') PRIMARY KEY,
		attempts		text NOT NULL,
		container		text NOT NULL,
		created_at 		timestamptz NOT NULL,
		depends_on		text NOT NULL,
		job_definition 	text NOT NULL,
		job_id 			uuid NOT NULL UNIQUE,
		job_name		text NOT NULL,
		job_queue		text NOT NULL,
		last_changed 	timestamptz NOT NULL,
		parameters		text NOT NULL,
		retry_strategy	text,
		started_at 		timestamptz,
		status 			job_status NOT NULL,
		status_reason	text,
		stopped_at		timestamptz
	);
	ALTER SEQUENCE job_id_sequence OWNED BY job.id;
EOSQL
