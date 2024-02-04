import json

from flask import Flask, jsonify, request

app = Flask(__name__)


@app.route("/", methods=["POST"])
def index():
    # Write the JSON data to the file
    with open("requests.log", "a") as file:
        file.write(json.dumps(request.get_json()) + "\n")

    # Return a 200 OK response
    return jsonify({"message": "Request received successfully!"}), 200


if __name__ == "__main__":
    app.run(port=5000)
