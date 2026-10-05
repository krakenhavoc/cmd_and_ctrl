package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Muster the Departed — Enchantment {2}{W}:
//
//	"When this enchantment enters, create a 1/1 white Spirit creature
//	 token with flying.
//	 Morbid — At the beginning of your end step, if a creature died
//	 this turn, populate. (Create a token that's a copy of a creature
//	 token you control.)"
//
// The enters trigger is an ordinary token. The morbid clause is an
// intervening if (CR 603.4) on the table-wide tally of creatures that
// died this turn (b11CreaturesDiedThisTurn: any player's, tokens
// included), checked as the end step begins and again as the trigger
// resolves, like Deathreap Ritual's. The Spirit makes the populate
// worth having, since it is a creature token you control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0ba9cca-b946-465c-acae-0b465d7e1622",
		Name:         "Muster the Departed",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Muster the Departed — create a 1/1 white Spirit with flying",
				Do(CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 1})),
			On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b11CreaturesDiedThisTurn(g) > 0
			}), "Muster the Departed — morbid, populate", func(g *game.Game, item *game.StackItem) error {
				if b11CreaturesDiedThisTurn(g) == 0 {
					return nil
				}
				return Populate{}.Apply(NewContext(g, item))
			}),
		},
	})
}
