# Description of the Archive.org scraping library

**arc** searches torrents on Archive.org

See [here the Go documentation](https://godoc.org/github.com/juliensalinas/torrengo/arc) of this library.

The **Lookup** function searches Archive.org and returns a clean list of torrents. For each torrent the following info is retrieved:

* name
* description page

The **FindAndDlFile** function opens a torrent description page, retrieves the torrent file url, and downloads the torrent file.
