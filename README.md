# stowaway

`stowaway` keeps your dotfiles in one repository and links the ones you want
into your home directory. Profiles decide which files belong to which machine,
so a work laptop and a personal box can share a repository without sharing a
setup.

It only ever creates symlinks. Your real files stay in the repository, where
they belong to whatever version control you already use.

## Install

```
go install stowaway@latest
```

Or from a checkout:

```
make install
```

## Repository layout

A repository is any directory containing a `profiles` folder. Profiles are
`.conf` files; each line maps a file in the repository onto a path in `$HOME`:

```
# ~/dotfiles/profiles/default.conf
.vimrc -> .vimrc
vim/colors -> .vim/colors
git/config -> .gitconfig
```

The left side is relative to the repository root, the right side relative to
your home directory. Both use forward slashes on every platform. `#` starts a
comment.

## Usage

```
stowaway install -profile work
stowaway diff -profile work
stowaway uninstall -profile work
stowaway list
```

`install` links every entry. A target that is already the right link is left
alone, so run it as often as you like. If something else is sitting at a target
path, `install` refuses and tells you which entry it was; pass `-force` to
replace it.

`diff` compares the profile against the live home directory and prints one line
per entry that drifted:

```
foreign    /home/me/.vimrc    regular file
elsewhere  /home/me/.gitconfig  points at /home/me/old-gitconfig
absent     /home/me/.vim/colors
```

It exits 0 when everything matches and 1 when anything drifted, which makes it
usable as a shell prompt hook or a cron check. A target that is a perfect copy of
the repository file is still drift — the point of the tool is the link.

`uninstall` removes only links that point back into the repository, then prunes
any directories it left empty. Files it does not manage are never touched.

## Finding the repository

`stowaway` walks up from the working directory until it finds a `profiles`
folder. Run it from inside your dotfiles repository, or set `STOWAWAY_REPO` to a
path to skip the search:

```
export STOWAWAY_REPO=~/dotfiles
```

## Dry runs

`install` and `uninstall` take `-dry-run`, which prints what would happen
without touching anything:

```
$ stowaway install -profile work -dry-run
link /home/me/dotfiles/.vimrc -> /home/me/.vimrc
replace /home/me/dotfiles/git/config -> /home/me/.gitconfig
```

## Development

```
make build
make test
make vet
```

The only dependency is the standard library. Go 1.17 or newer.

## License

MIT, see `LICENSE`.
