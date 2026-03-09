-- Seed more food items
INSERT INTO products (name, description, price, category) VALUES
('Sandwich de Pavo', 'Pechuga de pavo, queso suizo y mayonesa chipotle', 120.00, 'Food'),
('Ensalada César', 'Lechuga romana, croutones y aderezo césar clásico', 95.00, 'Food'),
('Panini Caprese', 'Tomate, mozzarella fresca y pesto de albahaca', 110.00, 'Food'),
('Tostado de Aguacate', 'Pan integral, aguacate machacado y un toque de chile', 85.00, 'Food'),
('Muffin de Arándanos', 'Esponjoso y lleno de arándanos frescos', 45.00, 'Pastry'),
('Galleta con Chispas de Chocolate', 'Horneada al momento, suave por dentro', 35.00, 'Pastry')
ON CONFLICT DO NOTHING;
