CREATE TABLE IF NOT EXISTS originals (
    name TEXT NOT NULL UNIQUE,
    created INTEGER NOT NULL,
    data BLOB
);

CREATE TABLE IF NOT EXISTS derived (
    original_name TEXT NOT NULL,
    transformation TEXT NOT NULL,
    created INTEGER NOT NULL,
    last_access INTEGER NOT NULL,
    data BLOB,
        
    CONSTRAINT fk_original_name
        FOREIGN KEY (original_name)
        REFERENCES originals(name)
        ON DELETE CASCADE,
        
    CONSTRAINT uk_original_transformation
        UNIQUE (original_name, transformation)
);