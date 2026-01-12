import os
import sys

# Tcl/Tk Path Fix for Windows Venv
python_path = os.path.dirname(sys.executable)
if "logic_env" in python_path:
    # If running in venv, point to the system Tcl/Tk if venv is missing them
    base_python = os.path.dirname(python_path) 
    # Usually system python is at C:\Users\...\AppData\Local\Programs\Python\Python313
    # Let's try to detect the common locations
    tcl_path = os.path.join(sys.base_prefix, 'tcl', 'tcl8.6')
    tk_path = os.path.join(sys.base_prefix, 'tcl', 'tk8.6')
    if os.path.exists(tcl_path):
        os.environ['TCL_LIBRARY'] = tcl_path
    if os.path.exists(tk_path):
        os.environ['TK_LIBRARY'] = tk_path

import customtkinter as ctk
from ui.app_gui import LogicApp

# Add the project directory to sys.path to allow imports from core/ui
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

def main():
    # Set appearance and theme
    ctk.set_appearance_mode("Dark")  # Options: "System" (standard), "Dark", "Light"
    ctk.set_default_color_theme("blue")  # Options: "blue" (standard), "green", "dark-blue"
    
    # Initialize the app
    app = LogicApp()
    
    # Run the application
    try:
        app.mainloop()
    except Exception as e:
        print(f"Critical error: {e}")

if __name__ == "__main__":
    main()
