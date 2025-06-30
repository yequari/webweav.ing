ALTER TABLE users RENAME COLUMN HashedPassword TO HashedPasswordOld;
ALTER TABLE users ADD COLUMN HashedPassword char(60) NULL;
UPDATE users SET HashedPassword=HashedPasswordOld;
ALTER TABLE users DROP COLUMN HashedPasswordOld;
