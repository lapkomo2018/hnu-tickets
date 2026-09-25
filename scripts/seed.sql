INSERT INTO users (name, email, password_hash, role) VALUES
    ('Administrator', 'admin@example.com', crypt('password123', gen_salt('bf')), 'ADMIN'),
    ('Specialist A', 'specialist@example.com', crypt('password123', gen_salt('bf')), 'SPECIALIST'),
    ('Specialist B', 'specialist2@example.com', crypt('password123', gen_salt('bf')), 'SPECIALIST'),
    ('Demo User', 'user@example.com', crypt('password123', gen_salt('bf')), 'USER'),
    ('Demo User 2', 'user2@example.com', crypt('password123', gen_salt('bf')), 'USER')
ON CONFLICT (email) DO NOTHING;

INSERT INTO categories (name, description) VALUES
    ('Software', 'Applications and operating systems'),
    ('Hardware', 'Computers and internal components'),
    ('Network', 'Network and connectivity'),
    ('Accounts', 'User account issues'),
    ('Access Rights', 'Permissions and access requests'),
    ('Peripherals', 'Printers, monitors, and accessories'),
    ('Email', 'Email services'),
    ('Other', 'Other support requests')
ON CONFLICT (name) DO NOTHING;
