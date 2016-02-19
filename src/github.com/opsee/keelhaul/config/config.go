package config

type Config struct {
	PublicHost            string
	PostgresConn          string
	EtcdAddr              string
	BastionConfigKey      string
	BastionCFTemplate     string
	VapeEmailEndpoint     string
	VapeUserInfoEndpoint  string
	VapeKey               string
	FieriEndpoint         string
	LaunchesSlackEndpoint string
	TrackerSlackEndpoint  string
	NSQDAddr              string
	NSQTopic              string
	NSQLookupds           string
	BartnetEndpoint       string
	BeavisEndpoint        string
	HugsEndpoint          string
	SpanxEndpoint         string
}
