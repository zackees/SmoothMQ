const AWS = require('sqs-legacy-adapter');
const dotenv = require('dotenv');
const { promisify } = require('util');
const path = require('path');

console.log(AWS);

// AWS.config.logger = console;

// .env file is located in the root of the project
const envFile = path.join(__dirname, '../../.env');

// dotenv.config();
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
        const listResponse = await sqs.listQueues({ QueueNamePrefix: queueName }).promise();
        console.log('Raw listQueues response:', listResponse);
        if (listResponse.QueueUrls && listResponse.QueueUrls.length > 0) {
            const queueUrl = listResponse.QueueUrls[0];
            console.log(`Queue already exists: ${queueUrl}`);
            return [queueUrl, false];
        } else {
            const createResponse = await sqs.createQueue({ QueueName: queueName }).promise();
            console.log('Raw createQueue response:', createResponse);
            const queueUrl = createResponse.QueueUrl;
            console.log(`Created queue: ${queueUrl}`);
            return [queueUrl, true];
        }
    } catch (error) {
        console.error('Error in createOrGetQueue:', error);
        if (error.$response) {
            console.error('Raw error response:', error.$response.body.toString());
        }
        throw error;
    }
}

async function runSqsTest(endpointUrl, awsSecretAccessKey, awsAccessId) {
    const params = {
        apiVersion: "2012-11-05",
        region: 'us-east-1',
        accessKeyId: awsAccessId,
        secretAccessKey: awsSecretAccessKey,
        endpoint: new AWS.Endpoint(endpointUrl),
        signatureVersion: 'v4',
    }
    console.log('Creating SQS client with params:', params)
    console.log(AWS.SQS);
    const sqs = new AWS.SQS(params);

    const queueName = 'my-test-que-for-testing';
    let queueUrl, queueCreated;

    try {
        [queueUrl, queueCreated] = await createOrGetQueue(sqs, queueName);
        console.log(`Queue URL: ${queueUrl}`);

        const sendResponse = await sqs.sendMessage({ QueueUrl: queueUrl, MessageBody: 'hello world' }).promise();
        console.log('Raw sendMessage response:', sendResponse);
        console.log('Sent a message to the queue');

        const receiveResponse = await sqs.receiveMessage({ QueueUrl: queueUrl, MaxNumberOfMessages: 1 }).promise();
        console.log('Raw receiveMessage response:', receiveResponse);
        if (receiveResponse.Messages && receiveResponse.Messages.length > 0) {
            const message = receiveResponse.Messages[0];
            console.log(`Received message: ${message.Body}`);
            const deleteResponse = await sqs.deleteMessage({ QueueUrl: queueUrl, ReceiptHandle: message.ReceiptHandle }).promise();
            console.log('Raw deleteMessage response:', deleteResponse);
            console.log('Deleted the message');
        } else {
            console.log('No messages in the queue');
        }

        await sleep(2000);
    } catch (error) {
        console.error('Error in runSqsTest:', error);
        if (error.$response) {
            console.error('Raw error response:', error.$response.body.toString());
        }
    } finally {
        if (queueUrl) {
            console.log(`Destroying queue: ${queueUrl}`);
            const deleteQueueResponse = await sqs.deleteQueue({ QueueUrl: queueUrl }).promise();
            console.log('Raw deleteQueue response:', deleteQueueResponse);
            console.log(`Destroyed queue: ${queueUrl}`);
        }
    }
}

async function main() {
    const awsSecretAccessKey = process.env.AWS_SECRET_ACCESS_KEY;
    const awsAccessId = process.env.AWS_ACCESS_KEY_ID;
    //const endpoints = ['http://localhost', 'https://jobs.kumquat.live'];
    const endpoints = ['http://localhost'];

    for (const endpointUrl of endpoints) {
        console.log(`\nTesting with endpoint: ${endpointUrl}`);
        await runSqsTest(endpointUrl, awsSecretAccessKey, awsAccessId);
    }
}

main().catch(error => console.error('Error in main:', error));
