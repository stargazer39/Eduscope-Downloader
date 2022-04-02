# Eduscope Downloader


## How to download a video

1. Clone this repo
2. Install [Go](https://go.dev/)
3. Install and add ffmpeg to the path. Refer [this guide](https://www.thewindowsclub.com/how-to-install-ffmpeg-on-windows-10)
4. Open a terminal inside the cloned folder
5. Run one of below commands (Refer to the usage):

## With Username & Password

```bash
go run . -url "https://lecturecapture.sliit.lk/neplayer.php?id=Vjk1eXF3TG5nSV8zNTY4NQ==" -u <username> -p <password>
```

## Without Username or Password

```bash
go run . -url "https://lecturecapture.sliit.lk/neplayer.php?id=Vjk1eXF3TG5nSV8zNTY4NQ=="
```

Very ALPHA state as of now. Need ffmpeg in the PATH.
Refer [this guide](https://www.thewindowsclub.com/how-to-install-ffmpeg-on-windows-10) to setup ffmpeg in the PATH

# Usage 

```
Usage of Eduscope Downloader:
  -high-quality
        Downloads video at a higher quality
  -p string
        Eduscope Password
  -u string
        Eduscope User Name
  -url string
        Eduscope URL
```
