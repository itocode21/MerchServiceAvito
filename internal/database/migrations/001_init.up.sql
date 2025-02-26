-- 0001_init.up.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    coins INT NOT NULL DEFAULT 1000,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    price INT NOT NULL
);

CREATE TABLE inventory (
    user_id INT REFERENCES users(id),
    item_id INT REFERENCES items(id),
    quantity INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, item_id)
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    from_user_id INT REFERENCES users(id),
    to_user_id INT REFERENCES users(id),
    amount INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_users_username ON users (username);
CREATE INDEX idx_inventory_user_id ON inventory (user_id);
CREATE INDEX idx_transactions_from_user_id ON transactions (from_user_id);
CREATE INDEX idx_transactions_to_user_id ON transactions (to_user_id);