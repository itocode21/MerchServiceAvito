CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    coins INTEGER NOT NULL CHECK (coins >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    price INTEGER NOT NULL
);

CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    item_id INTEGER REFERENCES items(id),
    quantity INTEGER NOT NULL CHECK (quantity >= 0)
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    from_user_id UUID REFERENCES users(id),
    to_user_id UUID REFERENCES users(id),
    amount INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO items (name, price) VALUES
('t-shirt', 80), ('cup', 20), ('book', 50), ('pen', 10),
('powerbank', 200), ('hoody', 300), ('umbrella', 200),
('socks', 10), ('wallet', 50), ('pink-hoody', 500);