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

CREATE TABLE guestbooks (
    Id integer primary key autoincrement,
    ShortId integer UNIQUE NOT NULL,
    SiteUrl varchar(512) NOT NULL,
    UserId blob(16) NOT NULL,
    Created datetime NOT NULL,
    IsDeleted boolean NOT NULL DEFAULT FALSE,
    IsActive boolean NOT NULL DEFAULT TRUE,
    FOREIGN KEY (UserId) REFERENCES users(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

CREATE TABLE guestbook_comments (
    Id integer primary key autoincrement,
    ShortId integer UNIQUE NOT NULL,
    GuestbookId blob(16) NOT NULL,
    ParentId blob(16),
    AuthorName varchar(256) NOT NULL,
    AuthorEmail varchar(256) NOT NULL,
    AuthorSite varchar(256),
    CommentText text NOT NULL,
    PageUrl varchar(256),
    Created datetime NOT NULL,
    IsPublished boolean NOT NULL DEFAULT TRUE,
    IsDeleted boolean NOT NULL DEFAULT FALSE,
    FOREIGN KEY (GuestbookId) 
        REFERENCES guestbooks(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,
    FOREIGN KEY (ParentId)
        REFERENCES guestbook_comments(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);
