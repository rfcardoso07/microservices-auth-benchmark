import datetime
import json
import random
import sys
import time

import numpy as np
import requests

## Global variables to keep track of existing IDs

customer_ids = []
account_ids = []
transaction_ids = []
notification_ids = []

## Helper functions to generate random values


def generate_random_email():
    domains = ["gmail.com", "yahoo.com", "hotmail.com", "outlook.com"]
    name = "".join(random.choices("abcdefghijklmnopqrstuvwxyz", k=8))
    domain = random.choice(domains)
    return f"{name}@{domain}"


def generate_random_name():
    names = ["Alice", "Bob", "Charlie", "David", "Emma", "Frank", "Grace", "Henry"]
    return random.choice(names)


def generate_random_customer_id():
    return random.choice(customer_ids)


def generate_random_account_id():
    return random.choice(account_ids)


def generate_random_customer_id_and_remove():
    global customer_ids
    id = random.choice(customer_ids)
    customer_ids.remove(id)
    return id


def generate_random_account_id_and_remove():
    global account_ids
    id = random.choice(account_ids)
    account_ids.remove(id)
    return id


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
        "customerEmail": generate_random_email(),
    }


def delete_customer_body():
    return {"customerID": generate_random_customer_id_and_remove()}


def get_customer_body():
    return {"customerID": generate_random_customer_id()}


def create_account_body():
    return {"customerID": generate_random_customer_id_and_remove()}


def delete_account_body():
    return {"accountID": generate_random_account_id_and_remove()}


def delete_accounts_by_customer_body():
    return {"customerID": generate_random_customer_id_and_remove()}


def get_account_body():
    return {"accountID": generate_random_account_id()}


def get_accounts_by_customer_body():
    return {"customerID": generate_random_customer_id()}


def add_to_balance_body():
    return {
        "accountID": generate_random_account_id(),
        "amount": generate_random_amount(),
    }


def subtract_from_balance_body():
    return {
        "accountID": generate_random_account_id(),
        "amount": generate_random_amount(),
    }


def transfer_amount_body():
    return {
        "senderID": generate_random_account_id(),
        "receiverID": generate_random_account_id(),
        "amount": generate_random_amount(),
    }


def transfer_amount_and_notify_body():
    return {
        "senderID": generate_random_account_id(),
        "receiverID": generate_random_account_id(),
        "amount": generate_random_amount(),
    }


def get_transaction_body():
    return {"transactionID": generate_random_transaction_id()}


def notify_body():
    return {
        "transactionID": generate_random_transaction_id(),
        "receiverID": generate_random_account_id(),
        "amount": generate_random_amount(),
    }


def get_notification_body():
    return {"notificationID": generate_random_notification_id()}


def get_balance_by_customer_body():
    return {"customerID": generate_random_customer_id_and_remove()}


def get_balance_history_body():
    return {
        "customerID": generate_random_customer_id(),
        "numberOfRecords": generate_random_number_of_records(),
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
    "notify": notify_body,
    "getNotification": get_notification_body,
    "getBalanceByCustomer": get_balance_by_customer_body,
    "getBalanceHistory": get_balance_history_body,
}

## Functions for making API calls and measuring response times


def fill_database_with_accounts(number_of_accounts: int, url: str):
    while number_of_accounts:
        json_data = create_account_body()
        try:
            response = requests.post(url, json=json_data)
            response.raise_for_status()  # Raise an exception for HTTP errors
        except Exception as e:
            print(f"Request for {url} with JSON {json_data} failed: {str(e)}")
            sys.exit(1)
        number_of_accounts -= 1

    return


def get_response_time(url: str, json_data: dict):
    start_time = time.time()
    try:
        response = requests.post(url, json=json_data)
        response.raise_for_status()  # Raise an exception for HTTP errors
    except Exception as e:
        if e.response.status_code == 401:
            pass
        else:
            print(f"Request for {url} with JSON {json_data} failed: {str(e)}")
            sys.exit(1)

    end_time = time.time()

    return end_time - start_time


def measure_response_times(
    number_of_requests: int,
    app_url: str,
    endpoint: str,
    app_version: str,
    auth: str,
    export_url: str,
):
    response_times = []
    start_time = time.time()

    i = 0

    while number_of_requests:
        response_time = get_response_time(app_url, endpoint_bodies[endpoint]())
        response_times.append(response_time)
        number_of_requests -= 1
        i += 1
        if (i % 1000) == 0:
            print(f"Reached {i} requests...")
            if i == 1000:
                # Start measuring resource consumption
                _ = requests.post(f"http://localhost:5000/start/{app_version}/{endpoint}")

    elapsed_time = time.time() - start_time

    # Calculate statistics
    min_response_time = np.min(response_times)
    max_response_time = np.max(response_times)
    avg_response_time = np.mean(response_times)
    std_dev_response_time = np.std(response_times)

    stats = {
        "application version": app_version,
        "auth": auth,
        "endpoint": endpoint,
        "number of requests": len(response_times),
        "min response time": min_response_time,
        "max response time": max_response_time,
        "avg response time": avg_response_time,
        "std dev response time": std_dev_response_time,
        "elapsed time": elapsed_time,
        "requests per second": (len(response_times) / elapsed_time),
    }

    # export_response = requests.post(export_url, json=stats)
    # export_response.raise_for_status()

    return stats


