package main

import (
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"path/filepath"
	"testing"
)

type result struct {
	name  string
	conf  string
	svCmd *svCmd
}

func TestParseSvCfg(t *testing.T) {
	var cfSets = []result{
		{
			name: "SmsTech",
			conf: `
[program:queueSms]
process_name=%(program_name)s_%(process_num)02d
command=php /var/www/sms-service/yii queue-sms/listen --verbose=1 --color=0
autostart=true
autorestart=true
user=www-set
numprocs=400
redirect_stderr=true
stdout_logfile=/var/log/supervisor/queue-sms.log
startretries=10
	`,
			svCmd: &svCmd{
				name:    "queueSms",
				command: []string{"php", "/var/www/sms-service/yii", "queue-sms/info"},
			},
		},
		{
			name: "Apiprofit",
			conf: `
[supervisord]
identifier = supervisor

[program:lead_queue_processing]
process_name = %(program_name)s_%(process_num)02d
command = php /var/www/html/yii2-main/console/../yii queue/listen lead_queue_processing --isolate=0 --verbose=1
autostart = true
autorestart = true
user = www-set
numprocs = 1000
redirect_stderr = true
stdout_logfile = /var/log/supervisor/lead_queue_processing.log
	`,
			svCmd: &svCmd{
				name:    "lead_queue_processing",
				command: []string{"php", "/var/www/html/yii2-main/console/../yii", "queue/info", "lead_queue_processing"},
			},
		},
	}

	for _, set := range cfSets {
		t.Run(set.name, func(t *testing.T) {
			f, err := os.OpenFile(filepath.Join("qobserver_test.conf"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = f.Close()
				_ = os.Remove(f.Name())
			}()

			if _, err := f.Write([]byte(set.conf)); err != nil {
				log.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}

			svCmd, err := parseSvCfg(f.Name())

			require.NoError(t, err)
			require.Equal(t, set.svCmd, svCmd)
		})
	}
}
