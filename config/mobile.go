package config

// MobileConfig - for external email services
type MobileConfig struct {
	Activate     string
	Provider     string
	APIToken     string
	AddrFrom     string
	TrackOpens   bool
	TrackLinks   string
	DeliveryType string

	// for templated email
	MobileVerificationTemplateID int64
	MobileVerificationCodeLength uint64
	MobileVerificationTag        string
	HTMLModel                    string
	MobileVerifyValidityPeriod   uint64 // in seconds
}
