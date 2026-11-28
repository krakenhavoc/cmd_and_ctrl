package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mana Web — Artifact {3} (Weatherlight):
//
//	"Whenever a land an opponent controls is tapped for mana, tap all lands that
//	 player controls that could produce any type of mana that land could
//	 produce."
//
// A triggered ability (not a mana ability) of the Web's controller, watching
// every land an opponent taps for mana. As it resolves it reads what the tapped
// land could produce (CR 106.7) and taps each untapped land that player controls
// that could produce at least one of those types; the land that was tapped is
// already tapped. The mana already made stays in the pool.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82909ff4-f6f7-44f1-bbd0-6e54afa5ee6d",
		Name:         "Mana Web",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventManaAbilityActivated},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.B20LandTappedForMana(ev, g) {
					return false
				}
				land, ok := g.LookupCardForEffect(ev.Source)
				return ok && land.Controller != source.Controller
			},
			Key: "Mana Web — tap all lands that player controls that could produce any type of mana that land could produce",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				tapped, ok := g.LookupCardForEffect(item.Trigger.Event.Source)
				if !ok {
					return nil
				}
				types := map[string]bool{}
				for _, m := range g.ProducibleManaLocked(tapped) {
					types[m] = true
				}
				var ids []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller != tapped.Controller || !c.IsLand() || c.Tapped {
						continue
					}
					for _, m := range g.ProducibleManaLocked(c) {
						if types[m] {
							ids = append(ids, c.InstanceID)
							break
						}
					}
				}
				for _, id := range ids {
					if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
