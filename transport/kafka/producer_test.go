package kafka_test

// func TestPublisher(t *testing.T) {
// 	var (
// 		testdata = "testdata"
// 		encode   = func(context.Context, *kafka.Message, interface{}) error { return nil }
// 		decode   = func(_ context.Context, msg kafka.Message) (interface{}, error) {
// 			return TestResponse{string(msg.Value), ""}, nil
// 		}
// 	)

// 	mockBroker := sarama.NewMockBroker(t, 1)
// 	defer mockBroker.Close()
// 	mockBroker.Returns(&sarama.MetadataResponse{})

// 	brokers := []string{mockBroker.Addr()}
// 	topic := fmt.Sprintf("kafka-go-%016x", rand.Int63())
// 	groupId := fmt.Sprintf("kafka-go-group-%016x", rand.Int63())

// 	writer := kafka.NewWriter(kafka.WriterConfig{
// 		Brokers:          brokers,
// 		Topic:            topic,
// 		Balancer:         &kafka.LeastBytes{},
// 		Dialer:           kafka.DefaultDialer,
// 		CompressionCodec: snappy.NewCompressionCodec(),
// 	})

// 	r := kafka.NewReader(kafka.ReaderConfig{
// 		Brokers:  brokers,
// 		Topic:    topic,
// 		GroupID:  groupId,
// 		MinBytes: 1,
// 		MaxBytes: 10e6,
// 		MaxWait:  100 * time.Millisecond,
// 	})
// 	defer r.Close()

// 	producer := kafkatransport.NewProducer(writer, topic, encode, decode)
// 	_, err := producer.Endpoint()(context.Background(), struct{ Value []byte }{Value: []byte(testdata)})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	m, err := r.FetchMessage(context.Background())
// 	if err != nil {
// 		t.Errorf("error fetching message: %s", err)
// 	}

// 	if string(m.Value) != testdata {
// 		t.Errorf("want %q, have %q", testdata, string(m.Value))
// 	}
// }
