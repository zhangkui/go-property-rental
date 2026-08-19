package handler

import "go-property-rental/internal/domain/valueobject"

func moneyValue(v int64) valueobject.Money { return valueobject.Money(v) }
