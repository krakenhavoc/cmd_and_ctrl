package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skybreen — Plane — Kaldheim (Planechase Anthology, #4870):
//
//	"Players play with the top card of their libraries revealed.
//	 Spells that share a card type with the top card of a library can't be
//	 cast.
//	 Whenever chaos ensues, target player loses life equal to the number of
//	 cards in their hand."
//
// A plane (ADR 0216 §6). The reveal is Field of Dreams' (every library, read
// from the command zone while the plane is face up, library_top.go). The cast
// restriction reads every library's top card, not only the caster's (the
// ruling), and covers permanent spells too; a spell cast from the top of a
// library is judged against the card that is then on top. The chaos ability
// counts the target's hand as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "61006acf-de38-420b-ab7a-685fbd64141c",
		Name:              "Skybreen",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealedEachPlayer,
		CastRestrictions: []game.CastRestriction{{
			Label: "Spells that share a card type with the top card of a library can't be cast.",
			Forbids: func(q game.CastQuery) bool {
				for _, p := range q.Game.Seats {
					if p == nil || p.Library == nil {
						continue
					}
					if top, ok := skybreenTopCard(p.Library, q.Card); ok && skybreenSharesType(q.Card, top) {
						return true
					}
				}
				return false
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Skybreen — target player loses life equal to the number of cards in their hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						p := g.PlayerByIDForEffect(t.ID)
						if p == nil || p.Hand == nil {
							return nil
						}
						return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -p.Hand.Size())
					}
					return nil
				}),
				TargetPlayer("target player")),
		},
	})
}

// skybreenTopCard is the top card of a library as a cast is judged: the
// card being cast is not on top of anything once it is announced, so when
// it is the top card the card beneath it is.
func skybreenTopCard(lib *game.Zone, casting game.Card) (game.Card, bool) {
	for i := len(lib.Cards) - 1; i >= 0; i-- {
		if lib.Cards[i].InstanceID != casting.InstanceID {
			return lib.Cards[i], true
		}
	}
	return game.Card{}, false
}

var skybreenCardTypes = []string{"artifact", "battle", "creature", "enchantment", "instant", "kindred", "land", "planeswalker", "sorcery"}

func skybreenSharesType(a, b game.Card) bool {
	for _, t := range skybreenCardTypes {
		if a.HasCardType(t) && b.HasCardType(t) {
			return true
		}
	}
	return false
}
