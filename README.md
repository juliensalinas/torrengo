**!!! This is a work in progress, this project is not working for the moment !!!**

# Description of Torrengo

## How To

### Purpose

Torrengo is a CLI (command line) program written in Go which searches torrent files from various websites. It's compatible under Windows, Linux and MacOS.

Current supported websites are the following:

1. <https://archive.org> (default)
1. <http://www.thepiratebay.se.net/>
1. <http://torrentdownloads.me/>
1. <http://1337x.to/>

**Caution!** Apart from <https://archive.org>, the websites above might host some illegal content and in some countries their use might be prohibited, consequently the search is disabled by default for these websites. Read [legal issues regarding The Pirate Bay](https://en.wikipedia.org/wiki/The_Pirate_Bay#Legal_issues) for example. If you decide to search these websites anyway you will find the procedure in the "Use" section below, but neither I, nor the tool shall be held responsible for any action taken against you for using Torrengo on the above-mentioned sites.

### Installation

...

### Use

...

-------------------------------

## Project's Structure

The project is made up of the following packages:

* **arc**: searches the Archive.org website, parses the 1st page of results, and returns a clean list of torrents (title, description, number of current leechers and seeders, torrent file address, magnet link url if any, etc.)
* **tpb**: same as above but for The Pirate Bay website 
* **td**: same as above but for the Torrent Downloads website
* **otts**: same as above but for the 1337x website
* **main**: gets the search input from the user and displays him a list of torrents matching his search. The search is made concurrently: each website is searched within an independent goroutine. Then gets input from the user about which torrent to download from the list and finally downloads the torrent file or prints the magnet link. Optionally the user can decide to open the file or magnet directly with Deluge.
