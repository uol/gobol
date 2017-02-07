package saw

import (
	"fmt"
	"log"
	"log/syslog"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/uol/gobol/rotate"
)

type devNull struct{}

func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
}

type LogfileSettings struct {
	Path           string
	LogLevel       string
	TimeFormat     string
	RotationPeriod string
	Symlink        bool
}

type Settings struct {
	File struct {
		WriteTo  bool
		Settings LogfileSettings
	}
	Syslog struct {
		Local         bool
		LocalSettings struct {
			Severity string
			Facility string
			Tag      string
		}
		UDP         bool
		UDPSettings struct {
			Severity string
			Facility string
			Tag      string
			Address  string
			Port     string
		}
		TCP         bool
		TCPSettings struct {
			Severity string
			Facility string
			Tag      string
			Address  string
			Port     string
		}
	}
}

func New(settings Settings) (*logrus.Logger, error) {
	if !settings.File.WriteTo && !settings.Syslog.Local && !settings.Syslog.UDP && !settings.Syslog.TCP {
		log.Fatalln(`AT LEAST ONE LOG OUTPUT MUST BE ENABLED`)
	}

	logger := logrus.New()
	if settings.File.WriteTo {
		if settings.File.Settings.Path != "" {
			if settings.File.Settings.RotationPeriod != "daily" && settings.File.Settings.RotationPeriod != "hourly" {
				return nil, fmt.Errorf(
					`Wrong RotationPeriod format: found: %v. Acceptable values are: daily and hourly`,
					settings.File.Settings.RotationPeriod,
				)
			}

			logLevel, err := getLogLevel(settings.File.Settings.LogLevel)
			if err != nil {
				return nil, err
			}
			logger.Level = logLevel

			var logRotator = new(rotate.Rotator)
			logRotator.Add(&rotate.LogFile{
				NamePrefix: settings.File.Settings.Path,
				TimeFormat: settings.File.Settings.TimeFormat,
				Symlink:    settings.File.Settings.Symlink,
				CallBack: func(f *os.File) {
					logger.Out = f
				},
			})
			logRotator.Start(settings.File.Settings.RotationPeriod)
		}
	} else {
		logger.Out = new(devNull)
	}

	if settings.Syslog.Local {
		severity, err := getSeverity(settings.Syslog.LocalSettings.Severity)
		if err != nil {
			return nil, err
		}
		facility, err := getFacility(settings.Syslog.LocalSettings.Facility)
		if err != nil {
			return nil, err
		}
		hook, err := logrus_syslog.NewSyslogHook("", "", severity|facility, settings.Syslog.LocalSettings.Tag)
		if err != nil {
			logger.Error("Unable to connect to local syslog daemon.")
		} else {
			logger.Hooks.Add(hook)
		}
	}

	if settings.Syslog.TCP {
		severity, err := getSeverity(settings.Syslog.TCPSettings.Severity)
		if err != nil {
			return nil, err
		}

		facility, err := getFacility(settings.Syslog.TCPSettings.Facility)
		if err != nil {
			return nil, err
		}

		hook, err := logrus_syslog.NewSyslogHook(
			settings.Syslog.TCPSettings.Address,
			settings.Syslog.TCPSettings.Port,
			severity|facility,
			settings.Syslog.TCPSettings.Tag,
		)

		if err != nil {
			logger.Error("Unable to connect to syslog daemon.")
		} else {
			logger.Hooks.Add(hook)
		}
	}

	if settings.Syslog.UDP {
		severity, err := getSeverity(settings.Syslog.UDPSettings.Severity)
		if err != nil {
			return nil, err
		}

		facility, err := getFacility(settings.Syslog.UDPSettings.Facility)
		if err != nil {
			return nil, err
		}

		hook, err := logrus_syslog.NewSyslogHook(
			settings.Syslog.UDPSettings.Address,
			settings.Syslog.UDPSettings.Port,
			severity|facility,
			settings.Syslog.UDPSettings.Tag,
		)

		if err != nil {
			logger.Error("Unable to connect to syslog daemon.")
		} else {
			logger.Hooks.Add(hook)
		}

	}

	return logger, nil
}

func getLogLevel(ll string) (logrus.Level, error) {
	level, ok := map[string]logrus.Level{
		"DEBUG": logrus.DebugLevel,
		"INFO":  logrus.InfoLevel,
		"WARN":  logrus.WarnLevel,
		"ERROR": logrus.ErrorLevel,
		"FATAL": logrus.FatalLevel,
		"PANIC": logrus.PanicLevel,
	}[ll]
	if !ok {
		return logrus.PanicLevel, fmt.Errorf(
			`Wrong logLevel format: found: %v Acceptable values are: DEBUG, INFO, WARN, ERROR, FATAL and PANIC`,
			ll,
		)
	}
	return level, nil
}

func getSeverity(severity string) (syslog.Priority, error) {
	priority, ok := map[string]syslog.Priority{
		"EMERGENCY": syslog.LOG_EMERG,
		"ALERT":     syslog.LOG_ALERT,
		"CRITICAL":  syslog.LOG_CRIT,
		"ERROR":     syslog.LOG_ERR,
		"WARNING":   syslog.LOG_WARNING,
		"NOTICE":    syslog.LOG_NOTICE,
		"INFO":      syslog.LOG_INFO,
		"DEBUG":     syslog.LOG_DEBUG,
	}[severity]
	if !ok {
		return syslog.LOG_EMERG, fmt.Errorf(
			`Wrong severity format: found: %v Acceptable values are: EMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFO and DEBUG`,
			severity,
		)
	}
	return priority, nil
}

func getFacility(facility string) (syslog.Priority, error) {
	priority, ok := map[string]syslog.Priority{
		"KERN":     syslog.LOG_KERN,
		"USER":     syslog.LOG_USER,
		"MAIL":     syslog.LOG_MAIL,
		"DAEMON":   syslog.LOG_DAEMON,
		"AUTH":     syslog.LOG_AUTH,
		"SYSLOG":   syslog.LOG_SYSLOG,
		"LPR":      syslog.LOG_LPR,
		"NEWS":     syslog.LOG_NEWS,
		"UUCP":     syslog.LOG_UUCP,
		"CRON":     syslog.LOG_CRON,
		"AUTHPRIV": syslog.LOG_AUTHPRIV,
		"FTP":      syslog.LOG_FTP,
		"LOCAL0":   syslog.LOG_LOCAL0,
		"LOCAL1":   syslog.LOG_LOCAL1,
		"LOCAL2":   syslog.LOG_LOCAL2,
		"LOCAL3":   syslog.LOG_LOCAL3,
		"LOCAL4":   syslog.LOG_LOCAL4,
		"LOCAL5":   syslog.LOG_LOCAL5,
		"LOCAL6":   syslog.LOG_LOCAL6,
		"LOCAL7":   syslog.LOG_LOCAL7,
	}[facility]
	if !ok {
		return syslog.LOG_EMERG, fmt.Errorf(
			`Wrong severity format: found: %v Acceptable values are: UUCP, CRON, NEWS, AUTHPRIV, FTP, LOCAL0, LOCAL1, LOCAL2, LOCAL3, LOCAL4, LOCAL5, LOCAL6 and LOCAL7`,
			facility,
		)
	}
	return priority, nil
}
