{{define "Title"}}
    Markdown Example
{{end}}
{{define "Content"}}
<main>

# Example

This is an example of a file generated from Markdown (html/example/index.md), 
with syntax highlighting built in to the page server-side. This is all cached 
and so forth, built for speed!


Here's the function  responsible for creating templates and performing calls
to the necessary md conversions. Yeah, it calls Node.js, as the only suitable 
syntax highlighting library is a Javascript one. But thanks to this, there's 
zero client-side Javascript! I think that makes the jank worth it.
```go
func (ts *TemplateServer) create(fp string) error {
	ts.mtx.Lock()
	defer ts.mtx.Unlock()
	var err error
	if ts.cache[fp] == nil {
		var tmpl *template.Template

		if filepath.Ext(fp) == ".md" {
			/* Markdown file, convert it */
			log.Println("Converting Markdown to HTML " + fp)

			cmd := exec.Command("node", "./markdown/convert.mjs", fp)
			err := cmd.Run()
			if err != nil {
				return err
			}

			tmpl = template.New(fp)

			f, err := os.ReadFile("buffer")
			_, err = tmpl.New(filepath.Base(fp)).Parse(string(f))
			if err != nil {
				return err
			}

			_, err = tmpl.ParseFiles("./html/template.html")
			if err != nil {
				return err
			}

		} else {
			/* Normal HTML file */
			tmpl, err = template.ParseFiles("./html/template.html", fp)
			if err != nil {
				return err
			}
		}

		ts.cache[fp] = tmpl
		log.Println("Cached a template for " + fp)
	}
	return nil
}
```


<div class="micros">
	<img src="/micro/invalidator.gif" alt="">
	<img src="/micro/vim3.gif" alt="">
</div>

</main>
{{end}}
{{template "template.html" .}}
