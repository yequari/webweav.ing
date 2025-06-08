DELETE FROM settings WHERE Description='remote_enabled';
DELETE FROM settings WHERE Description='is_visible';
DELETE FROM settings WHERE Description='reenable_comments';
DELETE FROM settings WHERE Description='commenting_enabled';
DELETE FROM settings WHERE Description='local_timezone';

DELETE FROM setting_data_types WHERE Description='boolean';
DELETE FROM setting_data_types WHERE Description='datetime';
DELETE FROM setting_data_types WHERE Description='integer';
DELETE FROM setting_data_types WHERE Description='alphanumeric';

DELETE FROM setting_groups WHERE Description='user';
DELETE FROM setting_groups WHERE Description='guestbook';

DELETE FROM allowed_setting_values WHERE Caption='commenting_enabled';
DELETE FROM allowed_setting_values WHERE Caption='commenting_disabled';
DELETE FROM allowed_setting_values WHERE Caption='guestbook_visible';
DELETE FROM allowed_setting_values WHERE Caption='guestbook_invisible';
DELETE FROM allowed_setting_values WHERE Caption='remote_enabled';
DELETE FROM allowed_setting_values WHERE Caption='remote_disabled';
