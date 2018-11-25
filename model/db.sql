CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- races
CREATE TABLE race (
    race_id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    name varchar NOT NULL,
    start_time timestamp,
    end_time timestamp
);

-- index to query event type by name, which must be unique
CREATE UNIQUE INDEX uix_race_name ON race USING btree (name);

-- teams
CREATE TABLE team (
    team_id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    bib_number varchar NOT NULL,
    name varchar NOT NULL,
    race_id uuid NOT NULL
);

-- Add a foreign key constraint to race
ALTER TABLE team add constraint team_race_fk foreign key (race_id) REFERENCES race (race_id);
CREATE UNIQUE INDEX uix_team_bibnumber ON team USING btree (bib_number);

-- laps
CREATE TABLE lap (
    lap_id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    time timestamp NOT NULL,
    race_id uuid NOT NULL,
    team_id uuid NOT NULL
);

-- Add a foreign key constraint to race
ALTER TABLE team add constraint lap_race_fk foreign key (race_id) REFERENCES race (race_id);
-- Add a foreign key constraint to team
ALTER TABLE team add constraint lap_team_fk foreign key (team_id) REFERENCES team (team_id);

-- sample data
insert into race(name) values ('course adultes');