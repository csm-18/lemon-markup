# Lemon Markup

Lemon Markup is a strict, compile-time-only static HTML extension language for building reusable markup components with zero client-side runtime.

<img src="lemon-logo.png" alt="Lemon Logo" width="400" height="400">

## Overview

Lemon Markup uses `.lm` files to compose HTML pages and reusable component templates. It compiles only at build time and emits plain HTML to `dist/`.

## Features

* HTML-first syntax with custom component templates.
* Compile-time expansion only — no runtime JS or hydration.
* Strict validation for duplicate templates, unused props, and circular component references.
* Supports raw literal blocks with `<raw>` to preserve markup and interpolation syntax.
* Page output only for files containing `<!DOCTYPE html>`.

## What Lemon Markup Omits

Lemon Markup intentionally does not provide:

* client-side reactivity or state management
* runtime component rendering
* scoped CSS or CSS-in-JS
* router or navigation primitives
* hydration or browser-side template logic

## Installation

Build the compiler with Go:

```bash
go build -o lemon
```

Then put the lemon binary in a place of your choice and add its path to the PATH env variable.



## CLI Usage

```bash
./lemon            # print version/about message
./lemon init       # initialize a new lemon project
./lemon build      # compile all .lm files and generate dist/*.html
./lemon version    # print version
./lemon help       # print commands list
```

## Project Structure

A Lemon project typically includes:

* `index.lm` or other page entrypoints
* `.lm` template files
* `dist/` generated output
* `assets/`, `styles/`, and `logic/` static folders

The compiler discovers `.lm` files recursively from the project root, but it ignores `dist/`, `assets/`, `styles/`, and `logic/` directories.

On each build, `assets/`, `styles/`, and `logic/` are synchronized into `dist/` so the generated `dist/` tree can be served directly without additional path-remapping.

The synchronization process checks each static folder, creates missing destination directories, copies new files, and updates changed files in `dist/`.

## File Output Rules

* Source files have extension `.lm`.
* Only `.lm` files containing `<!DOCTYPE html>` emit HTML output.
* Files without DOCTYPE act as template-only partials.
* Output files are written under `dist/`.
* If a source file path begins with `markup/`, the `markup/` prefix is stripped from the output path.
* The build cleans up stale `dist/*.html` files that no longer map to an active `.lm` source.

## Syntax

### Custom Components vs Native HTML

* Tags starting with an uppercase letter are custom components.
* Tags starting with a lowercase letter are native HTML elements.

```html
<Card Title="Hello"></Card>
<div class="container"></div>
```

### Template Declarations

Templates are declared with `<template name="Name">`.

```html
<template name="Card">
  <div class="card">
    <h1>{{ Title }}</h1>
  </div>
</template>
```

Template names must:

* begin with an uppercase letter
* be case-sensitive
* be globally unique across the project
* not use reserved names `template`, `children`, or `raw`

Templates cannot be nested inside other `<template>` declarations.

### Props

Component props are passed as attributes.

```html
<Card Title="Welcome" Theme="dark"></Card>
```

Props are interpolated inside templates with `{{ Name }}`:

```html
<h1>{{ Title }}</h1>
```

Raw expressions or JavaScript are not supported inside interpolation markers.

### Children

`{{ children }}` injects nested markup from a component invocation.

```html
<template name="Layout">
  <main>{{ children }}</main>
</template>

<Layout>
  <p>Hello World</p>
</Layout>
```

### Raw Blocks

Use `<raw>` to preserve literal content and prevent expansion.

```html
<raw>
  <p>Use {{ raw }} syntax literally.</p>
  <Sidebar></Sidebar>
</raw>
```

The `<raw>` wrapper is stripped from the final output.

### Tag Rules

* Native HTML void elements are supported without explicit closing tags: `meta`, `link`, `img`, `br`, `hr`, `input`, `col`, `area`, `base`, `embed`, `param`, `source`, `track`, and `wbr`.
* Non-void elements require an explicit closing tag.
* Self-closing shorthand like `<Button />` or `<div />` is invalid.

## Validation

The compiler performs strict validation during parse and expansion.

* Duplicate template names across files are fatal.
* Unknown custom component references are fatal.
* Circular component references are fatal.
* Unused props passed to a component are fatal.

## Example

```html
<!DOCTYPE html>
<html>
  <body>
    <Card Title="Hello Lemon"></Card>
  </body>
</html>

<template name="Card">
  <div class="card">
    <h1>{{ Title }}</h1>
  </div>
</template>
```

## Compiler Architecture

The compiler internals are separated into stages:

* `lexer.go` — tokenizes markup, raw blocks, variables, and DOCTYPE declarations.
* `parser.go` — builds the AST and validates syntax.
* `registry.go` — tracks global template definitions.
* `expander.go` — expands components, resolves props, and checks circular references.
* `generator.go` — emits the final HTML string.

## See Also

Look inside `specs/` for full language and compiler behavior details.
