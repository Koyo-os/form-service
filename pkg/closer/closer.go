package closer

type (
	Closer interface{
		Close() error
	}

	CloserGroup struct {
		closers []Closer
	}
)

func NewCloserGroup(closers ...Closer) *CloserGroup {
	return &CloserGroup{
		closers: closers,
	}
}

func (c *CloserGroup) Close() error {
	var err error
	
	for _, closer := range c.closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return err	
}