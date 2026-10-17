-- Decryption must be performed by an explicit data recovery procedure before
-- downgrading the application. This migration only removes the state marker.
DROP TABLE data_crypto_state;
