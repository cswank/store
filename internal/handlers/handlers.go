package handlers

func getNavbarLinks() []link {
	return []link{
		{Name: "Home", Link: "/"},
		{Name: "Shop", Link: "/shop"},
		{Name: "Wholesale", Link: "/wholesale"},
		{Name: "Contact Us", Link: "/contact"},
	}
}
