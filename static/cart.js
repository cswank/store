{{define "cart.js"}}

var price = {{.Price}};

var items = JSON.parse(localStorage.getItem("shopping-cart"));

var shopClient = ShopifyBuy.buildClient({
    apiKey: '{{.Shopify.APIKey}}',
    domain: '{{.Shopify.Domain}}',
    appId: '6'
});

var products = {};

function getProducts() {
    for (var title in items) {
        var item = items[title];
        shopClient.fetchProduct(item.id).then(function(product) {
            products[product.attrs.product_id] = product;
        })
    }
}

function loadCart() {
    getProducts();
    if (items == undefined) {
        items = {};
    }

    var i = 0;
    for (var title in items) {
        var item = items[title];
        var url = "/cart/lineitem/" + item.cat + "/" + item.subcat + "/" + title;
        $.get(url, {quantity: item.count}, function(html) {
            $("#items").append($(html));
        });
        i++;
    }

    updateTotal(items);
    
    showCart(i);
}

loadCart();

function showCart(i) {
    if (i == 0) {
        document.getElementById("cart").style.visibility = "hidden";
        document.getElementById("empty-cart").style.visibility = "visible";
    } else {
        document.getElementById("cart").style.visibility = "visible";
        document.getElementById("empty-cart").style.visibility = "hidden";
    }
}

function update(title, n) {
    item = items[title];
    item.count += n;
    if (item.count < 0) {
        item.count = 0;
    }
    doAddToCart(items, item, title);
    updateTotal(items)
    var sel = "#" + item.id + "-quantity";
    $(sel).val(item.count);

    sel = "#" + item.id + "-total";
    var itemPrice = item.count * price;
    $(sel).text("$" + itemPrice.toFixed(2));
    return false;
}

function updateTotal(items) {
    var total = 0.0;
    for (var title in items) {
        var item = items[title];
        total += item.count * price;
    }
    $("#grand-total").text("$" + total.toFixed(2));
}

function updateOnBlur(title) {
    var sel = "#" + item.id + "-quantity";
    var val = $(sel).val();
    update(title, val);
}

function removeItem(title) {
    var item = items[title];
    delete items[title];
    localStorage.setItem("shopping-cart", JSON.stringify(items));
    $("#" + item.id).remove();
    updateTotal(items);
    doInitCart(items);
}

function checkout() {
    shopClient.createCart().then(function(cart) {
        var variants = [];
        for (var title in items) {
            var item = items[title];
            var product = products[item.id];
            variants.push({variant: product.selectedVariant, quantity: item.count});
        }
        localStorage.removeItem("shopping-cart");
        updateCartLink(0);
        cart.createLineItemsFromVariants(...variants).then(function(cart) {
            window.open("/", '_self');
            window.open(cart.checkoutUrl, '_blank');
        })
    })
}

function clearCart() {
    localStorage.removeItem("shopping-cart");
    updateCartLink(0);
    items = {};
    doInitCart(items);

    var div = document.getElementById("items");
    while (div.firstChild) {
        div.removeChild(div.firstChild);
    }

    showCart(0);
}

{{end}}

