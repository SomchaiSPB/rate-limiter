package rate_limiter

const (
	maxMsgPerSen               = 5
	maxReqPerMin               = 10000
	maxFailedTransactionPerDay = 3
)

type Config struct {
	MaxMessagesPerSec           int
	MaxRequestsPerMin           int
	MaxFailedTransactionsPerDay int
}

func (c Config) WithMaxMessages(maxMsg int) Config {
	c.MaxMessagesPerSec = maxMsg
	return c
}

func (c Config) WithMaxRequests(maxReq int) Config {
	c.MaxRequestsPerMin = maxReq
	return c
}

func (c Config) WithMaxFailedTransactions(maxTrans int) Config {
	c.MaxFailedTransactionsPerDay = maxTrans
	return c
}

func NewConfig() Config {
	return Config{
		MaxMessagesPerSec:           maxMsgPerSen,
		MaxRequestsPerMin:           maxReqPerMin,
		MaxFailedTransactionsPerDay: maxFailedTransactionPerDay,
	}
}
