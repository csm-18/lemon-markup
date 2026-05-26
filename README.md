# Lemon Markup Compiler

Lemon Markup is a strict, compile-time-only static HTML extension language.

<img src="lemon-logo.png" style="display:block;margin:0 auto;" alt="Lemon Logo" width="500" height="500">

## Overview

This is a Go implementation of the Lemon Markup compiler that compiles `.lm` files to static HTML. It provides reusable components, strict prop validation, and compile-time error checking.

## Features

- **Components**: Reusable custom elements with uppercase names (e.g., `<Card>`)
- **Props**: Pass data to components with strict validation (missing/unused detection)
- **Children Injection**: Use `{{ children }}` to inject nested content
- **Raw Blocks**: Use `<raw></raw>` to escape compilation
- **Strict Validation**: Circular reference detection, prop checking
- **Void Elements**: Proper handling of HTML void elements (meta, link, img, etc.)
- **Single-pass Compilation**: Fast, predictable compilation

## Installation

```bash
go build -o lm .
```

## Usage

### Initialize a New Project
```bash
lm init
```
Creates a `markup/` folder with a hello page example.

### Compile Markup Files
```bash
lm compile
```
Compiles all `.lm` files in the current directory. Output goes to `dist/`.

### Show Version
```bash
lm version
```

### Show Help
```bash
lm help
```

## Example

### Simple Component

**markup/hello.lm:**
```html
<!DOCTYPE html>
<html>
<body>
    <Card Title="Welcome"></Card>
</body>
</html>

<template name="Card">
    <div class="card">
        <h1>{{ Title }}</h1>
    </div>
</template>
```

**Compiles to:**
```html
<!DOCTYPE html>
<html>
<body>
    
    <div class="card">
        <h1>Welcome</h1>
    </div>

</body>
</html>
```

### Children Injection

```html
<template name="Layout">
    <main>
        {{ children }}
    </main>
</template>

<Layout>
    <p>This content is injected!</p>
</Layout>
```

### Raw Blocks

```html
<raw>
    {{ this.variable.wont.be.processed }}
    <CustomTag></CustomTag>
</raw>
```

## Architecture

The compiler follows the recommended architecture from the spec:

- **Lexer** (`lexer/`): Tokenizes input into tags, variables, text, and raw blocks
- **Parser** (`parser/`): Builds an AST and validates balanced tags
- **Registry** (`registry/`): Stores templates globally and checks for collisions
- **Expander** (`expander/`): Validates props, detects circular references, and expands templates
- **Generator** (`generator/`): Serializes the expanded AST back to HTML

## Validation

The compiler enforces strict validation:

- **Missing Props**: Error if a component invocation doesn't provide required props
- **Unused Props**: Error if extra props are passed that aren't used in the template
- **Circular References**: Error if components reference each other in cycles
- **Self-Closing Tags**: Error if self-closing syntax (`/>`) is used (not allowed in Lemon Markup)
- **Nested Templates**: Error if template declarations appear inside other templates

## File Structure

- Entry point files without DOCTYPE are treated as templates/partials (no HTML output)
- Files with DOCTYPE compile into individual HTML files in the dist/ folder
- Directory structure is preserved (e.g., `markup/pages/home.lm` → `dist/markup/pages/home.html`)

## See Also

- See `specs/lemon-markup.md` for the complete language specification
