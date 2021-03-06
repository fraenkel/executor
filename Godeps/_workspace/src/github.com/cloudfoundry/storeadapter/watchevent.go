package storeadapter

type WatchEvent struct {
	Type EventType
	Node StoreNode
}

type EventType int

const (
	CreateEvent = EventType(iota)
	DeleteEvent
	ExpireEvent
	UpdateEvent
)
