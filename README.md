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

Bare is glue between your build process and the tool you use for deploying static websites. Given a website and with
some or no guidance from you, it outputs a directory with your HTML
pages and their assets, preemptively built, and takes care of the little details, like patching URLs, 404s, or executing
Javascript.

### Getting started

Create a default `bare.toml` configuration file:

```bash
bare init
```

And then export your site:

```bash
bare export http://localhost:8000 --output dist/
```

This will produce, alongside an export of your website, a detailed report of the pages Bare found and the assumptions it
made, use this report to guide further configuration.
