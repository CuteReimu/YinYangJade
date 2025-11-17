import matplotlib.pyplot as plt
import numpy as np
import io
import base64

def draw_chart(days, clipped_exps, dated_lvls, exp_flags, dated_exps):

    # Data
    x = np.array(days)
    bar_values = np.array(clipped_exps)
    line_values = np.array(dated_lvls)

    BAR_BASE_COLOR = "#4d6bff"
    BAR_OVER_COLOR = "#ff6b6b"
    BAR_UNDER_COLOR = "#ffd93b"
    _colors = [BAR_BASE_COLOR, BAR_OVER_COLOR, BAR_UNDER_COLOR]
    colors = [_colors[flag] for flag in exp_flags] # -1: under, 0: base, 1: over, this is a bad practice

    num_ticks = 6
    exp_ticks = [0, 2, 4, 6, 8, 10]
    line_min = min(v for v in line_values if v is not None)
    line_max = max(v for v in line_values if v is not None)
    line_range = line_max - line_min
    line_min_padded = line_min - 0.1 * line_range
    line_max_padded = line_max + 0.15 * line_range
    line_ticks = np.linspace(line_min_padded, line_max_padded, num_ticks)

    fig, ax1 = plt.subplots(figsize=(8, 5))

    # Background
    bg = "#050816"
    fig.patch.set_facecolor(bg)
    ax1.set_facecolor(bg)

    # Bars
    ax1.bar(x, bar_values, color=colors, zorder=3)
    ax1.set_yticks(exp_ticks)
    ax1.yaxis.set_major_formatter(lambda x, _: f"{x:.0f}T")

    # Line
    ax2 = ax1.twinx()
    line = ax2.plot(x, line_values, marker="o", linewidth=2, markersize=5, color="#9bff7a", zorder=4)[0]
    line.axes.update_datalim(np.array([[0, line_min_padded]]))
    line.axes.update_datalim(np.array([[len(x) - 1, line_max_padded]]))
    line.axes.autoscale_view()
    ax2.yaxis.set_major_formatter(lambda x, _: f"{x:.2f}")
    ax2.set_yticks(line_ticks)


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
