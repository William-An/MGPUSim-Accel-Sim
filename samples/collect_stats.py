#!/usr/bin/env python3
import csv
import json
from pprint import pprint
from copy import deepcopy

cache_template = {
    "read-hit": 0,
    "read-miss": 0,
    "read-mshr-hit": 0,
    "write-hit": 0,
    "write-miss": 0,
    "write-mshr-hit": 0,
}

stats = {
    "kernel_cycles": 0,
    "L1ICache": deepcopy(cache_template),
    "L1VCache": deepcopy(cache_template),
    "L1SCache": deepcopy(cache_template),
    "L2": deepcopy(cache_template),
    "DRAM": {
        "read_trans_count": 0,
        "write_trans_count": 0
    }
}

with open("metrics.csv") as metricsFile:
    metricsReader = csv.DictReader(metricsFile)
    for metric in metricsReader:
        # Strip white spaces
        tmp = dict()
        for key in metric:
            tmp[key.strip()] = metric[key].strip()
        metric = tmp
        if metric["where"] == "driver" and metric["what"] == "kernel_time":
            stats["kernel_cycles"] = eval(metric["value"]) * 1E9
        else:
            for key in stats:
                if key in metric["where"] and metric["what"] in stats[key]:
                    stats[key][metric["what"]] += eval(metric["value"])
    print(json.dumps(stats, sort_keys=True, indent=4))