package work

// WBS level constants.
const (
	LevelVision     = "vision"
	LevelRoadmap    = "roadmap"
	LevelInitiative = "initiative"
	LevelStory      = "story"
	LevelTask       = "task"
	LevelSubtask    = "subtask"
)

// ValidLevel reports whether level is a known WBS level.
func ValidLevel(level string) bool {
	switch level {
	case LevelVision, LevelRoadmap, LevelInitiative, LevelStory, LevelTask, LevelSubtask:
		return true
	default:
		return false
	}
}
