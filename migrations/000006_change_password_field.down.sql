ALTER TABLE users RENAME COLUMN HashedPassword TO HashedPasswordOld;
ALTER TABLE users ADD COLUMN HashedPassword char(60) NOT NULL DEFAULT '0000';
UPDATE users SET HashedPassword=HashedPasswordOld;
ALTER TABLE users DROP COLUMN HashedPasswordOld;
