package rotate

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Rotator struct {
	loggers    []*LogFile
	ticker     *time.Ticker
	running    bool
	updatelock sync.Mutex
	period     string
}

func (c *Rotator) Add(l *LogFile) error {

	c.updatelock.Lock()

	defer c.updatelock.Unlock()

	for _, v := range c.loggers {

		if v == l {
			return errors.New("Already exists")
		}

	}

	if err := l.Reopen(); err != nil {
		return err
	}

	if l.GraceTime == 0 {
		l.GraceTime = 1 * time.Millisecond
	}

	loggers_ := append(c.loggers, l)

	c.loggers = loggers_

	return nil
}

func (c *Rotator) Del(l *LogFile) error {

	c.updatelock.Lock()

	defer c.updatelock.Unlock()

	for i, v := range c.loggers {
		if v == l {
			c.loggers[i], c.loggers = c.loggers[len(c.loggers)-1], c.loggers[:len(c.loggers)-1]
			return nil
		}
	}

	return errors.New("Not found")

}

func (c *Rotator) Start(rotationPeriod string) {

	c.period = rotationPeriod

	if c.running == true {
		return
	}

	go c.tickTack()

}

func (c *Rotator) tickTack() {

	now := time.Now().UTC()

	var until time.Time

	switch c.period {
	case "daily":
		until = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	case "hourly":
		until = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, time.UTC)
	}

	d := until.Sub(now)

	c.ticker = time.NewTicker(d)

	for range c.ticker.C {

		for _, l := range c.loggers {

			go l.Reopen()

		}

		c.ticker.Stop()

		go c.tickTack()

		return

	}

}

func (c *Rotator) Stop() {
	c.ticker.Stop()
}

type LogFile struct {
	oldWriter   *os.File
	Writer      **os.File
	CurrentFile string
	NamePrefix  string
	TimeFormat  string
	Symlink     bool
	CallBack    func(*os.File)
	GraceTime   time.Duration
}

func (c *LogFile) Reopen() error {

	filename := fmt.Sprintf("%s-%s", c.NamePrefix, time.Now().UTC().Format(c.TimeFormat))

	_, stat_err := os.Stat(filename)

	if filename != c.CurrentFile || stat_err != nil {

		file, err_f := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

		if err_f != nil {
			log.Printf("rotator: Failed to create new logfile '%s': %s - ignoring this change", filename, err_f)
			return err_f
		}

		c.CurrentFile = filename

		if c.Writer != nil {
			(*c.Writer) = file
		}

		if c.Symlink {

			os.Remove(c.NamePrefix)

			err_s := os.Symlink(filename, c.NamePrefix)

			if err_s != nil {
				log.Printf("rotator: Failed to symlink '%s' to '%s': %s", filename, c.NamePrefix, err_s)
			}

		}

		if c.oldWriter != nil {
			oldWriter := c.oldWriter
			go func() {
				time.Sleep(c.GraceTime)
				oldWriter.Close()
			}()
		}
		c.oldWriter = file

		if c.CallBack != nil {
			go c.CallBack(file)
		}

	}

	return nil

}
