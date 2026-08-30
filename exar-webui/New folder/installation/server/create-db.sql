CREATE DATABASE IF NOT EXISTS NetChecker CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE NetChecker;

CREATE TABLE IF NOT EXISTS countries (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    `country-name` VARCHAR(128) NOT NULL,
    UNIQUE KEY uk_country_name (`country-name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS cities (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    `city-name` VARCHAR(128) NOT NULL,
    `country-id` BIGINT NOT NULL,
    UNIQUE KEY uk_city_country (`city-name`, `country-id`),
    CONSTRAINT fk_cities_country
        FOREIGN KEY (`country-id`) REFERENCES countries(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS provider (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    `provider-name` VARCHAR(128) NOT NULL,
    UNIQUE KEY uk_provider_name (`provider-name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS hosts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    host VARCHAR(255) NOT NULL,
    `host-type` INT NOT NULL DEFAULT 1,
    UNIQUE KEY uk_hosts_host (host)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `http-icmp-latency` (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    `host-id` BIGINT NOT NULL,
    `resolved-ip` VARCHAR(45),
    `http-latency` INT,
    `icmp-latency` INT,
    `checked-at` DATETIME NOT NULL,
    CONSTRAINT fk_latency_host
        FOREIGN KEY (`host-id`) REFERENCES hosts(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    INDEX idx_latency_host_checked (`host-id`, `checked-at`),
    INDEX idx_latency_checked (`checked-at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
