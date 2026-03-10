package decimal

func (d Decimal) Sub1(e Decimal) Decimal {
	val, _ := d.Sub(e)
	return val
}
}

func (d Decimal) Mul1(e Decimal) Decimal {
	val, _ := d.Mul(e)
	return val
}

func (d Decimal) Add1(e Decimal) Decimal {

	val, _ := d.Add(e)
	return val
}

func (d Decimal) Div(e Decimal) Decimal {

	val, _ := d.Quo(e)
	return val
}

func (d Decimal) GreaterThan(e Decimal) bool {
	return d.Cmp(e) == 1
}

func (d Decimal) LessThanOrEqual(e Decimal) bool {
	return d.Cmp(e) <= 0
}

func (d Decimal) GreaterThanOrEqual(e Decimal) bool {
	return d.Cmp(e) >= 0
}
