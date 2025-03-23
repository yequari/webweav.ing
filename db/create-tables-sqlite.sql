CREATE TABLE users (
    Id integer primary key autoincrement,
    ShortId integer UNIQUE NOT NULL,
    Username varchar(32) NOT NULL,
    Email varchar(256) UNIQUE NOT NULL,
    IsDeleted boolean NOT NULL DEFAULT FALSE,
    IsBanned boolean NOT NULL DEFAULT FALSE,
    HashedPassword char(60) NOT NULL,
    Created datetime NOT NULL
);

CREATE TABLE websites (
    Id integer primary key autoincrement,
    ShortId integer UNIQUE NOT NULL,
    Name varchar(256) NOT NULL,
    SiteUrl varchar(512) NOT NULL,
    AuthorName varchar(512) NOT NULL,
    UserId integer NOT NULL,
    Created datetime NOT NULL,
    Deleted datetime,
    FOREIGN KEY (UserId) REFERENCES users(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

CREATE TABLE guestbooks (
    Id integer primary key autoincrement,
    ShortId integer UNIQUE NOT NULL,
    WebsiteId integer UNIQUE NOT NULL,
    UserId integer NOT NULL,
    Created datetime NOT NULL,
    Deleted datetime,
    IsActive boolean NOT NULL DEFAULT TRUE,
    FOREIGN KEY (UserId) REFERENCES users(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
    FOREIGN KEY (WebsiteId) REFERENCES websites(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

CREATE TABLE guestbook_comments (
    Id integer primary key autoincrement,
    ShortId integer UNIQUE NOT NULL,
    GuestbookId integer NOT NULL,
    ParentId integer,
    AuthorName varchar(256) NOT NULL,
    AuthorEmail varchar(256) NOT NULL,
    AuthorSite varchar(256),
    CommentText text NOT NULL,
    PageUrl varchar(256),
    Created datetime NOT NULL,
    IsPublished boolean NOT NULL DEFAULT TRUE,
    Deleted datetime,
    FOREIGN KEY (GuestbookId) 
        REFERENCES guestbooks(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,
    FOREIGN KEY (ParentId)
        REFERENCES guestbook_comments(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

CREATE TABLE sessions (
    token CHAR(43) primary key,
    data BLOB NOT NULL,
    expiry TEXT NOT NULL
);
