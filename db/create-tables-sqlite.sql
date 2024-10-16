CREATE TABLE users (
    Id blob(16) primary key,
    Username varchar(32) NOT NULL,
    Email varchar(256) NOT NULL,
    IsDeleted boolean NOT NULL DEFAULT FALSE
) WITHOUT ROWID;

CREATE TABLE guestbooks (
    Id blob(16) primary key,
    SiteUrl varchar(512) NOT NULL,
    UserId blob(16) NOT NULL,
    IsDeleted boolean NOT NULL DEFAULT FALSE,
    IsActive boolean NOT NULL DEFAULT TRUE,
    FOREIGN KEY (UserId) REFERENCES users(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

CREATE TABLE guestbook_comments (
    Id blob(16) primary key,
    GuestbookId blob(16) NOT NULL,
    ParentId blob(16),
    AuthorName varchar(256) NOT NULL,
    AuthorEmail varchar(256) NOT NULL,
    AuthorSite varchar(256),
    CommentText text NOT NULL,
    PageUrl varchar(256),
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
