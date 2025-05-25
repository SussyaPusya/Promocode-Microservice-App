package redis

import "fmt"

const (
	promoKey = "promoId_%s"
)

func GetPromoKey(promoId string) string {
	return fmt.Sprintf(promoKey, promoId)
}
