## Quick Start

* n/N - next/prev pane
* j/k - cursor down/up
* c/d - copy code block
* i - edit snippet
* s - toggle snippet pane
* use "---" to separate sections
* each section needs a title

```bash {copyable}
echo "Charm.sh Rocks 🚀"
```

```bash {title="Custom Title"}
echo "https://minions.wiki"
```

---

## GitHub Repository

Get repo from GitHub by SSH:

```bash {copyable}
mdf get repo kugarocks/rockman
```

HTTPS URL is also supported:

```bash {copyable}
mdf get repo https://github.com/kugarocks/rockman.git
```

Switch repo:

```bash {copyable}
mdf set repo
```

---

## More Commands

Switch folder:

```bash {copyable}
mdf set folder
```

Fuzzy find snippet:

```bash {copyable}
mdf examp
```

List folders:

```bash {copyable}
mdf list folder
```

---

## Configuration

Checkout:

```bash {copyable}
https://github.com/kugarocks/markdown-finder
```

---

## Raycast Script Command

```bash {copyable}
LANG=en_US.UTF-8 \
MDF_HOME=/Users/kuga/mdf \
/Applications/Alacritty.app/Contents/MacOS/alacritty \
    --config-file /Users/kuga/alacritty.toml \
    -e /usr/local/bin/mdf "$1" \
    > /dev/null 2>&1
```
