CREATE TABLE IF NOT EXISTS route (
    id INT NOT NULL AUTO_INCREMENT,
    pattern VARCHAR(255) NOT NULL,
    backend_url VARCHAR(255) NOT NULL,
    method VARCHAR(15) NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6),
    inactivated_at TIMESTAMP(6),

    PRIMARY KEY (id)
);
