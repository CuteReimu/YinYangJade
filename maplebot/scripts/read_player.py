from file_utils import *
import sys

def try_encode_gb2312(name):
    try:
        encoded_name = name.encode('gb2312').decode('latin1')
        logging.info(f"Converted name {name} to latin1 {encoded_name}")
        return encoded_name
    except Exception as e:
        return name

def process_player_data(name):
    logging.info(f'Start processing data for player {name}')
    name = name.strip()
    name = try_encode_gb2312(name)

    if len(name) > 30: 
        logging.error(f"Player name {name} is too long.")
        return {}
    
    player_names_di = load_player_names()
    player_names = list(player_names_di.keys())
    player_names_lower = {n.lower(): i for i, n in enumerate(player_names)}
    if name.lower() not in player_names_lower:
        logging.info(f"Player name {name} not found. Adding to player list.")
        player_names_di[name] = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        save_player_names(player_names_di)
        return {}
    
    name = player_names[player_names_lower[name.lower()]]
    player_dict = load_dict(player_dict_fn.format(name))
    
    if len(player_dict) == 0:
        logging.info(f"No data found for player {name}.")
        return {}
    if 'data' not in player_dict:
        logging.info(f"No data found for player {name}.")
        return {}
    if len(player_dict['data']) == 0:
        logging.info(f"No data found for player {name}.")
        return {}
    
    data = {}
    exp_dict = load_dict(lvl_exp_file)
    single_lvl_exp = exp_dict['single']
    
    player_name = player_dict['data'][-1]['name']
    level = player_dict['data'][-1]['level']
    exp = player_dict['data'][-1]['exp']
    legion_lvl = player_dict['data'][-1]['legionLevel']
    raid_power = player_dict['data'][-1]['raidPower']
    
    cdata ={
        'Name': player_name,
        "Level": level,
        "EXP": exp,
        'Class': player_dict['data'][-1].get('jobName', ''),
        'ClassID': player_dict['data'][-1].get('jobID', -1),
        'ClassDetail': player_dict['data'][-1].get('jobDetail', '00'),
        "EXPPercent": round(exp / single_lvl_exp[str(level)] * 100, 1),
        "LegionLevel": legion_lvl,
        "RaidPower": raid_power,
        "Image": player_dict['img'],
    }
    data['CharacterData'] = cdata
    
    gdata = []
    for entry in player_dict['data']:
        gdata.append({
            "DateLabel": entry['datetime'].split(" ")[0],
            "Level": entry['level'],
            "CurrentEXP": entry['exp'],
        })
    data['CharacterData']['GraphData'] = gdata
    logging.info(f'Processed data for player {name}')
    
    return data

if __name__ == "__main__":
    player_name = sys.argv[1]
    data = process_player_data(player_name)
    print(json.dumps(data))
