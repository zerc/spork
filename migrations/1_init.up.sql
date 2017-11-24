CREATE SEQUENCE serial START 1;

CREATE TABLE urls (
    id  integer PRIMARY KEY DEFAULT nextval('serial'),
    original    text,
    hash        varchar(100)
)
