package sqs

/*
Docs:
https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_ReceiveMessage.html
https://docs.aws.amazon.com/cli/latest/reference/sqs/delete-message.html

Testing:
aws sqs list-queues --endpoint-url http://localhost:3001
aws sqs send-message --queue-url https://sqs.us-east-1.amazonaws.com/1/a --message-body "hello world" --endpoint-url http://localhost:3001
aws sqs receive-message --queue-url https://sqs.us-east-1.amazonaws.com/1/a --endpoint-url http://localhost:3001
aws sqs delete-message --receipt-handle x --queue-url https://sqs.us-east-1.amazonaws.com/1/a --endpoint-url http://localhost:3001
aws sqs create-queue --queue-name b --endpoint-url http://localhost:3001
aws sqs get-queue-attributes --queue-url https://sqs.us-east-1.amazonaws.com/1/a --endpoint-url http://localhost:3001
*/

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"q/models"
	"strconv"
	"strings"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type SQS struct {
	app           *fiber.App
	queue         models.Queue
	tenantManager models.TenantManager
}

func NewSQS(queue models.Queue, tenantManager models.TenantManager) *SQS {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(zerolog.ErrorLevel)

	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &logger,
	}))

	s := &SQS{
		app:           app,
		queue:         queue,
		tenantManager: tenantManager,
	}

	app.Use(s.authMiddleware)
	app.Post("/*", s.Action)

	return s
}

func (s *SQS) authMiddleware(c *fiber.Ctx) error {
	r, err := adaptor.ConvertRequest(c, false)
	if err != nil {
		return err
	}

	awsHeader, err := ParseAuthorizationHeader(r)
	if err != nil {
		return err
	}

	tenantId, secretKey, err := s.tenantManager.GetAWSSecretKey(awsHeader.AccessKey, awsHeader.Region)
	if err != nil {
		return err
	}

	err = ValidateAWSRequest(awsHeader, secretKey, r)
	if err != nil {
		return err
	}

	c.Locals("tenantId", tenantId)
	return c.Next()
}

func (s *SQS) Start() error {
	fmt.Println("SQS Endpoint: http://localhost:3001")
	return s.app.Listen(":3001")
}

func (s *SQS) Stop() error {
	return s.app.Shutdown()
}

func (s *SQS) Action(c *fiber.Ctx) error {
	awsMethodHeader, ok := c.GetReqHeaders()["X-Amz-Target"]
	if !ok {
		return errors.New("X-Amz-Target header not found")
	}
	awsMethod := awsMethodHeader[0]

	var r *http.Request = &http.Request{}
	fasthttpadaptor.ConvertRequest(c.Context(), r, false)

	tenantId := c.Locals("tenantId").(int64)

	// log.Println(awsMethod)

	switch awsMethod {
	case "AmazonSQS.SendMessage":
		return s.SendMessage(c, tenantId)
	case "AmazonSQS.ReceiveMessage":
		return s.ReceiveMessage(c, tenantId)
	case "AmazonSQS.DeleteMessage":
		return s.DeleteMessage(c, tenantId)
	case "AmazonSQS.ListQueues":
		return s.ListQueues(c, tenantId)
	case "AmazonSQS.CreateQueue":
		return s.CreateQueue(c, tenantId)
	case "AmazonSQS.GetQueueAttributes":
		return s.GetQueueAttributes(c, tenantId)
	case "AmazonSQS.PurgeQueue":
		return s.PurgeQueue(c, tenantId)
	case "AmazonSQS.DeleteQueue":
		return s.DeleteQueue(c, tenantId)
	default:
		return fmt.Errorf("SQS method %s not implemented", awsMethod)
	}

}

func (s *SQS) DeleteQueue(c *fiber.Ctx, tenantId int64) error {
	req := &DeleteQueueRequest{}

	err := json.Unmarshal(c.Body(), req)
	if err != nil {
		return err
	}

	tokens := strings.Split(req.QueueUrl, "/")
	queue := tokens[len(tokens)-1]

	err = s.queue.DeleteQueue(tenantId, queue)
	if err != nil {
		return err
	}

	return c.JSON(DeleteQueueResponse{})
}

func (s *SQS) PurgeQueue(c *fiber.Ctx, tenantId int64) error {
	req := &PurgeQueueRequest{}

	err := json.Unmarshal(c.Body(), req)
	if err != nil {
		return err
	}

	tokens := strings.Split(req.QueueUrl, "/")
	queue := tokens[len(tokens)-1]

	messages := s.queue.Filter(tenantId, queue, models.FilterCriteria{})
	for _, msg := range messages {
		s.queue.Delete(tenantId, queue, msg)
	}

	rc := PurgeQueueResponse{
		Success: true,
	}

	return c.JSON(rc)
}

func (s *SQS) GetQueueAttributes(c *fiber.Ctx, tenantId int64) error {
	req := &GetQueueAttributesRequest{}

	err := json.Unmarshal(c.Body(), req)
	if err != nil {
		return err
	}

	tokens := strings.Split(req.QueueUrl, "/")
	queue := tokens[len(tokens)-1]

	stats := s.queue.Stats(tenantId, queue)

	rc := GetQueueAttributesResponse{
		Attributes: map[string]string{
			"ApproximateNumberOfMessages": fmt.Sprintf("%d", stats.TotalMessages),
		},
	}

	return c.JSON(rc)
}

