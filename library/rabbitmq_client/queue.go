package rabbitmq_client

type Queues []QueueConfig

type QueueConfig struct {
	Name          string                 `json:"name"`
	Durable       bool                   `json:"durable"`
	AutoDelete    bool                   `json:"auto_delete"`
	Exclusive     bool                   `json:"exclusive"`
	Channel       int                    `json:"channel"`
	PrefetchCount int                    `json:"prefetch_count"`
	NoLocal       bool                   `json:"no_local"`
	NoWait        bool                   `json:"no_wait"`
	AutoAck       bool                   `json:"auto_ack"`
	WorkerCount   int8                   `json:"worker_count"`
	Arguments     map[string]interface{} `json:"arguments"`
	Exchange      ExchangeConfig         `json:"exchange"`
	RoutingKey    string                 `json:"routing_key"`
	Retry         RetryConfig            `json:"retry"`
}

type ExchangeConfig struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Durable    bool                   `json:"durable"`
	AutoDelete bool                   `json:"auto_delete"`
	Internal   bool                   `json:"internal"`
	NoWait     bool                   `json:"no_wait"`
	Arguments  map[string]interface{} `json:"arguments"`
}

type RetryConfig struct {
	Count uint8 `json:"count"`
	Delay int   `json:"delay"`
}
