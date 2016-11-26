{{define "confirm.js"}}

function doDelete() {
    $.ajax({
        url: {{.Resource}},
        type: 'DELETE',
        success: function(result) {
            console.log("success");
        }
    });
}

function back() {
    
}

{{end}}

