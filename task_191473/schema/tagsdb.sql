-- Drop tables if they already exist
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS item_tags;

-- Create table for items
CREATE TABLE items (
                       id INT PRIMARY KEY AUTO_INCREMENT,
                       name VARCHAR(255) NOT NULL,
                       description TEXT,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create table for tags
CREATE TABLE tags (
                      id INT PRIMARY KEY AUTO_INCREMENT,
                      name VARCHAR(100) NOT NULL UNIQUE,
                      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create a join table to link items with tags (Many-to-Many relationship)
CREATE TABLE item_tags (
                           item_id INT,
                           tag_id INT,
                           PRIMARY KEY (item_id, tag_id),
                           FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
                           FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Sample data insertion
-- Insert some sample items
INSERT INTO items (name, description) VALUES ('Sample Item 1', 'Description for Item 1');
INSERT INTO items (name, description) VALUES ('Sample Item 2', 'Description for Item 2');
INSERT INTO items (name, description) VALUES ('Sample Item 3', 'Description for Item 3');

-- Insert some sample tags
INSERT INTO tags (name) VALUES ('Technology');
INSERT INTO tags (name) VALUES ('Science');
INSERT INTO tags (name) VALUES ('Art');
INSERT INTO tags (name) VALUES ('History');

-- Associate items with tags
INSERT INTO item_tags (item_id, tag_id) VALUES (1, 1); -- Item 1 -> Technology
INSERT INTO item_tags (item_id, tag_id) VALUES (1, 2); -- Item 1 -> Science
INSERT INTO item_tags (item_id, tag_id) VALUES (2, 3); -- Item 2 -> Art
INSERT INTO item_tags (item_id, tag_id) VALUES (3, 4); -- Item 3 -> History