func (s *SQS) CreateQueue(c *fiber.Ctx, tenantId int64) error {
	req := &CreateQueueRequest{}

	err := json.Unmarshal(c.Body(), req)
	if err != nil {
		return err
	}

	err = s.queue.CreateQueue(tenantId, req.QueueName)
	if err != nil {
		return err
	}

	queueUrl := fmt.Sprintf("https://sqs.us-east-1.amazonaws.com/%d/%s", tenantId, req.QueueName)
	rc := CreateQueueResponse{
		QueueUrl: queueUrl,
	}

	return c.JSON(rc)
}

func (s *SQS) ListQueues(c *fiber.Ctx, tenantId int64) error {
	queues, err := s.queue.ListQueues(tenantId)
	if err != nil {
		return err
	}

	queueUrls := make([]string, len(queues))

	for i, queue := range queues {
		queueUrls[i] = fmt.Sprintf("https://sqs.us-east-1.amazonaws.com/%d/%s", tenantId, queue)
	}

	rc := ListQueuesResponse{
		QueueUrls: queueUrls,
	}

	return c.JSON(rc)
}

func (s *SQS) SendMessage(c *fiber.Ctx, tenantId int64) error {
	// TODO: make this configurable on queue
	visibilityTimeout := 30

	req := &SendMessagePayload{}

	err := json.Unmarshal(c.Body(), req)
	if err != nil {
		return err
	}

	tokens := strings.Split(req.QueueUrl, "/")
	queue := tokens[len(tokens)-1]

	kv := make(map[string]string)
	for k, v := range req.MessageAttributes {
		kv[k+"_DataType"] = v.DataType
		if v.DataType == "String" {
			kv[k] = v.StringValue
		} else if v.DataType == "Number" {
			kv[k] = v.StringValue
		} else if v.DataType == "Binary" {
			kv[k] = v.BinaryValue
		}
	}

	messageId, err := s.queue.Enqueue(tenantId, queue, req.MessageBody, kv, req.DelaySeconds, visibilityTimeout)
	if err != nil {
		return err
	}

	hasher := md5.New()
	hasher.Write([]byte(req.MessageBody))

	response := SendMessageResponse{
		MessageId:        fmt.Sprintf("%d", messageId),
		MD5OfMessageBody: hex.EncodeToString(hasher.Sum(nil)),
	}

	return c.JSON(response)
}

type DeleteQueueRequest struct {
	QueueUrl string `json:"QueueUrl"`
}

type DeleteQueueResponse struct {}

func (s *SQS) ReceiveMessage(c *fiber.Ctx, tenantId int64) error {
	req := &ReceiveMessageRequest{}

	err := json.Unmarshal(c.Body(), req)
	if err != nil {
		return err
	}

	if req.MaxNumberOfMessages == 0 {
		req.MaxNumberOfMessages = 1
	}

	// log.Println(req)

	tokens := strings.Split(req.QueueUrl, "/")
	queue := tokens[len(tokens)-1]

	messages, err := s.queue.Dequeue(tenantId, queue, req.MaxNumberOfMessages)
	if err != nil {
		return err
	}

	response := ReceiveMessageResponse{
		Messages: make([]Message, len(messages)),
	}

	hasher := md5.New()

	for i, message := range messages {
		hasher.Reset()
		hasher.Write(message.Message)

		response.Messages[i] = Message{
			MessageId:         fmt.Sprintf("%d", message.ID),
			ReceiptHandle:     fmt.Sprintf("%d", message.ID),
			Body:              string(message.Message),
			MessageAttributes: make(map[string]MessageAttribute),
			MD5OfBody:         hex.EncodeToString(hasher.Sum(nil)),
		}

		for k, v := range message.KeyValues {
			if strings.HasSuffix(k, "_DataType") {
				continue
			}
			attr := MessageAttribute{
				DataType: message.KeyValues[k+"_DataType"],
			}
			if attr.DataType == "String" {
				attr.StringValue = v
			} else if attr.DataType == "Number" {
				attr.StringValue = v
			} else if attr.DataType == "Binary" {
				data, err := base64.StdEncoding.DecodeString(v)
				if err != nil {
					log.Println(message.ID, err)
				} else {
					attr.BinaryValue = data
				}
			}

			response.Messages[i].MessageAttributes[k] = attr
		}
	}

	return c.JSON(response)
}

func (s *SQS) DeleteMessage(c *fiber.Ctx, tenantId int64) error {
	req := &DeleteMessageRequest{}

	err := json.Unmarshal(c.Body(), req)
	if err != nil {
		return err
	}

	tokens := strings.Split(req.QueueUrl, "/")
	queue := tokens[len(tokens)-1]

	messageId, err := strconv.ParseInt(req.ReceiptHandle, 10, 64)
	if err != nil {
		return err
	}

	err = s.queue.Delete(tenantId, queue, messageId)
	if err != nil {
		return err
	}

	return nil
}
