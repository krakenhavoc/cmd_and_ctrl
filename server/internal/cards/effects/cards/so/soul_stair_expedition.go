package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Stair Expedition — Enchantment {B} (Zendikar, S82 OLD b496):
//
//	"Landfall — Whenever a land you control enters, you may put a quest counter on
//	 this enchantment.
//	 Remove three quest counters from this enchantment and sacrifice it: Return up
//	 to two target creature cards from your graveyard to your hand."
//
// Khalni Heart Expedition's landfall counter and cost, with Baloth Null's
// up-to-two graveyard return (each target is checked again as it resolves, and
// the cost is paid before the targets are chosen).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24b11ede-26c8-40ed-b874-c0f6b21572c3",
		Name:         "Soul Stair Expedition",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Landfall("Soul Stair Expedition — put a quest counter on it",
				effects.Do(effects.CounterOnThis{Kind: "quest", N: 1})), "Put a quest counter on Soul Stair Expedition?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three quest counters from this enchantment and sacrifice it: Return up to two target creature cards from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost:    effects.Plus(effects.RemoveCountersFromThis("quest", 3), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 2),
			Effect:  effects.Do(effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand)),
		}},
	})
}
