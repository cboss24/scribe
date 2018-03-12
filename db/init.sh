#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE USER scribe;
    CREATE DATABASE scribe;
    GRANT ALL PRIVILEGES ON DATABASE scribe TO scribe;

	CREATE SEQUENCE job_id_sequence;
    CREATE TYPE job_status AS ENUM ('SUBMITTED', 'PENDING', 'RUNNABLE', 'STARTING', 'RUNNING', 'SUCCEEDED', 'FAILED');
	CREATE TABLE job (
		job_id 		bigint NOT NULL DEFAULT nextval('job_id_sequence') PRIMARY KEY,
		batch_id	uuid NOT NULL UNIQUE,
		status		job_status NOT NULL
	);
	ALTER SEQUENCE job_id_sequence OWNED BY job.job_id;
EOSQL
