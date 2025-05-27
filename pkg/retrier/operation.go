package retrier

import "time"

type Try func() error

func Do(number uint8, duration time.Duration, try Try) error {
	var err error

	for i := uint8(0); i < number; i++ {
		err = try()

		if err == nil {
			break
		}

		if i < number-1 {
			time.Sleep(duration)
		}
	}

	return err
}
