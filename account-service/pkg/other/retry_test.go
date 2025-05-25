package other_test

import (
	"errors"
	"fmt"
	"testing"

	"gitlab.com/pisya-dev/account-service/pkg/other"
)

func TestRetries(t *testing.T) {
	i := 1
	op := func() error {

		fmt.Println(i)
		if i == 3 {
			fmt.Println("3retryai phuis")
			return nil
		}
		i++

		return errors.New("govno")
	}

	if err := other.Retry(op, 5, 200); err != nil {
		t.Error(err)
	}

}
