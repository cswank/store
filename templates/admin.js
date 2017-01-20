{{define "admin.js"}}

function setBackground(title) {
    var img = document.getElementById("background-" + title);
    console.log(img);
    img.style.borderColor = "#C1E0FF";
    img.style.borderWdith = "1px";
    img.style.borderStyle = "solid";

    document.getElementById('background-input').value = title;
}


{{end}}
