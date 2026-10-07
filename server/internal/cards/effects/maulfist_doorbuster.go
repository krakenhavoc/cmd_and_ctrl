package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maulfist Doorbuster — Creature — Human Warrior {3}{R}, 4/2:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 Whenever this creature attacks, you may pay {E}. If you do, target
//	 creature can't block this turn."
//
// ADR 0129 §3 (#1995): the target, any creature, is chosen as the
// trigger goes on the stack (CR 603.3d); paid, it can't block this turn
// (a restriction until end of turn, CR 509.1b).
//
// No simplification.
func init() {
	attack := whenThisAttacksMayPayEnergy("Maulfist Doorbuster", 1, "stop it blocking",
		func(g *game.Game, item *game.StackItem) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return RestrictUntilEOT{Target: item.Targets[0].ID, Restrictions: game.CantBlock,
				Label: "Maulfist Doorbuster — can't block this turn"}.Apply(NewContext(g, item))
		})
	attack.Targets = TargetCreature("target creature")
	Register(Spec{
		OracleID:     "70a04e3f-e123-4925-9a02-9825488f1bc8",
		Name:         "Maulfist Doorbuster",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Maulfist Doorbuster", 2),
			attack,
		},
	})
}
