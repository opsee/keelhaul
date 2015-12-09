package config

type Config struct {
	PublicHost        string
	PostgresConn      string
	EtcdAddr          string
	BastionConfigKey  string
	BastionCFTemplate string
	VapeEndpoint      string
	VapeKey           string
	FieriEndpoint     string
	SlackEndpoint     string
	NSQDAddr          string
	NSQTopic          string
	NSQLookupds       string
	BartnetEndpoint   string
	BeavisEndpoint    string
}
