{{define "category-admin.js"}}

function confirm() {
    document.location.href = "/admin/confirm?name={{.Name}}&resource=/categories/{{.Name}}";
}

{{end}}
