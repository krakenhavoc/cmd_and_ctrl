package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quicksmith Genius — Creature — Human Artificer {2}{R}, 2/2 (issue
// #1112):
//
//	"Whenever an artifact you control enters, you may discard a card.
//	 If you do, draw a card."
//
// A one-mana-cheaper Bloodmark Prime that turns every Treasure,
// Vehicle or Signet in the deck into a loot. The printed text is "IF
// you do", not "WHEN you do": that is a same-ability continuation
// (CR 603.1), not a CR 603.12 reflexive triggered ability, so the
// #636 reflexive-trigger vocabulary is the wrong tool here — it would
// put a second object on the stack this card never prints.
//
// The seam this card used to wait on was the ORDERING: an older
// discard primitive queued the prompt and returned immediately, so a
// draw written on the next line resolved before the player had even
// chosen what to pitch, turning a rummage into a strictly-better
// loot. That is fixed engine-side by the #1019/#1027 prompt-run
// continuation — PlayerDiscardsThenForEffect settles the discard leg
// first and only then runs its `Then` callback — which is exactly
// what b39MayDiscardThenDraw(1, ...) already wraps for "you may
// discard N. If you do, draw M" (Thrilling Discovery, Cathartic
// Pyre). Declining is a legal answer and draws nothing, per CR
// 601.2a's "if you do" being a real condition.
//
// The trigger condition is ArtifactEnteredUnderYourControl — the
// source itself included, though Quicksmith Genius is not an
// artifact and never matches it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bca0c136-3845-4cb9-8d80-f504a01ad50a",
		Name:         "Quicksmith Genius",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, ArtifactEnteredUnderYourControl,
				"Quicksmith Genius — discard a card. If you do, draw a card.",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return b39MayDiscardThenDraw(1, false,
						"Quicksmith Genius — discard a card to draw a card?",
						func(discarded int) int {
							if discarded > 0 {
								return 1
							}
							return 0
						})(ctx)
				}),
		},
	})
}
