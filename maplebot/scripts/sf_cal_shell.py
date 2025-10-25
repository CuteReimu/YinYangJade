import sys
import numpy as np
import time
from functools import partial

debug = False
OPTION = ["Simulation", "Markov"][1]

job_zero = False
other_events = [True, False, True]  # 5/10/15, +2 events, star catching

odds = [
    {
        # current_star: [upgrade, fail_stay, fail_down, fail_break]
        0: [0.950, 0.050, 0.000, 0.000],
        1: [0.900, 0.100, 0.000, 0.000],
        2: [0.850, 0.150, 0.000, 0.000],
        3: [0.850, 0.150, 0.000, 0.000],
        4: [0.800, 0.200, 0.000, 0.000],
        5: [0.750, 0.250, 0.000, 0.000],
        6: [0.700, 0.300, 0.000, 0.000],
        7: [0.650, 0.350, 0.000, 0.000],
        8: [0.600, 0.400, 0.000, 0.000],
        9: [0.550, 0.450, 0.000, 0.000],
        10: [0.500, 0.500, 0.000, 0.000],
        11: [0.450, 0.550, 0.000, 0.000],
        12: [0.400, 0.600, 0.000, 0.000],
        13: [0.350, 0.650, 0.000, 0.000],
        14: [0.300, 0.700, 0.000, 0.000],
        15: [0.300, 0.679, 0.000, 0.021],
        16: [0.300, 0.000, 0.679, 0.021],
        17: [0.300, 0.000, 0.679, 0.021],
        18: [0.300, 0.000, 0.672, 0.028],
        19: [0.300, 0.000, 0.672, 0.028],
        20: [0.300, 0.630, 0.000, 0.070],
        21: [0.300, 0.000, 0.630, 0.070],
        22: [0.030, 0.000, 0.776, 0.194],
        23: [0.020, 0.000, 0.686, 0.294],
        24: [0.010, 0.000, 0.594, 0.396],
    },
    {
        # current_star: [upgrade, fail_stay, fail_down, fail_break]
        0: [0.950, 0.050, 0.000, 0.000],
        1: [0.900, 0.100, 0.000, 0.000],
        2: [0.850, 0.150, 0.000, 0.000],
        3: [0.850, 0.150, 0.000, 0.000],
        4: [0.800, 0.200, 0.000, 0.000],
        5: [0.750, 0.250, 0.000, 0.000],
        6: [0.700, 0.300, 0.000, 0.000],
        7: [0.650, 0.350, 0.000, 0.000],
        8: [0.600, 0.400, 0.000, 0.000],
        9: [0.550, 0.450, 0.000, 0.000],
        10: [0.500, 0.500, 0.000, 0.000],
        11: [0.450, 0.550, 0.000, 0.000],
        12: [0.400, 0.600, 0.000, 0.000],
        13: [0.350, 0.650, 0.000, 0.000],
        14: [0.300, 0.700, 0.000, 0.000],
        15: [0.300, 0.679, 0.000, 0.021],
        16: [0.300, 0.679, 0.000, 0.021],
        17: [0.150, 0.782, 0.000, 0.068],
        18: [0.150, 0.782, 0.000, 0.068],
        19: [0.150, 0.765, 0.000, 0.085],
        20: [0.300, 0.595, 0.000, 0.105],
        21: [0.150, 0.7225, 0.000, 0.1275],
        22: [0.150, 0.680, 0.000, 0.170],
        23: [0.100, 0.720, 0.000, 0.180],
        24: [0.100, 0.720, 0.000, 0.180],
        25: [0.100, 0.720, 0.000, 0.180],
        26: [0.070, 0.744, 0.000, 0.186],
        27: [0.050, 0.760, 0.000, 0.190],
        28: [0.030, 0.776, 0.000, 0.194],
        29: [0.010, 0.792, 0.000, 0.198],
    },
]


star_cap = {
    94: 5,
    107: 8,
    117: 10,
    127: 15,
    137: 20,
}

star_cost_divisor = {
    10: 40_000,
    11: 22_000,
    12: 15_000,
    13: 11_000,
    14: 7_500,
}


def format_number(value, decimal=2):
    if value is int:
        return f"{value:.0f}"
    return f"{value:.{decimal}f}"


def get_star_cap(eq_level, new_star_cap=None):
    cap = 25 if new_star_cap is None else new_star_cap
    for max_lvl, star in star_cap.items():
        if eq_level <= max_lvl:
            cap = star
            break
    return cap


