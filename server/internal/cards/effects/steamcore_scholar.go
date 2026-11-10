package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steamcore Scholar — Creature — Weird Detective {2}{U}, 2/2 (Murders at
// Karlov Manor, #4469):
//
//	"Flying, vigilance
//	 When this creature enters, draw two cards. Then discard two cards
//	 unless you discard an instant or sorcery card or a creature card
//	 with flying."
//
// Alpharael's draw-then-discard (discardTwoUnlessOne): the drawn cards
// are legal discards, and the modal accepts any two cards or a single
// card that is an instant, a sorcery or a creature with flying. A hand
// of one card discards that card whatever it is (CR 701.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0fb9a928-9adb-4dc1-9761-857df7e6c643",
		Name:         "Steamcore Scholar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Steamcore Scholar — draw two cards, then discard two cards unless you discard an instant or sorcery card or a creature card with flying",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 2, game.DrawThen{
						Ref: mkmSteamcoreDiscard, Player: item.Controller, Source: item.SourceCardID, N: 2,
					})
				}),
		},
	})
}

// mkmSteamcoreDiscard is Steamcore Scholar's "then discard two cards
// unless you discard an instant or sorcery card or a creature card with
// flying". An on-disk key: never renamed, never reused.
var mkmSteamcoreDiscard = discardTwoUnlessOne("steamcore-scholar-discard",
	"Steamcore Scholar — discard two cards, or one instant, sorcery or flying creature card",
	func(c game.Card) bool {
		return c.IsInstant() || c.IsSorcery() || (c.IsCreature() && game.HasKeyword(&c, "flying"))
	})
