package product

type Update struct {
	Name          *string
	Description   *string
	PriceMinor    *int
	StockQuantity *int
}

// IsEmpty reports whether the update carries nothing to write.
func (u Update) IsEmpty() bool {
	return u.Name == nil && u.Description == nil && u.PriceMinor == nil && u.StockQuantity == nil
}
