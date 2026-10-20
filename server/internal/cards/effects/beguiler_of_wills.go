package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beguiler of Wills — Creature — Human Wizard {3}{U}{U}, 1/1:
//
//	"{T}: Gain control of target creature with power less than or equal
//	 to the number of creatures you control."
//
// The CR 611.2a card again: the activated ability states no duration,
// so the theft lasts until the game ends or the creature stops being
// the object it was, and killing or bouncing the Beguiler gives
// nothing back (the effect was created by the ability's resolution and
// has no tie to its source). GainControl with an indefinite duration is
// Agent of Treachery's shape.
//
// The power bound is read twice, as the printed text says: when the
// target is chosen and again as the ability resolves (CR 608.2b), so a
// creature you lost in response, or one that grew, makes the target
// illegal. The count is every creature you control, the Beguiler
// included, taken from the battlefield rather than remembered.
//
// The {T} cost makes this an ordinary summoning-sick-gated activation
// (CR 302.6), and the stolen creature is itself summoning-sick under
// its new controller (CR 302.6), so it cannot attack that turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20c20c86-4435-49e0-bbf3-b51c49c2eadd",
		Name:         "Beguiler of Wills",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}: Gain control of target creature with power less than or equal to the number of creatures you control.",
			Cost:  TapCost(),
			Targets: TargetCreature("target creature with power less than or equal to the number of creatures you control",
				PowerAtMostCreaturesYouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target := FirstLegalBattlefieldTarget(ctx)
				if target == uuid.Nil {
					return nil
				}
				return GainControl{
					Target:     target,
					Controller: item.Controller,
					Duration:   game.IndefiniteDuration(),
					Label:      "Beguiler of Wills — gain control (no stated duration)",
				}.Apply(ctx)
			},
		}},
	})
}

// PowerAtMostCreaturesYouControl is "creature with power less than or
// equal to the number of creatures you control" (Beguiler of Wills),
// counted as the target is chosen and again as the ability resolves
// (CR 608.2b). "You" is the player choosing.
func PowerAtMostCreaturesYouControl() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		n := 0
		for _, p := range g.BattlefieldCardsForEffect() {
			if p.Controller == caster && p.IsCreature() {
				n++
			}
		}
		return c.CurrentPower() <= n
	}
}
