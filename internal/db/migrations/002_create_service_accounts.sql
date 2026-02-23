CREATE TABLE IF NOT EXISTS service_account (
    id INT NOT NULL AUTO_INCREMENT,
    org_id INT NOT NULL,
    identifier VARCHAR(255),
    client_id VARCHAR(255),
    client_secret VARCHAR(255),
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6),
    inactivated_at TIMESTAMP(6),

    PRIMARY KEY (id),
    CONSTRAINT fk_service_account_org FOREIGN KEY (org_id) REFERENCES org(id),
    CONSTRAINT uq_service_account_identifier UNIQUE (identifier)
);
