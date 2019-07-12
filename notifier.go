package notiz



var (
	defaultNotifier  Notifier
	criticalNotifier Notifier
	closeChan        chan bool = make(chan bool)
)


type ReportLog struct {
	ReportType string                 `json:"report_type,omitempty"`
	Priority   string                 `json:"priority,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

type NotifyMessage interface {
	MessageData() []byte
	Opts() map[string]string
	Hash() string
}

type Notifier interface {
	Format(log ReportLog) NotifyMessage
	Notify(NotifyMessage)
}
