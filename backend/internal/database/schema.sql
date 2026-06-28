CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	email TEXT UNIQUE NOT NULL,
	username TEXT UNIQUE NOT NULL,
	first_name TEXT,
	last_name TEXT,
	password_hash TEXT NOT NULL,
	avatar_url TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS recipes (
	id TEXT PRIMARY KEY,
	user_id TEXT,
	title TEXT NOT NULL,
	description TEXT,
	prep_time INTEGER,
	cook_time INTEGER,
	servings INTEGER,
	difficulty TEXT CHECK (difficulty IN ('', 'easy', 'medium', 'hard')),
	category TEXT,
	cuisine TEXT,
	image_url TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ingredients (
	id TEXT PRIMARY KEY,
	recipe_id TEXT NOT NULL,
	name TEXT NOT NULL,
	amount TEXT,
	unit TEXT,
	notes TEXT,
	position INTEGER NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS instructions (
	id TEXT PRIMARY KEY,
	recipe_id TEXT NOT NULL,
	text TEXT NOT NULL,
	position INTEGER NOT NULL,
	duration INTEGER,
	temperature INTEGER,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS recipe_tags (
	recipe_id TEXT NOT NULL,
	tag TEXT NOT NULL,
	PRIMARY KEY (recipe_id, tag)
);

CREATE TABLE IF NOT EXISTS collections (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	is_public INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS collection_recipes (
	collection_id TEXT NOT NULL,
	recipe_id TEXT NOT NULL,
	PRIMARY KEY (collection_id, recipe_id)
);
