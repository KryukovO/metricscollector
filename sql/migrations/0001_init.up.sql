--
BEGIN TRANSACTION;
--
DO $$
BEGIN
    IF EXISTS(
        SELECT 
            * 
        FROM information_schema.tables 
        WHERE table_name = '__metrics' AND table_schema = 'public'
    )
    THEN
        ALTER TABLE "__metrics" RENAME TO metrics;
    ELSE
        CREATE TABLE IF NOT EXISTS metrics(
            id INT GENERATED ALWAYS AS IDENTITY,
            mname TEXT NOT NULL,
            mtype TEXT NOT NULL,
            delta BIGINT,
            value DOUBLE PRECISION,
            PRIMARY KEY(id),
            UNIQUE(mname, mtype)
        );
    END IF;
    --
    CREATE INDEX IF NOT EXISTS metrics_mname_idx ON metrics USING hash(mname);
    CREATE INDEX IF NOT EXISTS metrics_mtype_idx ON metrics USING hash(mtype);
END $$;
--
--
COMMIT TRANSACTION;
