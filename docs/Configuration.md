```yaml
site:
  name: Geode
  suffix: " - Geode"
  base_url: https://example.com

build:
  output: public
  mode: draft

theme: default

ignorePatterns:
  - .git
  - .obsidian
  - .trash

socials:
  - title: Discord
    link: https://discord.gg/xxx
  - title: Github
    link: https://github.com/xxx
```

- `site`
  - `name`: site name
  - `suffix`: site suffix
  - `base_url`: site base url
- `build`
  - `output`: output directory
  - `mode`: `draft` or `explicit`. If `draft`, Geode will build all files except files with `draft: true` frontmatter. If `explicit`, Geode will only build files with `publish: true` frontmatter.
- `theme`: theme name (folder name in `themes` directory)
- `ignorePatterns`: patterns to ignore build
- `socials`: list your social links
