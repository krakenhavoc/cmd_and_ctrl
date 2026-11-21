package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tithe Taker — Creature {1}{W}, 2/1:
//
//	"During your turn, spells your opponents cast cost {1} more to cast and
//	 abilities your opponents activate cost {1} more to activate unless
//	 they're mana abilities.
//	 Afterlife 1 (When this creature dies, create a 1/1 white and black
//	 Spirit creature token with flying.)"
//
// Afterlife has no helper: it is the dies trigger it reminds (CR 702.135a).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c916b81f-700e-4a1a-8c8e-c4685ceaecd2",
		Name:         "Tithe Taker",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(1, "During your turn, spells your opponents cast cost {1} more to cast.",
				effects.OpponentsSpell(), effects.Cmdr2B03DuringSourcesControllersTurn()),
			effects.ActivationCostsMore(1, "During your turn, abilities your opponents activate cost {1} more to activate unless they're mana abilities.",
				effects.OpponentsSpell(), effects.Cmdr2B03DuringSourcesControllersTurn(), effects.Cmdr2B03NotAManaAbility()),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Tithe Taker — afterlife 1: create a 1/1 white and black Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white and black Spirit with flying"), N: 1})),
		},
	})
}
