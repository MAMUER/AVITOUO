package core

type Settings struct {
	Contacts               []string `json:"contacts"`
	Phones                 []string `json:"phones"`
	Addresses              []string `json:"addresses"`
	Companies              []string `json:"companies"`
	Emails                 []string `json:"emails"`
	DisableAddressAutoFill bool     `json:"disable_address_auto_fill"`
}

type AdRow struct {
	RowNum         int
	ID             string
	Title          string
	Description    string
	ContactPerson  string
	Phone          string
	Address        string
	Company        string
	Email          string
	Category       string
	ProductType    string
	SubProductType string
	Price          string
	PhotoNames     string
}
