# Description of the 1337x scraping library

**otts** searches torrents on 1337x.to

The **Lookup** function searches 1337x.to and returns a clean list of torrents. For each torrent the following info is retrieved:

* name
* description page
* size
* upload date
* number of seeders
* number of leechers

The **ExtractMag** function opens a torrent description page and retrieves the torrent magnet link.
