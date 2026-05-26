# Lemon Markup Specification v0.1

## Overview

Lemon Markup is a static HTML extension language.

### Goals

- Reusable custom elements
- HTML-first syntax
- Compile-time only
- Zero runtime
- Outputs plain HTML
- Simple parsing and compilation

### Not Included

Lemon Markup intentionally omits:

- Reactivity
- State management
- JavaScript execution
- Routing
- Scoped CSS
- Hydration / client runtime

It is a static compiler for reusable HTML structures.

## File Extension

`.lm`

Examples:

```
index.lm
about.lm
components.lm
```

## Output Rules

Any `.lm` file containing a DOCTYPE (`<!DOCTYPE html>`) is compiled into an HTML file.

Example:

```
index.lm → index.html
```

Files without a doctype are treated as partial/template files and are not emitted as HTML output.

## Basic Syntax

Normal HTML is valid Lemon Markup syntax.

Example:

```html
<div>
    <h1>Hello</h1>
</div>
```

### Custom Elements

Tags beginning with an uppercase letter are treated as custom elements (templates):

```html
<Card />
<Button />
<Layout />
```

Tags beginning with lowercase letters are treated as native HTML elements:

```html
<div>
<section>
<button>
```

## Template Declaration

Templates are declared using the `<template>` tag with a `name` attribute:

```html
<template name="Card">
    ...
</template>
```

Templates may appear anywhere inside a `.lm` file.

### Template Rules

#### Templates Cannot Be Nested

Valid:

```html
<template name="Card">
    <div></div>
</template>

<section>
    <Card />
</section>
```

Invalid (compiler error):

```html
<template name="Layout">
    <template name="Card">
    </template>
</template>
```

Compiler error: Nested templates are not allowed.

#### Template Names

Template names:

- Must begin with uppercase letters
- Are case-sensitive
- Must be unique

Valid examples: `Card`, `Button`, `MainLayout`

Invalid examples: `card`, `button`

#### Reserved Names

The names `template`, `children`, and `attrs` are reserved and cannot be used as template names.

## Template Usage

Templates are used like normal HTML elements:

```html
<Card />

<!-- or -->
<Card>
    <h1>Hello</h1>
</Card>
```

## Props

Attributes beginning with uppercase letters are treated as component props:

```html
<Button Text="Login" />
```

`Text` is a prop in the example above.

### Native HTML Attributes

Attributes beginning with lowercase letters are treated as native HTML attributes:

```html
<div class="box"></div>
```

### Mixed Attributes

```html
<Card class="main" Title="Welcome">
</Card>
```

Rules:

- `class` → native HTML attribute
- `Title` → component prop

## Variable Interpolation

Variables use double braces:

```text
{{ Title }}
```

Allowed sources for interpolation:

- Component props
- `children`
- `attrs`

Not allowed:

- Expressions
- JavaScript or function calls

Invalid examples:

```text
{{ x + y }}
{{ alert() }}
```

Compiler error: Expressions are not supported.

## Children Injection

Special variable `{{ children }}` represents nested child content.

Template example:

```html
<template name="Card">
    <div class="card">
        {{ children }}
    </div>
</template>
```

Usage:

```html
<Card>
    <p>Hello</p>
</Card>
```

Output:

```html
<div class="card">
    <p>Hello</p>
</div>
```

## Attribute Forwarding

Special variable `{{ attrs }}` contains all lowercase HTML attributes passed to the component. Uppercase prop attributes are excluded.

Template example:

```html
<template name="Card">
    <div {{ attrs }}>
        {{ children }}
    </div>
</template>
```

Usage:

```html
<Card class="box" id="main">
    Hello
</Card>
```

Output:

```html
<div class="box" id="main">
    Hello
</div>
```

## Self-Closing Components

Supported:

```html
<Button />
```

Equivalent to:

```html
<Button></Button>
```

## Template Resolution

All templates are globally available during compilation. The compiler scans all `.lm` files and registers templates before compilation.

Example layout:

```
src/
 ├── index.lm
 ├── ui/
 │    └── components.lm
```

## Compilation Process

Compilation steps:

1. Scan all `.lm` files
2. Register templates
3. Parse documents
4. Detect custom elements
5. Expand templates
6. Inject props
7. Inject children
8. Forward `attrs`
9. Generate HTML

## Component Expansion

When the compiler encounters:

```html
<Card Title="Hello">
    <p>World</p>
</Card>
```

The compiler:

- Finds template `Card`
- Clones template content
- Injects props
- Injects children
- Recursively expands nested components

### Recursive Expansion

Templates may use other templates. The compiler expands recursively until only native HTML remains.

Example:

```html
<template name="Page">
    <Layout>
        <Card>
            {{ children }}
        </Card>
    </Layout>
</template>
```

### Circular References

Circular template references are disallowed.

Invalid example:

```html
<template name="A">
    <B />
</template>

<template name="B">
    <A />
</template>
```

Compiler error: Circular template reference detected.

### Unknown Templates

Using an undefined template causes a compiler error. Example:

```html
<Card />
```

If `Card` is not defined: `Unknown template: Card`.

## HTML Preservation

Native HTML is preserved exactly unless modified through template expansion.

Example input:

```html
<template name="Card">
    <div class="card" {{ attrs }}>
        {{ children }}
    </div>
</template>

<!DOCTYPE html>
<html>
<body>

<Card class="main">
    <h1>Hello</h1>
</Card>

</body>
</html>
```

Output:

```html
<!DOCTYPE html>
<html>
<body>

<div class="card" class="main">
    <h1>Hello</h1>
</div>

</body>
</html>
```

## Compiler Errors

Common compiler errors include:

- Nested templates
- Unknown templates
- Duplicate template names
- Circular references
- Invalid interpolation
- Malformed HTML
- Invalid template names

Example error:

```text
Error: Duplicate template name "Card"
 --> components.lm:12:1
```

### Recommended AST Nodes

- `DocumentNode`
- `ElementNode`
- `TemplateNode`
- `TextNode`
- `VariableNode`
- `AttributeNode`

### Recommended Compiler Structure

```
lexer/      # Produces tokens: tags, attributes, text, interpolation blocks
parser/     # Builds AST
registry/   # Stores templates
expander/   # Resolves and expands custom elements
generator/  # Outputs final HTML
```

## Design Philosophy

Lemon Markup extends HTML minimally.

It is:

- not a replacement for HTML
- not a JavaScript framework
- not reactive

It is:

- a static HTML reuse language
- a compile-time templating system
- a lightweight component compiler

The final browser output is always plain HTML.