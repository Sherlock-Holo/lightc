package libnetwork

type CreateNetworkConflict struct {
	ExistName   string
	ExistSubnet string
}

func (e CreateNetworkConflict) Error() string {
	return "network " + e.ExistName + " exists, subnet is " + e.ExistSubnet
}
