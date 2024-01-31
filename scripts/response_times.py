import requests
import time
import numpy as np
import random

last_created_account_id = 0

# Function to generate a random email address
def generate_random_email():
    domains = ["gmail.com", "yahoo.com", "hotmail.com", "example.com"]
    name = ''.join(random.choices('abcdefghijklmnopqrstuvwxyz', k=8))
    domain = random.choice(domains)
    return f"{name}@{domain}"

# Function to generate a random name
def generate_random_name():
    names = ["Alice", "Bob", "Charlie", "David", "Emma", "Frank", "Grace", "Henry"]
    return random.choice(names)

# Function to generate a random customer ID
def generate_random_customer_id():
    return random.randint(1, 1000)

# Function to generate a random sender or receiver ID
def generate_random_sender_receiver_id():
    return random.randint(1, last_created_account_id)

# Function to generate a random amount
def generate_random_amount():
    return random.randint(10, 1000)

# Function to make API calls and measure response time
def get_response_time(url, json_data):
    global last_created_account_id
    start_time = time.time()
    try:
        response = requests.post(url, json=json_data)
        response.raise_for_status()  # Raise an exception for HTTP errors
        response_json = response.json()  # Convert response to JSON format
        account_id = response_json.get('accountID')  # Retrieve accountID from response
        if account_id:
            last_created_account_id = int(account_id)
        end_time = time.time()
        return end_time - start_time, True
    except requests.exceptions.RequestException as e:
        print(json_data)
        end_time = time.time()
        print(f"Request failed: {e}")
        return end_time - start_time, False

# Function to perform multiple requests and calculate statistics
def measure_response_times(url, timeout, use_randomized_body):
    num_requests = 0
    successful_requests = 0
    response_times = []

    start_time = time.time()
    while True:
        num_requests += 1
        if use_randomized_body == "createCustomer":
            json_data = {
                "customerName": generate_random_name(),
                "customerEmail": generate_random_email()
            }
        elif use_randomized_body == "createAccount":
            json_data = {
                "customerID": generate_random_customer_id()
            }
        else:
            json_data = {
                "senderID": generate_random_sender_receiver_id(),
                "receiverID": generate_random_sender_receiver_id(),
                "amount": generate_random_amount()
            }
        response_time, success = get_response_time(url, json_data)
        response_times.append(response_time)
        if success:
            successful_requests += 1
        elapsed_time = time.time() - start_time
        if elapsed_time >= timeout:
            break

    # Calculate statistics
    avg_response_time = np.mean(response_times)
    max_response_time = np.max(response_times)
    std_dev_response_time = np.std(response_times)

    print("\n--- Response Time Statistics ---")
    print(f"Avg Response Time: {avg_response_time} seconds")
    print(f"Max Response Time: {max_response_time} seconds")
    print(f"Std Dev Response Time: {std_dev_response_time} seconds")
    print(f"Successful Requests: {successful_requests} out of {num_requests}")
    print(f"Success Rate: {successful_requests / num_requests * 100}%")
    print(f"Requests Per Second: {num_requests / timeout}")
    

if __name__ == "__main__":
    timeout = 10  # Specify the time for testing in seconds

    url = "http://localhost:8000/createAccount/john/12345"
    #url = "http://localhost:8000/createAccount"
    use_randomized_body = "createAccount"
    measure_response_times(url, timeout, use_randomized_body)

    url = "http://localhost:8000/createCustomer/john/12345"
    #url = "http://localhost:8000/createCustomer"
    use_randomized_body = "createCustomer"
    measure_response_times(url, timeout, use_randomized_body)

    url = "http://localhost:8000/transferAmount/john/12345"
    #url = "http://localhost:8000/transferAmount"
    use_randomized_body = "transferAmount"
    measure_response_times(url, timeout, use_randomized_body)

    url = "http://localhost:8000/transferAmountAndNotify/john/12345"
    #url = "http://localhost:8000/transferAmountAndNotify"
    use_randomized_body = "transferAmountAndNotify"
    measure_response_times(url, timeout, use_randomized_body)