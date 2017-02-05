{{define "base.js"}}

function updateCartLink(n, animate) {
    var $element = $('#Cart');
    var text = "";
    var updatedText;
    if (n > 0) {
        updatedText = "Cart (" + n + ")";
        if (animate) {
            $element.css('background-color', '#FD69E4');
            $element.css('color', 'white');
            setTimeout(function(){
                $element.css('background-color', 'white');
                $element.css('color', 'black');
            }, 1500);
        }
    } else {
        updatedText = "Cart";
    }
    
    $element.text(updatedText);
}

function initCart() {
    var cart = JSON.parse(localStorage.getItem("shopping-cart"));
    if (cart == undefined) {
        cart = {};
    }
    doInitCart(cart, false);
}

function doInitCart(cart, animate) {
    var n = 0;
    for (var p in cart) {
        n += cart[p].count;
    }
    updateCartLink(n, animate);
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
    doAddToCart(cart, item, title, true);
}

function doAddToCart(cart, item, title, animate) {
    cart[title] = item;
    localStorage.setItem("shopping-cart", JSON.stringify(cart));
    doInitCart(cart, animate);
}

$(document).ready(function() {
    initCart();
});

{{end}}
