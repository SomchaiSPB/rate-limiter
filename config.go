package rate_limiter

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
		MaxMessagesPerSec:           5,
		MaxRequestsPerMin:           10000,
		MaxFailedTransactionsPerDay: 3,
	}
}
