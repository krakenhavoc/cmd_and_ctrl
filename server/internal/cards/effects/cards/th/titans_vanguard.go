package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Titans' Vanguard — Creature — Eldrazi {3}{R}{G}, 5/5 (Modern Horizons 3):
//
//	"Devoid
//	 When you cast this spell and whenever this creature attacks, put a +1/+1
//	 counter on each colorless creature you control.
//	 Trample"
//
// "Cast and attacks" is one printed ability; it is registered as the two rows
// the engine needs (stack and battlefield), sharing one effect.
func init() {
	const label = "Titans' Vanguard — put a +1/+1 counter on each colorless creature you control"
	effects.Register(effects.Spec{
		OracleID:        "3da6d24b-ecbd-4a41-a37e-84a19b243414",
		Name:            "Titans' Vanguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell(label, effects.GapB52CounterOnEachColorlessCreature),
			effects.WheneverThisAttacks(label, effects.GapB52CounterOnEachColorlessCreature),
		},
	})
}
