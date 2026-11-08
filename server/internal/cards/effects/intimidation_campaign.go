package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Intimidation Campaign — Enchantment {1}{U}{B} (Outlaws of Thunder
// Junction, #4468):
//
//	"When this enchantment enters, each opponent loses 1 life, you gain 1
//	 life, and you draw a card.
//	 Whenever you commit a crime, you may return this enchantment to its
//	 owner's hand."
//
// The enter trigger is the drain (EachOpponentLosesLife, GainLife) and a
// draw, as footprinted Do steps. The crime trigger (CR 700.13, Kaervek's
// YouCommittedCrime, every crime and not once a turn) is a CR 603.5 "you
// may" asked as it resolves; the bounce does nothing if the enchantment
// has already left or come back as a new object (CR 400.7). Recasting it
// drains and draws again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5915039b-3dbd-4a49-8bd3-23ff0fd9c90a",
		Name:         "Intimidation Campaign",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Intimidation Campaign — each opponent loses 1 life, you gain 1 life, and you draw a card",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1}, DrawCards{N: 1})),
			Optional(On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return YouCommittedCrime(ev, source)
			}, "Intimidation Campaign — return this enchantment to its owner's hand (crime)", returnThisPermanentToOwnersHand),
				"Intimidation Campaign — return this enchantment to its owner's hand?"),
		},
	})
}
