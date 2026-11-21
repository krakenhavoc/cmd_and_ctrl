package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Part in Friendship — Enchantment {4}{G} (The Hobbit):
//
//	"Whenever a nontoken creature you control dies, reveal cards from the
//	 top of your library until you reveal a creature card. If its mana
//	 value is less than or equal to the number of lands you control, put
//	 it onto the battlefield. Otherwise, put it into your hand. Put the
//	 rest on the bottom of your library in a random order. This ability
//	 triggers only once each turn."
//
// Midnight Reaper's nontoken test (the creature is judged as it last
// existed), Exemplar of Light's once-each-turn count (b11TriggeredThisTurn
// reads the per-turn tally of triggers that went on the stack, so a
// countered one still counts, as printed), and Artificer Class's
// reveal-until walk. The hit's mana value is compared with the lands its
// controller has as the ability resolves, after the death that triggered it.
// Either way the rest of the revealed cards go on the bottom in a random
// order (the library-to-battlefield batch and the take to hand each carry
// that as their tail). With no creature card in the library all of it is
// revealed and put on the bottom, and nothing else happens.
//
// No simplification.
const partInFriendshipLabel = "Part in Friendship — reveal until a creature card; mana value at most your lands: battlefield, otherwise hand; the rest on the bottom in a random order"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1a9c467-2d64-4e11-adf3-7333c923b187",
		Name:         "Part in Friendship",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && !effects.IsToken(dead) &&
					!effects.B11TriggeredThisTurn(g, source.InstanceID, partInFriendshipLabel)
			}, partInFriendshipLabel, partInFriendshipReveal),
		},
	})
}

// partInFriendshipReveal is the ability's body.
func partInFriendshipReveal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Controller()
	run, hit := effects.RevealUntil(ctx, player, func(c game.Card) bool { return c.IsCreature() },
		"Part in Friendship — reveal until a creature card")
	isHit := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit }
	if hit != uuid.Nil {
		if card, ok := g.LookupCardForEffect(hit); ok {
			if mv, ok := g.ManaValueForEffect(card); ok && mv > effects.B10LandsControlled(g, player) {
				return effects.TakeFromLibraryToHand{
					Player: player,
					Cards:  run,
					Match:  isHit,
					All:    true,
					Label:  "Part in Friendship — put the creature card into your hand",
					Key:    effects.TakeRestBottomRandom,
				}.Apply(ctx)
			}
		}
	}
	return effects.PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    isHit,
		All:      true,
		ThenTail: effects.PutRestOnBottomInRandomOrderTail,
	}.Apply(ctx)
}
