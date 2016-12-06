{{define "admin-product.js"}}


function confirm() {

    document.location.href = "/admin/confirm?resource={{.URI}}&name={{.Product.Title}}";
    document.getElementById('form').onsubmit = function() {
        return false;
    };
    return false;
}

$("#background-images div div img").click(function(e) {
    console.log(e);
})

{{end}}
