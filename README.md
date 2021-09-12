# BeatList
### Go library for beatsaber playlists
<img alt="image" height="256" src="./cmd/gui/Icon.png" width="256"/>

BeatList is a Beatsaber playlist creation and editing tool that can be deployed unto any device.

you can find more info on playlists on the [PlaylistManager](https://github.com/rithik-b/PlaylistManager/wiki) wiki page

## Library
[![Go Reference](https://pkg.go.dev/badge/github.com/zivoy/BeatList/pkg/playlist.svg)](https://pkg.go.dev/github.com/zivoy/BeatList/pkg/playlist)

You can load playlists using the `Load` function and save by calling `Save` or `SavePretty` on the Playlist.

The `Cover` type has functions for dealing with base64 encoded images.

more info in the [godoc](https://pkg.go.dev/github.com/zivoy/BeatList/pkg/playlist)

## GUI
##### if you find a bug, please report it using the `Report Bug` button in the Help menu.
Get the latest GUI app from [Releases](/releases/latest)


#### features:
- Keyboard shortcuts
  - <kbd>Ctrl</kbd> + <kbd>S</kbd> for saving current playlist
  - <kbd>Ctrl</kbd> + <kbd>O</kbd> for opening a new playlist
  - <kbd>Ctrl</kbd> + <kbd>N</kbd> for starting a new playlist
- Add songs using the `+` icon and remove them using the `x` icon under the songlist
- Rearrange songs using the `↑` and `↓` arrows

[comment]: <> (##Examples)

![Example with list](./assets/top200rankedExample.png)