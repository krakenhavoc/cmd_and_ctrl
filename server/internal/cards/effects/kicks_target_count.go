package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// CountedByKicks is "choose <target>, then choose another <target> for
// each time this spell was kicked" (#6150, CR 601.2c, 702.33c-d): Comet
// Storm, Strength of the Tajuru. The clause takes one target unkicked
// and one more for each kick announced with the optional costs, which
// the engine binds wherever the announcement is judged
// (game.TargetSpecUnderOptionalCosts). The targets are distinct, as
// "another" says (the clause's AllowSame stays off). Mutates and returns
// the clause, as WithCount does.
func CountedByKicks(t *game.TargetSpec) *game.TargetSpec {
	t.Min, t.Max = 1, 1
	t.CountFromKicks = true
	return t
}

// checkCountFromKicks refuses a kick-counted clause the engine cannot
// size: it binds only a spell's single card-level clause, from the kicks
// the spell's own kicker or multikicker announces.
func checkCountFromKicks(spec Spec) {
	if spec.Targets == nil || !spec.Targets.CountFromKicks {
		return
	}
	if spec.Targets.ClauseCount() != 1 {
		panic(fmt.Sprintf("effects.Register: %q counts a clause of a multi-clause statement by its kicks — only a single clause is bound", spec.Name))
	}
	for _, oc := range spec.OptionalCosts {
		if oc.Key == game.KickerKey || oc.Key == game.MultikickerKey {
			return
		}
	}
	panic(fmt.Sprintf("effects.Register: %q counts its targets by its kicks but has no kicker or multikicker", spec.Name))
}
