package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Balrog, Durin's Bane — Legendary Creature — Avatar Demon {5}{B}{R},
// 7/5 (The Lord of the Rings: Tales of Middle-earth):
//
//	"This spell costs {1} less to cast for each permanent sacrificed this
//	 turn.
//	 Haste
//	 The Balrog can't be blocked except by legendary creatures.
//	 When The Balrog dies, destroy target artifact or creature an opponent
//	 controls."
//
// The discount is Bedlam Reveler's self cost modifier over the caster's
// per-turn sacrifice tally (the count Pyre-Sledge Arsonist reads), which
// counts every permanent the caster sacrificed this turn, tokens included.
// The block rule is Deluxe Dragster's shape with the allowed set Legendary.
// The dies trigger targets an artifact or creature an opponent controls and
// destroys it as it resolves, if it is still a legal target (CR 608.2b).
//
// No simplification.
func init() {
	dies := effects.Targeting(
		effects.WhenThisDies("The Balrog, Durin's Bane — destroy target artifact or creature an opponent controls", effects.DestroyTheFirstLegalTarget),
		effects.TargetPermanent("target artifact or creature an opponent controls", effects.Or(effects.Artifact(), effects.Creature()), effects.OpponentControls()),
	)
	effects.Register(effects.Spec{
		OracleID:        "d113af33-d75d-44e4-a8dc-31f8984edc42",
		Name:            "The Balrog, Durin's Bane",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return q.Game.TurnTallyFor(q.Controller).PermanentsSacrificed
			}, "This spell costs {1} less to cast for each permanent sacrificed this turn."),
		},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.Legendary(), "legendary creatures"),
		},
		Triggered: []game.TriggeredAbility{dies},
	})
}
