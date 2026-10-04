package main

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
)

// thinkingFlags resolves the thinking experiment (#2196) the same way
// for probe, suite run and arena: --think or CMDCTRL_BOT_THINK turns
// it on, --max-tokens or CMDCTRL_BOT_MAX_TOKENS sets the reply budget,
// and a thinking run that named no budget gets
// model.DefaultThinkingMaxTokens — written back into the result so the
// report records the number that was sent rather than a zero.
//
// A malformed environment value is an error, not a default: a run that
// silently fell back to thinking off would report the baseline under
// the experiment's name.
func thinkingFlags(think bool, maxTokens int) (bool, int, error) {
	envThink, envTokens, err := model.ThinkingFromEnv()
	if err != nil {
		return false, 0, err
	}
	think = think || envThink
	if maxTokens <= 0 {
		maxTokens = envTokens
	}
	if think && maxTokens <= 0 {
		maxTokens = model.DefaultThinkingMaxTokens
	}
	return think, maxTokens, nil
}

// thinkingNote is the line a report header carries about the reply
// budget and thinking. Empty on a default run.
func thinkingNote(think bool, maxTokens int) string {
	switch {
	case think:
		return fmt.Sprintf("thinking on, max tokens %d", maxTokens)
	case maxTokens > 0:
		return fmt.Sprintf("thinking off, max tokens %d", maxTokens)
	}
	return ""
}
