# Description of the Ygg Torrent scraping library

**ygg** searches torrents on www.yggtorrent.is

See [here the Go documentation](https://godoc.org/github.com/juliensalinas/torrengo/ygg) of this library.

Torrents can be searched freely on Ygg Torrent, but an account is needed to download the torrent file. This library authenticates the user before downloading the torrent file.

The **Lookup** function searches www.yggtorrent.is and returns a clean list of torrents. For each torrent the following info is retrieved:

* name
* description page
* size
* upload date
* number of seeders
* number of leechers

The **FindAndDlFile** function takes the Ygg Torrent user id and password, authenticates the user, opens a torrent description page, retrieves the torrent file url, and downloads the torrent file.
