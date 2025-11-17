from file_utils import *
import sys
from exp_utils import *
from draw_exp_graph import *

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

def process_player_data_full(name):
    data = process_player_data(name)
    
    lvl_exp = load_dict(lvl_exp_file)
    lvl_single = lvl_exp['single']
    lvl_culm = lvl_exp['cumulative']
    
    status = 'success'
    try:
        name = data['CharacterData']['Name']
        job_name = data['CharacterData']['Class']
        job_name = class_translations.get(job_name, job_name)
        _level = data['CharacterData']['Level']
        _exp = data['CharacterData']['EXP']
        gdata = data['CharacterData']['GraphData']
        legion_level = data['CharacterData']['LegionLevel']
        avatar_img64 = data['CharacterData']['Image']
    except KeyError:
        status = 'No Data'
        return {
            'status': status,
            'profile': '',
            'text': '尚未收录角色数据',
            'chart': '',
        }
        
    times = [datetime.fromisoformat(di['DateLabel']) for di in gdata]
    lvls = [di['Level'] for di in gdata]
    exps = [di['CurrentEXP'] for di in gdata]

    _exps, _lvls = get_processed_y(
        exps, lvls, lvl_single, lvl_culm
    )
    days, dated_exps, dated_lvls = format_series_data(
        times, _exps, _lvls
    )
    
    hasChange = check_exp_has_change(dated_exps)
    if not hasChange:
        logging.info(f'No EXP change detected for player {name}')
        exp_text = '近期经验无变化\r\n'
        imgb64 = ''
    else:
        days_to_lvl, exp_percent = days_to_level(
            dated_exps, _exp, _level, lvl_single
        )
        clipped_exps, exp_flags = clip_exps(dated_exps)
        logging.info(f'Processed EXP and Level data for player {name}')

        imgb64 = draw_chart(days, clipped_exps, dated_lvls, exp_flags, dated_exps)
        logging.info(f'Drew EXP graph for player {name}')
        exp_text =  f'预计还有{days_to_lvl}天升级\r\n'

    message_txt = (
        f'角色名：{name}\r\n' +
        f'职业：{job_name}\r\n' +
        f'等级：{_level} ({exp_percent}%)\r\n' +
        f'联盟：{legion_level}\r\n' +
        exp_text
    )
    result = {
        'status': status,
        'profile': avatar_img64,
        'text': message_txt,
        'chart': imgb64,
    }
    return result

if __name__ == "__main__":
    player_name = sys.argv[1]
    if len(sys.argv) > 2 and sys.argv[2] == 'silence':
        process_player_data(player_name)
    else:
        data = process_player_data_full(player_name)
        print(json.dumps(data))