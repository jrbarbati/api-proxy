CREATE TABLE IF NOT EXISTS audit_log (
    id INT NOT NULL AUTO_INCREMENT,
    entity_id INT NOT NULL,
    entity_type VARCHAR(255) NOT NULL,
    performed_by_id INT NOT NULL,
    action VARCHAR(255) NOT NULL,
    created_at TIMESTAMP(6) NOT NULL,

    PRIMARY KEY (id)
)