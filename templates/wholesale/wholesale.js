{{define "wholesale.js"}}

var cart;

function updateQuantities(id, title, category, subcategory, quantity) {
    var item = cart[title];
    if (item == undefined) {
        item = {
            id: id,
            count: 0,
            cat: category,
            subcat: subcategory
        };
    }

    item.count += quantity;

    if (item.count < 1) {
        item.count = 1;
    }
    
    cart[title] = item;
    $("#" + id).val(item.count);
}

function addItemsToCart() {
    localStorage.setItem("shopping-cart", JSON.stringify(cart));
    doInitCart(cart, true);
}

$(document).ready(function() {
    cart = JSON.parse(localStorage.getItem("shopping-cart"));
    
    if (cart == null) {
        cart = {};
    }

    for (var id in cart) {
        var item = cart[id];
        $("#" + item.id).val(item.count);
    }
});

{{end}}
