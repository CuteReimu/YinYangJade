import requests
import json
import time
import datetime
import base64
import os

from file_utils import *

player_url = "https://www.nexon.com/api/maplestory/no-auth/ranking/v2/na?type=overall&id=legendary&page_index=1&character_name={}"
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
    
    count = data['totalCount'] if data is not None else 1 # Make no conclusion if error occurs
    return count > 0


def try_request(url, name, retries=3, wait=10):
    data = None
    for retry in range(retries):
        try:
            response = requests.get(url.format(name))
            data = response.json()
            logging.info(f"Requested for {name} successfully")
            break
        except Exception as e:
            print(f"Error fetching player data for {name}: {e}, retrying ({retry + 1}/3)..., request status: {response.status_code if 'response' in locals() else 'N/A'}")
            if retry == 2:
                print(f"Waiting for {wait} seconds before retrying...")
                logging.warning(f"Fetch is too fast, waiting for {wait} seconds")
                time.sleep(wait)
            continue
    return data


def request_from_name_list():
    names_dict = load_player_names()
    names = list(names_dict.keys())
    names_to_del = []

    for i, name in enumerate(names):
        player_dict = load_dict(player_dict_fn.format(name))
        if len(player_dict) == 0:
            player_dict['data'] = []
            player_dict['img'] = ""
            
        if i % 50 == 0:
            print(f"Processing player {i}...")

        data = try_request(legion_url, name)
        count = data['totalCount'] if data is not None else 0
        if count == 0:
            data = try_request(player_url, name)
            count = data['totalCount'] if data is not None else 0

            if count == 0:
                update_time = names_dict[name]
                time_in_days = (datetime.datetime.now() - datetime.datetime.strptime(update_time, "%Y-%m-%d %H:%M:%S")).days
                # names_dict[name] = update_time  # Removed: no-op
                if time_in_days > 3:
                    names_to_del.append(name)
                    del names_dict[name]
                print(f"Player {name} does not exist for {time_in_days} days.")
                logging.info(f"Player {name} does not exist for {time_in_days} days.")
                time.sleep(sleep_per_request * 2)  # Avoid hitting rate limits
                continue
            
        logging.info(f'{name} data found')
        
        player_name = data['ranks'][0]['characterName']
        exp = data['ranks'][0]['exp']
        lvl = data['ranks'][0]['level']
        img_url = data['ranks'][0]['characterImgURL']
        jobID = data['ranks'][0].get('jobID', -1)
        job_detail = data['ranks'][0].get('jobDetail', -1)
        job_name = data['ranks'][0].get('jobName', '')
        legion_lvl = data['ranks'][0].get('legionLevel', 0)
        legion_raid = data['ranks'][0].get('raidPower', 0)
        
        try:
            response = requests.get(img_url)
            img64 = base64.b64encode(response.content).decode('utf-8')
        except Exception as e:
            img64 = ""
            logging.warning(f"Error fetching image for {player_name}: {e}")
            
        cur_dict = {
            "name": player_name,
            "datetime": datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            "exp": exp,
            "level": lvl,
            "jobID": jobID,
            "jobDetail": job_detail,
            "jobName": job_name,
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
        names_dict[name] = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        logging.info(f'Updated data for player {name}')
        
        time.sleep(sleep_per_request)  # Avoid hitting rate limits
    remove_player_names(names_to_del, names_dict)

if __name__ == "__main__":
    sta = time.time()
    logging.info("Starting data scrape...")
    request_from_name_list()
    end = time.time()
    print(f"Total time taken: {(end - sta)/60} minutes")
    print("Done")
    logging.info(f"Data scrape completed in {(end - sta)/60} minutes")