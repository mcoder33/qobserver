# Qobserver (monitoring Yii2 queue workers managed by Supervisor)

> **What it does**
>
> Qobserver reads Supervisor configs and periodically executes Yii2 queue “info” commands (e.g. `php /path/to/project/yii queue/info` or `php /path/to/project/yii queue/info name`).
> If the observed metrics exceed the configured thresholds for **waiting** and/or **delay**, it sends an alert to **Telegram**.

---

## Supervisor config requirements

> **Important**
> The `command` field in a Supervisor program config must contain the word `queue`.

### Valid config examples

```ini
[program:queueSms]
process_name=%(program_name)s_%(process_num)02d
command=php /var/www/sms-service/yii queue-sms/listen --verbose=1 --color=0
autostart=true
autorestart=true
user=user
numprocs=400
redirect_stderr=true
stdout_logfile=/var/log/svr/queue-sms.log
startretries=10
```

```ini
[supervisord]
identifier = svr

[program:lead_queue_processing]
process_name = %(program_name)s_%(process_num)02d
command = php /var/www/html/yii2-main/console/../yii queue/listen lead_queue_processing --isolate=0 --verbose=1
autostart = true
autorestart = true
user = user
numprocs = 1000
redirect_stderr = true
stdout_logfile = /var/log/svr/lead_queue_processing.log
```

---

## Installation & usage (Linux)

```bash
git clone git@github.com:mcoder33/qobserver.git
make build-linux
```

Then copy the binary from `./bin` and run it:

```bash
./qobserver -tg-token=YOUR_BOT_TOKEN -tg-chat-id=YOUR_CHAT_ID -sleep=15m -max-wait=1000 -max-delay=80000 -verbose=true
```

Or just copy prebuild binary for Linux and use it 

```bash
./bin/qobserver_linux -tg-token=YOUR_BOT_TOKEN -tg-chat-id=YOUR_CHAT_ID -sleep=15m -max-wait=1000 -max-delay=80000 -verbose=true
```

---

## CLI flags

```
Usage of qobserver:
  -config string
        Path to supervisor conf.d directory (default "/etc/supervisor/conf.d")
  -max-delay int
        Threshold for delayed alert (default 10000)
  -max-wait int
        Threshold for waiting alert (default 1000)
  -sleep duration
        Sleep between info executing in seconds; use 1s,2s,Ns... (default 1s)
  -tg-chat-id string
        Telegram chat ID
  -tg-token string
        Telegram bot token
  -ttl duration
        Command execution ttl; use 1s,2s,Ns... (default 5s)
  -verbose
        Verbose mode
```

---

## Makefile targets

Run:

```bash
make help
```

Targets:

- `download-deps` — download Go module dependencies.
- `build` — build the binary for the current OS/arch into `./bin/qobserver`.
- `build-linux` — build a Linux amd64 binary into `./bin/qobserver_linux`.
- `test` — run tests (with `-race`) from `./internal/...`.
- `install-lint-deps` — install `golangci-lint` if missing.
- `lint` — run `golangci-lint` on the project.
- `help` — print this help.
