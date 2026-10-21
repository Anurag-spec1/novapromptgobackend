DROP INDEX IF EXISTS idx_images_public_id;
ALTER TABLE images DROP COLUMN IF EXISTS cloudinary_public_id;