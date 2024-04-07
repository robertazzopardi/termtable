-- Users Table
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  username VARCHAR(50) NOT NULL UNIQUE,
  email VARCHAR(100) NOT NULL UNIQUE,
  password_hash CHAR(60) NOT NULL
);

-- Products Table
CREATE TABLE products (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  price DECIMAL(10,2) NOT NULL,
  stock INTEGER NOT NULL DEFAULT 0
);

-- Orders Table
CREATE TABLE orders (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  total_price DECIMAL(10,2) NOT NULL
);

-- Order Items Table (many-to-many relationship between Orders and Products)
CREATE TABLE order_items (
  order_id INTEGER NOT NULL REFERENCES orders(id),
  product_id INTEGER NOT NULL REFERENCES products(id),
  quantity INTEGER NOT NULL DEFAULT 1,
  PRIMARY KEY (order_id, product_id)
);

-- Categories Table (for product categorization)
CREATE TABLE categories (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE
);

-- Add Category Foreign Key to Products Table
ALTER TABLE products
ADD COLUMN category_id INTEGER REFERENCES categories(id);

-- Insert Users
INSERT INTO users (username, email, password_hash)
VALUES ('john_doe', 'john.doe@example.com', '$2y$10$...'),  -- Replace with hashed passwords
       ('jane_smith', 'jane.smith@example.com', '$2y$10$...'),
       ('alice_cooper', 'alice.cooper@example.com', '$2y$10$...'),
       ('bob_marley', 'bob.marley@example.com', '$2y$10$...'),
       ('charlie_chaplin', 'charlie.chaplin@example.com', '$2y$10$...'),
       ('david_bowie', 'david.bowie@example.com', '$2y$10$...'),
       ('einstein', 'albert.einstein@example.com', '$2y$10$...'),
       ('marie_curie', 'marie.curie@example.com', '$2y$10$...'),
       ('friedrich_nietzsche', 'friedrich.nietzsche@example.com', '$2y$10$...'),
       ('stephen_hawking', 'stephen.hawking@example.com', '$2y$10$...');

-- Insert Categories (already exists from previous data)
INSERT INTO categories (name)
VALUES ('Clothing'),
      ('Kitchen'),
      ('Electronics');

-- Insert Products
INSERT INTO products (name, description, price, stock, category_id)
VALUES ('T-Shirt', 'Comfortable cotton T-Shirt', 19.99, 100, 1),
       ('Coffee Mug', 'Great for your morning coffee', 9.99, 50, 2),
       ('Wireless Headphones', 'High-quality sound for music and calls', 79.99, 25, 3),
       ('Laptop Sleeve', 'Protects your laptop in style', 14.99, 75, 3),
       ('Travel Mug', 'Keeps your drink hot or cold on the go', 16.99, 30, 2),
       ('Running Shoes', 'Comfortable shoes for your workout', 69.99, 40, 1),
       ('Baseball Cap', 'Sporty cap for any occasion', 12.99, 20, 1),
       ('Mouse Pad', 'Improves mouse control and comfort', 7.99, 50, 3),
       ('Water Bottle', 'Stay hydrated throughout the day', 12.99, 60, 2),
       ('Backpack', 'Spacious and comfortable for everyday use', 49.99, 20, 1),
       ('Keyboard', 'Ergonomic design for improved typing comfort', 39.99, 35, 3),
       ('Desk Lamp', 'Provides adjustable lighting for your workspace', 24.99, 45, 3),
       ('Notebook', 'Lined pages for writing and note-taking', 7.99, 80, 1),
       ('Pen Set', 'Includes a variety of pens for different writing needs', 9.99, 25, 1);

-- Insert Orders (assuming Users table has data)
INSERT INTO orders (user_id, created_at, total_price)
VALUES (1, CURRENT_TIMESTAMP, 42.98),
       (2, CURRENT_TIMESTAMP, 19.99),
       (3, CURRENT_TIMESTAMP, 29.99),
       (4, CURRENT_TIMESTAMP, 89.99),
       (5, CURRENT_TIMESTAMP, 24.99),
       (6, CURRENT_TIMESTAMP, 57.98),
       (7, CURRENT_TIMESTAMP, 32.98),
       (8, CURRENT_TIMESTAMP, 14.99);

-- Insert Order Items (assuming Orders and Products tables have data)
INSERT INTO order_items (order_id, product_id, quantity)
VALUES (1, 1, 2),  -- Order 1: 2 T-Shirts
       (1, 3, 1),  -- Order 1: 1 Running Shoes
       (2, 2, 1),  -- Order 2: 1 Coffee Mug
       (2, 8, 1),  -- Order 2: 1 Mouse Pad
       (3, 4, 1),  -- Order 3: 1 Baseball Cap
       (3, 7, 2),  -- Order 3: 2 Travel Mugs
       (4, 5, 1),  -- Order 4: 1 Wireless Headphones
       (4, 6, 1),  -- Order 4: 1 Laptop Sleeve
       (5, 1, 1),  -- Order 5: 1 T-Shirt
       (5, 9, 1),  -- Order 5: 1 Water Bottle
       (6, 3, 2),  -- Order 6: 2 Running Shoes
       (6, 10, 1),  -- Order 6: 1 Backpack
       (7, 2, 1),  -- Order 7: 1 Coffee Mug
       (7, 11, 1),  -- Order 7: 1 Keyboard
       (8, 12, 2),  -- Order 8: 2 Notebooks
       (8, 13, 1);  -- Order 8: 1 Pen Set

