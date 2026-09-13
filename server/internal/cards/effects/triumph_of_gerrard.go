package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Triumph of Gerrard — Enchantment — Saga for {1}{W}:
//
//	"I, II — Put a +1/+1 counter on target creature you control with
//	         the greatest power.
//	 III — Target creature you control with the greatest power gains
//	       flying, first strike, and lifelink until end of turn."
//
// "With the greatest power" picks out a SET, not a card: on a tie
// every tied creature is a legal target and the controller chooses
// (CR 700.3). GreatestPowerYouControl is written as a target
// predicate for exactly that reason, and it is re-evaluated at
// resolution like any other, so a pump in response can legally take
// the chosen creature out of the set and fizzle the chapter.
func init() {
	Register(Spec{
		OracleID:     "426a5f3f-f161-49b3-97d0-02cda554752b",
		Name:         "Triumph of Gerrard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, "Triumph of Gerrard — I: +1/+1 counter on your biggest creature",
				targetGreatestPowerYouControl(), gerrardCounter),
			ChapterTriggerTargeting(2, "Triumph of Gerrard — II: +1/+1 counter on your biggest creature",
				targetGreatestPowerYouControl(), gerrardCounter),
			ChapterTriggerTargeting(3, "Triumph of Gerrard — III: flying, first strike and lifelink",
				targetGreatestPowerYouControl(), gerrardKeywords),
		},
	})
}

// targetGreatestPowerYouControl builds a fresh spec per chapter.
// Fresh, not shared: a TargetSpec is handed to the engine and stored
// on the stack item, and three chapters sharing one pointer would
// make any future per-item mutation of it leak across chapters.
func targetGreatestPowerYouControl() *game.TargetSpec {
	return TargetCreature("target creature you control with the greatest power",
		GreatestPowerYouControl())
}

func gerrardCounter(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return AddCounter{
		Target: item.Targets[0].ID,
		Kind:   game.CounterPlusOne,
		N:      1,
	}.Apply(NewContext(g, item))
}

func gerrardKeywords(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   item.Targets[0].ID,
		Keywords: []string{"flying", "first strike", "lifelink"},
		Label:    "Triumph of Gerrard — flying, first strike, lifelink",
	}.Apply(NewContext(g, item))
}
