package other

import (
	"errors"
	"fmt"
	"time"
)

func Retry(op func() error, maxRetries int, baseDelay int) error {

	for i := 1; i <= maxRetries; i++ {

		duration := baseDelay * 2 * i

		err := op()
		if err != nil {
			time.After(time.Millisecond * time.Duration(duration))
			continue
		}

		return nil

	}

	return errors.New(fmt.Sprint("Retryaing error  tries:", maxRetries))

}
