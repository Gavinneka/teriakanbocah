import sqlite3
import os
import json
from datetime import datetime

class LogicDatabase:
    def __init__(self, db_path="logic_history.db"):
        self.db_path = db_path
        self._init_db()

    def _init_db(self):
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS history (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                expression TEXT NOT NULL,
                variables TEXT NOT NULL,
                result_type TEXT NOT NULL,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        ''')
        conn.commit()
        conn.close()

    def add_entry(self, expression, variables, rows):
        # Calculate result type
        results = [row[-1] for row in rows]
        if all(results):
            res_type = "Tautology"
        elif not any(results):
            res_type = "Contradiction"
        else:
            res_type = "Contingency"

        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute('''
            INSERT INTO history (expression, variables, result_type)
            VALUES (?, ?, ?)
        ''', (expression, ", ".join(variables), res_type))
        conn.commit()
        conn.close()

    def get_history(self):
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute('SELECT id, expression, variables, result_type, timestamp FROM history ORDER BY timestamp DESC')
        rows = cursor.fetchall()
        conn.close()
        return rows

    def clear_history(self):
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute('DELETE FROM history')
        conn.commit()
        conn.close()
