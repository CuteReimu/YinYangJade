import matplotlib.pyplot as plt
import numpy as np
import io
import base64

def draw_chart(days, clipped_exps, dated_lvls, exp_flags, dated_exps):

    # Data
    T = 1e12
    x = np.array(days)
    bar_values = np.array(clipped_exps)
    line_values = np.array(dated_lvls)

    BAR_BASE_COLOR = "#4d6bff"
    BAR_OVER_COLOR = "#ff6b6b"
    BAR_UNDER_COLOR = "#ffd93b"
    _colors = [BAR_BASE_COLOR, BAR_OVER_COLOR, BAR_UNDER_COLOR]
    colors = [_colors[flag] for flag in exp_flags] # -1: under, 0: base, 1: over, this is a bad practice

    num_ticks = 6
    exp_ticks = np.linspace(0, 10, num_ticks)
    line_min = min(v for v in line_values if v is not None)
    line_max = max(v for v in line_values if v is not None)
    # line_range = line_max - line_min
    # if line_range == 0:
    #     line_range = 1  # Avoid ymin == ymax issue
    # line_min_padded = line_min - 0.1 * line_range
    # line_max_padded = line_min + 2.1 * line_range
    # line_ticks = np.linspace(line_min_padded, line_max_padded, num_ticks)
    line_min_padded, line_max_padded, line_ticks = get_nearest_spread(line_min, line_max)

    fig, ax1 = plt.subplots(figsize=(8, 5))

    # Background
    bg = "#050816"
    fig.patch.set_facecolor(bg)
    ax1.set_facecolor(bg)

    # Bars
    ax1.bar(x, bar_values, color=colors, zorder=3)
    ax1.set_yticks(exp_ticks)
    ax1.yaxis.set_major_formatter(lambda x, _: f"{x:.0f}T")
    ax1.set_ylim(0, 10)

    # Line
    ax2 = ax1.twinx()
    line = ax2.plot(x, line_values, marker="o", linewidth=2, markersize=5, color="#9bff7a", zorder=4)[0]
    # line.axes.update_datalim(np.array([[0, line_min_padded]]))
    # line.axes.update_datalim(np.array([[len(x) - 1, line_max_padded]]))
    # line.axes.autoscale_view()
    ax2.yaxis.set_major_formatter(lambda x, _: f"{x:.2f}")
    ax2.set_yticks(line_ticks)
    ax2.set_ylim(line_min_padded, line_max_padded)


    # Disable all default grids except horizontal lines
    ax1.grid(False)
    ax2.grid(False)
    ax1.yaxis.grid(True, linestyle="-", linewidth=0.6, color="#262b3e", alpha=0.9, zorder=-1)

    # Ticks/labels colors
    tick_color = "#a7adc4"
    for ax in (ax1, ax2):
        ax.tick_params(axis="both", colors=tick_color, labelsize=15)
        ax.tick_params(axis='x', labelrotation=60)
        ax.yaxis.label.set_color(tick_color)

    # Clean up spines
    for spine in ["top", "right", "left"]:
        ax1.spines[spine].set_visible(False)
        ax2.spines[spine].set_visible(False)
    ax1.spines["bottom"].set_color("#20253a")

    # Show exact text for overflown bars
    for i, (val, flag) in enumerate(zip(dated_exps, exp_flags)):
        if flag == 1:
            display_val = val / T
            text = f"{display_val:.1f}T"
            ax1.text(x=i, y=bar_values[i] + 0.3, s=text, ha="center", va="bottom", color="#ffffff", fontsize=12, zorder=5)

    buf = io.BytesIO()
    plt.savefig(buf, format='png', bbox_inches='tight')
    b64 = base64.b64encode(buf.getvalue()).decode('ascii')
    return b64


def get_nearest_spread(min_lvl, max_lvl):
    steps = [0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 20]
    if max_lvl >= 260:
        steps = steps[:-1] # Remove 20
    if max_lvl >= 270:
        steps = steps[:-1] # Remove 10
    if max_lvl >= 280:
        steps = steps[:-1] # Remove 5
    if max_lvl >= 290:
        steps = steps[:-1] # Remove 2
        
    LVL_MAX = 300
    spreads = [step * 5 for step in steps]
    lvl_spread_double = (max_lvl - min_lvl) * 2
    
    step_using = steps[-1]
    for i, spread in enumerate(spreads):
        if spread >= lvl_spread_double:
            step_using = steps[i]
            break
    
    mid_lvl = round(max_lvl / step_using) * step_using
    end_lvl = mid_lvl + step_using * 2
    if end_lvl > LVL_MAX:
        end_lvl = LVL_MAX
        mid_lvl = end_lvl - step_using * 2
    start_lvl = mid_lvl - step_using * 3
    ticks = [start_lvl + i * step_using for i in range(6)]
    
    print(f'Min level: {min_lvl}, Max level: {max_lvl}, Using step: {step_using}, Start: {start_lvl}, End: {end_lvl}, Ticks: {ticks}')
    return start_lvl, end_lvl, ticks
