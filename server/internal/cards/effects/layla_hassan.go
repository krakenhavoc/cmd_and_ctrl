package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Layla Hassan — Legendary Creature — Human Assassin {3}{W}, 3/4 (Assassin's
// Creed, #5948):
//
//	"First strike
//	 When Layla Hassan enters and whenever one or more Assassins you control
//	 deal combat damage to a player, return target historic card from your
//	 graveyard to your hand."
//
// One printed ability on two events: Layla entering, and a damage batch
// (once per player hit per step, however many Assassins connect, CR 603.2c).
// The target is a historic card (Coal Hill School's b14Historic: artifact,
// legendary or Saga) in your graveyard, chosen as the trigger goes on the stack
// and re-checked as it resolves; with none there is no trigger (CR 603.3d).
//
// No simplification.
func init() {
	t := OncePerBatchPerPlayer(OnAny([]game.EventKind{game.EventETB, game.EventDealDamage},
		func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
			switch ev.Kind {
			case game.EventETB:
				return Self(ev, source, ch, g)
			case game.EventDealDamage:
				if !combatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && HasSubtype("Assassin")(g, source.Controller, game.ThroughTextOf(source, dealer))
			}
			return false
		}, "Layla Hassan — return target historic card from your graveyard to your hand", returnTargetCardToHand))
	Register(Spec{
		OracleID:        "e27feb69-3120-4e8d-8ee6-0f87b58e2ae3",
		Name:            "Layla Hassan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(t, TargetCardInGraveyard("target historic card from your graveyard", b14Historic(), YouOwn())),
		},
	})
}
