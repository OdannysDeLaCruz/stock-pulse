package stock

type StockExternalRepository interface {
	FetchData() ([]Stock, error)
}