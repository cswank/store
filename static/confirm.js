{{define "confirm.js"}}

function doDelete() {
    $.ajax({
        url: {{.Resource}},
        type: 'DELETE',
        success: function(result) {
            document.location.href = "/admin";
        },
        failure: function(result) {
            console.log("fail", result);
        }
    });
}

function back() {
    document.location.href = "/admin";
}

{{end}}

