# 🐙 Kraken-DCA

**Automated dollar cost averaging on the Kraken exchange**

![Docker](https://img.shields.io/github/actions/workflow/status/jagstangch/krakendca/docker-publish.yml?label=docker)
![Gosec](https://img.shields.io/github/actions/workflow/status/jagstangch/krakendca/gosec.yml?label=gosec)
![License](https://img.shields.io/github/license/jagstangch/krakendca)

> This is a fork of [Primexz/KrakenDCA](https://github.com/Primexz/KrakenDCA). See [Differences from upstream](#-differences-from-upstream) for what changed.

## Table of Contents
1. ➤ [About the project](#-about-the-project)
    - [Orders](#-orders)
    - [Post-only orders](#-post-only-orders)
2. ➤ [Configuration](#-configuration)
3. ➤ [Notifications](#-notifications)
4. ➤ [Metrics](#-metrics)
5. ➤ [Run with Docker](#-run-with-docker)
    - [Docker-CLI](#docker-cli)
    - [Docker-Compose](#-docker-compose)
6. ➤ [Run without Docker](#-run-without-docker)
7. ➤ [Differences from upstream](#-differences-from-upstream)

## 🔍 About the project

Since Kraken has extremely high fees for its crypto savings plan, this tool was developed to invest regularly in cryptocurrencies. Since Kraken offers two platforms: ‘Kraken’ and ‘Kraken Pro’, we take advantage of the fact that Kraken Pro has very low fees. All you have to do is regularly deposit fiat currencies on Kraken Pro and the bot does everything else for you.

### 💰 Orders

#### When will orders be placed?

Bitcoin orders are placed as often as possible. To illustrate the behaviour in more detail, let's look at the following example:
You deposit €500 per month on Kraken-Pro.
The bot calculates how often you can buy Bitcoin this month, as Kraken has a minimum purchase limit of 0.00006 BTC. Your orders are executed as often as possible throughout the month, thereby achieving dollar-cost averaging.

Kraken changes this minimum from time to time. You can check the current value with:

```bash
curl -s "https://api.kraken.com/0/public/AssetPairs?pair=XBTUSD" | jq '.result[].ordermin'
```

### 🧾 Post-only orders

Many DCA bots place market orders, or limit orders that cross the spread, so the order fills right away. Those always pay Kraken's **taker** fee. This bot places **post-only** limit orders by default (`LIMIT_ORDER_MODE=true`): Kraken only accepts the order if it rests on the order book, so it always pays the lower **maker** fee.

How it works:
1. At startup, the bot asks Kraken to validate a post-only order without placing it (`validate=true`). If the API key, pair, order size or price is rejected, you see it in the log right away instead of at the first purchase.
2. For each purchase, a post-only buy order is placed 0.1 below the current ask price and expires after 4 minutes.
3. If Kraken cancels it because it would have crossed the spread (the price moved), a new order is placed 0.1 lower. If it expires unfilled, a new order is placed at the new price. This repeats up to `LIMIT_ORDER_RETRY_COUNT` times.
4. A filled order logs the pair, the fee type (`liquidity=maker` or `liquidity=taker`), the fee and the fee in percent, so you can check you paid the maker rate:
   ```
   INFO kraken: Successfully bought btc, paid maker fee   cost=3.69829 fee=0.01479 feePercent=0.40% liquidity=maker orderType=limit pair=XBTEUR price=61638.1 volume=6e-05
   ```
   A post-only order can only fill as maker (Kraken cancels it otherwise), and a market order always fills as taker.
5. If no order filled after all retries, the purchase is skipped and you get a notification. Set `LIMIT_ORDER_MARKET_FALLBACK=true` to buy with a market order instead.

The bot never places a new order while the previous one might still be open, so a network error can't cause a double purchase.

## ⚙️ Configuration

This tool is configured via environment variables. Some environment variables are required and some activate additional functionalities.


| Variable | Description | Required | Default |
| --- | --- | --- | --- |
| `KRAKEN_PUBLIC_KEY` | Your Kraken API public key | ✅ | |
| `KRAKEN_PRIVATE_KEY` | Your Kraken API private key | ✅ | |
| `CURRENCY` | Your fiat currency to be used, e.g. USD, EUR, GBP or CHF | ❌  | `USD` |
| `KRAKEN_ORDER_SIZE` | The order size to be used. It must not be lower than Kraken's minimum order size. This value should only be edited if you know exactly what you are doing. | ❌ | `0.00006` |
| `LIMIT_ORDER_MODE` | If set to true, post-only limit orders are placed, so you pay the maker fee instead of the taker fee. See [Post-only orders](#-post-only-orders). If set to false, market orders are placed. | ❌ | `true` |
| `LIMIT_ORDER_RETRY_COUNT` | How many times a post-only order is retried before giving up. | ❌ | `8` |
| `LIMIT_ORDER_MARKET_FALLBACK` | If set to true, a market order (taker fee) is placed when no post-only order filled after all retries. | ❌ | `false` |
| `CHECK_DELAY` | How often the algorithm should be executed, in seconds. | ❌ | `60` |
| `METRIC_PORT` | Port of the metrics server. | ❌ | `3000` |
| `GOTIFY_URL` | Host (and optional port) of your Gotify server, without `http://` | ❌ |  |
| `GOTIFY_APP_TOKEN` | App token for the app on the Gotify server | ❌ |  |
| `NOTIFY_ON_PURCHASE` | Send a notification for every purchase. Errors and the startup message are always sent. | ❌ | `true` |
| `SMTP_HOST` | SMTP server for e-mail notifications, e.g. `smtp.gmail.com` | ❌ |  |
| `SMTP_PORT` | SMTP port. `465` uses TLS, any other port uses STARTTLS. | ❌ | `587` |
| `SMTP_USERNAME` | SMTP login | ❌ |  |
| `SMTP_PASSWORD` | SMTP password or app password | ❌ |  |
| `EMAIL_FROM` | Sender address | ❌ | `SMTP_USERNAME` |
| `EMAIL_TO` | Recipient address(es), comma-separated | ❌ |  |

## 📱 Notifications

Notifications are sent when the bot starts, for every purchase (unless `NOTIFY_ON_PURCHASE=false`) and when a purchase fails. Each purchase notification includes the pair, order type, volume, price, cost and fee, and whether the maker or taker fee was applied. At startup, the log shows which channels are enabled:

```
INFO notification: Notifications enabled   channels=gotify, e-mail
```

### Gotify

The environment variables `GOTIFY_URL` and `GOTIFY_APP_TOKEN` can be used to activate Gotify Push Notifications.

In this fork, notifications are sent over plain **HTTP** (e.g. `GOTIFY_URL=192.168.1.10:8080`). This is meant for a Gotify server on your local network. Don't point it at a server on the internet, as the app token would be sent unencrypted.

### E-mail

Set `SMTP_HOST` and `EMAIL_TO` to activate e-mail notifications. The connection is always encrypted: if the server doesn't support TLS, no e-mail is sent. Example for Gmail (create an [app password](https://myaccount.google.com/apppasswords) first, your normal password won't work):

```yaml
    environment:
      - SMTP_HOST=smtp.gmail.com
      - SMTP_PORT=587
      - SMTP_USERNAME=you@gmail.com
      - SMTP_PASSWORD=your-app-password
      - EMAIL_TO=you@gmail.com
```

Restart the bot after setting these. You should receive a "Kraken DCA started" e-mail within a few seconds. If not, look for `prefix=email` errors in the log.

## 📊 Metrics

The bot serves a small JSON endpoint at `http://<host>:3000/metrics` with the time of the next planned order (Unix timestamp in seconds):

```json
{"nextOrder": 1767225600}
```

## 🐳 Run with Docker

Images are built by GitHub Actions and published to `ghcr.io/jagstangch/krakendca`:

| Tag | Built from |
| --- | --- |
| `master` | Every push to `master` |
| `nightly` | Daily scheduled build |

### Docker-CLI

```bash
docker run -d --name kraken_dca \
  -e KRAKEN_PUBLIC_KEY=your-public-key \
  -e KRAKEN_PRIVATE_KEY=your-private-key \
  -e CURRENCY=EUR \
  --restart always \
  ghcr.io/jagstangch/krakendca:master
```


### 🚀 Docker-Compose

Create a `docker-compose.yml`:

```yaml
services:
  kraken_dca:
    image: ghcr.io/jagstangch/krakendca:master
    environment:
      - KRAKEN_PUBLIC_KEY=your-public-key
      - KRAKEN_PRIVATE_KEY=your-private-key
      - CURRENCY=EUR
    restart: always
```

```bash
docker compose up -d
```

To update to the latest image:

```bash
docker compose pull && docker compose up -d
```


## 💻 Run without Docker

This tool can be run directly with Go (1.23 or newer) for development.

```bash
KRAKEN_PUBLIC_KEY=your-public-key KRAKEN_PRIVATE_KEY=your-private-key go run .
```

## 🔀 Differences from upstream

| Change | Why |
| --- | --- |
| Default `KRAKEN_ORDER_SIZE` is `0.00006` | Matches Kraken's current BTC minimum order size |
| Post-only limit orders are on by default, checked at startup, with an optional market fallback | Pay the maker fee instead of the taker fee |
| Gotify notifications use HTTP instead of HTTPS | For a Gotify server on the local network |
| E-mail notifications, notifications for every purchase, and the fee type (maker/taker) and pair in the logs | Know what every purchase cost |
| The Dockerfile compiles the binary itself (multi-stage build) | So `docker build .` works without building with GoReleaser first |
| Docker image published by `docker-publish.yml` | Images under `ghcr.io/jagstangch/krakendca` |
| Startup log line says "modded by Jag" | To tell the fork's builds apart |
