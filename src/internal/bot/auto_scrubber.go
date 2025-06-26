package bot

import (
	"log/slog"
	"time"
)

// scrubberInterval - How often we check
const scrubberInterval = 1 * time.Minute

var (
	// scrubbers - Map of auto scrubbers to monitor
	scrubbers map[string]map[string]time.Duration

	// Chan to use on shutdown
	scrubberStop = make(chan struct{})
)

// initScrubber - Loads the scrubber configs from the DB
func initScrubber() error {
	// Load config
	allScrubbers, err := b.Db.GetAllAutoScrubbers()
	if err != nil {
		return err
	}

	// map[GuildID][UserID]Interval
	scrubbers = make(map[string]map[string]time.Duration)
	for _, scrubber := range allScrubbers {
		if _, ok := scrubbers[scrubber.GuildID]; !ok {
			scrubbers[scrubber.GuildID] = make(map[string]time.Duration)
		}

		scrubbers[scrubber.GuildID][scrubber.UserID] = scrubber.Duration
	}

	runScrubber()

	return nil
}

// runScrubber - Start the auto scrubber system
func runScrubber() {
	ticker := time.NewTicker(scrubberInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				// Loop through all scrubbers, clone for funsies
				tmpScrub := scrubbers
				for guildID, users := range tmpScrub {
					for userID, interval := range users {
						emojis, err := b.Db.GetRecentEmojisForUser(guildID, userID, 12)
						if err != nil {
							slog.Error("Error getting recent emojis for user", "guild_id", guildID, "user_id", userID, "err", err)
							continue
						}

						// If older creation time + interval is before now, remove it
						for _, e := range emojis {
							if e.Timestamp.Add(interval).Before(time.Now()) {
								// Ignore errors here as it's likely bad data
								err := b.DiscordSession.MessageReactionRemove(e.ChannelID, e.MessageID, e.EmojiID, e.UserID)
								if err != nil {
									slog.Error("Error removing emoji reaction", "err", err, "emoji", e.EmojiID, "user", e.UserID)
								}

								// We care if we can't delete from our DB..
								err = b.Db.DeleteEmojiUsageById(e.ID)
								if err != nil {
									slog.Error("Error deleting emoji usage", "err", err, "emoji", e)
									continue
								}
							}
						}
					}
				}
			// Handle shutdown
			case <-scrubberStop:
				ticker.Stop()
				return
			}
		}
	}()
}

// startScrubbingUser - Start an auto scrubber for a user in a guild
func startScrubbingUser(guildID string, userID string, interval time.Duration) error {
	err := b.Db.AddAutoScrubber(guildID, userID, interval)
	if err != nil {
		slog.Error("Failed to add auto scrubber", "err", err)
		return err
	}

	// Add to map
	if _, ok := scrubbers[guildID]; !ok {
		scrubbers[guildID] = make(map[string]time.Duration)
	}
	scrubbers[guildID][userID] = interval

	return nil
}

// stopScrubbingUser - Stop an auto scrubber for a user in a guild
func stopScrubbingUser(guildID string, userID string) error {
	// Remove from instant
	delete(scrubbers[guildID], userID)
	return b.Db.RemoveAutoScrubber(guildID, userID)
}
