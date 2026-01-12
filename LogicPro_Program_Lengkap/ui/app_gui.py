import customtkinter as ctk
from core.logic_engine import LogicEngine
from core.database import LogicDatabase
from datetime import datetime
import tkinter as tk
from tkinter import messagebox

from PIL import Image
import os

class LogicApp(ctk.CTk):
    def __init__(self):
        super().__init__()

        self.title("LogicPro - Media Pembelajaran Logika")
        self.geometry("1100x700")
        
        # Paths
        self.base_path = os.path.dirname(os.path.abspath(__file__))
        self.assets_path = os.path.join(os.path.dirname(self.base_path), "assets")
        
        # Initialize Core
        self.engine = LogicEngine()
        self.db = LogicDatabase(os.path.join(os.path.dirname(self.base_path), "logic_history.db"))
        
        # Configure Grid
        self.grid_columnconfigure(1, weight=1)
        self.grid_rowconfigure(0, weight=1)

        # Sidebar
        self.sidebar_frame = ctk.CTkFrame(self, width=200, corner_radius=0)
        self.sidebar_frame.grid(row=0, column=0, sticky="nsew")
        self.sidebar_frame.grid_rowconfigure(4, weight=1)

        self.logo_label = ctk.CTkLabel(self.sidebar_frame, text="LOGIC PRO", font=ctk.CTkFont(size=24, weight="bold"))
        self.logo_label.grid(row=0, column=0, padx=20, pady=(20, 10))

        self.dashboard_button = ctk.CTkButton(self.sidebar_frame, text="Dashboard", command=self.show_dashboard)
        self.dashboard_button.grid(row=1, column=0, padx=20, pady=10)

        self.input_button = ctk.CTkButton(self.sidebar_frame, text="Input Logika", command=self.show_input)
        self.input_button.grid(row=2, column=0, padx=20, pady=10)

        self.history_button = ctk.CTkButton(self.sidebar_frame, text="Riwayat", command=self.show_history)
        self.history_button.grid(row=3, column=0, padx=20, pady=10)

        self.theory_button = ctk.CTkButton(self.sidebar_frame, text="Materi Logika", command=self.show_theory)
        self.theory_button.grid(row=4, column=0, padx=20, pady=10)

        self.appearance_mode_label = ctk.CTkLabel(self.sidebar_frame, text="Appearance:", anchor="w")
        self.appearance_mode_label.grid(row=5, column=0, padx=20, pady=(10, 0))
        self.appearance_mode_optionemenu = ctk.CTkOptionMenu(self.sidebar_frame, values=["Dark", "Light", "System"],
                                                                       command=self.change_appearance_mode_event)
        self.appearance_mode_optionemenu.grid(row=6, column=0, padx=20, pady=(10, 10))

        self.exit_button = ctk.CTkButton(self.sidebar_frame, text="Keluar", fg_color="transparent", border_width=2, text_color=("gray10", "#DCE4EE"), command=self.quit)
        self.exit_button.grid(row=7, column=0, padx=20, pady=20)

        # Main Content Area
        self.main_view = ctk.CTkFrame(self, corner_radius=0, fg_color="transparent")
        self.main_view.grid(row=0, column=1, padx=20, pady=20, sticky="nsew")
        self.main_view.grid_columnconfigure(0, weight=1)
        self.main_view.grid_rowconfigure(0, weight=1)

        self.show_dashboard()

    def clear_view(self):
        for widget in self.main_view.winfo_children():
            widget.destroy()

    def change_appearance_mode_event(self, new_appearance_mode: str):
        ctk.set_appearance_mode(new_appearance_mode)

    def show_dashboard(self):
        self.clear_view()
        dash_frame = ctk.CTkFrame(self.main_view, fg_color="transparent")
        dash_frame.pack(expand=True, fill="both")

        # Load Icon
        try:
            icon_image = ctk.CTkImage(light_image=Image.open(os.path.join(self.assets_path, "icon.png")),
                                     dark_image=Image.open(os.path.join(self.assets_path, "icon.png")),
                                     size=(200, 200))
            icon_label = ctk.CTkLabel(dash_frame, image=icon_image, text="")
            icon_label.pack(pady=(50, 20))
        except Exception:
            pass

        welcome_label = ctk.CTkLabel(dash_frame, text="LOGIC PRO", font=ctk.CTkFont(size=48, weight="bold"))
        welcome_label.pack(pady=10)

        subtitle_label = ctk.CTkLabel(dash_frame, text="Media Pembelajaran Logika Matematika Modern", font=ctk.CTkFont(size=20))
        subtitle_label.pack(pady=10)

        btn_container = ctk.CTkFrame(dash_frame, fg_color="transparent")
        btn_container.pack(pady=40)

        ctk.CTkButton(btn_container, text="Mulai Perhitungan", width=250, height=60, font=ctk.CTkFont(size=16, weight="bold"), command=self.show_input).grid(row=0, column=0, padx=15)
        ctk.CTkButton(btn_container, text="Materi Pembelajaran", width=250, height=60, font=ctk.CTkFont(size=16, weight="bold"), command=self.show_theory).grid(row=0, column=1, padx=15)

    def show_input(self):
        self.clear_view()
        input_frame = ctk.CTkFrame(self.main_view)
        input_frame.pack(fill="both", expand=True, padx=20, pady=20)

        title = ctk.CTkLabel(input_frame, text="Input Ekspresi Logika", font=ctk.CTkFont(size=24, weight="bold"))
        title.pack(pady=20)

        guide = ctk.CTkLabel(input_frame, text="Operator: ! (NOT), & (AND), | (OR), -> (IF), <-> (IFF)\nContoh: (P & Q) -> R", font=ctk.CTkFont(size=14))
        guide.pack(pady=5)

        self.expr_entry = ctk.CTkEntry(input_frame, width=500, placeholder_text="Masukkan ekspresi di sini...")
        self.expr_entry.pack(pady=20)
        self.expr_entry.bind("<Return>", lambda e: self.calculate())

        calc_btn = ctk.CTkButton(input_frame, text="Hitung Tabel Kebenaran", command=self.calculate)
        calc_btn.pack(pady=10)

        # Table Container
        self.scrollable_frame = ctk.CTkScrollableFrame(input_frame, label_text="Hasil Tabel Kebenaran")
        self.scrollable_frame.pack(fill="both", expand=True, pady=20)

    def calculate(self):
        expr = self.expr_entry.get()
        if not expr:
            messagebox.showwarning("Warning", "Ekspresi tidak boleh kosong!")
            return

        table, err = self.engine.generate_truth_table(expr)
        if err:
            messagebox.showerror("Error", f"Gagal memproses ekspresi: {err}")
            return

        # Clear table
        for widget in self.scrollable_frame.winfo_children():
            widget.destroy()

        # Render Header
        for i, col in enumerate(table['header']):
            h = ctk.CTkLabel(self.scrollable_frame, text=col, font=ctk.CTkFont(weight="bold"), 
                            fg_color="#1f538d" if i == len(table['header'])-1 else "transparent",
                            corner_radius=5)
            h.grid(row=0, column=i, padx=10, pady=5, sticky="nsew")

        # Render Rows
        for r, row in enumerate(table['rows']):
            for c, val in enumerate(row):
                color = "#2fa572" if val is True else "#c0392b"
                text = "T" if val is True else "F"
                l = ctk.CTkLabel(self.scrollable_frame, text=text, text_color=color, font=ctk.CTkFont(weight="bold"))
                l.grid(row=r+1, column=c, padx=10, pady=2)

        # Save to DB
        self.db.add_entry(expr, table['variables'], table['rows'])

    def show_history(self):
        self.clear_view()
        hist_frame = ctk.CTkFrame(self.main_view)
        hist_frame.pack(fill="both", expand=True, padx=20, pady=20)

        title = ctk.CTkLabel(hist_frame, text="Riwayat Perhitungan", font=ctk.CTkFont(size=24, weight="bold"))
        title.pack(pady=20)

        self.hist_scroll = ctk.CTkScrollableFrame(hist_frame)
        self.hist_scroll.pack(fill="both", expand=True, pady=10)

        history_data = self.db.get_history()
        
        # Header
        cols = ["Ekspresi", "Variabel", "Tipe", "Waktu"]
        for i, col in enumerate(cols):
            h = ctk.CTkLabel(self.hist_scroll, text=col, font=ctk.CTkFont(weight="bold"))
            h.grid(row=0, column=i, padx=20, pady=10, sticky="w")

        for r, data in enumerate(history_data):
            for c, val in enumerate(data[1:]): # Skip ID
                l = ctk.CTkLabel(self.hist_scroll, text=str(val))
                l.grid(row=r+1, column=c, padx=20, pady=5, sticky="w")

        clear_btn = ctk.CTkButton(hist_frame, text="Bersihkan Riwayat", fg_color="#c0392b", hover_color="#962d22", command=self.clear_history)
        clear_btn.pack(pady=20)

    def show_theory(self):
        self.clear_view()
        theory_frame = ctk.CTkScrollableFrame(self.main_view, label_text="Materi Logika Matematika")
        theory_frame.pack(fill="both", expand=True, padx=20, pady=20)

        materi = [
            ("Negasi (NOT)", "Simbol: !, ~, ¬. Membalikkan nilai kebenaran. T -> F, F -> T."),
            ("Konjungsi (AND)", "Simbol: &, ∧, *. Benar hanya jika kedua pernyataan benar."),
            ("Disjungsi (OR)", "Simbol: |, ∨, +. Benar jika salah satu atau kedua pernyataan benar."),
            ("Implikasi (IF-THEN)", "Simbol: ->, →, =>. P → Q berarti 'Jika P maka Q'. Salah hanya jika T → F."),
            ("Biimplikasi (IFF)", "Simbol: <->, ↔, <=>. P ↔ Q berarti 'P jika dan hanya jika Q'. Benar jika nilai sama.")
        ]

        for title, desc in materi:
            f = ctk.CTkFrame(theory_frame, fg_color=("gray90", "gray20"))
            f.pack(fill="x", pady=10, padx=10)
            ctk.CTkLabel(f, text=title, font=ctk.CTkFont(size=18, weight="bold")).pack(anchor="w", padx=15, pady=(10, 5))
            ctk.CTkLabel(f, text=desc, font=ctk.CTkFont(size=14), wraplength=700, justify="left").pack(anchor="w", padx=15, pady=(0, 10))

    def clear_history(self):
        if messagebox.askyesno("Confirm", "Hapus semua riwayat?"):
            self.db.clear_history()
            self.show_history()

if __name__ == "__main__":
    app = LogicApp()
    app.mainloop()
