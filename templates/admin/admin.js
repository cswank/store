{{define "admin.js"}}

function setBackground(title) {
    var img = document.getElementById("background-" + title);
    console.log(img);
    img.style.borderColor = "#C1E0FF";
    img.style.borderWdith = "1px";
    img.style.borderStyle = "solid";

    document.getElementById('background-input').value = title;
}

function confirmCategory() {
    document.location.href = "/admin/confirm?resource={{.Resource}}&name={{.ResourceName}}";
    document.getElementById('category-delete-form').onsubmit = function() {
        return false;
    };
    return false;
}

{{end}}
