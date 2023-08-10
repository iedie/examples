-- This file will provision a new database with the desired schema.

------ Tables ------
 
-- Sessions, a simple table for storing encrypted credentials and expiration
CREATE TABLE sessions (
    id             SERIAL                     PRIMARY KEY, 
    encryptedcreds BYTEA                      NOT NULL,
    expiration     TIMESTAMP WITH TIME ZONE   NOT NULL,
    endoflife      TIMESTAMP WITH TIME ZONE   NOT NULL         
); 

-- Users, a simple table for storing User records
CREATE TABLE users (
    id    SERIAL   PRIMARY KEY,
    first TEXT     NOT NULL,
    last  TEXT     NOT NULL,
    email TEXT     NOT NULL,
)