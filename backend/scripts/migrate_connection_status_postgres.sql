-- Migration: Add connection status fields to storage_systems table
-- This migration adds connection monitoring fields to track system health

-- Add connection_status column if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'storage_systems' 
        AND column_name = 'connection_status'
    ) THEN
        ALTER TABLE storage_systems 
        ADD COLUMN connection_status VARCHAR(50) DEFAULT 'unknown';
        RAISE NOTICE 'Added connection_status column';
    ELSE
        RAISE NOTICE 'connection_status column already exists';
    END IF;
END $$;

-- Add last_connection_check column if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'storage_systems' 
        AND column_name = 'last_connection_check'
    ) THEN
        ALTER TABLE storage_systems 
        ADD COLUMN last_connection_check TIMESTAMP;
        RAISE NOTICE 'Added last_connection_check column';
    ELSE
        RAISE NOTICE 'last_connection_check column already exists';
    END IF;
END $$;

-- Add connection_error column if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'storage_systems' 
        AND column_name = 'connection_error'
    ) THEN
        ALTER TABLE storage_systems 
        ADD COLUMN connection_error TEXT;
        RAISE NOTICE 'Added connection_error column';
    ELSE
        RAISE NOTICE 'connection_error column already exists';
    END IF;
END $$;

-- Verify the migration
SELECT 
    column_name, 
    data_type, 
    column_default
FROM information_schema.columns 
WHERE table_name = 'storage_systems' 
AND column_name IN ('connection_status', 'last_connection_check', 'connection_error')
ORDER BY column_name;
