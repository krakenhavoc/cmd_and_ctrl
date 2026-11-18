package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zurzoth, Chaos Rider — Legendary Creature — Devil {2}{R}, 2/3 (S82
// reprint pass):
//
//	"Whenever an opponent draws their first card each turn, if it's not
//	 their turn, you create a 1/1 red Devil creature token with "When this
//	 token dies, it deals 1 damage to any target."
//	 Whenever one or more Devils you control attack one or more players, you
//	 and those players each draw a card, then discard a card at random."
//
// The first trigger reads the player's per-turn draw tally, which the engine
// bumps before the event (Faerie Mastermind's shape), and the "if it's not
// their turn" intervening clause (CR 603.4) is checked as the card is drawn
// and again as the trigger resolves. The second is one trigger per
// declaration (CR 603.2c, OncePerBatch): its players are the ones the
// Devils were declared against, a planeswalker or battle not being a
// player. You draw and discard first, then each of those players in the
// order the attacks were declared; each draws a card and then discards one
// at random.
//
// No simplification.
func init() {
	notTheirTurn := func(g *game.Game, ev game.Event, you game.Card) bool {
		return ev.Kind == game.EventDrawCard && ev.Actor != uuid.Nil && ev.Actor != you.Controller &&
			g.TurnTallyFor(ev.Actor).CardsDrawn == 1 && !isActivePlayer(g, ev.Actor)
	}
	Register(Spec{
		OracleID:     "a7da9ec5-f82f-4a6d-9737-06a1556e2dd6",
		Name:         "Zurzoth, Chaos Rider",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return notTheirTurn(g, ev, *source)
			}, "Zurzoth, Chaos Rider — create a 1/1 red Devil creature token",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger != nil && isActivePlayer(g, item.Trigger.Event.Actor) {
						return nil
					}
					return CreateToken{Template: devilToken(), N: 1}.Apply(NewContext(g, item))
				}),
			OncePerBatchPerAttacker(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && attackerHadSubtype(g, ev.CardID, source.TextSubtype("Devil")) && attackedPlayer(g, ev) != uuid.Nil
			}, "Zurzoth, Chaos Rider — you and those players each draw a card, then discard a card at random",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := NewContext(g, item)
					players := append([]uuid.UUID{item.Controller}, zurzothDevilAttackedPlayers(g, item.Trigger.Event, item.Controller, ctx.TextSubtype("Devil"))...)
					for _, id := range players {
						if p := g.PlayerByIDForEffect(id); p == nil || p.Eliminated {
							continue
						}
						if err := (DrawCards{Player: id, N: 1}).Apply(ctx); err != nil {
							return err
						}
						if err := (DiscardCards{Player: id, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})),
		},
	})
}
