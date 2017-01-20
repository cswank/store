{{define "base.js"}}

function updateCartLink(n) {
    if (n > 0) {
        $('#Cart').trigger('mouseenter');
        $('#Cart').text("Cart (" + n + ")");
    } else {
        $('#Cart').text("Cart");
    }
}

function initCart() {
    var cart = JSON.parse(localStorage.getItem("shopping-cart"));
    if (cart == undefined) {
        cart = {};
    }
    doInitCart(cart);
}

function doInitCart(cart) {
    var n = 0;
    for (var p in cart) {
        n += cart[p].count;
    }
    updateCartLink(n);
}

function addToCart(title) {
    var cart = JSON.parse(localStorage.getItem("shopping-cart"));
    if (cart == undefined) {
        cart = {};
    }

    var item = cart[title];
    if (item == undefined) {
        item = {
            id: id,
            count: quantity,
            cat: category,
            subcat: subcategory
        };
    }
    item.count = quantity;
    doAddToCart(cart, item, title);
}

function doAddToCart(cart, item, title) {
    cart[title] = item;
    localStorage.setItem("shopping-cart", JSON.stringify(cart));
    doInitCart(cart);
}

initCart();

{{end}}
