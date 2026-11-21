package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wight — Creature — Zombie Soldier {1}{B}, 3/2 (Adventures in the
// Forgotten Realms, S82):
//
//	"This creature enters tapped.
//	 Life Drain — Whenever a creature dealt damage by this creature this turn
//	 dies, create a tapped 2/2 black Zombie creature token and exile that
//	 card."
//
// The entry is Oathsworn Vampire's. The death trigger is Predator Ooze's, over
// the turn's record of damaged objects. The token is made whether or not the
// card can still be exiled; "that card" is exiled only if it is still in its
// owner's graveyard (a token that died is no card).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a07a676-890f-49a2-9a2d-17d7b1dc6868",
		Name:         "Wight",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && effects.DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Wight — a creature it damaged died: create a tapped 2/2 black Zombie token and exile that card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					dead := ctx.Trigger().Event.CardID
					if err := (effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("2/2 black Zombie")).EntersTapped(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					if z := g.FindCardZoneForEffect(dead); z != nil && z.Kind == game.ZoneGraveyard {
						return g.ExileCardForEffect(dead)
					}
					return nil
				}),
		},
	})
}
