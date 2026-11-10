package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nanoform Sentinel — Artifact Creature — Robot {2}{U}, 3/2 (Edge of
// Eternities, #4461):
//
//	"Whenever this creature becomes tapped, untap another target
//	 permanent. This ability triggers only once each turn."
//
// "Becomes tapped" is the tap event and the attack that taps it
// (ThisBecameTapped, #4534), so a vigilance attacker does not trigger it.
// The once-per-turn limit is the object's own record
// (TriggersOnlyOnceEachTurn): a second tap the same turn does not trigger,
// and a trigger that was countered or had no target still spent it. The
// target is any permanent but the Sentinel itself, its own tapped state
// included, and is rejudged as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "07155470-c5ab-41ae-ade3-bdc06d357379",
		Name:         "Nanoform Sentinel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(Targeting(
				OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
					"Nanoform Sentinel — untap another target permanent",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							return UntapTarget{Target: t.ID}.Apply(ctx)
						}
						return nil
					}),
				Another(TargetPermanent("another target permanent")))),
		},
	})
}
