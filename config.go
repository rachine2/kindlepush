package kindlepush

// The configuration options for kindlepush.
type Config struct {
	Debug bool
	// MaxNumber specifies the maximum number of list entries
	// in each one of subscribe.
	MaxNumber int
	// Subscribes is the channel list to subscribe.
	Subscribes []string
	// KindleAddr specifies the Kindle email account to receives.
	KindleAddr string
	// The email client config.
	Email EmailConfig
}

type EmailConfig struct {
	From     string
	Username string
	Password string
	SMTP     string
}
