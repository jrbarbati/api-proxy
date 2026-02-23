CREATE TABLE IF NOT EXISTS rate_limit (
    id INT NOT NULL AUTO_INCREMENT,
    org_id INT NOT NULL,
    service_account_id INT NULL,
    limit_per_minute INT NOT NULL,
    limit_per_day INT NOT NULL,
    limit_per_month INT NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6),
    inactivated_at TIMESTAMP(6),

    PRIMARY KEY (id),
    CONSTRAINT fk_rate_limit_org FOREIGN KEY (org_id) REFERENCES org(id),
    CONSTRAINT fk_rate_limit_service_account FOREIGN KEY (service_account_id) REFERENCES service_account(id),
    CONSTRAINT uq_rate_limit_org UNIQUE (org_id, service_account_id)
);
