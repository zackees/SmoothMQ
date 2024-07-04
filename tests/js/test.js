const { SQSClient, ListQueuesCommand, CreateQueueCommand, SendMessageCommand, ReceiveMessageCommand, DeleteMessageCommand, DeleteQueueCommand } = require('@aws-sdk/client-sqs');
const dotenv = require('dotenv');
const { promisify } = require('util');
const path = require('path');

// .env file is located in the root of the project
const envFile = path.join(__dirname, '../../.env');

dotenv.config({ path: envFile });

// assert that the AWS_SECRET_ACCESS_KEY is set
if (!process.env.AWS_SECRET_ACCESS_KEY) {
    console.error('AWS_SECRET_ACCESS_KEY is not set');
    process.exit(1);
} else {
    // print out the contents of the env file
    const parsed = dotenv.parse(require('fs').readFileSync(envFile, 'utf-8'));
    console.log("\nContents of .env file:");
    Object.entries(parsed).forEach(([key, value]) => {
        console.log(`${key}=${value}`);
    });
    console.log("");
}

const sleep = promisify(setTimeout);

async function createOrGetQueue(sqs, queueName) {
    try {
        const options = { QueueNamePrefix: queueName };
        const listResponse = await sqs.send(new ListQueuesCommand(options));
        if (listResponse.QueueUrls && listResponse.QueueUrls.length > 0) {
            const queueUrl = listResponse.QueueUrls[0];
            console.log(`Queue already exists: ${queueUrl}`);
            return [queueUrl, false];
        } else {
            const createResponse = await sqs.send(new CreateQueueCommand({ QueueName: queueName }));
            const queueUrl = createResponse.QueueUrl;
            console.log(`Created queue: ${queueUrl}`);
            return [queueUrl, true];
        }
    } catch (error) {
        console.error('Error in createOrGetQueue:', error);
        throw error;
    }
}

async function runSqsTest(endpointUrl, awsSecretAccessKey, awsAccessId) {
    const clientConfig = {
        region: 'us-east-1',
        credentials: {
            accessKeyId: awsAccessId,
            secretAccessKey: awsSecretAccessKey,
        },
        endpoint: endpointUrl,
    };
    console.log('Creating SQS client with config:', clientConfig);
    const sqs = new SQSClient(clientConfig);

    const queueName = 'my-test-que-for-testing';
    let queueUrl, queueCreated;

    try {
        [queueUrl, queueCreated] = await createOrGetQueue(sqs, queueName);
        console.log(`Queue URL: ${queueUrl}`);

        await sqs.send(new SendMessageCommand({ QueueUrl: queueUrl, MessageBody: 'hello world' }));
        console.log('Sent a message to the queue');

        const receiveResponse = await sqs.send(new ReceiveMessageCommand({ QueueUrl: queueUrl, MaxNumberOfMessages: 1 }));
        if (receiveResponse.Messages && receiveResponse.Messages.length > 0) {
            const message = receiveResponse.Messages[0];
            console.log(`Received message: ${message.Body}`);

            await sqs.send(new DeleteMessageCommand({ QueueUrl: queueUrl, ReceiptHandle: message.ReceiptHandle }));
            console.log('Deleted the message');
        } else {
            console.log('No messages in the queue');
        }

        await sleep(2000);
    } catch (error) {
        console.error('Error in runSqsTest:', error);
    } finally {
        if (queueUrl) {
            console.log(`Destroying queue: ${queueUrl}`);
            await sqs.send(new DeleteQueueCommand({ QueueUrl: queueUrl }));
            console.log(`Destroyed queue: ${queueUrl}`);
        }
    }
}

async function main() {
    const awsSecretAccessKey = process.env.AWS_SECRET_ACCESS_KEY;
    const awsAccessId = process.env.AWS_ACCESS_KEY_ID;
    const endpoints = ['http://localhost', 'https://jobs.kumquat.live'];

    for (const endpointUrl of endpoints) {
        console.log(`\nTesting with endpoint: ${endpointUrl}`);
        await runSqsTest(endpointUrl, awsSecretAccessKey, awsAccessId);
    }
}

main().catch(error => console.error('Error in main:', error));
