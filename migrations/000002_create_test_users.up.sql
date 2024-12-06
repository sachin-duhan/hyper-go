-- Password for both users is 'password123'
INSERT INTO users (email, password, role) VALUES
('admin@example.com', '$2a$10$zYuWGzO9kLCRxGy5.9HBVeGe3A8bMVSBgZB3zqF9yQ9O4qyXGBcbO', 'admin'),
('user@example.com', '$2a$10$zYuWGzO9kLCRxGy5.9HBVeGe3A8bMVSBgZB3zqF9yQ9O4qyXGBcbO', 'user')
ON CONFLICT (email) DO NOTHING; 