CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    coins INTEGER NOT NULL DEFAULT 1000 CHECK (coins >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    price INTEGER NOT NULL
);

CREATE TABLE inventory (
    user_id INTEGER REFERENCES users(id),
    item_id INTEGER REFERENCES items(id),
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    PRIMARY KEY (user_id, item_id)
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    from_user_id INTEGER REFERENCES users(id),
    to_user_id INTEGER REFERENCES users(id),
    amount INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO items (name, price) VALUES
('t-shirt', 80), ('cup', 20), ('book', 50), ('pen', 10),
('powerbank', 200), ('hoody', 300), ('umbrella', 200),
('socks', 10), ('wallet', 50), ('pink-hoody', 500);