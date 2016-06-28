# Luzifer / badge-gen

Ever ran into this scenario? You wanted to add a link to something to your GitHub project using a nice button like the Godoc or the Travis-CI button but you was not able to find a button for this having the text you wanted? I did. I wanted to add a button "API Documentation" to one of my projects and did not find any button with that caption. So I wrote it myself.

And I wasn't myself if I would allow me to do the same work twice or more often so I wrote a small webserver able to generate those buttons with a customizable text in SVG (I did not care about older browser since long ago)…

## Usage

### Using my version

Simple use the raw-API URL below or one of the URLs listed on the [demo page](https://badge.luzifer.io/):

```
https://badge.luzifer.io/v1/badge?title=API&text=Documentation&color=4c1
```

Parameters `title` and `text` are free-text strings while `color` has to be 3- or 6-letter hex notation for colors like that one you use in CSS.

To embed them into Markdown pages like this `README.md`:

```
![YourTitle](https://badge.luzifer.io/static/API/Documentation/4c1)
```

### Using your own hosted version

- There is a [Docker container](https://quay.io/repository/luzifer/badge-gen) for it. Just start it and use your own URL
- You also can download the binary from [GoBuilder.me](https://gobuilder.me/github.com/Luzifer/badge-gen) and use that one

### Popular buttons rebuilt

Hint: To get the source just look into the source of this README.md

![godoc reference](https://badge.luzifer.io/static/godoc/reference/5d79b5)
![API Documentation](https://badge.luzifer.io/static/API/Documentation/4c1)
![gratipay support](https://badge.luzifer.io/static/gratipay/support%20myproject/4c1)
![gitter chat](https://badge.luzifer.io/static/GITTER/JOIN%20CHAT/1dce73)
![achievement](https://badge.luzifer.io/static/Achievement/You%20found%20a%20badge!/911)

Yeah, sure you even could fake your Travis-CI build status but seriously: Why should you do that? Shame on you if you do!
