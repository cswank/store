package templates

type ThumbPage struct {
	Page
	Products []Product
}

type Product struct {
	Title       string
	Cat         string
	Subcat      string
	Price       string
	Total       string
	Quantity    int
	Description string
	ID          string
	Image       string
	Thumb       string
	Link        string
}

func NewProduct(title, cat, subcat string) Product {
	return Product{
		Title:    title,
		Cat:      cat,
		Subcat:   subcat,
		Price:    cfg.DefaultPrice,
		Quantity: 1,
	}
}
