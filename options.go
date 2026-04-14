package kvwrap

type RemoteConfig struct {
	Endpoint          string
	ConnectionTimeout int // in seconds
	RequestTimeout    int // in seconds
}
