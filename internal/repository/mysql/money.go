package mysql

import "go-property-rental/internal/domain/valueobject"

func entityMoney(v int64) valueobject.Money { return valueobject.Money(v) }
