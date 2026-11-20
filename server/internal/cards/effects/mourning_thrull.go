package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mourning Thrull — Creature — Thrull {1}{W/B}, 1/1 (Guildpact, #5599):
//
//	"({W/B} can be paid with either {W} or {B}.)
//	 Flying
//	 Whenever this creature deals damage, you gain that much life."
//
// Horned Cheetah's trigger on a flyer: any damage the Thrull deals,
// combat or not, read off the event's Amount (the damage actually dealt,
// after prevention). It is a triggered ability, not lifelink, so the life
// comes when it resolves, to the Thrull's controller (CR 109.5). The
// hybrid symbol is printed mana cost, paid by the shared payer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0406a534-c69c-40bb-801f-228a662ac1a2",
		Name:            "Mourning Thrull",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && ev.Source == source.InstanceID
			},
			Key: "Mourning Thrull — you gain that much life",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
