CREATE TABLE households (
  id VARCHAR(64) PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE household_members (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  household_id VARCHAR(64) NOT NULL,
  role_code VARCHAR(32) NOT NULL,
  display_name VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_household_role (household_id, role_code),
  CONSTRAINT fk_member_household FOREIGN KEY (household_id) REFERENCES households(id)
);

CREATE TABLE weight_profiles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  household_id VARCHAR(64) NOT NULL,
  role_code VARCHAR(32) NOT NULL,
  metric_key VARCHAR(64) NOT NULL,
  weight_value DECIMAL(8,2) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_profile_metric (household_id, role_code, metric_key),
  CONSTRAINT fk_weight_household FOREIGN KEY (household_id) REFERENCES households(id)
);

CREATE TABLE houses (
  id VARCHAR(64) PRIMARY KEY,
  household_id VARCHAR(64) NOT NULL,
  community_name VARCHAR(128) NOT NULL,
  listing_name VARCHAR(128) NOT NULL,
  view_date DATE NULL,
  total_price DECIMAL(10,2) NOT NULL DEFAULT 0,
  unit_price DECIMAL(12,2) NOT NULL DEFAULT 0,
  area DECIMAL(10,2) NOT NULL DEFAULT 0,
  house_age DECIMAL(10,2) NOT NULL DEFAULT 0,
  floor_text VARCHAR(64) NOT NULL DEFAULT '',
  orientation VARCHAR(64) NOT NULL DEFAULT '',
  house_type VARCHAR(32) NOT NULL DEFAULT '商品房',
  renovation VARCHAR(32) NOT NULL DEFAULT '简装',
  commute_time DECIMAL(10,2) NOT NULL DEFAULT 0,
  metro_time DECIMAL(10,2) NOT NULL DEFAULT 0,
  monthly_fee DECIMAL(10,2) NOT NULL DEFAULT 0,
  living_convenience DECIMAL(10,2) NOT NULL DEFAULT 6,
  efficiency_rate DECIMAL(10,2) NOT NULL DEFAULT 75,
  light_score DECIMAL(10,2) NOT NULL DEFAULT 6,
  noise_score DECIMAL(10,2) NOT NULL DEFAULT 6,
  layout_score DECIMAL(10,2) NOT NULL DEFAULT 6,
  property_score DECIMAL(10,2) NOT NULL DEFAULT 6,
  community_score DECIMAL(10,2) NOT NULL DEFAULT 6,
  comfort_score DECIMAL(10,2) NOT NULL DEFAULT 6,
  parking_score DECIMAL(10,2) NOT NULL DEFAULT 6,
  notes TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_houses_household (household_id),
  CONSTRAINT fk_house_household FOREIGN KEY (household_id) REFERENCES households(id)
);

CREATE TABLE house_bonus_selections (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  house_id VARCHAR(64) NOT NULL,
  bonus_key VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_house_bonus (house_id, bonus_key),
  CONSTRAINT fk_bonus_house FOREIGN KEY (house_id) REFERENCES houses(id) ON DELETE CASCADE
);

CREATE TABLE house_risk_selections (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  house_id VARCHAR(64) NOT NULL,
  risk_key VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_house_risk (house_id, risk_key),
  CONSTRAINT fk_risk_house FOREIGN KEY (house_id) REFERENCES houses(id) ON DELETE CASCADE
);