def get_meso_cost(
    cur_star: int,
    eq_level: int,
    safe_guard=False,
    job_zero=False,
    discount=False,
    new_system=False,
):
    multiplier = 1
    if eq_level <= 130:
        raise ValueError("D")
    if discount:
        multiplier -= 0.3
    if safe_guard and not new_system:
        if cur_star != 15 and cur_star != 16:
            raise ValueError("Only safeguard 15 and 16 star items")
        multiplier += 1
    elif safe_guard and new_system:
        if cur_star < 15 or cur_star > 18:
            raise ValueError(
                "Only safeguard 15, 16 and 17 star items in new star system"
            )
        multiplier += 2

    if job_zero:
        eq_level = min(150, eq_level)
    if eq_level in [
        108,
        109,
    ]:
        cur_star = min(7, cur_star)
    if eq_level in [
        118,
        119,
    ]:
        cur_star = min(9, cur_star)
    if eq_level in [
        128,
        129,
    ]:
        cur_star = min(14, cur_star)

    eq_level //= 10
    eq_level *= 10

    new_sf_mult = 1
    if new_system:
        if cur_star == 17:
            new_sf_mult = 4 / 3
        elif cur_star == 18:
            new_sf_mult = 20 / 7
        elif cur_star == 19:
            new_sf_mult = 40 / 9
        elif cur_star == 21:
            new_sf_mult = 8 / 5
        else:
            new_sf_mult = 1

    if cur_star < 10:
        meso_cost = 100 * round(eq_level**3 * (cur_star + 1) / 2500 + 10)
    else:
        divisor = star_cost_divisor.get(cur_star, 20_000)
        meso_cost = 100 * round(
            new_sf_mult * eq_level**3 * (cur_star + 1) ** 2.7 / divisor + 10
        )
    return round(meso_cost * multiplier)


def calculate_markov(P, C, init_idx, absorb_count=1):
    """
    Calculate expected mean, variance, skewness, and quantiles of total cost in a Markov chain.

    Parameters:
    P (numpy.ndarray): Transition probability matrix, i*2 as star state, 
                       i*2+1 as star state after downgrade from star i+1, -1 as success state by default.
    C (numpy.ndarray): Cost matrix (same shape as P).
    init_idx (int): Index of initial state.
    absorb_count (int): Number of absorbing states at the end of the matrix.

    Returns:
    tuple: (mean_cost, var_cost, skew, quartiles)
    """

    # print(P.shape, C.shape)
    n = P.shape[0]
    Q = P[:-absorb_count, :-absorb_count]  # transient-to-transient
    P_full = P[:-absorb_count, :]  # transient-to-all
    C_full = C[:-absorb_count, :]  # cost from transient to all

    # r[i] = expected immediate cost at state i
    r = np.sum(P_full * C_full, axis=1)

    I = np.eye(n - 1)
    g = np.linalg.solve(I - Q, r)

    mean_cost = g[init_idx]

    return mean_cost


def calculate_no_boom_chance(P, init_idx):
    """
    Calculate expected mean and variance of total cost in a Markov chain.

    Parameters:
    P (numpy.ndarray): Transition probability matrix, with -2 as boom state and -1 as success state.
    init_idx (int): Index of initial state.

    Returns:
    float: probability of reaching success state with no boom
    """

    Q = P[:-2, :-2]  # transient-to-transient
    I = np.eye(Q.shape[0])
    R = P[:-2, -2:]

    B_absorb = np.linalg.solve(I - Q, R)
    absorb_probs = B_absorb[init_idx]  # probabilities for each absorbing state
    return absorb_probs[1]


