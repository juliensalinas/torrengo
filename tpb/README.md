# Description of the ThePirateBay scraping library

**tpb** searches torrents on all The Pirate Bay proxies located on https://proxybay.bz

See [here the Go documentation](https://godoc.org/github.com/juliensalinas/torrengo/tpb) of this library.

The **Lookup** function retrieves all The Pirate Bay urls located on https://proxybay.bz, launches a search on all thoses urls concurrently, and returns a clean list of torrents from the url that responded first. The returned url is also checked in-depth because some proxies sometimes return a page with no error but the page actually does not have any result. For each torrent the following info is retrieved:

* name
* magnet link
* size
* upload date
* number of seeders
* number of leechers
