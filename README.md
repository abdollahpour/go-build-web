![go-build-web logo](/static/img/logo.png)

go-build-web
======
**go-build-web** is a simple powerful (enough) tools to build website using [Markdown], JSON and HTTP2. You don't need any special tools or database, you don't need to learn any technology. You can use the existing HTML knowledge you have and build your website. The application server supports required header for HTTP2 and you can be sure that your website runs fast on modern browsers.

Install/Run
======
There's no setup, just download the latest version from `releases`, extract the zip file and run `./go-build-web-<OS>`. Your website is ready [http://localhost:3000](http://localhost:3000)

To generate the static HTML files, run `./go-build-web-<OS> clean build`. You HTML files are ready in `build` folder.

Get started
===========
To create a new pages, you need to create at least two files in **templates** folder:
* \<path\>.json: This file is a meta data about your page.
* \<path\>.html: Template file using [golang template]

For example for **http://yourdomain/about** you need **templates/about.json** and **templates/index.html**

\<path\>.json contains some information about the page, for example title, keyword, description & etc.

```
{
    "templates": ["layouts/default.html"],
	"preloads": ["/css/main.css", "/js/main.js", "/img/logo.png"],
	"language": "en",
	"description": "Simple micro CMS using Go, Markdown and HTTP2",
	"keywords": ["CMS", "Golang", "HTTP2", "Markdown"],
	"author": "Hamed Abdollahpour",
	"image": "/img/logo.png",
	"title": "go-build-web: Build easier and faster"
}
```
Templates is a list of the templates that load before the \<path\>.html tempalte file. You can use these tamplates as layout of your page template.

The template file itself it based on [golang template] with two extra methods:

```
{{Markdown "some **Markown** text"}}
{{MarkdownFile "some_file.md"}}
```

You can use these two methods to make part of your page using [Markdown] which is more convineit for large contents.

More
====
You can change the running port using **PORT** environment variable. For example in bash:

    PORT=8000 ./go-build-web-<OS>

You can use [nginx] as reverse proxy to active **HTTP2** and **SSL** support. For example:

```
server {
    # Ensure that HTTP/2 is enabled for the server        
    listen 443 ssl http2;

    ssl_certificate ssl/certificate.pem;
    ssl_certificate_key ssl/key.pem;

    root /var/www/html;

    # Intercept Link header and initiate requested Pushes
    location = /myapp {
        proxy_pass http://upstream;
        http2_push_preload on;
    }
}
```

[Markdown]: https://github.com/adam-p/markdown-here/wiki/Markdown-Cheatsheet
[golang template]: https://golang.org/pkg/text/template/
[nginx]: https://www.nginx.com/blog/nginx-1-13-9-http2-server-push/