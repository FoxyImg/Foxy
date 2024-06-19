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

![Screenshot](screenshot.webp)

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

- [Background Removal](#background-removal)
- [Crop/Resizing](#cropresizing)
  - [Source Crop](#source-crop)
  - [Crop Modes](#crop-modes)
  - [Width/Height](#width2Fheight)
  - [Aspect Ratio](#aspect-ratio)
  - [Zoom](#zoom)
  - [Gravity](#gravity)
  - [Face Index](#face-index)
  - [Face Gravity](#face-gravity)
  - [Face Padding](#face-padding)
  - [Face Zoom](#face-zoom)
  - [Person Index](#person-index)
  - [Person Gravity](#person-gravity)
  - [Person Padding](#person-padding)
  - [Person Zoom](#person-zoom)
- [Adjustments](#adjustments)
  - [Brightness](#brightness)
  - [Saturation](#saturation)
  - [Exposure](#exposure)
  - [Gamma](#gamma)
  - [Hue](#hue)
- [Stylize](#stylize)
  - [Blur](#blur)
  - [Pixelate](#pixelate)
  - [Stylize Order](#stylize-order)
- [Gradient Map](#gradient-map)
  - [Monochrome](#monochrome)
  - [Blend Mode](#blend-mode)
  - [Blur](#blur)
  - [Opacity](#opacity)
  - [Stops](#stops)
- [Image Properties](#image-properties)
  - [Background Color](#background-color)
- [Box/Padding](#boxpadding)
  - [Padding](#padding)
  - [Border](#border)
- [Masking](#masking)
  - [Mask](#mask)
- [Redactions](#redactions)
  - [Redact](#redact)
- [Overlays](#overlays)
  - [Overlay](#overlay)

### Background Removal

#### Replace Background With Color
```html
/bgr:c:<method>:<color>
```
Replace the background with a solid RGBA color.

#### Replace Background With Image
```html
/bgr:img:<method>:<base64_encoded_image_key>
```
Replace the background with an image from the source.

In both cases, `<method>` is one of `photoroom`, `fg` or `person`.  `photoroom` will use the Photoroom API to remove the background.  `fg` will use the foreground ML model to remove the background and `person` will use the person ML model to remove the background.

### Crop/Resizing

#### Source Crop
```html
/src:<left>,<top>,<width>,<height>
```
Crops the source image before any additional processing is done.  This occurs AFTER vision has done any face detection.

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
/w:<width>/h:<height>
```
The width or height of the target image in pixels.

The crop modes `face`, `person`, `fill`, `smart` and `fit` require both width and height to be set OR either width or height to be set and an aspect ratio (see below).

#### Aspect Ratio
```html
/ar:<width>:<height>
```
The aspect ratio of the image, eg 16:9.

You must specify width OR height for aspect ratio to work. If you specify both width and height, only width will be used.

#### Zoom
```html
/zoom:<zoom_factor>
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
/face:gravity:<face_horizontal_anchor>:<face_vertical_anchor>
```
This controls how the face is anchored in the crop.  By default, it is `center:top`.

#### Face Padding
```html
/face:pad:<face_horizontal_anchor>:<face_vertical_anchor>
```
This controls the distance of any edge of the face's bounding box from the edge of the cropped image.  This value is in pixels relative to a 512x512 image. For example, if you specify a padding of 24px, on a crop of 960x960 the padding would actually be 12px as 12 is 50% of 24 and 960 is 50% of 1920.

The default value is 48px.

#### Face Zoom
```html
/face:zoom:<zoom>
```
Controls how much the bounding box of the face fills the crop.  This value is the percentage of the crop to fill.  For example, specifying `/face:zoom:100` scale the bounding box so that it filled 100% of the crop (proportionally of course).

#### Person Index
```html
/person:index:<person_index>
```
The index of the person to crop.  If the specified index is out of range, the bounding box of all people will be used.

If you pass `largest` as the value, the bounding box of the largest person will be used.

If you pass `smallest` as the value, the bounding box of the smallest person will be used.

#### Person Gravity
```html
/person:gravity:<person_horizontal_anchor>:<person_vertical_anchor>
```
This controls how the person is anchored in the crop.  By default, it is `center:center`.

#### Person Padding
```html
/person:pad:<person_horizontal_anchor>:<person_vertical_anchor>
```
This controls the distance of any edge of the person's bounding box from the edge of the cropped image.  This value is in pixels relative to a 512x512 image. For example, if you specify a padding of 24px, on a crop of 256x256 the padding would actually be 12px as 12 is 50% of 24 and 256 is 50% of 512.

The default value is 0px.

#### Person Zoom
```html
/person:zoom:<zoom>
```
Controls how much the bounding box of the person fills the crop.  This value is the percentage of the crop to fill.  For example, specifying `/person:zoom:100` scale the bounding box so that it filled 100% of the crop (proportionally of course).

### Adjustments

#### Brightness
```html
/bri:<brightness>
```
The brightness of the image, 0 to 200.

#### Saturation
```html
/sat:<saturation>
```
The saturation of the image, 0 to 200.

#### Exposure
```html
/exp:<exposure>
```
The exposure of the image, -100 to 100.

#### Gamma
```html
/gamma:<gamma>
```
The gamma of the image, 0 to 10.  Default is 1

#### Hue
```html
/hue:<hue>
```
The hue of the image, -360 to 360.

### Stylize

#### Blur
```html
/blur:<blur>
```
Apply a gaussian blur to the image, 0 to 512.

#### Pixelate
```html
/px:<pixelate>
```
Apply a pixelate effect to the image, 0 to 512.

#### Stylize Order
```html
/stylize:<order>
```
The order of the stylize effects to apply separated by a comma.  Valid values are `blur`, `px`.

### Gradient Map

#### Monochrome
```html
/gm:mono:<true|false>
```
Before applying the gradient map, convert the image to monochrome.  Defaults to true.

#### Blend Mode
```html
/gm:blend:<blend_mode>
```
The blend mode to use when applying the gradient map.  Valid values are:

| Mode       | Value |
|------------|-------|
| Clear      | 0     |
| Source     | 1     |
| Over       | 2     |
| In         | 3     |
| Out        | 4     |
| Atop       | 5     |
| Dest       | 6     |
| DestOver   | 7     |
| DestIn     | 8     |
| DestOut    | 9     |
| DestAtop   | 10    |
| XOR        | 11    |
| Add        | 12    |
| Saturate   | 13    |
| Multiply   | 14    |
| Screen     | 15    |
| Overlay    | 16    |
| Darken     | 17    |
| Lighten    | 18    |
| ColorDodge | 19    |
| ColorBurn  | 20    |
| HardLight  | 21    |
| SoftLight  | 22    |
| Difference | 23    |
| Exclusion  | 24    |

#### Blur
```html
/gm:blur:<blur>
```
Apply a gaussian blur to the mapped image, 0 to 512.

#### Opacity
```html
/gm:opacity:<opacity>
```
The opacity of the mapped image when composited with the source image, 0 to 100.  Defaults to 100.

#### Stops
```html
/gm:stops:<stop1>,<color1>:<stop2>,<color2>:...
```
The stops and colors to use when applying the gradient map.  Minimum two stops are required.
  
### Image Properties

#### Background Color
```html
/bg:<background_color>
```
The background color of the image.  The color is specified as a hexadecimal color code without the leading '#'.  It can be a 6 (RGB) or 8 (RGBA) character hex string.

### Box/Padding

#### Padding
```html
/pad:<pad_color>:<padding>
/pad:<pad_color>:<horizontal_padding>:<vertical_padding>
/pad:<pad_color>:<left_padding>:<top_padding>:<right_padding>:<bottom_padding>
```

Adds interior padding to the image.

#### Border
```html
/border:<border_color>:<border_width>
/border:<border_color>:<horizontal_border_width>:<vertical_border_width>
/border:<border_color>:<left_border_width>:<top_border_width>:<right_border_width>:<bottom_border_width>
```
Adds an external border to the image.

#### Masking

### Mask
```html
/mask:rect:<corner_radius>
/mask:square:<corner_radius>
/mask:ellipse
/mask:circle
/mask:image:<base64 encoded source key or path>:<fit>
```
Masks the image with a mask.  The mask type is one of `circle`, `ellipse`, `image`, `rect`, or `square`.

### Redactions

#### Redact
```html
/redact:faces:<face_list>
/redact:people:<people_list>
/redact:region:<corner_radius>:<rotation>:<normalized_left>,<normalized_top>,<normalized_width>,<normalized_height>
/redact:color:<color>
/redact:blur:<blur_amount>
/redact:pixelate:<pixelate_amount>
/redact:mask:blur:<blur_amount>
/redact:mask:expand:<expand_percent>
/redact:mask:pixelate:<pixelate_amount>
```
Blurs or pixelates the faces, people, or regions in the image.

### Overlays
Overlays are images, text or shapes (rectangles and ellipses) that are drawn on top of the image.  Overlays are the last step in the processing pipeline and can/should be used for watermarks and similar.

You can have any number of overlays, though you are limited to the maximum length of the URL which is 4096 characters on most (but not all) systems.

It's recommended for complicated overlays to consolidate what you can in an SVG and use that as part of the overlay.

#### Overlay
```html
/ov:<overlay_index>:type:<image|text>
/ov:<overlay_index>:url:<base64 encoded source key or path>
/ov:<overlay_index>:text:<base64 encoded text>
/ov:<overlay_index>:font:<base64 encoded fontname>
/ov:<overlay_index>:xy:<px|rel>:<x>:<y>
/ov:<overlay_index>:a:<left|center|right>:<top|center|bottom>
/ov:<overlay_index>:sz:<px|rel>:<width>:<height>
/ov:<overlay_index>:minsz:<width>:<height>
/ov:<overlay_index>:maxsz:<width>:<height>
/ov:<overlay_index>:rot:<angle>
/ov:<overlay_index>:o:<opacity%>
/ov:<overlay_index>:fit:<fit|fill|crop>
/ov:<overlay_index>:tc:<text color>
/ov:<overlay_index>:fc:<fill color>
/ov:<overlay_index>:sc:<stroke color>
/ov:<overlay_index>:sw:<stroke width>
/ov:<overlay_index>:ds:o:<drop shadow opacity%>
/ov:<overlay_index>:ds:c:<drop shadow color>
/ov:<overlay_index>:ds:bl:<drop shadow blur>
/ov:<overlay_index>:ds:xy:<x offset>:<y offset>
/ov:<overlay_index>:bg:c:<background color>
/ov:<overlay_index>:bg:bl:<background blur>
/ov:<overlay_index>:bg:sat:<background saturation>
/ov:<overlay_index>:bg:con:<background contrast>
/ov:<overlay_index>:bg:bri:<background brightness>
/ov:<overlay_index>:bg:pad:<px|rel>:<h padding>:<v padding>
```
