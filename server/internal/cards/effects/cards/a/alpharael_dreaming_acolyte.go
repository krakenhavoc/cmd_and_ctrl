package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alpharael, Dreaming Acolyte — Legendary Creature — Human Cleric
// {1}{U}{B}, 2/3 (Edge of Eternities, #4461):
//
//	"When Alpharael enters, draw two cards. Then discard two cards
//	 unless you discard an artifact card.
//	 During your turn, Alpharael has deathtouch."
//
// The entry trigger is Thirst for Knowledge's draw-then-discard
// (discardTwoUnlessOne) with two draws: the drawn cards are legal
// discards, and the discard modal accepts two cards of any kind or a
// single artifact card. It has its own key so the prompt names
// Alpharael. The deathtouch is Cid, Freeflier Pilot's shape: a layer 6
// grant on Alpharael alone, on only while it is its controller's turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4be4071-7910-460f-90f3-6c4d8cc9db65",
		Name:         "Alpharael, Dreaming Acolyte",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Alpharael, Dreaming Acolyte — draw two cards, then discard two cards unless you discard an artifact card",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 2, game.DrawThen{
						Ref: eoeABAlpharaelDiscard, Player: item.Controller, Source: item.SourceCardID, N: 2,
					})
				}),
		},
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.IsCreature() && effects.IsActivePlayer(g, source.Controller)
			}, "deathtouch"),
		},
	})
}

// eoeABAlpharaelDiscard is Alpharael's "then discard two cards unless
// you discard an artifact card". An on-disk key: never renamed, never
// reused.
var eoeABAlpharaelDiscard = effects.DiscardTwoUnlessOne("alpharael-discard-two-unless-artifact",
	"Alpharael, Dreaming Acolyte — discard two cards, or one artifact card",
	func(c game.Card) bool { return c.IsArtifact() })
