CREATE TABLE sessions (
    token CHAR(43) primary key,
    data BLOB NOT NULL,
    expiry TEXT NOT NULL
);
