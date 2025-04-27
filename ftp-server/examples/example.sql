-- Create a new database (if you haven't already)
CREATE DATABASE shop_db;

-- Switch to the new database
\c shop_db;

-- Create users table
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create products table
CREATE TABLE products (
    product_id SERIAL PRIMARY KEY,
    product_name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock_quantity INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create orders table
CREATE TABLE orders (
    order_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending',
    total_amount DECIMAL(10, 2) NOT NULL
);

-- Create order_items table (many-to-many relationship between orders and products)
CREATE TABLE order_items (
    order_item_id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(order_id) ON DELETE CASCADE,
    product_id INT REFERENCES products(product_id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);


INSERT INTO users (username, email, password_hash)
VALUES
    ('john_doe', 'john@example.com', 'hashedpassword1'),
    ('alice_smith', 'alice@example.com', 'hashedpassword2'),
    ('bob_jones', 'bob@example.com', 'hashedpassword3'),
    ('emma_white', 'emma@example.com', 'hashedpassword4');


    INSERT INTO products (product_name, description, price, stock_quantity)
    VALUES
        ('Laptop', 'A high-performance laptop with 16GB RAM and 512GB SSD', 1200.00, 50),
        ('Wireless Mouse', 'Ergonomic wireless mouse with USB receiver', 25.99, 150),
        ('Keyboard', 'Mechanical keyboard with RGB backlighting', 80.00, 100),
        ('Smartphone', 'Latest model with 128GB storage and 5G support', 899.99, 200),
        ('Headphones', 'Noise-cancelling over-ear headphones', 199.99, 80);

        
        INSERT INTO orders (user_id, total_amount, status)
        VALUES 
            (1, 1280.99, 'completed'),
            (2, 250.99, 'pending'),
            (3, 899.99, 'shipped'),
            (4, 179.99, 'completed');

            
            
            INSERT INTO order_items (order_id, product_id, quantity, price)
            VALUES 
                (1, 1, 1, 1200.00),  -- John bought 1 Laptop
                (1, 2, 1, 25.99),    -- John bought 1 Wireless Mouse
                (2, 3, 1, 80.00),    -- Alice bought 1 Keyboard
                (3, 4, 1, 899.99),   -- Bob bought 1 Smartphone
                (4, 5, 1, 179.99);   -- Emma bought 1 Headphones
