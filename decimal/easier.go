package decimal

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
