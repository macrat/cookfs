package cookfs

type SimpleDiscoverPlugin struct {
	Self_  *Node
	Nodes_ []*Node
}

func (sd SimpleDiscoverPlugin) Bind(cook *CookFS) {
}

func (sd SimpleDiscoverPlugin) Run(stop chan struct{}) error {
	return nil
}

func (sd SimpleDiscoverPlugin) Self() *Node {
	return sd.Self_
}

func (sd SimpleDiscoverPlugin) Nodes() []*Node {
	return sd.Nodes_
}
