CREATE TABLE IF NOT EXISTS weekly_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    start_date TEXT NOT NULL UNIQUE,
    end_date TEXT NOT NULL,
    total_commits INTEGER NOT NULL,
    active_days INTEGER NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS daily_commits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    weekly_stats_id INTEGER NOT NULL,
    date TEXT NOT NULL,
    commits INTEGER NOT NULL,
    FOREIGN KEY (weekly_stats_id) REFERENCES weekly_stats(id),
    UNIQUE(weekly_stats_id, date)
);

CREATE TABLE IF NOT EXISTS hourly_activity (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    weekly_stats_id INTEGER NOT NULL,
    hour INTEGER NOT NULL,
    commits INTEGER NOT NULL,
    FOREIGN KEY (weekly_stats_id) REFERENCES weekly_stats(id),
    UNIQUE(weekly_stats_id, hour)
);

CREATE TABLE IF NOT EXISTS repo_details (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    weekly_stats_id INTEGER NOT NULL,
    repo_name TEXT NOT NULL,
    commits INTEGER NOT NULL,
    bar_width INTEGER NOT NULL,
    FOREIGN KEY (weekly_stats_id) REFERENCES weekly_stats(id),
    UNIQUE(weekly_stats_id, repo_name)
);

CREATE TABLE IF NOT EXISTS language_commits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    weekly_stats_id INTEGER NOT NULL,
    language TEXT NOT NULL,
    commits INTEGER NOT NULL,
    is_main BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (weekly_stats_id) REFERENCES weekly_stats(id),
    UNIQUE(weekly_stats_id, language)
);