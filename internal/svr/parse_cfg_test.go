package svr

import (
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestParseSvCfg(t *testing.T) {
	testSet := []struct {
		name string
		conf string
		Cmd  *Cmd
	}{
		{
			name: "queueSms",
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
			Cmd: &Cmd{
				name:    "queueSms",
				command: []string{"php", "/var/www/sms-service/yii", "queue-sms/info"},
			},
		},
		{
			name: "lead_queue_processing",
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
			Cmd: &Cmd{
				name:    "lead_queue_processing",
				command: []string{"php", "/var/www/html/yii2-main/console/../yii", "queue/info", "lead_queue_processing"},
			},
		},
	}

	tempDir := t.TempDir()
	for _, set := range testSet {
		t.Run(set.name, func(t *testing.T) {
			fullName := filepath.Join(tempDir, "qobserver_test.conf")
			f, err := os.OpenFile(fullName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
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

			cmd, err := ParseCfg(fullName, nil)

			require.NoError(t, err)
			require.Equal(t, set.Cmd, cmd)
		})
	}
}
