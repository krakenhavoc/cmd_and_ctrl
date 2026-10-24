package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mezzio Mugger — Creature — Lizard Rogue {4}{R}, 3/3:
//
//	"Whenever this creature attacks, exile the top card of each player's
//	 library. You may play those cards this turn, and you may spend mana
//	 as though it were mana of any color to cast those spells.
//	 Blitz {2}{R}"
//
// Blitz (CR 702.152, #3772). The attack trigger is Etali's walk over the
// seats with the impulse grant ExileTopWithPermission builds: "play", so
// a land is playable (CastOnly false), until end of turn, any colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "068699a9-dd5c-428b-96eb-fd085dda71e9",
		Name:             "Mezzio Mugger",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Blitz("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Mezzio Mugger — exile the top card of each player's library",
				mezzioMuggerExile),
		},
	})
}

// mezzioMuggerExile exiles each player's top card with a play-this-turn
// grant to the trigger's controller.
func mezzioMuggerExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	// One library at a time in APNAP order (#4266): a CR 616.1 prompt one
	// exile asks is the only one open, and the next library waits for it.
	return exileTopEachPlayerInTurn(ctx, apnapPlayers(g),
		ExileTopWithPermission{GrantTo: item.Controller, N: 1, AnyColor: true}, nil)
}
