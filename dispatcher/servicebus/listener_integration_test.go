package servicebus

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"pack.ag/amqp"

	"github.com/lawrencegripper/mlops/dispatcher/helpers"
	"github.com/lawrencegripper/mlops/dispatcher/types"
	log "github.com/sirupsen/logrus"
)

func prettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}

// TestNewListener performs an end-2-end integration test on the listener talking to Azure ServiceBus
func TestIntegrationNewListener(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Paniced: %v", prettyPrintStruct(r))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	config := &types.Configuration{
		ClientID:            os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret:        os.Getenv("AZURE_CLIENT_SECRET"),
		ResourceGroup:       os.Getenv("AZURE_RESOURCE_GROUP"),
		SubscriptionID:      os.Getenv("AZURE_SUBSCRIPTION_ID"),
		TenantID:            os.Getenv("AZURE_TENANT_ID"),
		ServiceBusNamespace: os.Getenv("AZURE_SERVICEBUS_NAMESPACE"),
		Hostname:            "Test",
		ModuleName:          helpers.RandomName(8),
		SubscribesToEvent:   "ExampleEvent",
		LogLevel:            "Debug",
		JobConfig: &types.JobConfig{
			JobRetryCount: 1337,
		},
	}
	listener := NewListener(ctx, config)
	// Remove topic to ensure each test has a clean topic to work with
	defer deleteSubscription(listener, config)

	nonce := time.Now().String()
	sender := createAmqpSender(listener)
	err := sender.Send(ctx, &amqp.Message{
		Value: nonce,
	})
	if err != nil {
		t.Error(err)
	}

	message, err := listener.AmqpReceiver.Receive(ctx)
	if err != nil {
		t.Error(err)
	}

	message.Accept()
	if message.Value != nonce {
		t.Errorf("value not as expected in message Expected: %s Got: %s", nonce, message.Value)
	}

	depth, err := listener.GetQueueDepth()
	if err != nil || depth == nil {
		t.Error("Failed to get queue depth")
		t.Error(err)
	}

	derefDepth := *depth

	if derefDepth != 0 {
		t.Errorf("Expected queue depth of 0 Got:%v", derefDepth)
		t.Fail()
	}
}

// todo: Fix this integration test
// Currently calling reject causes the message to be deadlettered in SB and never redelivered.
// dispite the fact it's delivery count is under the maxDeliveryCount value.
func TestIntegrationRequeueRejectedMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Paniced: %v", prettyPrintStruct(r))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	config := &types.Configuration{
		ClientID:            os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret:        os.Getenv("AZURE_CLIENT_SECRET"),
		ResourceGroup:       os.Getenv("AZURE_RESOURCE_GROUP"),
		SubscriptionID:      os.Getenv("AZURE_SUBSCRIPTION_ID"),
		TenantID:            os.Getenv("AZURE_TENANT_ID"),
		ServiceBusNamespace: os.Getenv("AZURE_SERVICEBUS_NAMESPACE"),
		Hostname:            "Test",
		ModuleName:          helpers.RandomName(8),
		SubscribesToEvent:   "ExampleEvent",
		LogLevel:            "Debug",
		JobConfig: &types.JobConfig{
			JobRetryCount: 1337,
		},
	}
	listener := NewListener(ctx, config)
	// Remove topic to ensure each test has a clean topic to work with
	defer deleteSubscription(listener, config)

	nonce := time.Now().String()
	sender := createAmqpSender(listener)
	err := sender.Send(ctx, &amqp.Message{
		Value: nonce,
	})
	if err != nil {
		t.Error(err)
	}

	message, err := listener.AmqpReceiver.Receive(ctx)
	if err != nil {
		t.Error(err)
	}

	if message.Header.DeliveryCount != 0 {
		t.Error("first delivery has wrong count")
	}

	message.Release()

	checkUntil := time.Now().Add(time.Minute * 8)
	checkCtx, cancel := context.WithDeadline(context.Background(), checkUntil)
	defer cancel()

	messageSecondDelivery, err := listener.AmqpReceiver.Receive(checkCtx)
	if err != nil {
		t.Error(err)
	}

	if messageSecondDelivery.Value != message.Value {
		t.Error("redelivered message value different from original")
	}

	if messageSecondDelivery.Header.DeliveryCount != 1 {
		t.Error("redelivered message count doens't have correct delivery count set")
	}

	messageSecondDelivery.Accept()
}

// createAmqpSender exists for e2e testing.
func createAmqpSender(listener *Listener) *amqp.Sender {
	if listener.AmqpSession == nil {
		log.WithField("currentListener", listener).Panic("Cannot create amqp listener without a session already configured")
	}

	sender, err := listener.AmqpSession.NewSender(
		amqp.LinkTargetAddress("/" + listener.TopicName),
	)
	if err != nil {
		log.Fatal("Creating receiver:", err)
	}

	return sender
}

func deleteSubscription(listener *Listener, config *types.Configuration) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*45)
	defer cancel()
	_, err := listener.subsClient.Delete(ctx, config.ResourceGroup, config.ServiceBusNamespace, listener.TopicName, listener.SubscriptionName)
	if err != nil {
		panic(err)
	}
}
