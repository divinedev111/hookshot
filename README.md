[![CI](https://github.com/divinedev111/hookshot/actions/workflows/ci.yml/badge.svg)](https://github.com/divinedev111/hookshot/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/divinedev111/hookshot)](https://goreportcard.com/report/github.com/divinedev111/hookshot)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

# hookshot

Webhook testing from your terminal. Capture, inspect, replay, and export webhook requests with provider-aware detection and SQLite-backed persistence.

## Install

```bash
go install github.com/divinedev111/hookshot@latest
```

## Quick Start

```bash
hookshot listen :8080                    # start capturing webhooks
hookshot history                         # list captured events
hookshot inspect 1                       # view full event details
hookshot replay 1 --to localhost:3000    # replay to your app
hookshot export 1 --as curl              # get a runnable curl command
```

## Commands

### `hookshot listen <addr>`

Start a webhook capture server.

```bash
hookshot listen :8080
hookshot listen :8080 --forward localhost:3000/webhooks
hookshot listen :8080 --json
```

Captures every incoming request, pretty-prints it live, and stores it in SQLite. Use `--forward` to proxy each request to your local app and see response statuses. Use `--json` for structured JSON line output.

### `hookshot history`

List captured webhook events.

```bash
hookshot history
hookshot history --provider stripe --limit 10
hookshot history --json
```

### `hookshot inspect <id>`

Show full details of a captured event including headers, body, and forward response.

```bash
hookshot inspect 1
hookshot inspect 1 --json
```

### `hookshot replay <id> --to <url>`

Replay a captured request to any URL with the original headers and body.

```bash
hookshot replay 1 --to localhost:3000/webhooks
hookshot replay 1 --to https://staging.example.com/hooks --json
```

### `hookshot export <id> --as <format>`

Export a captured request as an executable command.

```bash
hookshot export 1 --as curl
```

## Provider Detection

hookshot auto-detects webhook providers from request headers:

| Provider | Detection Header | Event Type Source |
|----------|-----------------|-------------------|
| Stripe | `Stripe-Signature` | `type` field in JSON body |
| GitHub | `X-Github-Event` | Header value |
| Shopify | `X-Shopify-Hmac-Sha256` | `X-Shopify-Topic` header |
| Slack | `X-Slack-Signature` | `type` field in JSON body |
| Discord | `X-Signature-Ed25519` | `type` field in JSON body |
| Twilio | `X-Twilio-Signature` | `EventType` form field |

Unrecognized providers show as `[unknown]`.

## Why hookshot

- **Single binary** -- no runtime dependencies, no Docker, no cloud accounts
- **Provider-aware** -- auto-detects Stripe, GitHub, Shopify, Slack, Discord, Twilio
- **Persistent** -- SQLite storage means your webhook history survives restarts
- **Replay** -- resend any captured webhook to any URL with one command
- **Export** -- get runnable curl commands for any captured event
- **Forward mode** -- proxy webhooks to your local app while capturing them

Compared to webhook.site (cloud-only), ngrok (tunnel-focused), or Stripe CLI (single-provider), hookshot gives you a local-first, provider-agnostic tool for webhook development.

## License

MIT
