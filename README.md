# SmoothMQ Fork

SmoothMQ is a drop-in replacement for SQS with a much smoother developer experience.
It has a functional UI, observability, tracing, message scheduling, and rate-limiting.
SmoothMQ lets you run a private SQS instance on any cloud.

## Fork Info

This fork has a number of improvements which allow it to be more easily deployed.

  * port 3000 -> localhost:80/ui and is protected with basic http authorization.
    * user: user
    * password: <AMAZON_SECRET_KEY>
  * port 3001 -> localhost:80/ and is protected with the amazon secret key 
  * Can be dropped into Render.com / Digital Ocean
    * If you allocate a disk at /var/data then the SQS queue will be persistant across reboots/deploys
  * This code base has a test suite
    * javascript
    * python
  * Deleting of Queues is implemented

## Testing

  * We have three tests
    * Javascript
      * `test.js` shows how to connect using the new Amazon SQS client driver
      * `test-legacy.js` shows how to connect using a wrapper for the new Amazon SQS Driver
        * `const AWS = require('sqs-legacy-adapter');`
        * instead of `const AWS = require('aws-sdk');`
    * python
      * `test.py`
     
## Footguns

  * This SQS package only supports signing v4.


## Login Credentials

/ui will go to the ui of the service.
  * user: "user"
  * pass: (see .env file for password)

Note that the AWS_SECRET_ACCESS_KEY you should use in the api is located in the `.env` file.

It will be the same as the password used to access the /ui page.


## Quick Start

Clone the repo and then invoke:

```bash
./dev
```

Then optionally invoke the tests

```
./test
```

## Getting Started

SmoothMQ deploys as a single go binary and can be used by any existing SQS client.

## Running

This will run a UI on `:3000` and an SQS-compatible server on `:3001`.

```
$ go run .
```

## UI

The UI lets you manage queues and search individual messages.

![Dashboard UI](docs/queue.gif)


## Other links

  * https://hub.docker.com/r/roribio16/alpine-sqs/
