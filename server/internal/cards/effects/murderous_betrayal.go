package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Murderous Betrayal — {B}{B}{B} Enchantment:
//
//	"{B}{B}, Pay half your life, rounded up: Destroy target nonblack
//	 creature. It can't be regenerated."
//
// A repeatable Doom Blade whose price is half your life each time. The
// life is a computed cost (#1594, ADR 0020 Decision 47): the half is
// taken of the total the activator has AS THEY ACTIVATE (CR 601.2f–g
// via CR 602.2b), so at 20 life the first activation costs 10, the
// second 5, and at 1 life it still costs 1 — CR 119.4 allows paying
// down to exactly 0, which is lethal. At 0 life it costs 0.
//
// "It can't be regenerated" is #667's rider on the shared destroy, the
// same body Terminate and Big Game Hunter use.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2fe1e2b-b621-4406-b5e0-f2eebd12c2ba",
		Name:         "Murderous Betrayal",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}{B}, Pay half your life, rounded up: Destroy target nonblack creature. It can't be regenerated",
			Cost:    Plus(ManaCost("{B}{B}"), PayLifeCount(LifeHalfYoursRoundedUp)),
			Targets: TargetCreature("target nonblack creature", NonBlack()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return destroyTheTargetPermanentNoRegen(item, NewContext(g, item))
			},
		}},
	})
}
