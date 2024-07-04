package sqs

type SendMessagePayload struct {
	QueueUrl               string                           `json:"QueueUrl"`
	MessageBody            string                           `json:"MessageBody"`
	DelaySeconds           int                              `json:"DelaySeconds,omitempty"`
	MessageAttributes      map[string]MessageAttributeValue `json:"MessageAttributes,omitempty"`
	MessageDeduplicationId string                           `json:"MessageDeduplicationId,omitempty"`
	MessageGroupId         string                           `json:"MessageGroupId,omitempty"`
}

type MessageAttributeValue struct {
	StringValue     string   `json:"StringValue,omitempty"`
	BinaryValue     string   `json:"BinaryValue,omitempty"`
	StringListValue []string `json:"StringListValue,omitempty"`
	BinaryListValue [][]byte `json:"BinaryListValue,omitempty"`
	DataType        string   `json:"DataType"`
}

type SendMessageResponse struct {
	MD5OfMessageBody       string `json:"MD5OfMessageBody"`
	MD5OfMessageAttributes string `json:"MD5OfMessageAttributes"`
	MessageId              string `json:"MessageId"`
	SequenceNumber         string `json:"SequenceNumber,omitempty"`
}

type ReceiveMessageRequest struct {
	QueueUrl                string   `json:"QueueUrl"`
	AttributeNames          []string `json:"AttributeNames,omitempty"`
	MessageAttributeNames   []string `json:"MessageAttributeNames,omitempty"`
	MaxNumberOfMessages     int      `json:"MaxNumberOfMessages,omitempty"`
	VisibilityTimeout       int      `json:"VisibilityTimeout,omitempty"`
	WaitTimeSeconds         int      `json:"WaitTimeSeconds,omitempty"`
	ReceiveRequestAttemptId string   `json:"ReceiveRequestAttemptId,omitempty"`
}

type ReceiveMessageResponse struct {
	Messages []Message `json:"Messages"`
}

type Message struct {
	MessageId              string                      `json:"MessageId"`
	ReceiptHandle          string                      `json:"ReceiptHandle"`
	MD5OfBody              string                      `json:"MD5OfBody"`
	Body                   string                      `json:"Body"`
	Attributes             map[string]string           `json:"Attributes"`
	MD5OfMessageAttributes string                      `json:"MD5OfMessageAttributes,omitempty"`
	MessageAttributes      map[string]MessageAttribute `json:"MessageAttributes,omitempty"`
}

type MessageAttribute struct {
	StringValue      string   `json:"StringValue,omitempty"`
	BinaryValue      []byte   `json:"BinaryValue,omitempty"`
	StringListValues []string `json:"StringListValues,omitempty"`
	BinaryListValues [][]byte `json:"BinaryListValues,omitempty"`
	DataType         string   `json:"DataType"`
}

type DeleteMessageRequest struct {
	QueueUrl      string `json:"QueueUrl"`
	ReceiptHandle string `json:"ReceiptHandle"`
}

type ListQueuesRequest struct {
	QueueNamePrefix string `json:"QueueNamePrefix,omitempty"`
}

type ListQueuesResponse struct {
	QueueUrls []string `json:"QueueUrls"`
}

type CreateQueueRequest struct {
	QueueName  string            `json:"QueueName"`
	Attributes map[string]string `json:"Attributes,omitempty"`
	Tags       map[string]string `json:"Tags,omitempty"`
}

type CreateQueueResponse struct {
	QueueUrl string `json:"QueueUrl"`
}

type GetQueueAttributesRequest struct {
	QueueUrl       string   `json:"QueueUrl"`
	AttributeNames []string `json:"AttributeNames,omitempty"`
}

type GetQueueAttributesResponse struct {
	Attributes map[string]string `json:"Attributes"`
}

type PurgeQueueRequest struct {
	QueueUrl string `json:"QueueUrl"`
}

type PurgeQueueResponse struct {
	Success bool `json:"Success"`
}
