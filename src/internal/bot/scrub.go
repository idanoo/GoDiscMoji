package bot

import (
	"log/slog"
	"sync"
)

type Scrubber struct {
	scrubs map[string]map[string]bool
	mutex  sync.RWMutex
}

var scrub *Scrubber

// shouldScrub - Check if a user is scrubbing in a guild
func (s *Scrubber) shouldScrub(guildID string, userID string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if _, ok := s.scrubs[guildID]; !ok {
		return false
	}

	if _, ok := s.scrubs[guildID][userID]; !ok {
		return false
	}

	return true
}

// initScrub - Loads the scrubber configs from the DB
func initScrub() error {
	s := Scrubber{
		scrubs: make(map[string]map[string]bool),
		mutex:  sync.RWMutex{},
	}
	scrub = &s

	// Load config
	allScrubbers, err := b.Db.GetAllScrubs()
	if err != nil {
		return err
	}

	// map[GuildID][UserID]bool
	scrubs := make(map[string]map[string]bool)
	for _, scrubber := range allScrubbers {
		if _, ok := scrubs[scrubber.GuildID]; !ok {
			scrubs[scrubber.GuildID] = make(map[string]bool)
		}

		scrubs[scrubber.GuildID][scrubber.UserID] = true
	}

	scrub.scrubs = scrubs

	return nil
}

// startScrubbingUser - Start an auto scrubber for a user in a guild
func (s *Scrubber) startScrubbingUser(guildID string, userID string) error {
	err := b.Db.AddScrub(guildID, userID)
	if err != nil {
		slog.Error("Failed to add auto scrubber", "err", err)
		return err
	}

	// Add to map
	if _, ok := s.scrubs[guildID]; !ok {
		s.scrubs[guildID] = make(map[string]bool)
	}
	s.scrubs[guildID][userID] = true

	return nil
}

// stopScrubbingUser - Stop an auto scrubber for a user in a guild
func (s *Scrubber) stopScrubbingUser(guildID string, userID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.scrubs[guildID], userID)
	return b.Db.RemoveScrub(guildID, userID)
}
