import requests
import time
import numpy as np
import random
import sys

## Global variables to keep track of existing IDs

customer_ids = []
account_ids = []
transaction_ids = []
notification_ids = []

## Helper functions to generate random values

def generate_random_email():
    domains = ["gmail.com", "yahoo.com", "hotmail.com", "outlook.com"]
    name = ''.join(random.choices('abcdefghijklmnopqrstuvwxyz', k=8))
    domain = random.choice(domains)
    return f"{name}@{domain}"

def generate_random_name():
    names = ["Alice", "Bob", "Charlie", "David", "Emma", "Frank", "Grace", "Henry"]
    return random.choice(names)

def generate_random_customer_id():
    return random.choice(customer_ids)

def generate_random_account_id():
    return random.choice(account_ids)

def generate_random_transaction_id():
    return random.choice(transaction_ids)

def generate_random_notification_id():
    return random.choice(notification_ids)

def generate_random_amount():
    return random.randint(1, 1000)

def generate_random_number_of_records():
    return random.randint(1, 10)

## Functions for generating random request bodies

def create_customer_body():
    return {
            "customerName": generate_random_name(),
            "customerEmail": generate_random_email()
    }

def delete_customer_body():
    return {
            "customerID": generate_random_customer_id()
    }

def get_customer_body():
    return {
            "customerID": generate_random_customer_id()
    }

def create_account_body():
    return {
            "customerID": generate_random_customer_id()
    }

def delete_account_body():
    return {
            "accountID": generate_random_account_id()
    }

def delete_accounts_by_customer_body():
    return {
            "customerID": generate_random_customer_id()
    }

def get_account_body():
    return {
            "accountID": generate_random_account_id()
    }

def get_accounts_by_customer_body():
    return {
            "customerID": generate_random_customer_id()
    }

def add_to_balance_body():
    return {
            "accountID": generate_random_account_id(),
            "amount": generate_random_amount()
    }

def subtract_from_balance_body():
    return {
            "accountID": generate_random_account_id(),
            "amount": generate_random_amount()
    }

def transfer_amount_body():
    return {
            "senderID": generate_random_account_id(),
            "receiverID": generate_random_account_id(),
            "amount": generate_random_amount()
    }

def transfer_amount_and_notify_body():
    return {
            "senderID": generate_random_account_id(),
            "receiverID": generate_random_account_id(),
            "amount": generate_random_amount()
    }

def get_transaction_body():
    return {
            "transactionID": generate_random_transaction_id()
    }

def notify_request_body():
    return {
            "transactionID": generate_random_transaction_id(),
            "receiverID": generate_random_account_id(),
            "amount": generate_random_amount()
    }

def get_notification_body():
    return {
            "notificationID": generate_random_notification_id()
    }

def get_balance_by_customer_body():
    return {
            "customerID": generate_random_customer_id()
    }

def get_balance_history_body():
    return {
            "customerID": generate_random_customer_id(),
            "numberOfRecords": generate_random_number_of_records
    }

## Dict of endpoints and their body generation functions

endpoint_bodies = {
    "createCustomer": create_customer_body,
    "deleteCustomer": delete_customer_body,
    "getCustomer": get_customer_body,
    "createAccount": create_account_body,
    "deleteAccount": delete_account_body,
    "deleteAccountsByCustomer": delete_accounts_by_customer_body,
    "getAccount": get_account_body,
    "getAccountsByCustomer": get_accounts_by_customer_body,
    "addToBalance": add_to_balance_body,
    "subtractFromBalance": subtract_from_balance_body,
    "transferAmount": transfer_amount_body,
    "transferAmountAndNotify": transfer_amount_and_notify_body,
    "getTransaction": get_transaction_body,
    "notifyRequest": notify_request_body,
    "getNotification": get_notification_body,
    "getBalanceByCustomer": get_balance_by_customer_body,
    "getBalanceHistory": get_balance_history_body
}


## Functions for making API calls and measuring response times

def get_response_time(url: str, json_data: dict):
    start_time = time.time()
    try:
        response = requests.post(url, json=json_data)
        response.raise_for_status()  # Raise an exception for HTTP errors
    except Exception as e:
        print(f"Request for {url} with JSON {json_data} failed: {str(e)}")
        sys.exit(1)
    
    end_time = time.time()

    return end_time - start_time

def measure_response_times(number_of_requests: int, app_url: str, endpoint: str, app_version: str, export_url: str):
    response_times = []
    start_time = time.time()

    while number_of_requests:
        response_time = get_response_time(url, endpoint_bodies[endpoint]())
        response_times.append(response_time)
        number_of_requests -= 1

    elapsed_time = time.time() - start_time

    # Calculate statistics
    min_response_time = np.min(response_times)
    max_response_time = np.max(response_times)
    avg_response_time = np.mean(response_times)
    std_dev_response_time = np.std(response_times)

    stats = {
        "application version": app_version, 
        "endpoint": endpoint,
        "number of requests": number_of_requests,
        "min response time": min_response_time,
        "max response time": max_response_time,
        "avg response time": avg_response_time,
        "std dev response time": std_dev_response_time,
        "requests per second": (number_of_requests / elapsed_time)
    }

    _ = requests.post(export_url, json=stats)
    return
    

## Main code

if __name__ == "__main__":
    app_version = "noauth"
    number_of_requests = 100
    export_url = "webhook.site"

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