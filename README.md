## Retry messages using SQS

### Produce messages to queue

You may produce messages to the selected queue
```
go run main.go produce --queue-name Votes --candidate-id andrew-yang --candidate-position mayor
```

### Consume messages from the queue

You may consume messages from the selected queue
```
go run main.go consume --queue-name Votes
```

### Retry messages

You may pull queue messages to be retried (from a retry queue)
```
go run main.go retry --queue-name Votes --retry-queue-name Votes-Retry
```
