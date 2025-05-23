package retrier

type RetrierOpts struct {
	Count    uint
	Interval uint
}

func MuliConnects[T any](count uint8, connFunc func() (T, error), retrierOpts *RetrierOpts) ([]T, error) {
	conns := make([]T, count)

	var err error

	for i := range count {
		if retrierOpts != nil {
			conns[i], err = Connect(uint8(retrierOpts.Count), retrierOpts.Interval, func() (T, error) {
				return connFunc()
			})
			if err != nil {
				return nil, err
			}
		} else {
			conns[i], err = connFunc()
			if err != nil {
				return nil, err
			}
		}
	}

	return conns, nil
}
