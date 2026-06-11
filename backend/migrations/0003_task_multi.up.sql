-- Tasks: support multiple photos and multiple executors.
-- Backfill the new jsonb arrays from the existing single-value columns so old
-- rows keep showing their photo/executor after the UI starts reading the arrays.
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS photos       JSONB DEFAULT '[]'::jsonb;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS executor_ids JSONB DEFAULT '[]'::jsonb;

UPDATE tasks
   SET photos = jsonb_build_array(photo)
 WHERE photo IS NOT NULL
   AND photo <> ''
   AND (photos IS NULL OR photos = '[]'::jsonb);

UPDATE tasks
   SET executor_ids = jsonb_build_array(executor_id)
 WHERE executor_id IS NOT NULL
   AND (executor_ids IS NULL OR executor_ids = '[]'::jsonb);
