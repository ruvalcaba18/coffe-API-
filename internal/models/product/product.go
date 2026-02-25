package product

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

type Filter struct {
	Query    string
	Category string
	MinPrice float64
	MaxPrice float64
}
