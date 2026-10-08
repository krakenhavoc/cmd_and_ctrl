package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Caustic Bronco — Creature — Snake Horse Mount {1}{B}:
//
//	"Whenever this creature attacks, reveal the top card of your library
//	 and put it into your hand. You lose life equal to that card's mana
//	 value if this creature isn't saddled. Otherwise, each opponent
//	 loses that much life.
//	 Saddle 3"
//
// Dark Confidant's reveal-then-move flip (a reveal is not a draw), with
// the payee decided by whether the Bronco was saddled. That is read when
// the trigger is built, off the live permanent, and carried on the item:
// a Bronco removed in response to its own trigger is judged as it last
// existed (CR 608.2h), which is saddled if it was. Saddle is sorcery
// speed, so the designation cannot change while the trigger waits.
//
// No simplification.
func init() {
	t := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		Key:     "Caustic Bronco — reveal the top card and put it into your hand",
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return attackDeclared(ev, source)
		},
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, "Caustic Bronco — reveal the top card and put it into your hand")
			if source.Saddled {
				item.Params.Amount = 1
			}
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			var revealed []uuid.UUID
			if err := (RevealTopOfLibrary{
				Player:   item.Controller,
				N:        1,
				Reason:   "Caustic Bronco — reveal the top card of your library",
				Revealed: &revealed,
			}).Apply(ctx); err != nil {
				return err
			}
			if len(revealed) == 0 {
				return nil
			}
			flipped := revealed[0]
			life := 0
			if c, ok := g.LookupCardForEffect(flipped); ok {
				life = c.ManaValue()
			}
			// The payment is the move's continuation (#993): a commander
			// returned to hand can pause on the CR 903.9 question, and
			// the life is paid after the card is in hand either way.
			return BounceToHand{Target: flipped, Then: func(ctx *Context, _ bool) error {
				if life == 0 {
					return nil
				}
				if item.Params.Amount == 1 {
					return eachOpponentLosesLife(ctx.Game, item, life)
				}
				return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -life)
			}}.Apply(ctx)
		},
	}
	Register(Spec{
		OracleID:     "165bdc68-1da0-47db-a069-978fc4c04b3f",
		Name:         "Caustic Bronco",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(3)},
		Triggered:    []game.TriggeredAbility{t},
	})
}
