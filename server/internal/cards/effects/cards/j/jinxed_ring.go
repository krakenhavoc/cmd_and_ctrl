package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jinxed Ring — Artifact {2} (Stronghold, S82 OLD batch 747, #5627):
//
//	"Whenever a nontoken permanent is put into your graveyard from the
//	 battlefield, this artifact deals 1 damage to you.
//	 Sacrifice a creature: Target opponent gains control of this artifact.
//	 (This effect lasts indefinitely.)"
//
// Scrapheap's put-into-your-graveyard reading: any nontoken permanent whose
// OWNER is the Ring's controller (so a permanent of yours an opponent stole
// counts, and one of theirs you stole does not), one trigger per permanent
// (CR 603.2c), the Ring seeing itself go too (CR 603.10a); the Ring is the
// damage source. The second ability is Jinxed Idol's: the control change is
// indefinite (CR 611.2a) and made only while the Ring is still the permanent
// the ability came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29cb8f47-7bb5-4881-98ef-59fc858cf9d7",
		Name:         "Jinxed Ring",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !c.IsToken() && c.Owner == source.Controller
			}, "Jinxed Ring — 1 damage to you", func(g *game.Game, item *game.StackItem) error {
				return effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: Target opponent gains control of this artifact.",
			Cost:    effects.SacrificeACreature(),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:  effects.OldB586TargetOpponentGainsControlOfThis("Jinxed Ring — target opponent gains control of it"),
		}},
	})
}
