import requests
import sys

from datetime import datetime, date

URL = {"anylp":"https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/zd39j4nd?var-ylq4yvzn=qzne828q&var-rn1kmmvl=qj70747q",
       "anyrp": "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/zd39j4nd?var-ylq4yvzn=qzne828q&var-rn1kmmvl=10vzvmol",
       "te":"https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/n2y0m18d?var-dloed1dn=qyzod221",
       "100noab":"https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/rkl6zprk?var-rn1k7xol=lx5o7641&var-38dg4448=1w4p4dmq",
       "100ab":"https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/rkl6zprk?var-rn1k7xol=lx5o7641&var-38dg4448=qoxpx35q",
       "judgement":"https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/wk6544o2?var-jlz631q8=1w4ozxvq",
       "low":"https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/wkp4r60k?var-9l7geqpl=1397dnx1",
       "ab":"https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/w206ox52?var-kn0eyxz8=10vzo8wl",
       "twisted":"https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/9kvvl0ok?var-yn26pzel=le2z97kl"}

CATEGORY_NAMES = {
    "anylp": "Any% Later Patches",
    "anyrp": "Any% Release Patch",
    "te": "True Ending",
    "100noab": "100% No AB",
    "100ab": "100% All Bosses",
    "judgement": "Judgement",
    "low": "Low%",
    "ab": "All Bosses",
    "twisted": "Twisted%",
}

def get_player_name(player, players):
    for p in players:
        if p["id"] == player["id"]:
            return p["names"]["international"]
    return player.get("name", "Unknown")

def format_time(t):
    t = int(t)
    m = t // 60
    s = t - m * 60
    h = m // 60
    if h > 0:
        m = m - h * 60
        return f"{h:d}:{m:02d}:{s:02d}"
    return f"{m:02d}:{s:02d}"

def format_relative_date(date_str):
    """
    将日期字符串转换为相对时间描述

    参数:
        date_str: 日期字符串，格式为 'YYYY-MM-DD'

    返回:
        str: 相对时间描述，如"今天"、"昨天"、"前天"、"X天前"、"X个月前"
    """
    try:
        input_date = datetime.strptime(date_str, '%Y-%m-%d').date()
        today = date.today()
        delta = today - input_date
        days_diff = delta.days

        if days_diff == 0:
            return "今天"
        elif days_diff == 1:
            return "昨天"
        elif days_diff == 2:
            return "前天"
        elif 3 <= days_diff < 30:
            return f"{days_diff}天前"
        elif 30 <= days_diff < 60:
            return "上个月"
        elif days_diff >= 60:
            months = days_diff // 30
            return f"{months}个月前"
        else:
            # 如果是未来日期，返回原日期
            return date_str
    except ValueError:
        # 如果日期格式错误，返回原字符串
        return date_str

def main(user_input):
    resp = requests.get(URL[user_input] + "&embed=players&top=3", timeout=60)
    resp.raise_for_status()

    data = resp.json()["data"]
    runs = data["runs"]
    if len(runs) > 3:
        runs = runs[:3]

    players = data.get("players", {}).get("data", [])

    print(f"=== 丝之歌 — {CATEGORY_NAMES[user_input]} - NMG ===")
    for entry in runs:
        place = entry["place"]
        run = entry["run"]
        time_sec = run["times"]["primary_t"]

        player = get_player_name(run["players"][0], players)
        time_str = format_time(time_sec)
        relative_date = " - " + format_relative_date(run["date"]) if "date" in run else ""

        print(f"{place}. {player} — {time_str}{relative_date}")

if __name__ == "__main__":
    if len(sys.argv) > 1:
        arg = sys.argv[1]
    else:
        arg = input("输入您想查询的榜单(any,te,100,judgement,low,ab,twisted): ")
    arg = arg.replace("%", "").lower()
    if arg == "any":
        main("anyrp")
        main("anylp")
    elif arg in ("all bosses", "all boss", "allbosses", "allboss"):
        main("ab")
    elif arg == "100":
        main("100noab")
        main("100ab")
    else:
        main(arg)