def fill_transition_cost_boom(
    new_system,
    get_meso_cost,
    end_star,
    eq_level,
    job_zero,
    discount,
    safe_guard,
    arr,
    weights,
    no_boom_mat,
    get_odds_and_inc,
    _5_10_15,
    kms_new,
):
    for i in range(end_star):
        upgrade, fail_stay, fail_down, fail_break, increment = get_odds_and_inc(i)
        if (not new_system and (i not in [15, 16])) or (
            new_system and (i not in [15, 16, 17])
        ):
            _safe_guard = False
        else:
            if not new_system and _5_10_15 and i == 15:
                _safe_guard = False
            else:
                _safe_guard = safe_guard
        cost = get_meso_cost(
            i, eq_level, _safe_guard, job_zero, discount, new_system=new_system
        )

        # Upgrade
        a = 2 * i
        b = 2 * i + 2 * increment
        arr[a, b] = upgrade
        weights[(a, b)] = cost
        if i + increment == end_star:
            no_boom_mat[a, -1] = upgrade
        else:
            no_boom_mat[a, b] = upgrade

        # Stay
        a = 2 * i
        b = 2 * i
        arr[a, b] = fail_stay
        no_boom_mat[a, b] = fail_stay
        if fail_stay > 0:
            weights[(a, b)] = cost

        # Down
        if i > 0:
            if i in [16, 21]:
                a = 2 * i
                b = 2 * i - 2
            else:
                a = 2 * i
                b = 2 * i - 1
            arr[a, b] = fail_down
            no_boom_mat[a, b] = fail_down
            if fail_down > 0:
                weights[(a, b)] = cost

        # Chance time
        if i >= 16 and i != 19 and i != 20 and (not kms_new):
            upgrade, fail_stay, fail_down, fail_break, increment = get_odds_and_inc(i)

            # Upgrade
            a = 2 * i + 1
            b = 2 * i + 2 * increment
            arr[a, b] = upgrade
            no_boom_mat[a, b] = upgrade
            weights[(a, b)] = cost

            # Down then comes back
            assert fail_down > 0, f"Chance time at {i} must have fail down"
            a = 2 * i + 1
            b = 2 * i
            lower_cost = get_meso_cost(
                i - 1, eq_level, False, job_zero, discount, new_system=new_system
            )
            arr[a, b] = fail_down
            no_boom_mat[a, b] = fail_down
            weights[(a, b)] = cost + lower_cost

            # Boom
            a = 2 * i + 1
            b = 2 * 12
            arr[a, b] = fail_break
            no_boom_mat[a, -2] += fail_break
            weights[(a, b)] = cost

        # Boom
        if fail_break > 0:
            a = 2 * i
            b = 2 * 12
            arr[a, b] = fail_break
            no_boom_mat[a, -2] = fail_break
            weights[(a, b)] = cost


def get_tap_cost_mat(end_star):
    size = end_star * 2 + 1
    tap_cost_mat = np.zeros((size, size))
    for i in range(end_star):
        for j in range(2):
            a = i * 2 + j
            # Up
            b = (i + 1) * 2
            tap_cost_mat[a, b] = 1

            # Down
            if i > 15:
                if i == 16 or i == 21:
                    b = (i - 1) * 2
                else:
                    b = (i - 1) * 2 + 1
                tap_cost_mat[a, b] = 1
                if j == 1:
                    b = i * 2
                    tap_cost_mat[a, b] = 2

            # Stay
            if i <= 15 or i == 20:
                b = i * 2
                tap_cost_mat[a, b] = 1

            # Boom
            if i >= 15:
                b = 12 * 2
                tap_cost_mat[a, b] = 1

    return tap_cost_mat


def get_boom_cost_mat(end_star):
    size = end_star * 2 + 1
    boom_cost_mat = np.zeros((size, size))
    for i in range(end_star):
        for j in range(2):
            a = i * 2 + j

            # Boom
            if i >= 15:
                b = 12 * 2
                boom_cost_mat[a, b] = 1

    return boom_cost_mat


def get_odds_and_inc(i, safe_guard, kms_new):
    upgrade, fail_stay, fail_down, fail_break = odds[kms_new][i]
    increment = 1

    # 5 10 15
    if other_events[0] and i in [5, 10, 15]:
        upgrade = 1.0
        fail_stay = 0.0
        fail_down = 0.0
        fail_break = 0.0

    # 1+1
    if other_events[1] and i <= 10:
        increment = 2

    # star catch
    if other_events[2]:
        upgrade *= 1.05
        if upgrade > 1.0:
            upgrade = 1.0
        multiplier = (1 - upgrade) / (1 - odds[kms_new][i][0])
        fail_break *= multiplier
        fail_stay *= multiplier
        fail_down *= multiplier

    # boom events
    # NOTE: This event might not be the same as other servers
    if other_events[3]:
        if i <= 21:
            perk = fail_break * 0.3
            fail_break -= perk
            fail_stay += perk

    if kms_new:
        if safe_guard and i in [15, 16, 17]:
            fail_stay += fail_break
            fail_break = 0.0
    else:
        if safe_guard and i in [15, 16]:
            if i == 15:
                fail_stay += fail_break
            elif i == 16:
                fail_down += fail_break
            fail_break = 0.0
    # print(f"Star {i}: Odds after events and safeguard: {upgrade, fail_stay, fail_down, fail_break}, increment: {increment}")
    assert (
        abs(sum([upgrade, fail_stay, fail_down, fail_break]) - 1) < 1e-6
    ), f"Odds do not sum to 1 at star {i}, sum: {upgrade + fail_stay + fail_down + fail_break}, odds: {upgrade, fail_stay, fail_down, fail_break}"
    return upgrade, fail_stay, fail_down, fail_break, increment


