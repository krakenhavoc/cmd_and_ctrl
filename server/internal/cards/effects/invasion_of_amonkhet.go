package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Amonkhet // Lazotep Convert — Battle — Siege, defense 4, for
// {1}{U}{B} (March of the Machine, S82 batch 23):
//
//	"(As a Siege enters, choose an opponent to protect it. You and others
//	  can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, each player mills three cards, then each
//	 opponent discards a card and you draw a card."
//
// Three sequential instructions of one resolution (CR 608.2c): the mills,
// one library at a time in APNAP order; then each opponent chooses a card
// of their own hand to discard (an empty hand is skipped); then the
// controller draws. A prompt any step asks holds the later ones.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfAmonkhetOracleID,
		Name:         "Invasion of Amonkhet",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 4,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Amonkhet — each player mills three cards, then each opponent discards a card and you draw a card", Do(
				Instruction(func(ctx *Context) error {
					return millEachPlayerInTurn(ctx, apnapPlayers(ctx.Game), constantMill(3), game.KeyedTail{})
				}),
				Instruction(func(ctx *Context) error { return eachOpponentDiscardsOne(ctx.Game, ctx.Item) }),
				DrawCards{N: 1},
			)),
			DefeatedTrigger("Invasion of Amonkhet — defeated: exile it, then cast Lazotep Convert", SiegeDefeated()),
		},
	})
}

// invasionOfAmonkhetOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfAmonkhetOracleID = "fd9b8889-a78a-4cda-9a6b-39f5536c3a0b"
