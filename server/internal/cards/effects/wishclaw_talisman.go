package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wishclaw Talisman — Artifact {1}{B} (EDHREC rank 525):
//
//	"This artifact enters with three wish counters on it.
//	 {1}, {T}, Remove a wish counter from this artifact: Search your
//	 library for a card, put it into your hand, then shuffle. An
//	 opponent gains control of this artifact. Activate only during your
//	 turn."
//
// The counters are an entry replacement (Braided Net's shape) and the
// removal is a cost, so a Talisman with no wish counters left can't be
// activated at all (the 2019-10-04 ruling). "Activate only during your
// turn" is DuringYourTurn: whoever controls it may use it on their own
// turn, which is how it passes round the table.
//
// The search is Demonic Tutor's. Once it has finished, the activating
// player chooses an opponent as the ability resolves (the ruling) and
// that player gains control of the Talisman with no end (CR 611.2a),
// still tapped, so they can use it on their next turn. A Talisman
// that left the battlefield and came back is a new object and is not
// handed over (CR 400.7). If the new controller leaves the game, the
// control effect ends with them (CR 800.4a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81c70ae7-3c18-4c9b-8505-e4db9e0e6518",
		Name:         "Wishclaw Talisman",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters("wish", 3, "Wishclaw Talisman: enters with three wish counters"),
		},
		Activated: []ActivatedAbility{{
			Label:     "{1}, {T}, Remove a wish counter from this artifact: Search your library for a card, put it into your hand, then shuffle. An opponent gains control of this artifact. Activate only during your turn.",
			Cost:      Plus(ManaCost("{1}"), TapCost(), RemoveCountersFromThis("wish", 1)),
			Condition: DuringYourTurn(),
			Purpose:   game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Effect:    wishclawTutorThenHandOver,
		}},
	})
}

// wishclawTutorThenHandOver searches, then gives the Talisman away.
func wishclawTutorThenHandOver(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	talisman := item.SourceCardID
	return SearchLibrary{
		Player:   ctx.Controller(),
		Dest:     game.ZoneHand,
		Limit:    1,
		Shuffle:  true,
		Source:   talisman,
		Reason:   "Wishclaw Talisman — search your library for a card",
		ThenTail: ctx.ItemTail(wishclawTalismanHandOver),
	}.Apply(ctx)
}

// wishclawTalismanHandOver is "an opponent gains control of Wishclaw
// Talisman": its controller chooses which. An on-disk key.
var wishclawTalismanHandOver = CardTail("wishclaw-talisman/hand-it-over", func(ctx *Context, o game.TailOutcome) error {
	g, item, controller, talisman := ctx.Game, ctx.Item, o.Tail.Player, o.Tail.Source
	g.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  controller,
		Among:    ctx.Opponents(),
		Source:   talisman,
		Question: "Wishclaw Talisman — choose an opponent to gain control of it",
		Item:     item,
		Then: func(g *game.Game, chosen uuid.UUID) error {
			if chosen == uuid.Nil || sourceIsNewObject(g, item) {
				return nil
			}
			g.GainControlForEffect(talisman, talisman, chosen, game.IndefiniteDuration(),
				"Wishclaw Talisman — an opponent gains control of this artifact")
			return nil
		},
	})
	return nil
})
