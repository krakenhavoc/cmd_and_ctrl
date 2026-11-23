package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckless Stormseeker // Storm-Charged Slasher — {2}{R} Creature — Human
// Werewolf 2/3 // Creature — Werewolf 3/4 (#2586, ADR 0132):
//
//	Front: "At the beginning of combat on your turn, target creature you
//	        control gets +1/+0 and gains haste until end of turn.
//	        Daybound"
//	Back:  "At the beginning of combat on your turn, target creature you
//	        control gets +2/+0 and gains trample and haste until end of
//	        turn.
//	        Nightbound"
//
// A targeted beginning-of-combat trigger (Avabruck Caretaker's shape):
// the creature is picked as the trigger goes on the stack and re-checked
// as it resolves, so one that left in response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	const oracle = "ea5fd21a-c23a-49ee-aab8-0a9618d65c11"
	boost := func(name string, power int, keywords ...string) game.TriggeredAbility {
		return effects.Targeting(
			effects.AtBeginningOfYourCombat(name+" — target creature you control gets a boost and gains keywords",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: target, Power: power, Label: name + " — power boost"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: target, Keywords: keywords, Label: name + " — keywords"}.Apply(ctx)
				}),
			effects.TargetCreature("target creature you control", effects.YouControl()))
	}
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Reckless Stormseeker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Triggered:       []game.TriggeredAbility{boost("Reckless Stormseeker", 1, "haste")},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Storm-Charged Slasher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Triggered:       []game.TriggeredAbility{boost("Storm-Charged Slasher", 2, "trample", "haste")},
	})
}
