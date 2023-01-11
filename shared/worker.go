package shared

import (
	"eve/service/model"
	"eve/utils"
	"sync"
	"time"

	"github.com/go-pg/pg"
)

// Worker ...
func SupportAccount() error {
	dbc := utils.Env.Db
	log := utils.Env.Log

	user := []model.User{}

	for {
		_, err := dbc.Query(&user, "select * from public.user where type = 7 and support_account = true and status = 1")
		if err != nil {
			if err == pg.ErrNoRows {
				// no tasks to perform sleep
				time.Sleep(30 * time.Second)
				continue
			}

			log.Debug(err)
			return err
		}

		if len(user) == 0 {
			// no tasks to perform sleep
			// log.Debug("sleep 10.1")
			time.Sleep(30 * time.Second)
			continue
		}

		if len(user) > 0 {
			// sleep for 1 hr once a user is found
			time.Sleep(60 * time.Minute)
			// create workers to handle the tasks
			disableUsers(user)
			continue
		}

	}

}

// Worker ...
func Worker() error {

	dbc := utils.Env.Db
	log := utils.Env.Log

	records := []model.TaskQueue{}
	gMon := GroupMonitor{}

	for {
		_, err := dbc.Query(&records, "select * from task_queue where status = 0")
		if err != nil {
			if err == pg.ErrNoRows {
				// no tasks to perform sleep
				log.Debug("sleep 10")
				time.Sleep(20 * time.Second)
				continue
			}

			log.Debug(err)
			return err
		}

		if len(records) == 0 {
			// no tasks to perform sleep
			// log.Debug("sleep 10.1")
			time.Sleep(20 * time.Second)
			continue
		}

		// create workers to handle the tasks
		distribute(records, &gMon, 10)
	}

}

func disableUsers(users []model.User) error {

	dbc := utils.Env.Db

	for _, user := range users {
		password := utils.MakeRandText(7)
		passwordHash, err := utils.HashPassword(password)
		if err != nil {
			return err
		}

		user.Password = passwordHash
		user.Status = model.IsDisabled

		_, err = dbc.Model(&user).Set("password =?password, status =?status").Where("id = ?id").Update()
		if err != nil {
			return err
		}
	}

	return nil

}

func distribute(tasks []model.TaskQueue, gMon *GroupMonitor, maxWorkers int) error {

	// check if the worker limit has been reached
	if gMon.Count() == maxWorkers {
		return nil
	}

	for i := 0; i < len(tasks); i++ {
		t := tasks[i]
		switch t.Type {
		case 1:
			// send email
			gMon.Add(1)
			func() {
				setTaskMode(sendEmail(t, gMon), t.ID)
			}()
		}

		// if worker limit has been reached, wait for a free worker
		if gMon.Count() >= maxWorkers {
			gMon.Available()
		}
	}

	return nil
}

func setTaskMode(failed bool, id int) {
	dbc := utils.Env.Db
	log := utils.Env.Log

	if failed {
		_, err := dbc.Exec("update task_queue set status = status - 1 where id = ?", id)
		if err != nil {
			log.Debug(err)
			return
		}
	} else {
		_, err := dbc.Exec("update task_queue set status = 1 where id = ?", id)
		if err != nil {
			log.Debug(err)
			return
		}
	}
}

// GroupMonitor ...
type GroupMonitor struct {
	mt    sync.Mutex
	count int
}

// Add ...
func (s *GroupMonitor) Add(delta int) {
	s.mt.Lock()
	defer func() {
		s.mt.Unlock()
	}()

	s.count += delta
}

// Done ...
func (s *GroupMonitor) Done() {
	s.Add(-1)
}

// Count ...
func (s *GroupMonitor) Count() int {
	s.mt.Lock()
	defer func() {
		s.mt.Unlock()
	}()

	return s.count
}

// Wait sleep until all workers are done
func (s *GroupMonitor) Wait() {

	for s.Count() != 0 {
		time.Sleep(1 * time.Second)
	}
}

// Available sleeps until a worker is available
func (s *GroupMonitor) Available() {
	cur := s.Count()
	for s.Count() >= cur {
		time.Sleep(1 * time.Second)
	}
}
