"use strict";
import hljs from 'highlight.js';
import Markdown from 'markdown-it';
import fs from 'node:fs/promises'

/* Must be ran with a mutex!! */

const md = Markdown({
	html: true,
	highlight: (str, lang) => {
		const code = lang && hljs.getLanguage(lang)
		  ? hljs.highlight(str, {
			  language: lang,
			  ignoreIllegals: true,
			}).value
		  : md.utils.escapeHtml(str);
		return `<pre class="hljs"><code>${code}</code></pre>`;
	},
});

async function main(path){
	if (!path) {
		process.exit(1)
	}
	let file;
	try {
		file = await fs.readFile(path, "utf8");
	} catch (e) {
		process.exit(1)
	}
	
	file = file.split("<main>")

	let head = file[0]
	let body = file[1].split("</main>")[0]
	let tail = file[1].split("</main>")[1]

	body = md.render(body);

	let output = head+"<main>"+body+"</main>"+tail

	await fs.writeFile("buffer", output)

	process.exit(0)
}

main(process.argv[2])
