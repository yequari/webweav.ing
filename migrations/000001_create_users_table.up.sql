CREATE TABLE users (
    Id integer primary key autoincrement,
    ShortId integer UNIQUE NOT NULL,
    Username varchar(32) NOT NULL,
    Email varchar(256) UNIQUE NOT NULL,
    Deleted datetime,
    IsBanned boolean NOT NULL DEFAULT FALSE,
    HashedPassword char(60) NOT NULL,
    Created datetime NOT NULL
);

CREATE TABLE sessions (
    token CHAR(43) primary key,
    data BLOB NOT NULL,
    expiry TEXT NOT NULL
);
