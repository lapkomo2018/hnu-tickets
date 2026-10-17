CREATE TABLE data_crypto_state (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id),
    encrypted BOOLEAN NOT NULL DEFAULT FALSE
);
INSERT INTO data_crypto_state (id, encrypted) VALUES (TRUE, FALSE);
