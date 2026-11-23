package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Conjurer's Mantle — Artifact — Equipment {1}{W} (March of the Machine
// Commander, #5477):
//
//	"Equipped creature gets +1/+1 and has vigilance.
//	 Whenever equipped creature attacks, look at the top six cards of your
//	 library. You may reveal a card that shares a creature type with that
//	 creature from among them and put it into your hand. Put the rest on
//	 the bottom of your library in a random order.
//	 Equip {1}"
//
// Batterbone's pump and vigilance, and Adaptive Omnitool's attack trigger
// (attachedCreatureAttacked, TakeFromLibraryToHand: only the card taken is
// revealed, the rest go to the bottom in a random order). "That creature"
// is the attacker as the ability resolves, or as it last existed there if
// it has gone (CR 608.2h); the card to take shares a creature type with it
// (s82MomB14SharesCreatureType), a changeling shares with any card that has a
// creature type (CR 702.73a), and a creature with no creature type shares
// with nothing, so every card goes to the bottom. A card need not be a creature
// card: a Kindred spell's creature types count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21b520c6-c94a-475a-a824-dd96394ea55b",
		Name:         "Conjurer's Mantle",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source)
			}, "Conjurer's Mantle — look at the top six cards, you may reveal a card that shares a creature type with the equipped creature and put it into your hand, the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					word := ""
					if info, ok := ctx.TriggeringPermanent(); ok {
						word = effects.Cmdr2B03TypesWord(info.Characteristic.Subtypes, info.Characteristic.AllCreatureTypes)
					}
					shares := effects.S82MomB14SharesCreatureType(word)
					return effects.TakeFromLibraryToHand{
						Player: item.Controller,
						Cards:  g.LookAtTopOfLibraryForEffect(item.Controller, 6),
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return shares(c)
						},
						Max:      1,
						Optional: true,
						Label:    "Conjurer's Mantle — you may reveal a card that shares a creature type with the equipped creature and put it into your hand",
						Key:      effects.TakeRevealRestBottomRandom,
					}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
