# feedfoo - do things based on new RSS feed items

Here's the current use case:

You have an RSS feed and want to send a status update about it to the fediverse.

There's an easy way and there's this tool.

## Requirements

  * some kind of go compiler, 1.17.2 works
  * [madonctl](https://github.com/McKael/madonctl)
  * an RSS feed as understood by [gofeed](https://github.com/mmcdole/gofeed)

## How to install

First, get `madonctl` to run. See [the README](https://github.com/McKael/madonctl#usage)

```
go get -u github.com/McKael/madonctl
$GOPATH/bin/madonctl config dump -i mastodon.social -L username@domain -P password

$GOPATH/bin/madonctl toot "Yay, test"
```

Then install `feedfoo`:

```
go get -u github.com/winks/feedfoo
$GOPATH/bin/feedfoo
```

## How to build

Get `feedfoo` with deps:

```
git clone https://github.com/winks/feedfoo
cd feedfoo
go get -u github.com/mmcdole/gofeed
go build
./feedfoo -help
```

## How to use

Try it out safely:

```
./feedfoo --feed https://f5n.org/blog/atom.xml --cache ./dump.json -- echo "%%TEXT%%"
```

Now try it out for real:

```
./feedfoo --feed https://f5n.org/blog/atom.xml --cache ./dump.json -- madonctl toot "%%TEXT%%"
```

This shouldn't do anything if you already ran it once to fill the cache file.
Maybe manually delete the lastest post from the json or update the feed.
I guess "cache" is the wrong term. More like "ignored" or "done".

## TODO

  * templated string instead of hardcoded "New blog post: TITLE LINK"

## License

ISC
