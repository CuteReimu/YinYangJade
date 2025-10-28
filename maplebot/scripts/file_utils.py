import json
import time
import os
import datetime

name_file = "player_name.txt"
player_dict_fn = "player_data/player_{}.json"
lvl_exp_file = "lvl_data.json"

if not os.path.exists("./player_data"):
    os.makedirs("./player_data")
assert os.path.isdir("./player_data")
assert os.path.exists("./lvl_data.json")

def read_with_retry(path, encoding="utf-8", attempts=3, delay=0.05, default=""):
    for i in range(attempts):
        try:
            with open(path, "r", encoding=encoding) as f:
                return f.read()
        except FileNotFoundError:
            if i < attempts - 1:
                time.sleep(delay)
            else:
                with open(path, "w", encoding=encoding) as f:
                    f.write(default)  # Create an empty file
                return default

def load_player_names():
    content = read_with_retry(name_file, encoding="utf-8", default="")
    names = [line.strip() for line in content.splitlines() if line.strip()]
    return names
    
def save_player_names(names):
    temp_name = f"{name_file}.tmp"
    with open(temp_name, "w", encoding="utf-8") as f:
        for name in names:
            f.write(name + "\n")
        f.flush()
        os.fsync(f.fileno())
    os.replace(temp_name, name_file)

def save_dict(fn, _dict):
    temp_name = f"{fn}.tmp"
    with open(temp_name, "w", encoding="utf8") as f:
        json.dump(_dict, f, ensure_ascii=False, indent=4)
        f.flush()
        os.fsync(f.fileno())
    os.replace(temp_name, fn)

def load_dict(fn):    
    content = read_with_retry(fn, encoding="utf8", default=json.dumps({}))
    try:
        _dict = json.loads(content)
    except json.JSONDecodeError:
        _dict = {}
    return _dict

def same_dict(dict1, dict2):
    keys = ["exp", "level", "jobID", "legionLevel", "raidPower"]
    for key in keys:
        if dict1.get(key) != dict2.get(key):
            return False
    date1 = datetime.datetime.strptime(dict1.get("datetime"), "%Y-%m-%d %H:%M:%S")
    date2 = datetime.datetime.strptime(dict2.get("datetime"), "%Y-%m-%d %H:%M:%S")
    if abs((date1 - date2)) >= datetime.timedelta(hours=24):
        return False
    return True
