-- races
CREATE TABLE race (
    race_id serial primary key,
    name varchar NOT NULL CHECK (name <> ''),
    start_time timestamp,
    end_time timestamp
);

-- index to query event type by name, which must be unique
CREATE UNIQUE INDEX uix_race_name ON race USING btree (name);

-- teams
CREATE TABLE team (
    team_id serial primary key,
    race_id int NOT NULL,
    bib_number int NOT NULL CHECK (bib_number > 0),
    name varchar NOT NULL CHECK (name <> ''),
    gender varchar(1) NOT NULL CHECK (gender <> ''),
    challenge varchar NOT NULL CHECK (challenge <> ''),
    age_category varchar NOT NULL CHECK (age_category <> ''),
    member1_first_name varchar NOT NULL CHECK (member1_first_name <> ''),
    member1_last_name varchar NOT NULL CHECK (member1_last_name <> ''),
    member1_date_of_birth date NOT NULL CHECK (member1_last_name <> ''),
    member1_age_category varchar NOT NULL CHECK (member1_age_category <> ''),
    member1_gender varchar(1) NOT NULL CHECK (member1_date_of_birth > '0001-01-01 00:00:00'),
    member1_club varchar,
    member2_first_name varchar NOT NULL CHECK (member2_first_name <> ''),
    member2_last_name varchar NOT NULL CHECK (member2_last_name <> ''),
    member2_date_of_birth date NOT NULL CHECK (member2_date_of_birth > '0001-01-01 00:00:00'),
    member2_age_category varchar NOT NULL CHECK (member2_age_category <> ''),
    member2_gender varchar(1) NOT NULL CHECK (member2_gender <> ''),
    member2_club varchar
);

-- Add a foreign key constraint to race
ALTER TABLE team add constraint team_race_fk foreign key (race_id) REFERENCES race (race_id);
CREATE UNIQUE INDEX uix_team_bibnumber ON team USING btree (race_id, bib_number);

-- laps
CREATE TABLE lap (
    lap_id serial primary key,
    time timestamp NOT NULL CHECK (time > '0001-01-01 00:00:00'),
    race_id int NOT NULL,
    team_id int NOT NULL
);

-- Add a foreign key constraint to race
ALTER TABLE team add constraint lap_race_fk foreign key (race_id) REFERENCES race (race_id);
-- Add a foreign key constraint to team
ALTER TABLE team add constraint lap_team_fk foreign key (team_id) REFERENCES team (team_id);

-- sample data
insert into race(race_id, name) values (1, 'Bike & Run XS');
insert into race(race_id, name) values (2, 'Bike & Run Jeunes 10-13');
insert into race(race_id, name) values (3, 'Bike & Run Jeunes 6-9');