CREATE INDEX websites_users ON websites(UserId);
CREATE INDEX guestbooks_websites ON guestbooks(WebsiteId);
CREATE INDEX comments_guestbooks ON guestbook_comments(GuestbookId);
CREATE INDEX web_settings_websites ON guestbook_settings(GuestbookId);
CREATE INDEX user_settings_users ON user_settings(UserId);