Qobserver (мониторинг очередей yii2, запущенных с помощью supervisor)
---
> [!NOTE] 
> Описание работы.
> Читаются конфиги, и с указанным слипом делает вызовы команд (``php /path/to/your/project/yii queue-any/info``).
> Далее если были пересечения допустимых границ для waiting и delay, отправляется сообщение в Telegram.


## Примеры корректных конфигов
> [!WARNING]
> Конфиг супервизора в поле command должен содержать слово queue
---
```
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
```
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

Установка и использование (Linux)
---
```shell
git clone ...
make build-linux
```
Далее копируем бинарник из `bin` и пользуемся
```shell
./qobserver -tg-token=YOUR_BOT_TOKEN -tg-chat-id=YOUR_CHAT_ID -sleep=15m -max-wait=1000 -max-delay=80000 -verbose=true
```
Параметры для использования команды
---
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