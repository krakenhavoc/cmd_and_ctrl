package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightfoot Rogue — Creature — Halfling Rogue {1}{B}, 2/1 (Adventures in
// the Forgotten Realms, #5499):
//
//	"Sneak Attack — Whenever this creature attacks, roll a d20.
//	 1—9 | This creature gains deathtouch until end of turn.
//	 10—19 | It gets +1/+0 and gains deathtouch until end of turn.
//	 20 | It gets +3/+0 and gains first strike and deathtouch until end of turn."
//
// The changes are made to this creature after the roll; if it has left the
// battlefield by then there is nothing to change.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a825e4c-4a87-4b8d-84f7-3e50b896d7cb",
		Name:         "Lightfoot Rogue",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Sneak Attack — roll a d20", effects.Do(effects.RollDice{Sides: 20, N: 1, Then: lightfootRogueTable})),
		},
	})
}

// lightfootRogueTable is the results table, after the d20.
var lightfootRogueTable = effects.DiceTail("dice/lightfoot-rogue-table", func(ctx *effects.Context, rolls []int) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	keywords := []string{"deathtouch"}
	power := 0
	if rolls[0] >= 10 {
		power = 1
	}
	if rolls[0] == 20 {
		power = 3
		keywords = []string{"first strike", "deathtouch"}
	}
	if power > 0 {
		if err := (effects.BoostUntilEOT{Target: ctx.Source(), Power: power, Label: "Lightfoot Rogue — Sneak Attack"}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: keywords, Label: "Lightfoot Rogue — Sneak Attack"}.Apply(ctx)
})
