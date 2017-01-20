{{define "confirm.js"}}

function doDelete() {
    $.ajax({
        url: "{{.Resource}}",
        type: "DELETE",
        success: function(result) {
            document.getElementById("success").style.visibility = "visible";
            document.getElementById("confirm").style.visibility = "hidden";
            return false;
        },
        failure: function(result) {
            console.log("fail", result);
        }
    });
}

function back(e) {
    window.location = "/";
    e.preventDefault();
    return false;
}

document.getElementById("success").style.visibility = "hidden";

{{end}}

