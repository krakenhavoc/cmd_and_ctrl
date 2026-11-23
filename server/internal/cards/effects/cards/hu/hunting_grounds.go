package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunting Grounds — Enchantment {G}{W} (Dominaria Remastered, #5948):
//
//	"Threshold — As long as there are seven or more cards in your graveyard,
//	 this enchantment has "Whenever an opponent casts a spell, you may put a
//	 creature card from your hand onto the battlefield.""
//
// The granted trigger is registered on the enchantment and only goes on the
// stack while its controller has seven or more cards in their graveyard, which
// is checked as the spell is cast: the ability is the enchantment's while the
// threshold holds and gone when it stops. The "may" is asked as the ability
// resolves (Anticausal Vestige's put-from-hand), and the creature card enters
// untapped under its controller's control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f3fec87-3dd2-4e5d-835c-eae535da4ecb",
		Name:         "Hunting Grounds",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B15OpponentCastSpell(ev, source) && effects.FraThresholdMet(g, source.Controller)
			},
			Key: "Hunting Grounds — you may put a creature card from your hand onto the battlefield",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutFromHandOntoBattlefield{
					Player:   item.Controller,
					Match:    effects.Creature(),
					Optional: true,
					Label:    "Hunting Grounds — you may put a creature card from your hand onto the battlefield",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
