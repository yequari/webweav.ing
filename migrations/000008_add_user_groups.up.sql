CREATE TABLE IF NOT EXISTS groups (
    Id integer PRIMARY KEY NOT NULL,
    Description varchar(256)
);

INSERT INTO groups (Id, Description)
    VALUES (1, 'admin'),(2, 'user');

CREATE TABLE IF NOT EXISTS users_groups (
    Id integer primary key autoincrement,
    UserId integer NOT NULL,
    GroupId integer NOT NULL,
    FOREIGN KEY (UserId) REFERENCES users(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,
    FOREIGN KEY (GroupId) REFERENCES groups(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);
