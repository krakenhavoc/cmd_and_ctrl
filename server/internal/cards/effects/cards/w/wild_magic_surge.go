package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wild Magic Surge — Instant {R}{R} (Commander Legends: Battle for Baldur's
// Gate, S82 GAP #5948):
//
//	"Destroy target permanent an opponent controls. Its controller reveals
//	 cards from the top of their library until they reveal a permanent card
//	 that shares a card type with that permanent. They put that card onto the
//	 battlefield and the rest on the bottom of their library in a random
//	 order."
//
// Proteus Staff's reveal. The permanent's controller and card types are read
// before it is destroyed (CR 608.2h: it is gone by the time the reveal runs),
// and the reveal runs whether or not the destroy happened (an indestructible
// target). A permanent card sharing one of those types (CR 205.2) is the stop;
// a library without one reveals itself and puts it all on the bottom. The
// player puts the card onto the battlefield, so it enters under their control.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41d16975-d6c0-4343-bea2-4f360ead9a1c",
		Name:         "Wild Magic Surge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent an opponent controls", effects.OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			id := item.Targets[0].ID
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			player := c.Controller
			var shared []string
			for _, ty := range []string{"artifact", "battle", "creature", "enchantment", "land", "planeswalker", "kindred"} {
				if c.HasCardType(ty) {
					shared = append(shared, ty)
				}
			}
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			run, hit := effects.RevealUntil(ctx, player, func(r game.Card) bool {
				if !effects.IsPermanentCard(r) {
					return false
				}
				for _, ty := range shared {
					if r.HasCardType(ty) {
						return true
					}
				}
				return false
			}, "Wild Magic Surge — reveal until a permanent card that shares a card type")
			return ctx.Game.PutByEnteringPlayerForEffect(func() error {
				return effects.PutFromLibraryOntoBattlefield{
					Player:   player,
					Cards:    run,
					Match:    func(_ *game.Game, _ uuid.UUID, r game.Card) bool { return hit != uuid.Nil && r.InstanceID == hit },
					All:      true,
					ThenTail: effects.PutRestOnBottomInRandomOrderTail,
				}.Apply(ctx)
			})
		},
	})
}
