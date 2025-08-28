package cmd

import (
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"path"
	"testing"
)

func TestCmdPool(t *testing.T) {
	testSet := map[string]struct {
		conf string
		cmd  Process
	}{
		"queueSms": {
			conf: `
	[program:queueSms]
	process_name=%(program_name)s_%(process_num)02d
	command=php /var/www/sms-service/yii queue-sms/listen --verbose=1 --color=0
	autostart=true
	autorestart=true
	user=www-set
	numprocs=400
	redirect_stderr=true
	stdout_logfile=/var/log/svr/queue-sms.log
	startretries=10
		`,
			cmd: Process{
				name:    "queueSms",
				command: []string{"php", "/var/www/sms-service/yii", "queue-sms/info"},
			},
		},
		"lead_queue_processing": {
			conf: `
	[supervisord]
	identifier = svr

	[program:lead_queue_processing]
	process_name = %(program_name)s_%(process_num)02d
	command = php /var/www/html/yii2-main/console/../yii queue/listen lead_queue_processing --isolate=0 --verbose=1
	autostart = true
	autorestart = true
	user = www-set
	numprocs = 1000
	redirect_stderr = true
	stdout_logfile = /var/log/svr/lead_queue_processing.log
		`,
			cmd: Process{
				name:    "lead_queue_processing",
				command: []string{"php", "/var/www/html/yii2-main/console/../yii", "queue/info", "lead_queue_processing"},
			},
		},
	}

	tempDir := t.TempDir()
	for _, set := range testSet {
		f, err := os.OpenFile(path.Join(tempDir, set.cmd.Name()+".conf"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := f.Write([]byte(set.conf)); err != nil {
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}

	pool := NewPool(nil)
	pool.Populate(tempDir)

	for _, cmd := range pool.GetAll() {
		require.Equal(t, testSet[cmd.Name()].cmd, *cmd)
	}
}
