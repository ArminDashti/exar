CREATE TABLE persons (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL
);

CREATE TABLE shops (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE expenses (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  person_id INTEGER NOT NULL,
  shop_id INTEGER NOT NULL,
  item_id INTEGER NOT NULL,
  date TEXT NOT NULL,
  amount REAL NOT NULL,
  FOREIGN KEY (person_id) REFERENCES persons(id),
  FOREIGN KEY (shop_id) REFERENCES shops(id),
  FOREIGN KEY (item_id) REFERENCES items(id)
);

CREATE TABLE expense_shares (
  expense_id INTEGER NOT NULL,
  person_id INTEGER NOT NULL,
  share REAL NOT NULL,
  PRIMARY KEY (expense_id, person_id),
  FOREIGN KEY (expense_id) REFERENCES expenses(id) ON DELETE CASCADE,
  FOREIGN KEY (person_id) REFERENCES persons(id)
);

CREATE INDEX idx_expenses_person ON expenses(person_id);
CREATE INDEX idx_expenses_date ON expenses(date);
CREATE INDEX idx_expenses_shop ON expenses(shop_id);
CREATE INDEX idx_expenses_item ON expenses(item_id);

INSERT INTO persons (id, name) VALUES (1, 'armin');
INSERT INTO persons (id, name) VALUES (2, 'ramin');
