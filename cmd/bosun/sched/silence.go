package sched

import (
	"fmt"
	"sync"
	"time"

	"bosun.org/cmd/bosun/expr"
	"bosun.org/models"
	"bosun.org/opentsdb"
)

// Silenced returns all currently silenced AlertKeys and the time they will be
// unsilenced.
func (s *Schedule) Silenced() map[expr.AlertKey]models.Silence {
	aks := make(map[expr.AlertKey]models.Silence)
	now := time.Now()
	silenceLock.RLock()
	defer silenceLock.RUnlock()
	for _, si := range s.Silence {
		if !si.ActiveAt(now) {
			continue
		}
		s.Lock("Silence")
		for ak := range s.status {
			if si.Silenced(now, ak.Name(), ak.Group()) {
				if aks[ak].End.Before(si.End) {
					aks[ak] = *si
				}
			}
		}
		s.Unlock()
	}
	return aks
}

var silenceLock = sync.RWMutex{}

func (s *Schedule) AddSilence(start, end time.Time, alert, tagList string, forget, confirm bool, edit, user, message string) (map[expr.AlertKey]bool, error) {
	if start.IsZero() || end.IsZero() {
		return nil, fmt.Errorf("both start and end must be specified")
	}
	if start.After(end) {
		return nil, fmt.Errorf("start time must be before end time")
	}
	if time.Since(end) > 0 {
		return nil, fmt.Errorf("end time must be in the future")
	}
	if alert == "" && tagList == "" {
		return nil, fmt.Errorf("must specify either alert or tags")
	}
	si := &models.Silence{
		Start:   start,
		End:     end,
		Alert:   alert,
		Tags:    make(opentsdb.TagSet),
		Forget:  forget,
		User:    user,
		Message: message,
	}
	if tagList != "" {
		tags, err := opentsdb.ParseTags(tagList)
		if err != nil && tags == nil {
			return nil, err
		}
		si.Tags = tags
	}
	silenceLock.Lock()
	defer silenceLock.Unlock()
	if confirm {
		delete(s.Silence, edit)
		s.Silence[si.ID()] = si
		return nil, nil
	}
	aks := make(map[expr.AlertKey]bool)
	for ak := range s.status {
		if si.Matches(ak.Name(), ak.Group()) {
			aks[ak] = s.status[ak].IsActive()
		}
	}
	return aks, nil
}

func (s *Schedule) ClearSilence(id string) error {
	silenceLock.Lock()
	defer silenceLock.Unlock()
	delete(s.Silence, id)
	return nil
}
