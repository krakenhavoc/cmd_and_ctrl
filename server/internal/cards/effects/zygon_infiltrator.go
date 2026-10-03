package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zygon Infiltrator — Creature — Alien Shapeshifter Soldier (3/2) for {2}{U}:
//
//	"Body-print — {2}{U}: Tap another target creature and put a stun counter on it. This creature becomes a copy of that creature for as long as that creature remains tapped. Activate only as a sorcery. (If a permanent with a stun counter would become untapped, remove one from it instead.)"
//
// ADR 0109 §3 (#1894): a copy effect (CR 707.2, layer 1) on this
// creature, timed by "for as long as THAT creature remains tapped" —
// WhilePinnedRemainsTapped, a duration about the creature copied
// rather than about the source. The stun counter means the first untap
// of that creature only removes the counter (CR 122.1d), so the copy
// lasts at least until the second. "Body-print" is an ability word
// with no rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "693c071a-d6b5-4a96-8fe3-8c61fa9406bc",
		Name:         "Zygon Infiltrator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "Body-print \u2014 {2}{U}: Tap another target creature and put a stun counter on it. This creature becomes a copy of that creature for as long as that creature remains tapped. Activate only as a sorcery. (If a permanent with a stun counter would become untapped, remove one from it instead.)",
			Cost:         ManaCost("{2}{U}"),
			SorcerySpeed: true,
			Targets:      Another(TargetCreature("another target creature")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				model := FirstLegalBattlefieldTarget(ctx)
				if model == uuid.Nil {
					return nil
				}
				if err := (TapTarget{Target: model}).Apply(ctx); err != nil {
					return err
				}
				if err := (AddCounter{Target: model, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return BecomeCopy{
					Targets:  []uuid.UUID{ctx.Source()},
					Of:       model,
					Duration: CopyWhileOfRemainsTapped,
					Label:    "Zygon Infiltrator — a copy of that creature while it remains tapped",
				}.Apply(ctx)
			},
		}},
	})
}
