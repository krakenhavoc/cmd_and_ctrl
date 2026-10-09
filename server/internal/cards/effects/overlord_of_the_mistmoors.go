package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overlord of the Mistmoors — Enchantment Creature — Avatar Horror
// {5}{W}{W}, 6/6:
//
//	"Impending 4—{2}{W}{W} (If you cast this spell for its impending
//	 cost, it enters with four time counters and isn't a creature
//	 until the last is removed. At the beginning of your end step,
//	 remove a time counter from it.)
//	 Whenever this permanent enters or attacks, create two 2/1 white
//	 Insect creature tokens with flying."
//
// Impending is the shared constructor; the card adds the trigger. While
// impending it is not a creature and cannot attack, but the enter half
// fires, which is the point of paying {2}{W}{W}.
//
// No simplification.
func init() {
	Register(Impending(4, "{2}{W}{W}", Spec{
		OracleID:     "7e64b1dc-a238-4bff-98ff-2bea44340568",
		Name:         "Overlord of the Mistmoors",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Overlord of the Mistmoors — create two 2/1 white Insect creature tokens with flying",
				b34CreateTokens(func() game.Card { return TokenCard("2/1 white Insect with flying") }, 2)),
		},
	}))
}
