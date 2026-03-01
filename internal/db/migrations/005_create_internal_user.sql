CREATE TABLE IF NOT EXISTS internal_user (
    id INT NOT NULL AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6),
    inactivated_at TIMESTAMP(6),

    PRIMARY KEY (id),
    CONSTRAINT uq_email UNIQUE (email)
)