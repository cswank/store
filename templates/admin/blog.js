{{define "blog.js"}}


function confirm() {
    document.location.href = "/admin/confirm?resource={{.URI}}&name={{.Blog.Title}}";
    document.getElementById('blog-form').onsubmit = function() {
        return false;
    };
    return false;
}

{{end}}
