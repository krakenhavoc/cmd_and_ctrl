package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyboon Evangelist — Creature — Bird Advisor {4}{W}, 3/3 (Streets of New
// Capenna Commander):
//
//	"Flying
//	 When this creature enters, support 6. (Put a +1/+1 counter on each of
//	 up to six other target creatures.)
//	 Whenever a creature with a counter on it attacks one of your
//	 opponents, that creature gains flying until end of turn."
//
// Support is the keyword action (support.go). The second ability is one
// trigger per attacking creature, whoever controls it: "attacks one of
// your opponents" is the player an opponent of the Evangelist's
// controller (CR 508.3e), not their planeswalker (Breena's reading), and
// the creature must carry a counter of any kind as it is declared. The
// flying is granted to that creature until end of turn, and it is read
// back from the trigger's event.
//
// No simplification.
func init() {
	const label = "Skyboon Evangelist — that creature gains flying until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "def2087a-5020-41e8-aa7b-d1f0c1f0fdb6",
		Name:            "Skyboon Evangelist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersSupport("Skyboon Evangelist", 6, true),
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				opp := effects.AttackedPlayer(g, ev)
				if opp == uuid.Nil || opp == source.Controller {
					return false
				}
				if p := g.PlayerByIDForEffect(opp); p == nil || p.Eliminated {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && effects.S82SncHasACounter(c)
			}, label, func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil || !effects.OnBattlefield(g, item.Trigger.Event.CardID) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: item.Trigger.Event.CardID, Keywords: []string{"flying"}, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
