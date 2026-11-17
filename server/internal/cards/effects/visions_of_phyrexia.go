package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Visions of Phyrexia — Enchantment {2}{R}{R} (The Brothers' War):
//
//	"At the beginning of your upkeep, exile the top card of your library.
//	 You may play that card this turn.
//	 At the beginning of your end step, if you didn't play a card from
//	 exile this turn, create a tapped Powerstone token."
//
// The upkeep ability is Outpost Siege's impulse exile. The end-step "if"
// reads a per-turn fact, the cards the controller played from exile
// (PlayerTurnTally.CardsPlayedFromExile: a spell cast from exile or a
// land played from exile, whichever card and whichever source put it
// there), and is checked as the step begins and again as the ability
// resolves (CR 603.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "634b46a6-ec8b-41f0-8eb6-1e9144035847",
		Name:         "Visions of Phyrexia",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Visions of Phyrexia — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1}.Apply(NewContext(g, item))
				}),
			AtYourEndStepIfNoCardPlayedFromExile("Visions of Phyrexia — create a tapped Powerstone token",
				Do(CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1})),
		},
	})
}
