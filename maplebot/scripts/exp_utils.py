from datetime import datetime, timedelta

def get_processed_y(exps, lvls, lvl_single, lvl_culm):
    exp_diffs = [0]
    lvl_decimals = [exps[0] / lvl_single[str(lvls[0])] + lvls[0]]
    for i in range(1, len(exps)):
        exp_prev = lvl_culm[str(lvls[i-1])] + exps[i-1]
        exp_curr = lvl_culm[str(lvls[i])] + exps[i]
        exp_diff = exp_curr - exp_prev
        exp_diffs.append(exp_diff)
        lvl_decimals.append(exps[i] / lvl_single[str(lvls[i])] + lvls[i])
    return exp_diffs, lvl_decimals

def format_series_data(times, exps, lvls):
    if isinstance(times[0], str):
        times = [datetime.fromisoformat(t) for t in times]
    series_max_length = 14
    last_day = times[-1]
    days = [last_day - timedelta(days=i) for i in range(series_max_length)][::-1]
    dated_exps = []
    dated_lvls = []
    prev_lvl = None     # Break the line if no data for starting days
    default_exp = None
    for i, day in enumerate(days):
        if day not in times:
            dated_exps.append(default_exp)
            dated_lvls.append(prev_lvl)
            continue
        idx = times.index(day)
        dated_exps.append(exps[idx])
        dated_lvls.append(lvls[idx])
        prev_lvl = lvls[idx]
        default_exp = 0
    days = [day.strftime('%m-%d') for day in days]
    return days, dated_exps, dated_lvls

def check_exp_has_change(dated_exps):
    filtered_exps = [val for val in dated_exps if val is not None and val != 0]
    return len(filtered_exps) > 0

def clip_exps(exps, min_value=0.2, max_value=10):
    T = 1e12
    clipped_exps = []
    flags = []
    for val in exps:
        if val == 0 or val is None:
            clipped_exps.append(0)
            flags.append(0)
        else:
            _val = val / T
            if _val < min_value:
                _val = min_value
                flags.append(-1)
            elif _val > max_value:
                _val = max_value
                flags.append(1)
            else:
                flags.append(0)
            clipped_exps.append(_val)
    return clipped_exps, flags

def days_to_level(dated_exps, current_exp, current_lvl, lvl_single):
    avg_exp = sum(val for val in dated_exps if val is not None) / sum(1 for val in dated_exps if val is not None)
    exp_needed = lvl_single[str(current_lvl)] - current_exp
    days_needed = round(exp_needed / avg_exp)
    cur_exp_percent = round(current_exp / lvl_single[str(current_lvl)] * 100, 1)
    return days_needed, cur_exp_percent