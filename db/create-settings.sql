INSERT INTO setting_groups (Description) VALUES ("guestbook");
INSERT INTO setting_groups (Description) VALUES ("user");

INSERT INTO setting_data_types (Description) VALUES ("alphanumeric")
INSERT INTO setting_data_types (Description) VALUES ("integer")
INSERT INTO setting_data_types (Description) VALUES ("datetime")

INSERT INTO settings (Description, Constrained, DataType, SettingGroup)
VALUES ("Local Timezone", 0, 1, 1);
