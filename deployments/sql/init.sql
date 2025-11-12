CREATE TABLE manifests
(
    file_id   UUID PRIMARY KEY,
    completed BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE chunk_locations
(
    file_id     UUID    NOT NULL,
    chunk_index INTEGER NOT NULL,
    repo_index  INTEGER NOT NULL,
    CONSTRAINT fk_manifests
        FOREIGN KEY (file_id)
            REFERENCES manifests (file_id)
            ON DELETE CASCADE
);