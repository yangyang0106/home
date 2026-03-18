ALTER TABLE users
ADD COLUMN is_admin TINYINT(1) NOT NULL DEFAULT 0 AFTER link_code;

UPDATE users
SET is_admin = 1
WHERE id = (
  SELECT id_keep
  FROM (
    SELECT id AS id_keep
    FROM users
    ORDER BY created_at ASC, id ASC
    LIMIT 1
  ) AS seed_admin
);
