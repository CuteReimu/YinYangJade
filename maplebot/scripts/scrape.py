import requests
import json
import time
import datetime
import base64
import os

from file_utils import *

player_url = "https://www.nexon.com/api/maplestory/no-auth/ranking/v2/na?type=overall&id=weekly&reboot_index=1&page_index=1&character_name={}"
legion_url = "https://www.nexon.com/api/maplestory/no-auth/ranking/v2/na?type=legion&id=45&page_index=1&character_name={}"
buffer_size = 15
sleep_per_request = 0.5  # seconds

def assert_player_onrank(name):
    url = player_url.format(name)
    try:
        response = requests.get(url)
        data = response.json()
    except Exception as e:
        print(f"Error fetching data for {name}: {e}")
        return True  # Assume online if error occurs
    
    count = data['totalCount']
    return count > 0


def try_request(url, name, retries=3, wait=60):
    for retry in range(retries):
        try:
            response = requests.get(url.format(name))
            data = response.json()

            break
        except Exception as e:
            print(f"Error fetching player data for {name}: {e}, retrying ({retry + 1}/3)...")
            if retry == 2:
                print(f"Waiting for {wait} seconds before retrying...")
                time.sleep(wait)
            continue
    return data


def request_from_name_list():
    names = load_player_names()

    for i, name in enumerate(names):
        player_dict = load_dict(player_dict_fn.format(name))
        if len(player_dict) == 0:
            player_dict['data'] = []
            player_dict['img'] = ""
            
        if i % 50 == 0:
            print(f"Processing player {i}...")

        data = try_request(legion_url, name)
        count = data['totalCount']
        if count == 0:
            data = try_request(player_url, name)
            count = data['totalCount']

            if count == 0:
                print(f"Player {name} does not exist.")
                continue
        
        player_name = data['ranks'][0]['characterName']
        exp = data['ranks'][0]['exp']
        lvl = data['ranks'][0]['level']
        jobID = data['ranks'][0]['jobID']
        img_url = data['ranks'][0]['characterImgURL']
        legion_lvl = data['ranks'][0]['legionLevel']
        legion_raid = data['ranks'][0]['raidPower']
        
        try:
            response = requests.get(img_url)
            img64 = base64.b64encode(response.content).decode('utf-8')
        except Exception as e:
            img64 = ""
            
        cur_dict = {
            "name": player_name,
            "datetime": datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            "exp": exp,
            "level": lvl,
            "jobID": jobID,
            "legionLevel": legion_lvl,
            "raidPower": legion_raid,
        }
        
        player_dict['img'] = img64
        
        last_dict = player_dict['data'][-1] if len(player_dict['data']) > 0 else None
        if last_dict is not None and same_dict(last_dict, cur_dict):
            continue
        player_dict['data'].append(cur_dict)
        player_dict['data'] = sorted(player_dict['data'], key=lambda x: x['datetime'])[-buffer_size:]
        
        save_dict(player_dict_fn.format(name), player_dict)

        time.sleep(sleep_per_request)  # Avoid hitting rate limits
        
if __name__ == "__main__":
    sta = time.time()
    request_from_name_list()
    end = time.time()
    print(f"Total time taken: {(end - sta)/60} minutes")
    print("Done")