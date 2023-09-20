# MICROGOPSTER

A microservice that generates Topster charts using last.fm user data. Made using [holedaemon/gopster](https://github.com/holedaemon/gopster). Part of the Never Stable Club.

# HOW-TO

Once running, send a POST to the root of the service with a JSON body that looks something like this:

```json
{
    "user": "...", // required
    "period": "overall|7day|1month|3month|6month|12month", // optional"
    "title": "My Topster Chart", // optional
    "background_color": "#FFFFFF", // optional
    "text_color": "#000000", // optional
    "gap": 20, // optional
    "show_titles": true, // optional
    "show_numbers": true // optional, only accepted when show_titles is true
}
```

The response is another JSON object, with a base64 encoded image. Do with it what you will.

```json
{
    "image": "base64 encoded image"
}
```

# LICENSE

[MIT](LICENSE)