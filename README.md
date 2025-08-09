# Bare

Bare is a tool for creating static websites using traditionally dynamic web
frameworks. Instead of restricting yourself to the idiosyncrasies and limitations of the
static-site generator (SSG) you use, you tap into the power of unrestricted rendering logic:

- **You need to fetch data to build your markup?** Dynamic applications need that too, so all web frameworks have great
  support for that. Does your SSG of choice?
- **You need templating at build-time?** You already know Blade, ERB, or whatever your framework uses.
- **Your website has evolved and needs to be dynamic?** It already is, stop using Bare and just deploy your favorite
  framework as you usually do.

So, how does Bare work?

Bare takes a ready-for-deployment snapshot of your locally-hosted dynamic hosted website. With some or no guidance from you, it extracts then downloads all the pages and their assets, that you preemptively built, and takes care of the little details, like patching URLs, 404s, or executing Javascript.

### Getting started

Install `bare` via `go`:
```
go install github.com/felixdorn/bare@latest
```
[How to install "go"?](https://go.dev/doc/install)

Or download a binary from the [releases page](https://github.com/felixdorn/bare/releases). For Linux (amd64):
```bash
curl -L https://github.com/felixdorn/bare/releases/latest/download/bare-linux-amd64 -o bare && chmod +x bare
```

Then export your site:
```bash
bare export http://localhost:8000 -o dist/
```

This will produce an export of your website and a detailed report of the pages Bare found, use it to guide further configuration.

Then, look at your exported website:
```bash
bare serve
```

### Configuring bare

Create a `bare.toml` configuration file:
```bash
bare init
```

The configuration looks like this:
```toml
# bare.toml
url = 'http://127.0.0.1:8000'
output = 'dist/'

[js]
enabled = false
wait_for = 2000 # milliseconds

[pages]
entrypoints = ['/']
extract-only = []
exclude = []
```

### Some pages or assets are missing.

* If you know their path in advance.

Add them to the `pages.crawl` configuration:
```toml
# bare.toml created by running `bare init`
[pages]
entrypoints = ['/', '/some-hidden-page']
```

* If you don't know their path in advance.

A common pattern is to create a new route server-side, for example, `/_/list-of-undiscoverable-pages`, which contains a list of links to all of the otherwise undiscoverable pages.

And then tell Bare to look for it:
```toml
# bare.toml created by running `bare init`
[pages]
entrypoints = []
extract-only = ['/_/list-of-undiscoverable-pages']
# ...
```

### How to handle JavaScript?

Bare can find pages and assets that would otherwise be missed by not executing your Javascript code.

* Option #1: the `--js-enabled` option
```bash
bare export http://localhost:8000 -o dist/ --js-enabled
```

* Option #2: Set `js.enabled` to true in `bare.toml`
```toml
# bare.toml created by running `bare init`
[pages]
# ...

[js]
enabled = true
wait_for = 2000 # milliseconds
executable_path = "/usr/bin/google-chrome-stable" # optional
flags = ["no-sandbox", "headless=new"] # optional
```

### How to exclude pages?

* Option #1: the `--exclude` option
```bash
bare export http://localhost:8000 -o dist/ --exclude "/internal/**" --exclude /api/*/internal --exclude /secret-page
```
> Tip: You can use the shorthand flag `-E` in place of `--exclude``

* Option #2: the `exclude` parameter in `bare.toml`
```toml
# bare.toml created by running `bare init`

[pages]
exclude = ['/internal/**', '/api/v1/internal', '/secret-page']
```

### Why the name?
Bare is named after my last philosophy professor, whose last name was Barrera. It also serves as a statement that vaguely gestures at the stupid incentives that lead to bloat. [Do](https://www.effectivealtruism.org/) [useful](https://www.givingwhatwecan.org/pledge) [things](https://veganoutreach.org/why-vegan/).
