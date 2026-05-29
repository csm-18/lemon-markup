# Lemon Markup Specification v0.2.0

## Overview

Lemon Markup is a strict, compile-time-only static HTML extension language.

### Goals

* Reusable custom components.
* HTML-first syntax.
* Compile-time only with zero client-runtime overhead.
* Outputs plain, predictable HTML.
* Fast, single-pass global parsing and compilation.

### Not Included

Lemon Markup intentionally omits reactivity, state management, client-side JavaScript execution, routing, scoped CSS, and hydration. It is strictly a static compiler for building structural, reusable HTML components.

---

## File Extension & Build Rules

* The file extension for Lemon Markup source files is `.lm`.
* The compiler discovers `.lm` files recursively from the project root.
* The directories `dist/`, `assets/`, `styles/`, and `logic/` are ignored during source discovery.
* Any `.lm` file containing a `<!DOCTYPE html>` declaration is treated as a target entry point and compiles into an individual HTML file.
* Files without a DOCTYPE are treated as template/partial files and do not emit standalone HTML output files.
* Output HTML files are written to `dist/`.
* Source paths are preserved relative to the project root, but a leading `markup/` directory prefix is stripped from the output path.
* On each build, `dist/` is created if missing, stale HTML files with no corresponding source are removed, and static folders are synchronized into `dist/`.
* The folders `assets/`, `styles/`, and `logic/` are synchronized into `dist/` so missing static files are copied and changed files are updated.

---

## Compiler Architecture

The compiler is organized into five distinct stages:

1. `lexer/` - tokenizes tags, text, variables, DOCTYPE declarations, and raw text blocks.
2. `parser/` - builds the AST, validates balanced tags, template declarations, and checks syntax.
3. `registry/` - globally indexes template definitions and detects naming collisions across files.
4. `expander/` - performs component expansion, prop resolution, circular dependency detection, and unused-prop validation.
5. `generator/` - serializes the final structural AST into raw HTML output.

---

## Core Syntax

### Component & Native Tag Distinction

* Tags beginning with an uppercase letter are treated as **custom components**.

```html
<Card></Card>
```

* Tags beginning with a lowercase letter are treated as **native HTML elements**.

```html
<div></div>
```

### Native Void Elements

Standard HTML void elements may be written without an explicit closing tag.

```html
<meta charset="utf-8">
<img src="logo.png">
<br>
```

Supported void elements include: `meta`, `link`, `img`, `br`, `hr`, `input`, `col`, `area`, `base`, `embed`, `param`, `source`, `track`, and `wbr`.

### No Self-Closing Shorthand

Every non-void element must have a matching explicit closing tag. Self-closing shorthand such as `<Button />` or `<div />` is invalid and will produce a compiler error.

```html
<Button></Button>
<div class="box"></div>

<Button />
<div />
```

---

## Template Declarations

Templates are declared using the `<template>` tag with a required `name` attribute.

```html
<template name="Card">
    <div class="card-layout">
        <h2>{{ Title }}</h2>
    </div>
</template>
```

A single `.lm` file may contain multiple `<template>` declarations.

### Template Constraints

1. **No Nesting:** `<template>` declarations cannot appear inside another template's body.
2. **Naming:** Template names must begin with an uppercase letter and are case-sensitive.
3. **Reserved Names:** Template names cannot be `template`, `children`, or `raw`.
4. **Global Uniqueness:** Template names must be unique across the entire compilation context. Duplicate template names across files cause a fatal compile error.

---

## Props (Attributes)

Lemon Markup uses a **props-only model**. Properties passed into custom components are declared as attributes on the component tag.

```html
<Card Title="Welcome Screen" Theme="dark"></Card>
```

* Attribute names may include letters, digits, hyphens, underscores, and dots.
* Attribute values are always plain strings.
* There is no runtime attribute forwarding or automatic root-element injection.

### Attribute Interpolation

Attribute values may also be interpolated from surrounding props using exact `{{ Name }}` syntax.

```html
<Card Title="{{ ParentTitle }}"></Card>
```

This interpolation is resolved during expansion if the surrounding props contain the referenced name.

---

## Variable Interpolation

Template variables are injected via double curly braces containing the exact case-sensitive prop name.

```html
<template name="Card">
    <div class="card {{ Theme }}">
        <h1>{{ Title }}</h1>
    </div>
</template>
```

* Logical expressions, math operations, and JavaScript execution are forbidden inside variable tags.
* Variables are treated strictly as identifiers; only the exact `{{ Name }}` form is supported.

### Missing Variables

The current compiler does not treat missing prop bindings as a fatal error. If a template variable has no matching prop, the unresolved token is preserved in the output as literal `{{ Name }}`.

---

## Children Injection

The special variable `{{ children }}` is reserved for nested markup provided between a component's opening and closing tags.

```html
<template name="Layout">
    <main>{{ children }}</main>
</template>

<Layout>
    <div>
        <p>Hello World</p>
    </div>
</Layout>
```

Nested children are preserved as written and expanded recursively when inserted.

---

## Escaping Compilation (`<raw>`)

Wrap markup in a `<raw>` block to prevent the compiler from evaluating variables or custom components. Content inside `<raw>` is treated as literal text.

```html
<raw>
    <p>In Vue.js, your interpolation looks like this: {{ user.name }}</p>
    <Sidebar></Sidebar>
</raw>
```

The `<raw>` wrapper itself is removed during generation, and its contents are emitted exactly as written.

---

## Compiler Diagnostics

The compiler emits strict fatal diagnostics during parsing and expansion. Errors include exact file, line, and column references.

### Duplicate Templates

Registering two templates with the same name anywhere in the project halts compilation.

### Unknown Component References

Referencing a custom component whose template is not registered causes a fatal error.

### Circular References

Components cannot reference themselves or form cyclic rendering chains. The compiler detects recursive component graphs and aborts.

### Unused Props

Passing a prop into a custom component that is not referenced by the component's template is a fatal error.

```html
<Card Title="Home" Link="/home"></Card>
```

This produces an error because `Link` is unused by `Card`.

---

## Output Behavior

* `lemon build` expands all templates and serializes final HTML output.
* Only files with an explicit `<!DOCTYPE html>` declaration emit `.html` output.
* Template-only `.lm` files contribute definitions to the global registry but do not emit output.
* Static folders `assets/`, `styles/`, and `logic/` are synchronized into `dist/` so the build output is directly servable.
* Output HTML is plain, generated HTML without runtime wrappers.

---

## Recommended Project Usage

* Use `.lm` files with DOCTYPE only for pages and entrypoints.
* Use `.lm` files without DOCTYPE for reusable template libraries and partials.
* Keep component definitions globally unique and avoid deeply nested cycles.

---

## Example

```html
<!DOCTYPE html>
<html>
<body>
    <Card Title="Hello World!"></Card>
</body>
</html>

<template name="Card">
    <div class="card">
        <h1>{{ Title }}</h1>
    </div>
</template>
```

Notes:
* `<Card>` is a custom component because it starts with an uppercase letter.
* The `Card` template is defined in the same file.
* The page emits `index.html` only when the source file includes `<!DOCTYPE html>`.
