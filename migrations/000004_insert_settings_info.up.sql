INSERT INTO setting_groups (Description) VALUES ('guestbook');
INSERT INTO setting_groups (Description) VALUES ('user');

INSERT INTO setting_data_types (Description) VALUES ('alphanumeric');
INSERT INTO setting_data_types (Description) VALUES ('integer');
INSERT INTO setting_data_types (Description) VALUES ('datetime');
INSERT INTO setting_data_types (Description) VALUES ('boolean');

INSERT INTO settings (Description, Constrained, DataType, SettingGroup)
    SELECT 'local_timezone', 0, t.id, g.id FROM setting_data_types as t INNER JOIN setting_groups AS g WHERE g.Description='user' AND t.Description='alphanumeric';

INSERT INTO settings (Description, Constrained, DataType, SettingGroup)
    SELECT 'commenting_enabled', 1, t.id, g.id FROM setting_data_types as t INNER JOIN setting_groups AS g WHERE g.Description='guestbook' AND t.Description='boolean';

INSERT INTO allowed_setting_values (SettingId, ItemValue, Caption)
    SELECT settings.id, 'true', 'commenting_enabled' FROM settings WHERE settings.Description='commenting_enabled';

INSERT INTO allowed_setting_values (SettingId, ItemValue, Caption)
    SELECT settings.id, 'false', 'commenting_disabled' FROM settings WHERE settings.Description='commenting_enabled';

INSERT INTO settings (Description, Constrained, DataType, SettingGroup)
    SELECT 'reenable_comments', 0, t.id, g.id FROM setting_data_types as t INNER JOIN setting_groups as g WHERE g.Description='guestbook' AND t.Description='alphanumeric';

INSERT INTO settings (Description, Constrained, DataType, SettingGroup)
    SELECT 'is_visible', 1, t.id, g.id FROM setting_data_types as t INNER JOIN setting_groups as g WHERE g.Description='guestbook' AND t.Description='boolean';

INSERT INTO allowed_setting_values (SettingId, ItemValue, Caption)
    SELECT settings.id, 'true', 'guestbook_visible' FROM settings WHERE settings.Description='is_visible';

INSERT INTO allowed_setting_values (SettingId, ItemValue, Caption)
    SELECT settings.id, 'false', 'guestbook_invisible' FROM settings WHERE settings.Description='is_visible';

INSERT INTO settings (Description, Constrained, DataType, SettingGroup)
    SELECT 'remote_enabled', 1, t.id, g.id FROM setting_data_types as t INNER JOIN setting_groups as g WHERE g.Description='guestbook' AND t.Description='boolean';

INSERT INTO allowed_setting_values (SettingId, ItemValue, Caption)
    SELECT settings.id, 'true', 'remote_enabled' FROM settings WHERE settings.Description='remote_enabled';

INSERT INTO allowed_setting_values (SettingId, ItemValue, Caption)
    SELECT settings.id, 'false', 'remote_disabled' FROM settings WHERE settings.Description='remote_enabled';
