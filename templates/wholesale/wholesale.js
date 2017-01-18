{{define "wholesale.js"}}

var items = {{.Items}};

function addItemsToCart() {
    $(".quantity").each(function() {
        var id = $(this).attr('id');
        var quantity = $(this).val();
        var item = items[id];
        console.log("id", id, quantity, item);
        item.count = parseInt(quantity);
        if (item.count > 0) {
            cart[item.title] = item;
        }
    })
    
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
