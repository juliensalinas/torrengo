# Description of Torrengo

## How To

### Purpose

Torrengo is a CLI (command line) program written in Go which concurrently searches torrent files from various sources. I really liked the [torrench](https://github.com/kryptxy/torrench) program which is an equivalent written in Python so I figured it could be nice to write a similar program in Go in order to increase speed thanks to concurrency.

Current supported sources are the following:

1. <https://archive.org> (called **arc** internally)
1. <http://pirateproxy.mx> (called **tpb** internally)
1. <http://torrentdownloads.me> (called **td** internally)
1. <http://1337x.to> (called **otts** internally)

**Caution!** Apart from Archive.org, the websites above might host some illegal content and in some countries their use might be prohibited. Read [legal issues regarding The Pirate Bay](https://en.wikipedia.org/wiki/The_Pirate_Bay#Legal_issues) for example. Neither I, nor the tool shall be held responsible for any action taken against you for using Torrengo on the above-mentioned sites.

### Installation

The whole program can be installed and compiled with the usual Go tools:

1. Set your Go environment variables properly
1. `go get github.com/juliensalinas/torrengo`
1. `go build torrengo`

Each website's scraper is an independent library that can be installed and reused. For example if you only want to use the Archive.org scraping library, simply do:

* `go get github.com/juliensalinas/torrengo/arc`

### Usage

Searching "Dumas Montecristo" from all sources is as simple as:

`torrengo Dumas Montecristo`

If you want to search from a specific source (let's say Archive.org):

`torrengo -s arc Dumas Montecristo`

Sources names:

* <https://archive.org>: arc
* <https://pirateproxy.mx>: tpb
* <http://torrentdownloads.me>: td
* <http://1337x.to>: otts

If you want to search from a several specific sources (let's say Archive.org and ThePirateBay), use commas:

`torrengo -s arc,tpb Dumas Montecristo`

Some sources give both a magnet link and a torrent file (you can choose which one you want), some only give a torrent file, and some only give a magnet link.

Optionally you can open the torrent file or magnet link directly in your torrent client (only **Deluge** is supported for the moment).

