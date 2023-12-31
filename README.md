# minimalist photos downloader from unsplash

## run
```
go run main.go params (all params are optional, except access key)
```

## how it work

 - get list of photos by query
 - download in img folder
 - skip downloaded by filename (filename == photo.ID, but "-" replaced to "_"). You may get original by this url: https://unsplash.com/photos/photo.ID
 - itterate page by page with count 30 and randon sleep between itterations (api limitations)
 - if rerun - will itterate page by page from latest

## params

 -c access key, get it here: https://unsplash.com/oauth/applications/ (its free)

 -captions, default flase, add .caption file with image description

 -q, query string, default photos

 query examples:
   - users/babakasotona/likes - get photos liked by user babakasotona
   - topics/people/photo - get all photos from "people" topic
   - "search/photos?query=office" - search all photos by query office

 -iq, image query, default: &w=256&h=256&fit=crop&crop=faces 

 image query examples:
   - &w=1024&h=1024&fit=crop&crop=faces
   - more examples https://unsplash.com/documentation#example-image-use

## Full example 

```
go run main.go --captions -c= -q="search/photos?order_by=editorial&plus=none&license=free&query=closeup%2Baesthetic%2Beditorial%2Bportrait%2Bfeatured%2Banimals%2Bwallpapers" -iq="&max-w=1280&max-h=1280&fm=png&q=95&fit=fil"
```
https://unsplash.com/s/photos/editorial?license=free