package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lifecreed Duo — Creature — Bat Bird {1}{W}, 1/2 (EDHREC rank 3796):
//
//	"Flying
//	 Whenever another creature you control enters, you gain 1 life."
//
// The Soul Warden that flies. Flying rides PrintedKeywords; the
// trigger is "another creature you control enters" — a token, a
// reanimated creature and a land something animated before it
// entered all count; the Duo's own entry does not, as printed. One
// trigger per creature, so a batch of tokens gains one life each.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "63686c6b-9051-4002-aa8e-da8a3021330f",
		Name:            "Lifecreed Duo",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.CreatureEnterPaying(effects.WheneverAnotherCreatureEntersUnderYourControl("Lifecreed Duo — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
				game.CreatureEnterPayoff{LifeGain: 1}),
		},
	})
}
