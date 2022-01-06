package outputs

type Event struct {
	Name        string
	Description string
	RekorURL    string
}

type Output interface {
	Send(Event) error
}
