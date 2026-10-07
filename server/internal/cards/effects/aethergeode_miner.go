package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aethergeode Miner — Creature — Dwarf Scout {1}{W}, 3/1:
//
//	"Whenever this creature attacks, you get {E}{E} (two energy counters).
//	 Pay {E}{E}: Exile this creature, then return it to the battlefield
//	 under its owner's control."
//
// ADR 0129 PR 1 (#1995). The blink is the Flicker primitive on the
// source: it returns as a new object (CR 400.7), untapped and out of
// combat, under its owner's control. A Miner that left and came back
// before the ability resolves is a different object and is not
// blinked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92045f6c-9ec2-4bf0-826c-e21ef253011c",
		Name:         "Aethergeode Miner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Aethergeode Miner — you get {E}{E}", ebYouGetEnergy(2)),
		},
		Activated: []ActivatedAbility{{
			Label: "Pay {E}{E}: Exile this creature, then return it to the battlefield under its owner's control.",
			Cost:  PayEnergy(2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return Flicker{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
