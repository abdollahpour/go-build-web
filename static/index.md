![go-build-web logo](/img/logo.png)

go-build-web
======
**go-build-web** is a simple powerful (enough) tools to build a website using [Markdown], JSON and [Golang template]. You don't need any special tools or database, you don't need to learn any technology. You can use the existing HTML knowledge you have and build your website. Unlike most of the static website generator, go-build-web does not pack the resources since HTTP2 has a better approach for loading and cashing resource.

Install/Run
======
There's no setup, just download the latest version from `releases`, extract the zip file and run `go-build-web-<OS>`. Your website is ready [http://localhost:3000](http://localhost:3000)

To generate the static HTML files, run `./go-build-web-<OS> clean build`. You HTML files are ready in `build` folder.

Get started
===========
To create a new page, you need to create at least one `index.html` file in **static** folder. This file can be a simple HTML file or uses the [Golang template] tags to make it more functional. Using the template syntax you can:

* Build layout using `layout.html` file.
* Render markdown using `MarkdownFile` function.
* Load data from an external JSON file, `index.json`. You can use this file to load some classified metadata using in your pages.

To create a new folder, you just need to make a folder with the same name and put `index.html` inside.

For example, if you want to make `http://foo.com/page1/` you need to make `static/page1/index.html`. You must remember in go-build-web all the pages finish by `/`.

For more information check the list of samples:

* Build CV: https://github.com/abdollahpour/go-build-web-sample-cv

Build the static files
==========
Simply run `go-build-web-<OS> clean build` and you website generetes whitin **build** directory.

Doing more
====
By default application run on port **3000** You can change the port using **PORT** environment variable. For example in bash:

    PORT=8000 ./go-build-web-<OS>

[Markdown]: https://github.com/adam-p/markdown-here/wiki/Markdown-Cheatsheet
[Golang template]: https://golang.org/pkg/text/template/
[nginx]: https://www.nginx.com/blog/nginx-1-13-9-http2-server-push/