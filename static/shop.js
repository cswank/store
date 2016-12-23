{{define "shop.js"}}

var category = {{.Product.Cat}};
var subcategory = {{.Product.Subcat}};
var id = {{.Product.ID}};
var quantity = {{.Product.Quantity}};


function updateQuantity(n) {
    quantity += n;
    if (quantity < 1) {
        quantity = 1;
    }
    $("#quantity").val(quantity);
}


var modal = document.getElementById('product-modal');

// Get the image and insert it inside the modal - use its "alt" text as a caption
var img = document.getElementById('img');
var modalImg = document.getElementById("img01");
img.onclick = function(){
    modal.style.display = "block";
    modalImg.src = this.src;
}

// Get the <span> element that closes the modal
var span = document.getElementById("product-modal");

// When the user clicks on <span> (x), close the modal
span.onclick = function() { 
    modal.style.display = "none";
}

{{end}}
