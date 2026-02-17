-- +goose Up
-- Add display_name column to users table (required field)
ALTER TABLE users ADD COLUMN display_name TEXT NOT NULL DEFAULT 'User';

-- Update existing users to have a display name based on their email prefix
UPDATE users 
SET display_name = COALESCE(
    NULLIF(split_part(email, '@', 1), ''),
    'User'
)
WHERE display_name = 'User' OR display_name = '';

-- Add constraint to ensure display name is not empty (must have content after trimming)
ALTER TABLE users ADD CONSTRAINT check_display_name_not_empty 
    CHECK (length(trim(display_name)) > 0 AND length(display_name) <= 100);

-- Remove the default after adding the constraint (force explicit values going forward)
ALTER TABLE users ALTER COLUMN display_name DROP DEFAULT;

-- Create index for efficient lookups
CREATE INDEX idx_users_display_name ON users(display_name) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_users_display_name;
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_display_name_not_empty;
ALTER TABLE users DROP COLUMN IF EXISTS display_name;
