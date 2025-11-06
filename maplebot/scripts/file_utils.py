import json
import time
import os
import datetime
import logging


name_file = "player_name.txt"
new_name_file = "player_name.json"
player_dict_fn = "player_data/player_{}.json"
lvl_exp_file = "lvl_data.json"
response_file = "response/response_{}.json"
log_file = f"pylogs/{datetime.datetime.now().strftime('%Y-%m-%d')}.log"


if not os.path.exists("./player_data"):
    os.makedirs("./player_data")
assert os.path.isdir("./player_data")
if not os.path.exists('./response'):
    os.makedirs('./response')
assert os.path.isdir("./response")
if not os.path.exists("./pylogs"):
    os.makedirs("./pylogs")
assert os.path.isdir("./pylogs")
assert os.path.exists("./lvl_data.json")

logging.basicConfig(level=logging.INFO,
                    format='%(asctime)s - %(levelname)s - %(message)s',
                    filename=log_file,
                    filemode='a',
                    encoding='utf-8'
)
logging.info(f"Program started at {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")

def read_with_retry(path, encoding="utf-8", attempts=3, delay=0.05, default=""):
    for i in range(attempts):
        try:
            with open(path, "r", encoding=encoding) as f:
                logging.info(f"Successfully read {path}")
                return f.read()
        except FileNotFoundError:
            if i < attempts - 1:
                time.sleep(delay)
            else:
                with open(path, "w", encoding=encoding) as f:
                    f.write(default)  # Create an empty file
                logging.warning(f"File {path} not found. Created new file with default content.")
                return default

def load_player_names():
    if os.path.exists(new_name_file):
        content = read_with_retry(new_name_file, encoding="utf-8", default="[]")
        names = json.loads(content)
        logging.info(f'Loaded player name from json')
        return names
    else:
        content = read_with_retry(name_file, encoding="utf-8", default="")
        _names = [line.strip() for line in content.splitlines() if line.strip()]
        names = {name: datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S") for name in _names}
        save_dict(new_name_file, names)
        logging.info(f'Loaded player name from txt and converted to json')
    return names
    
def save_player_names(names):
    temp_name = f"{new_name_file}.tmp"
    with open(temp_name, "w", encoding="utf-8") as f:
        json.dump(names, f, ensure_ascii=False, indent=4)
        f.flush()
        os.fsync(f.fileno())
    os.replace(temp_name, new_name_file)
    logging.info(f'Saved file {new_name_file}')

def remove_player_names(names_to_remove, updated_names):
    names = load_player_names()
    for name in names_to_remove:
        if name in names:
            del names[name]
    names.update(updated_names)
    save_player_names(names)
    logging.info(f'Removed {len(names_to_remove)} player names')

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
    if abs((date1 - date2)) >= datetime.timedelta(hours=20):
        return False
    return True
