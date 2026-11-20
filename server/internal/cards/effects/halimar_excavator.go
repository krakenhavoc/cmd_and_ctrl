package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Halimar Excavator — Creature — Human Wizard Ally {1}{U}, 1/3 (Worldwake, #5582):
//
//	"Whenever this creature or another Ally you control enters, target
//	 player mills X cards, where X is the number of Allies you control."
//
// The Ally trigger is South Pole Voyager's (wwkThisOrAnotherAllyEntered).
// X is counted as the trigger resolves (CR 608.2h): every Ally permanent
// its controller controls then, the Excavator included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd3e37c9-93bf-4f3e-a279-22afbffd8d43",
		Name:         "Halimar Excavator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, wwkThisOrAnotherAllyEntered,
				"Halimar Excavator — target player mills X cards, where X is the number of Allies you control",
				halimarExcavatorMill), TargetPlayer("target player")),
		},
	})
}

// halimarExcavatorMill mills the still-legal target player once for each
// Ally its controller controls as it resolves.
func halimarExcavatorMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	p, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	allies := permanentsControlledByMatching(g, item.Controller, Subtype(ctx.TextSubtype("Ally")))
	return MillCards{Player: p, N: len(allies)}.Apply(ctx)
}
