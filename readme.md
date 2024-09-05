
# goico

Implements basic support for encoding and decoding ICO files.

## Supported

* .ico and .cur
* BMP Images
* 32bit Alpha
* CUR Hotspots

## Unsupported

* PNG Images

## Usage

### Config Decoding

```go
f, err := os.Open("favicon.ico")
if err != nil {
    panic(err)
}
defer f.Close()

config, err := goico.DecodeConfig(f)
if err != nil {
    panic(err)
}
```

### Decoding

```go
f, err := os.Open("favicon.ico")
if err != nil {
    panic(err)
}
defer f.Close()

config, imgs, err := goico.Decode(f)
if err != nil {
    panic(err)
}
```

### Encoding

```go
f, err := os.Create("favicon.ico")
if err != nil {
    panic(err)
}
defer f.Close()

imgs := []image.Image{...} // images with a max size of 256x256
ico, err := goico.NewICOConfig(imgs)
if err != nil {
    panic(err)
}

err = goico.Encode(f, ico, imgs)
if err != nil {
    panic(err)
}
```
