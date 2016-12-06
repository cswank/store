{{define "shop.js"}}

var category = {{.Product.Cat}};
var subcategory = {{.Product.Subcat}};
var id = {{.Product.ID}};
var quantity = {{.Product.Quantity}};


function updateQuantity(n) {
    quantity += n;
    $("#quantity").val(quantity);
}

{{end}}
