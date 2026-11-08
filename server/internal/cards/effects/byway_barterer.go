package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Byway Barterer — Creature — Raccoon Rogue {2}{R}, 3/3 (Bloomburrow,
// #4467):
//
//	"Menace
//	 Whenever you expend 4, you may discard your hand. If you do, draw
//	 two cards. (You expend 4 as you spend your fourth total mana to cast
//	 spells during a turn.)"
//
// Expend is WheneverYouExpend (expend.go). The "you may" is Vaultguard
// Trooper's keyed yes/no and its body (discardHandThenDrawTwo): a yes
// discards the whole hand, even an empty one, and draws two. Menace
// rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a864b032-6e1a-4211-ad6e-54a5a5c7321b",
		Name:            "Byway Barterer",
		ReadsManaSpent:  game.ReadsYourCasts,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverYouExpend(4, "Byway Barterer — you may discard your hand; if you do, draw two cards",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Byway Barterer — discard your hand? If you do, draw two cards.",
						YesLabel: "Discard and draw two",
						NoLabel:  "Keep your hand",
						Then:     bywayBartererDiscardHand,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// bywayBartererDiscardHand is the yes. An on-disk key: never renamed,
// never reused.
var bywayBartererDiscardHand = MayYes("confirm/byway-barterer-discard-hand", discardHandThenDrawTwo)
