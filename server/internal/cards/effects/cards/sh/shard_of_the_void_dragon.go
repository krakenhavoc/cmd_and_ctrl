package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shard of the Void Dragon — Creature — C'tan {4}{B}{B}{B}, 7/7 (Warhammer
// 40,000 Commander, #5948):
//
//	"Flying
//	 Spear of the Void Dragon — Whenever this creature attacks, each opponent
//	 sacrifices a nonland permanent of their choice.
//	 Matter Absorption — Whenever an artifact is put into a graveyard from the
//	 battlefield or is put into exile from the battlefield, put two +1/+1
//	 counters on this creature."
//
// Butcher of Malakir's edict (each opponent chooses their own nonland
// permanent) on attack, and Disciple of the Vault's artifact-to-graveyard
// condition widened to exile, anyone's artifacts (the Dragon's own included,
// tokens included). One trigger per artifact that leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d04d2434-121e-4420-abfc-fa21ad5577c3",
		Name:            "Shard of the Void Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WheneverThisAttacks("Shard of the Void Dragon — each opponent sacrifices a nonland permanent",
				effects.Do(effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Nonland(), Label: "a nonland permanent"})),
				game.Purpose{Answers: game.AnswerRemove}),
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.GapB189ArtifactLeftBattlefieldToGraveyardOrExile(ev, g)
			}, "Shard of the Void Dragon — put two +1/+1 counters on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
