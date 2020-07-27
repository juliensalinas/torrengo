# Description of Torrengo

## How To

### Purpose

Torrengo is a CLI (command line) program written in Go which concurrently searches torrent files from various sources. I really liked the [torrench](https://github.com/kryptxy/torrench) program which is an equivalent written in Python so I figured it could be nice to write a similar program in Go in order to increase speed thanks to concurrency.

Nice supported features:

* the user decides which sources he wants to search (all sources are searched by default) and the search is done **concurrently**
* given that The Pirate Bay urls are changing quite often, this program concurrently launches a search on all The Pirate Bay urls found on <https://proxybay.bz> and retrieves torrents from the fastest response (the returned url is also checked in-depth because some proxies sometimes return a page with no error but the page actually does not have any result)
* torrent file search and download on Ygg Torrent, The Pirate Bay, and 1337, are protected by the Cloudflare bot detection. In order to comply, a Google Chrome browser is used under the hood (which means that you need to have Google Chrome installed in order for this program to work).
* <http://www.yggtorrent.**> can be searched freely, but an account is needed to download the torrent file, so the program authenticates the user before downloading the torrent file
* downloaded torrents can be launched in Deluge, QBittorrent, or Transmission
* a timeout can be set so long-running requests are ignored

Current supported sources are the following:

1. <https://archive.org> (called **arc** internally)
1. all The Pirate Bay urls located on <https://proxybay.bz> (called **tpb** internally)
1. <http://1337x.to> (called **otts** internally)
1. <http://www.yggtorrent.**> (previously t411, called **ygg** internally)

**Caution!** Apart from Archive.org, the websites above might host some illegal content and in some countries their use might be prohibited. Read [legal issues regarding The Pirate Bay](https://en.wikipedia.org/wiki/The_Pirate_Bay#Legal_issues) for example. Neither I, nor the tool shall be held responsible for any action taken against you for using Torrengo on the above-mentioned sites.

### Installation

**Prerequisite:** you need to have Google Chrome installed on your system. Torrengo needs a real Google Chrome browser in order to behave like any real browser and then properly deal with Javascript.

For security reasons I don't provide with compiled binaries. The program can be easily installed and compiled with the usual Go tools:

1. `go get github.com/juliensalinas/torrengo`
2. `go build github.com/juliensalinas/torrengo`

Each website's scraper is an independent library that can be installed and reused. For example if you only want to use the Archive.org scraping library, simply do:

* `go get github.com/juliensalinas/torrengo/arc`

### Usage

Searching "Dumas Montecristo" from all sources is as simple as:

`torrengo Dumas Montecristo`

![Torrgengo output](https://juliensalinas.com/en/images/torrengo-example_201809171014.png)

If you want to search from a specific source (let's say Archive.org):

`torrengo -s arc Dumas Montecristo`

Sources names:

* <https://archive.org>: arc
* all The Pirate Bay urls located on <https://proxybay.bz>: tpb
* <http://1337x.to>: otts
* <https://www.yggtorrent.**>: ygg

If you want to search from multiple sources (let's say Archive.org and ThePirateBay), use commas:

`torrengo -s arc,tpb Dumas Montecristo`

If some sources are too slow to respond, use a timeout. For example the following stops every HTTP requests that take more than 2 seconds and returns the other results found:

`./torrengo -t 2000 Dumas Montecristo`

Some sources give both a magnet link and a torrent file (you can choose which one you want), some only give a torrent file, and some only give a magnet link.

Optionally you can open the torrent file or magnet link directly in your torrent client (**Deluge**, **QBittorrent** or **Transmission** are supported for the moment).
