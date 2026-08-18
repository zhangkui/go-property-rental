package service

import "go-property-rental/internal/domain/valueobject"

func money(v int64) valueobject.Money { return valueobject.Money(v) }
