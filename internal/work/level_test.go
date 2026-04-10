package work

import "testing"

func TestValidLevel(t *testing.T) {
	valid := []string{
		LevelVision,
		LevelRoadmap,
		LevelInitiative,
		LevelStory,
		LevelTask,
		LevelSubtask,
	}
	for _, level := range valid {
		if !ValidLevel(level) {
			t.Errorf("ValidLevel(%q) = false, want true", level)
		}
	}

	invalid := []string{
		"",
		"milestone",
		"Vision",
		"TASK",
		"Subtask",
	}
	for _, level := range invalid {
		if ValidLevel(level) {
			t.Errorf("ValidLevel(%q) = true, want false", level)
		}
	}
}