def measure_response_times_randomizing_validity(
    number_of_requests: int,
    app_invalid_url: str,
    app_valid_url: str,
    endpoint: str,
    app_version: str,
    auth: str,
    export_url: str,
):
    urls = [app_invalid_url, app_valid_url]
    counts = [0, 0]
    target_count = int(number_of_requests / 2)
    response_times = []
    start_time = time.time()

    while number_of_requests:
        r = random.randint(0, 1)
        valid = r if counts[r] < target_count else int(not r)
        counts[valid] += 1

        response_time = get_response_time(urls[valid], endpoint_bodies[endpoint]())
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
        "auth": auth,
        "endpoint": endpoint,
        "number of requests": len(response_times),
        "min response time": min_response_time,
        "max response time": max_response_time,
        "avg response time": avg_response_time,
        "std dev response time": std_dev_response_time,
        "elapsed time": elapsed_time,
        "requests per second": (len(response_times) / elapsed_time),
    }

    # export_response = requests.post(export_url, json=stats)
    # export_response.raise_for_status()

    return stats


## Main code

app_version = "decentralized"
auth = "valid"
number_of_requests = 15000
host = "http://localhost:8000"
export_url = "http://webhook.site"

valid_user = "john"
valid_user_password = "12345"
invalid_user = "bob"
invalid_user_password = "34567"

customer_ids = [i for i in range(1, number_of_requests + 1)]
account_ids = [i for i in range(1, number_of_requests + 1)]
transaction_ids = [i for i in range(1, number_of_requests + 1)]
notification_ids = [i for i in range(1, number_of_requests + 1)]

ordered_endpoints = [
    "createCustomer",
    "getCustomer",
    "createAccount",
    "getAccount",
    "getAccountsByCustomer",
    "addToBalance",
    "subtractFromBalance",
    "transferAmount",
    "transferAmountAndNotify",
    "getTransaction",
    "notify",
    "getNotification",
    "getBalanceByCustomer",
    "getBalanceHistory",
    "deleteCustomer",
    "deleteAccount",
    "deleteAccountsByCustomer",
]

if app_version == "noauth":
    for endpoint in ordered_endpoints:

        if endpoint == "deleteAccount":
            ## need to fill DB again before running
            print(f"Filling DB before running {endpoint}...")
            fill_database_with_accounts(number_of_requests, f"{host}/createAccount")
            account_ids = [
                i for i in range(number_of_requests + 1, 2 * number_of_requests + 1)
            ]

        if endpoint == "deleteAccountsByCustomer":
            ## need to fill DB again before running
            print(f"Filling DB before running {endpoint}...")
            fill_database_with_accounts(number_of_requests, f"{host}/createAccount")
            customer_ids = [i for i in range(1, number_of_requests + 1)]
            account_ids = [
                i for i in range(2 * number_of_requests + 1, 3 * number_of_requests + 1)
            ]

        print(
            f"Measuring for {endpoint} with {app_version} version started at {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]}"
        )
        app_url = f"{host}/{endpoint}"
        stats = measure_response_times(
            number_of_requests, app_url, endpoint, app_version, auth, export_url
        )
        # Write the JSON data to the file
        with open("times.log", "a") as file:
            file.write(json.dumps(stats) + "\n")
            file.close()
        print(
            f"Measuring for {endpoint} with {app_version} version finished at {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]}"
        )
        print(stats)

        if endpoint in [
            "createAccount",
            "getBalanceByCustomer",
            "deleteCustomer",
            "deleteAccount",
        ]:
            ## need to restore customer IDs list after running these
            print(f"Restoring customer_ids after running {endpoint}...")
            customer_ids = [i for i in range(1, number_of_requests + 1)]
