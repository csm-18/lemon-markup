# Lemon Markup Specification v0.2.0

## Overview

Lemon Markup is a strict, compile-time-only static HTML extension language.

### Goals

* Reusable custom elements.
* HTML-first syntax.
* Compile-time only with zero client-runtime overhead.
* Outputs plain, predictable HTML.
* Fast, single-pass global parsing and compilation.

### Not Included

Lemon Markup intentionally omits reactivity, state management, client-side JavaScript execution, routing, scoped CSS, and hydration. It is strictly a static compiler for building structural, reusable HTML components.

## File Extension & Output Rules

* The file extension for Lemon Markup is `.lm`.
* Any `.lm` file containing a DOCTYPE (`<!DOCTYPE html>`) is treated as a target entry point and compiles into an individual HTML file (e.g., `index.lm` $\rightarrow$ `index.html`).
* Files without a DOCTYPE are treated as template/partial files and do not emit standalone HTML output files.

---

## Core Syntax

### Component & Native Tag Distinction

* Tags beginning with an uppercase letter are treated as **Custom Components**:
```html
<Card></Card>

```


* Tags beginning with a lowercase letter are treated as **Native HTML Elements**:
```html
<div></div>

```



### No Self-Closing Tags

Every element—whether a native HTML element or a custom component—**must have a matching explicit closing tag**. Self-closing shorthand expressions (e.g., `<Button />` or `<div />`) are invalid syntax and will throw a compiler error.

```html
<Button></Button>
<div class="box"></div>

<Button />
<div />

```

---

## Template Declarations

Templates are defined using the `<template>` tag with a unique `name` attribute. They can appear anywhere inside an `.lm` file, and a single `.lm` file can host multiple template declarations.

```html
<template name="Card">
    <div class="card-layout">
        <h2>{{ Title }}</h2>
    </div>
</template>

<template name="Button">
    <button class="btn-primary"></button>
</template>

```

### Template Constraints

1. **No Nesting:** Component templates cannot contain other `<template>` declarations inside their body.
2. **Naming Conventions:** Template names must begin with an uppercase letter, are case-sensitive, and cannot use the reserved names `template`, `children`, or `raw`.
3. **Strict Global Scope:** All template names must be unique across the entire project compilation context. If two files declare a template with the exact same name, the compiler throws an error.

---

## Props (Variables)

Lemon operates on a **Props-Only Model**. All attributes passed into custom components must start with an uppercase letter. There is no concept of raw HTML attribute forwarding or automatic root element injection.

```html
<Card Title="Welcome Screen" Theme="dark"></Card>

```

### Variable Interpolation

Properties passed into a component are injected via double curly braces containing the exact, case-sensitive prop name:

```html
<template name="Card">
    <div class="card {{ Theme }}">
        <h1>{{ Title }}</h1>
    </div>
</template>

```

*Note: Logical expressions, math operations, and JavaScript execution are strictly forbidden inside variable tags (e.g., `{{ 1 + 2 }}` will throw a compiler error).*

---

## Children Injection

The special variable `{{ children }}` is a reserved keyword representing any nested markup placed between the opening and closing tags of a custom element. The compiler preserves all whitespace, tabs, and newline characters exactly as written when copying children into the template.

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

---

## Escaping Compilation (`<raw>`)

To prevent the compiler from evaluating variables or processing custom elements, wrap the content in a `<raw>` block. Everything inside this block is treated as literal, raw text. The `<raw>` tags themselves are stripped out during final generation.

```html
<raw>
    <p>In Vue.js, your interpolation looks like this: {{ user.name }}</p>
    <Sidebar></Sidebar>
</raw>

```

---

## Strict Compiler Diagnostics

The compiler performs strict validation checks during the expansion phase. Any validation failure completely halts the compilation process and prints an error message containing the exact file, line, and column reference.

### 1. Missing Props (Fatal Error)

If a template expects a variable but that property is missing from the component invocation, it is a fatal error. To intentionally pass an empty value, developers must pass an explicit empty string `""`.

```html
<Card></Card> 

<Card Title=""></Card>

```

```text
Error: Missing required Prop "Title" for template "Card".
 --> src/index.lm:14:1

```

### 2. Unused Props (Fatal Error)

If a developer passes a property that is not declared or utilized inside the component's template body, it is treated as dead code or a typo, and compilation fails.

```html
<Card Title="Home" Link="/home"></Card>

```

```text
Error: Unused Prop "Link" passed to template "Card".
 --> src/index.lm:14:14

```

### 3. Circular References (Fatal Error)

Components cannot invoke themselves or form cyclic rendering chains (e.g., Template `A` calling Template `B`, which calls Template `A`). The compiler traces dependencies during parsing and aborts if an infinite loop is detected.

---

## Recommended Compiler Architecture

```text
lexer/      # tokenizes tags, variables, text, and toggles state for <raw> blocks
parser/     # builds AST; validates balanced tags and blocks
registry/   # globally indexes template definitions; checks for naming collisions
expander/   # handles strict validation (missing/unused props) and runs the AST swap
generator/  # serializes the pure structural AST back into raw, beautiful HTML strings

```