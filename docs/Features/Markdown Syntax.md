Markdown syntax with obsidian extension

# Text formatting

- **Bold**: `**Bold**`
- _Italic_: `*Italic*`
- `Code`: \`Code\`
- ~Strikethrough~: `~Strikethrough~`
- ==highlight==: `==highlight==`

# Heading

```markdown
# Heading 1

## Heading 2

### Heading 3

#### Heading 4

##### Heading 5

###### Heading 6
```

# Link

```markdown
- [External Link](https://www.google.com)
- [[Internal Link]]
- [[Internal Link|Custom Name]]
- [[Internal Link#Heading|Custom Name]]
```

- [Obsidian Link Syntax](https://help.obsidian.md/link-notes)
- [[Configuration|Geode Config File]]

# Image with Dynamic Size

```markdown
- ![cloud image](https://go.dev/images/gophers/motorcycle.svg)
- ![[Go.svg]]

- ![cloud image|200](https://go.dev/images/gophers/motorcycle.svg)
- ![[Go.svg|200]]
```

> [!info] Cloud Image with Dynamic Size
> ![cloud image|200](https://go.dev/images/gophers/motorcycle.svg)

> [!bug] Local Image with Dynamic Size
> ![[Go.svg|150]]

# Syntax Highlight

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
```

# Callouts

> Default title

> [!question]+ Can callouts be _nested_?
>
> > [!todo]- Yes!, they can. And collapsed!
> >
> > > [!example] You can even use multiple layers of nesting.

> [!note]
> Aliases: "note"

> [!abstract]
> Aliases: "abstract", "summary", "tldr"

> [!info]
> Aliases: "info"

> [!todo]
> Aliases: "todo"

> [!tip]
> Aliases: "tip", "hint", "important"

> [!success]
> Aliases: "success", "check", "done"

> [!question]
> Aliases: "question", "help", "faq"

> [!warning]
> Aliases: "warning", "attention", "caution"

> [!failure]
> Aliases: "failure", "missing", "fail"

> [!danger]
> Aliases: "danger", "error"

> [!bug]
> Aliases: "bug"

> [!example]
> Aliases: "example"

> [!quote]
> Aliases: "quote", "cite"
