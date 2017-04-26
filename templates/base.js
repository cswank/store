{{define "base.js"}}

var quantities;

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

function updateQuantities(id, title, category, subcategory, quantity) {
    var item = quantities[title];
    if (item == undefined) {
        item = {
            id: id,
            count: 0,
            cat: category,
            subcat: subcategory
        };
    }

    console.log("update quant", quantity, item);
    item.count += quantity;

    if (item.count < 1) {
        item.count = 1;
    }
    
    quantities[title] = item;
    $("#" + id).val(item.count);
}

function addItemsToCart() {
    var cart = JSON.parse(localStorage.getItem("shopping-cart"));
    if (cart == null) {
        cart = {};
    }

    console.log("cart:", cart);
    for (var key in quantities) {  
        cart[key] = quantities[key];
    }
    
    localStorage.setItem("shopping-cart", JSON.stringify(cart));
    doInitCart(cart, true);
}

function doAddToCart(cart, item, title, animate) {
    cart[title] = item;
    localStorage.setItem("shopping-cart", JSON.stringify(cart));
    doInitCart(cart, animate);
}

$(document).ready(function() {
    quantities = {};
    initCart();
});

{{end}}
