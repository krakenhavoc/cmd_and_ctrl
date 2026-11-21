package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Impending(4, "{2}{W}{W}", effects.Spec{
		OracleID:     "7e64b1dc-a238-4bff-98ff-2bea44340568",
		Name:         "Overlord of the Mistmoors",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Overlord of the Mistmoors — create two 2/1 white Insect creature tokens with flying",
				effects.B34CreateTokens(func() game.Card { return effects.TokenCard("2/1 white Insect with flying") }, 2)),
		},
	}))
}