if __name__ == "__main__":
    import time

    # 装备等级、当前星、目标星、保护、 是否抓星星、是否为kms新规、七折、必成、10星以下升2星、韩服加概率活动
    (
        eq_level,
        cur_star,
        end_star,
        safe_guard,
        star_catch,
        kms_new,
        discount,
        _5_10_15,
        _1_plus_1,
        boom_events,
    ) = sys.argv[1:11]
    eq_level = int(eq_level)
    init_star = int(cur_star)
    end_star = int(end_star)
    safe_guard = safe_guard.lower() in ["true", "1", "yes", "y"]
    star_catch = star_catch.lower() in ["true", "1", "yes", "y"]
    kms_new = kms_new.lower() in ["true", "1", "yes", "y"]
    discount = discount.lower() in ["true", "1", "yes", "y"]
    _5_10_15 = _5_10_15.lower() in ["true", "1", "yes", "y"]
    _1_plus_1 = _1_plus_1.lower() in ["true", "1", "yes", "y"]
    boom_events = boom_events.lower() in ["true", "1", "yes", "y"]

    other_events = [_5_10_15, _1_plus_1, star_catch, boom_events]

    print(f"{eq_level}: {init_star} -> {end_star}")
    print(
        f"safe_guard: {safe_guard}, star_catch: {star_catch}, kms_new: {kms_new}, discount: {discount}, 5/10/15: {_5_10_15}, 1+1: {_1_plus_1}, boom_events: {boom_events}"
    )
    print()

    if OPTION == "Markov":
        sta = time.time()

        size = end_star * 2 + 1
        arr = np.zeros((size, size))
        weights = {}
        weights_mat = np.zeros((size, size))
        no_boom_mat = np.zeros((size + 1, size + 1))

        arr[end_star * 2, end_star * 2] = 1
        get_odds_and_inc = partial(
            get_odds_and_inc, safe_guard=safe_guard, kms_new=kms_new
        )

        fill_transition_cost_boom(
            kms_new,
            get_meso_cost,
            end_star,
            eq_level,
            job_zero,
            discount,
            safe_guard,
            arr,
            weights,
            no_boom_mat,
            get_odds_and_inc,
            _5_10_15,
            kms_new,
        )
        for (i, j), cost in weights.items():
            weights_mat[i, j] = cost
        tap_count_mat = get_tap_cost_mat(end_star)
        boom_count_mat = get_boom_cost_mat(end_star)
        P = arr

        total_mean = calculate_markov(
            P, weights_mat, 2 * init_star
        )
        
        mids = []
        mid_results = []
        for mid in range(init_star + 1, end_star):
            mid_mean = calculate_markov(
                P[: 2 * mid + 1, : 2 * mid + 1], 
                weights_mat[ : 2 * mid + 1, : 2 * mid + 1], 
                2 * init_star
            )
            mids.append(str(mid))
            mid_results.append(f"{format_number(mid_mean)}")
            
        print(f"({', '.join(mids)}) Midway costs: <{', '.join(mid_results)}>")
        
        tap_total_mean = calculate_markov(
            P, tap_count_mat, 2 * init_star
        )
        boom_mean = calculate_markov(
            P, boom_count_mat, 2 * init_star
        )
        no_boom_chance = calculate_no_boom_chance(no_boom_mat, 2 * init_star)
        # boom_mean, boom_var = edge_counts_to_node_24(P)

        print(
            f"Cost Mean: <{format_number(total_mean)}>"
        )
        print(
            f"Boom mean: <{format_number(boom_mean)}>>"
        )
        print(f"Chance of no boom: <{no_boom_chance*100:4f}%>")
        print(
            f"Tap mean: <{format_number(tap_total_mean)}>>"
        )
        print(f"Time taken: {time.time() - sta:.2f}s")

    else:
        raise ValueError("Invalid OPTION")
