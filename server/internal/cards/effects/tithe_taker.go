package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c916b81f-700e-4a1a-8c8e-c4685ceaecd2",
		Name:         "Tithe Taker",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsMore(1, "During your turn, spells your opponents cast cost {1} more to cast.",
				OpponentsSpell(), cmdr2B03DuringSourcesControllersTurn()),
			ActivationCostsMore(1, "During your turn, abilities your opponents activate cost {1} more to activate unless they're mana abilities.",
				OpponentsSpell(), cmdr2B03DuringSourcesControllersTurn(), cmdr2B03NotAManaAbility()),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Tithe Taker — afterlife 1: create a 1/1 white and black Spirit creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 white and black Spirit with flying"), N: 1})),
		},
	})
}
