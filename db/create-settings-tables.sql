CREATE TABLE setting_groups (
    Id integer primary key autoincrement,
    Description varchar(256) NOT NULL
);

CREATE TABLE setting_data_types (
    Id integer primary key autoincrement,
    Description varchar(64) NOT NULL
);

CREATE TABLE settings (
    Id integer primary key autoincrement,
    Description varchar(256) NOT NULL,
    Constrained boolean NOT NULL,
    DataType integer NOT NULL,
    SettingGroup int NOT NULL,
    MinValue varchar(6),
    MaxValue varchar(6),
    FOREIGN KEY (DataType) REFERENCES setting_data_types(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,
    FOREIGN KEY (SettingGroup) REFERENCES setting_groups(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

CREATE TABLE allowed_setting_values (
    Id integer primary key autoincrement,
    SettingId integer NOT NULL,
    ItemValue varchar(256),
    Caption varchar(256),
    FOREIGN KEY (SettingId) REFERENCES settings(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

CREATE TABLE user_settings (
    Id integer primary key autoincrement,
    UserId integer NOT NULL,
    SettingId integer NOT NULL,
    AllowedSettingValueId integer,
    UnconstrainedValue varchar(256),
    FOREIGN KEY (UserId) REFERENCES users(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,
    FOREIGN KEY (SettingId) REFERENCES settings(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,
    FOREIGN KEY (AllowedSettingValueId) REFERENCES allowed_setting_values(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

CREATE TABLE guestbook_settings (
    Id integer primary key autoincrement,
    GuestbookId integer NOT NULL,
    SettingId integer NOT NULL,
    AllowedSettingValueId integer,
    UnconstrainedValue varchar(256),
    FOREIGN KEY (GuestbookId) REFERENCES guestbooks(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,
    FOREIGN KEY (SettingId) REFERENCES settings(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,
    FOREIGN KEY (AllowedSettingValueId) REFERENCES allowed_setting_values(Id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

INSERT INTO setting_groups (Description) VALUES ('guestbook');
INSERT INTO setting_groups (Description) VALUES ('user');

INSERT INTO setting_data_types (Description) VALUES ('alphanumeric');
INSERT INTO setting_data_types (Description) VALUES ('integer');
INSERT INTO setting_data_types (Description) VALUES ('datetime');
