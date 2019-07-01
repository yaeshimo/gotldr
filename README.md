# gotldr

tldr client.

Upstream tldr-pages: <https://github.com/tldr-pages/tldr/>.

## Usage

Display help.

```sh
gotldr -help
```

Update or download tldr pages into user cache directory with `git`.

```sh
gotldr -update
```

Display usage of tar.

```sh
gotldr tar
```

Edit or create your tldr pages into user cache directory.

```sh
# make USERCACHEDIR/gotldr/user/gotldr.md if not exist
gotldr -edit gotldr

# use vim
EDITOR=vim gotldr -edit gotldr

# use emacs
EDITOR=emacs gotldr -edit gotldr
```

## Requirements

- User cache directory e.g. `$XDG_CACHE_HOME/gotldr/` on Linux.
  If not exist then to create by first run.
- `git` if use `-update`.

## Installation

```sh
go get github.com/yaeshimo/gotldr
```

## License

MIT
