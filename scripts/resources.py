import json
import numpy as np
import subprocess
import threading

from flask import Flask

app = Flask(__name__)

containers = {}

# Run docker stats command and capture output
stats_output = subprocess.check_output(["docker", "stats", "--no-stream"])

# Split the output into lines and extract container stats
container_stats = [line.split() for line in stats_output.decode().split('\n')[1:] if line]

for stat in container_stats:
    containers[stat[1]] = {
        "cpu_values": [],
        "memory_values": [],
        "net_in_rates": [],
        "net_out_rates": [],
        "total_in_net": float(0),
        "total_out_net": float(0),
        "initial_in_net": float(0),
        "initial_out_net": float(0),
    }

@app.route("/start/<version>/<endpoint>", methods=["POST"])
def start(version, endpoint):
    # Start the routine in a separate thread
    background_thread = threading.Thread(target=capture_metrics, args=(str(version), str(endpoint)))
    background_thread.start()
    
    # Respond to the request with status code 200
    return 'Routine triggered!', 200

def capture_metrics(version: str, endpoint: str):
    # Continuously query docker stats and collect statistics for 30 seconds
    i = 30
    while i:
        print(f"Iteration {i}")
        
         # Run docker stats command and capture output
        stats_output = subprocess.check_output(["docker", "stats", "--no-stream"])

        # Split the output into lines and extract container stats
        container_stats = [line.split() for line in stats_output.decode().split('\n')[1:] if line]

        for stat in container_stats:
            containers[stat[1]]["cpu_values"].append(float(stat[2][:-1]))
            containers[stat[1]]["memory_values"].append(float(stat[3][:-3]))
            containers[stat[1]]["net_in_rates"].append(float(stat[7][:-2]) - containers[stat[1]]["total_in_net"])
            containers[stat[1]]["total_in_net"] = float(stat[7][:-2])
            containers[stat[1]]["net_out_rates"].append(float(stat[9][:-2]) - containers[stat[1]]["total_out_net"])
            containers[stat[1]]["total_out_net"] = float(stat[9][:-2])

        i -= 1

    final_stats = {}
        
    for container_name, stats in containers.items():
        final_stats[container_name] = {
            "cpu_min_value": min(stats["cpu_values"]),
            "cpu_max_value": max(stats["cpu_values"]),
            "cpu_mean_value": np.mean(stats["cpu_values"]),
            "cpu_std_dev_value": np.std(stats["cpu_values"]),

            "memory_min_value": min(stats["memory_values"]),
            "memory_max_value": max(stats["memory_values"]),
            "memory_mean_value": np.mean(stats["memory_values"]),
            "memory_std_dev_value": np.std(stats["memory_values"]),

            "net_in_min_value": min(stats["net_in_rates"]),
            "net_in_max_value": max(stats["net_in_rates"]),
            "net_in_mean_value": np.mean(stats["net_in_rates"]),
            "net_in_std_dev_value": np.std(stats["net_in_rates"]),
            "total_net_in": stats["total_in_net"] - stats["initial_in_net"],

            "net_out_min_value": min(stats["net_out_rates"]),
            "net_out_max_value": max(stats["net_out_rates"]),
            "net_out_mean_value": np.mean(stats["net_out_rates"]),
            "net_out_std_dev_value": np.std(stats["net_out_rates"]),
            "total_net_out": stats["total_out_net"] - stats["initial_out_net"],
        }

    total_cpu_mean = 0
    total_cpu_std_dev = 0
    total_memory_mean = 0
    total_memory_std_dev = 0
    total_net_in_mean = 0
    total_net_in_std_dev = 0
    total_net_out_mean = 0
    total_net_out_std_dev = 0

    for container_name, stats in final_stats.items():
        total_cpu_mean += stats["cpu_mean_value"]
        total_cpu_std_dev += stats["cpu_std_dev_value"]
        total_memory_mean += stats["memory_mean_value"]
        total_memory_std_dev += stats["memory_std_dev_value"]
        total_net_in_mean += stats["net_in_mean_value"]
        total_net_in_std_dev += stats["net_in_std_dev_value"]
        total_net_out_mean += stats["net_out_mean_value"]
        total_net_out_std_dev += stats["net_out_std_dev_value"]

    final_stats["total"] = {
        "app_version": str(version),
        "endpoint": str(endpoint),
        "total_cpu_mean": total_cpu_mean,
        "total_cpu_std_dev": total_cpu_std_dev,
        "total_memory_mean": total_memory_mean,
        "total_memory_std_dev": total_memory_std_dev,
        "total_net_in_mean": total_net_in_mean,
        "total_net_in_std_dev": total_net_in_std_dev,
        "total_net_out_mean": total_net_out_mean,
        "total_net_out_std_dev": total_net_out_std_dev,
    }

    # Write the JSON data to the file
    with open("resources.log", "a") as file:
        file.write(json.dumps(final_stats) + "\n")
        print(f"Wrote results to file for {endpoint} and version {version}")
        file.close()

    return

if __name__ == "__main__":
    app.run(port=5000)