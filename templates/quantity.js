{{define "quantity.js"}}

function updateQuantity(id, n) {
    var quantity = parseInt($("#" + id).val());
    quantity += n;
    if (quantity < 1) {
        quantity = 1;
    }
    $("#" + id).val(quantity);
}

{{end}}
