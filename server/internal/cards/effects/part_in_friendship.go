package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "e1a9c467-2d64-4e11-adf3-7333c923b187",
		Name:         "Part in Friendship",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && leftUnderControlOf(ev, dead) == source.Controller && !IsToken(dead) &&
					!b11TriggeredThisTurn(g, source.InstanceID, partInFriendshipLabel)
			}, partInFriendshipLabel, partInFriendshipReveal),
		},
	})
}

// partInFriendshipReveal is the ability's body.
func partInFriendshipReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	run, hit := revealUntil(ctx, player, func(c game.Card) bool { return c.IsCreature() },
		"Part in Friendship — reveal until a creature card")
	isHit := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit }
	if hit != uuid.Nil {
		if card, ok := g.LookupCardForEffect(hit); ok {
			if mv, ok := g.ManaValueForEffect(card); ok && mv > b10LandsControlled(g, player) {
				return TakeFromLibraryToHand{
					Player: player,
					Cards:  run,
					Match:  isHit,
					All:    true,
					Label:  "Part in Friendship — put the creature card into your hand",
					Key:    takeRestBottomRandom,
				}.Apply(ctx)
			}
		}
	}
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    isHit,
		All:      true,
		ThenTail: PutRestOnBottomInRandomOrderTail,
	}.Apply(ctx)
}
