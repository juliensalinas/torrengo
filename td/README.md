# Description of the TorrentDownloads scraping library

**td** searches torrents on torrentdownloads.me

See [here the Go documentation](https://godoc.org/github.com/juliensalinas/torrengo/td) of this library.

The following dependencies are required if you want to be able to download torrent files from <http://torrentdownloads.me> in order to bypass the Cloudlare protection:

* Python (2 or 3)
* Python's cfscrape library (`pip install cfscrape`)
* NodeJS

The **Lookup** function searches torrentdownloads.me and returns a clean list of torrents. For each torrent the following info is retrieved:

* name
* description page
* size
* number of seeders
* number of leechers

The **ExtractTorAndMag** function opens a torrent description page and retrieves the torrent file url + the torrent magnet link.

The **DlFile** function downloads a torrent.