else:
    if auth == "valid":
        ## only valid accesses
        for endpoint in ordered_endpoints:

            if endpoint == "deleteAccount":
                ## need to fill DB again before running
                print(f"Filling DB before running {endpoint}...")
                fill_database_with_accounts(
                    number_of_requests,
                    f"{host}/createAccount/{valid_user}/{valid_user_password}",
                )
                account_ids = [
                    i for i in range(number_of_requests + 1, 2 * number_of_requests + 1)
                ]

            if endpoint == "deleteAccountsByCustomer":
                ## need to fill DB again before running
                print(f"Filling DB before running {endpoint}...")
                fill_database_with_accounts(
                    number_of_requests,
                    f"{host}/createAccount/{valid_user}/{valid_user_password}",
                )
                customer_ids = [i for i in range(1, number_of_requests + 1)]
                account_ids = [
                    i
                    for i in range(
                        2 * number_of_requests + 1, 3 * number_of_requests + 1
                    )
                ]

            print(
                f"Measuring for {endpoint} with {app_version} version started at {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]}"
            )
            app_url = f"{host}/{endpoint}/{valid_user}/{valid_user_password}"
            stats = measure_response_times(
                number_of_requests, app_url, endpoint, app_version, auth, export_url
            )
            # Write the JSON data to the file
            with open("times.log", "a") as file:
                file.write(json.dumps(stats) + "\n")
                file.close()
            print(
                f"Measuring for {endpoint} with {app_version} version finished at {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]}"
            )
            print(stats)

            if endpoint in [
                "createAccount",
                "getBalanceByCustomer",
                "deleteCustomer",
                "deleteAccount",
            ]:
                ## need to restore customer IDs list after running these
                print(f"Restoring customer_ids after running {endpoint}...")
                customer_ids = [i for i in range(1, number_of_requests + 1)]

    if auth == "invalid":
        ## only invalid accesses
        for endpoint in ordered_endpoints:

            if endpoint == "deleteAccount":
                account_ids = [
                    i for i in range(number_of_requests + 1, 2 * number_of_requests + 1)
                ]

            if endpoint == "deleteAccountsByCustomer":
                customer_ids = [i for i in range(1, number_of_requests + 1)]
                account_ids = [
                    i
                    for i in range(
                        2 * number_of_requests + 1, 3 * number_of_requests + 1
                    )
                ]

            print(
                f"Measuring for {endpoint} with {app_version} version started at {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]}"
            )
            app_url = f"{host}/{endpoint}/{invalid_user}/{invalid_user_password}"
            stats = measure_response_times(
                number_of_requests, app_url, endpoint, app_version, auth, export_url
            )
            # Write the JSON data to the file
            with open("times.log", "a") as file:
                file.write(json.dumps(stats) + "\n")
                file.close()
            print(
                f"Measuring for {endpoint} with {app_version} version finished at {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]}"
            )
            print(stats)

            if endpoint in [
                "createAccount",
                "getBalanceByCustomer",
                "deleteCustomer",
                "deleteAccount",
            ]:
                ## need to restore customer IDs list after running these
                print(f"Restoring customer_ids after running {endpoint}...")
                customer_ids = [i for i in range(1, number_of_requests + 1)]

    if auth == "mixed":
        ## random mixed accesses
        for endpoint in ordered_endpoints:
            if endpoint == "getCustomer":
                customer_ids = [i for i in range(1, (int(number_of_requests / 2)) + 1)]

            if endpoint == "createAccount":
                customer_ids = [i for i in range(1, number_of_requests + 1)]

            if endpoint == "getAccount":
                account_ids = [i for i in range(1, (int(number_of_requests / 2)) + 1)]

            if endpoint == "deleteAccount":
                ## need to fill DB again before running
                print(f"Filling DB before running {endpoint}...")
                fill_database_with_accounts(
                    number_of_requests,
                    f"{host}/createAccount/{valid_user}/{valid_user_password}",
                )
                account_ids = [
                    i for i in range(number_of_requests + 1, 2 * number_of_requests + 1)
                ]

            if endpoint == "deleteAccountsByCustomer":
                ## need to fill DB again before running
                print(f"Filling DB before running {endpoint}...")
                fill_database_with_accounts(
                    number_of_requests,
                    f"{host}/createAccount/{valid_user}/{valid_user_password}",
                )
                customer_ids = [i for i in range(1, number_of_requests + 1)]
                account_ids = [
                    i
                    for i in range(
                        2 * number_of_requests + 1, 3 * number_of_requests + 1
                    )
                ]

            print(
                f"Measuring for {endpoint} with {app_version} version started at {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]}"
            )
            app_invalid_url = (
                f"{host}/{endpoint}/{invalid_user}/{invalid_user_password}"
            )
            app_valid_url = f"{host}/{endpoint}/{valid_user}/{valid_user_password}"
            stats = measure_response_times_randomizing_validity(
                number_of_requests,
                app_invalid_url,
                app_valid_url,
                endpoint,
                app_version,
                auth,
                export_url,
            )
            # Write the JSON data to the file
            with open("times.log", "a") as file:
                file.write(json.dumps(stats) + "\n")
                file.close()
            print(
                f"Measuring for {endpoint} with {app_version} version finished at {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]}"
            )
            print(stats)

            if endpoint in [
                "createAccount",
                "getBalanceByCustomer",
                "deleteCustomer",
                "deleteAccount",
            ]:
                ## need to restore customer IDs list after running these
                print(f"Restoring customer_ids after running {endpoint}...")
                customer_ids = [i for i in range(1, number_of_requests + 1)]
