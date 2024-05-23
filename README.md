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
The [Foxy Preview](https://github.com) repository contains a Vue.js app that can be used to preview images and mess around with the API.

## Building URLs
The structure of the URL is as follows:

`http://localhost:8080/<source key>/<source image url or relative path base64 encoded>/<parameters>?s=<signature>`

The `source key` is the same as the `DEFAULT_USER_SOURCE_KEY` env variable.
The `source image url or relative path base64 encoded` is the image url or relative path to the image that you want to process.  If you are using an S3 source, then this would be the key, eg `path/to/file.jpg`.
The `parameters` is a list of parameters that you want to pass to the API separated by `/`.

So an example URL would be:

`http://localhost:8080/my-source-key/path/to/file.jpg/w:1000/h:1000/face:0/person:1/fmt:webp?s=<signature>`

The `signature` is a base64 encoded HMAC SHA256 of the URL using your `DEFAULT_USER_SECRET` as the key.

### Parameters
TBD