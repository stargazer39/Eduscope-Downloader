# Eduscope Downloader


## How to download a video

1. Go to releases and download the latest
2. Unpack the 7z archive
3. Run eduscope-downloader
4. Follow the on-screen instructions

## If the download fails, you can download again and it will resume where it left off

# CLI Usage 

```
Usage of Eduscope Downloader:
  -ff string
        Custom FFMPEG arguments (default "-c copy")
  -high-quality
        Downloads video at a higher quality    
  -p string
        Eduscope Password
  -t int
        Set Thread Count for the HTTP client (default 8)
  -u string
        Eduscope User Name
  -url string
        Eduscope URL
```

# Example
```
.\eduscope-downloader-win.exe --high-quality -p 123 -u it20603618@my.sliit.lk -url https://lecturecapture.sliit.lk/neplayer.php?id=anBzV3hiYXdXV181NzU5OA== -t 8
```

# If you are getting this error, put ffmpeg in your PATH. 
```
cannot run executable found relative to current directory
```

[How to add ffmpeg to PATH](https://phoenixnap.com/kb/ffmpeg-windows) 
