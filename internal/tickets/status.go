package tickets

type Status string

const (
	StatusNew            Status = "NEW"
	StatusAssigned       Status = "ASSIGNED"
	StatusInProgress     Status = "IN_PROGRESS"
	StatusWaitingForUser Status = "WAITING_FOR_USER"
	StatusResolved       Status = "RESOLVED"
	StatusClosed         Status = "CLOSED"
	StatusCancelled      Status = "CANCELLED"
)

var transitions = map[Status]map[Status]struct{}{
	StatusNew: {
		StatusAssigned: {}, StatusCancelled: {},
	},
	StatusAssigned: {
		StatusInProgress: {}, StatusCancelled: {},
	},
	StatusInProgress: {
		StatusWaitingForUser: {}, StatusResolved: {}, StatusCancelled: {},
	},
	StatusWaitingForUser: {
		StatusInProgress: {},
	},
	StatusResolved: {
		StatusInProgress: {}, StatusClosed: {},
	},
}

func (s Status) Valid() bool {
	_, ok := map[Status]struct{}{
		StatusNew: {}, StatusAssigned: {}, StatusInProgress: {}, StatusWaitingForUser: {},
		StatusResolved: {}, StatusClosed: {}, StatusCancelled: {},
	}[s]
	return ok
}

func (s Status) Terminal() bool {
	return s == StatusClosed || s == StatusCancelled
}
