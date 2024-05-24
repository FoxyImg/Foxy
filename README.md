# Foxy
An image processing service.  Similar to [ImgProxy](https://github.com/imgproxy/imgproxy) but open source.

This is very much a work in progress.  Don't use in production yet.

## Features
- Cropping
- Resizing
- Other standard image manipulation (blur, brightness, contrast, saturation, hue, etc.)
- Face detection (via AWS Rekognition)
- People detection (via AWS Rekognition)
- Watermarking
- Export to PNG, WebP, AVIF, or JPEG
- Supports AWS S3, web and local filesystem as sources

## Running Foxy

### Database/Cache
Foxy requires a Postgres database and a redis instance.  The `docker-compose.yml` file in the `/docker` directory will get you started:

```bash
cd docker
docker-compose up -d
```

### Env Variables
Copy `.env.sample` to `env`. 

Below is a list of the env variables you can set:

- `DB_URL` - The URL of the Postgres database
- `REDIS_URL` - Redis host and port in the form `host/port`
- `USE_CACHE` - Whether to use the cache or not
- `CACHE_DIR` - The directory to store cached source images and metadata in
- `USE_RENDER_CACHE` - Whether to use the render cache or not
- `RENDER_CACHE_DIR` - The directory to store cached renders in
- `REQUIRE_SIGNATURE` - Whether to require a signature in the URL or not
- `DEFAULT_USER_NAME` - The default user name
- `DEFAULT_USER_EMAIL` - The default user email
- `DEFAULT_USER_NAME` - The default user name
- `DEFAULT_USER_PASSWORD` - The default user password (not used)
- `DEFAULT_USER_EMAIL` - The default user email
- `DEFAULT_USER_SECRET` - The default user secret
- `DEFAULT_USER_SOURCE_KEY` - The default user access key
- `DEFAULT_USER_SOURCE_CONFIG_FILE` - The path to the config json file for the default source

## Preview App
The [Foxy Preview](https://github.com/jawngee/foxy-preview) repository contains a Vue.js app that can be used to preview images and mess around with the API.

## Building URLs
The structure of the URL is as follows:

`http://localhost:8080/<source key>/<source image url or relative path base64 encoded>/<parameters>?s=<signature>`

The `source key` is the same as the `DEFAULT_USER_SOURCE_KEY` env variable.
The `source image url or relative path base64 encoded` is the image url or relative path to the image that you want to process.  If you are using an S3 source, then this would be the key, eg `path/to/file.jpg`.
The `parameters` is a list of parameters that you want to pass to the API separated by `/`.

So an example URL would be:

`http://localhost:8080/my-source-key/path/to/file.jpg/w:1000/h:1000/face:0/person:1/fmt:webp?s=<signature>`

The `signature` is a base64 encoded HMAC SHA256 of the URL using your `DEFAULT_USER_SECRET` as the key.

## Parameters
- **Crop/Resizing**
	- [Crop Modes](#crop-modes)
	- [Width/Height](#width2Fheight)
	- [Aspect Ratio](#aspect-ratio)
	- [Zoom](#zoom)
	- [Gravity](#gravity)
	- [Face Index](#face-index)
	- [Face Padding](#face-padding)
    - [Face Zoom](#face-zoom)

### Crop/Resizing

#### Crop Modes
```html
/crop:face,person,fit,smart,crop
```
The list crop modes to try in the order they should be tried.

- `face` - Crop to a face or the bounding box of all the detected faces
- `person` - Crop to a person or the bounding box of all the detected people
- `fit` - Resize the image proportionally to fit the specified width and height.  Any extra space will be filled with the `bg` background color (see below).
- `smart` - Crops the image using lipvip's smart algorithms.  These are really hit and miss.
- `crop` - Straight up crop the image.

If crop mode is omitted, the image is resized proportionally to the specified width and/or height.

The order of the crop modes is important.  If you specify:

```html
/crop:face,person,crop
```

Foxy will try to crop to faces first.  If no faces are detected, it will try to crop to people.  If no people are detected, it will do a regular crop.

#### Width/Height
```html
/w:<width>
```
The width of the target image in pixels.

```html
/h:<height>
```
The height of the image in pixels.

The crop modes `face`, `person`, `fill`, `smart` and `fit` require both width and height to be set OR either width or height to be set and an aspect ratio (see below).

#### Aspect Ratio
```html
/ar:<width>:<height>
```
The aspect ratio of the image, eg 16:9.

You must specify width OR height for aspect ratio to work. If you specify both width and height, only width will be used.

#### Zoom
```html
/zoom:<zoom factor>
```
The zoom factor of the crop in the range of `1` to `12`.

#### Gravity
```html
/gravity:<horizontal_gravity>:<vertical_gravity>
```
The anchor point of the crop.  Valid horizontal gravity values are `left`, `center`, `right`.  Valid vertical gravity values are `top`, `center`, `bottom`.

The default is `center:center`.

#### Face Index
```html
/face:index:<face_index>
```
The index of the face to crop.  If the specified index is out of range, the bounding box of all faces will be used.

If you pass `largest` as the value, the bounding box of the largest face will be used.

If you pass `smallest` as the value, the bounding box of the smallest face will be used.

#### Face Gravity
```html
/face:gravity:<face_horizontal_anchor?>:<face_vertical_anchor?>
```
This controls how the face is anchored in the crop.  By default, it is `center:top`.

#### Face Padding
```html
/face:pad:<face_horizontal_anchor?>:<face_vertical_anchor?>
```
This controls the distance of any edge of the face's bounding box from the edge of the cropped image.  This value is in pixels relative to a 1920x1920 image. For example, if you specify a padding of 24px, on a crop of 960x960 the padding would actually be 12px as 12 is 50% of 24 and 960 is 50% of 1920.

The default value is 48px.

#### Face Zoom
```html
/face:zoom:<zoom>
```
Controls how much the bounding box of the face fills the crop.  This value is the percentage of the crop to fill.  For example, specifying `/face:zoom:100` scale the bounding box so that it filled 100% of the crop (proportionally of course).
