# Music sources

## YouTube catalog

Afterhours Music searches music-category videos with the official YouTube Data API v3. YouTube supplies the titles, channel names, thumbnails, durations and public playlists. The music category includes music-related videos as well as songs. A channel name identifies the creator on YouTube.

Playback opens in the browser. The API provides catalog metadata and no direct audio stream, so these results never enter the app’s simulated playback queue. OAuth, account playlists and private playlists are outside this milestone.

The app sends the key as a request header only to the official API endpoint. Thumbnail requests go to approved YouTube image hosts without credentials. Search and playlist results stay in memory for the current view. Thumbnails are bounded and transient, and the app saves no API data to disk. Each explicitly requested results page uses a search or playlist-items request and a batched video lookup when there are videos. Superseded requests are canceled, and terminal errors such as quota exhaustion are never automatically retried.

[YouTube API reference](https://developers.google.com/youtube/data/reference/rest) · [YouTube developer policies](https://developers.google.com/youtube/terms/developer-policies)

## Offline catalog

Choose Mock to browse the fictional sample catalog and use simulated playback offline. The CAVA spectrum independently monitors the configured system audio output. It may show sound from the browser or from other applications.
