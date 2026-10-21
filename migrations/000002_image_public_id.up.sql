ALTER TABLE images ADD COLUMN cloudinary_public_id TEXT;

CREATE UNIQUE INDEX idx_images_public_id
    ON images (cloudinary_public_id)
    WHERE cloudinary_public_id IS NOT NULL;