package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treeguard Duo — Creature — Frog Rabbit {3}{G}, 3/4 (Bloomburrow,
// #4467):
//
//	"When this creature enters, until end of turn, target creature you
//	 control gains vigilance and gets +X/+X, where X is the number of
//	 creatures you control."
//
// X is Frontline Rush's count, the creatures you control as the trigger
// resolves, the Duo included while it is still there; it is read once
// and does not change later in the turn (CR 608.2h). The Duo may target
// itself. A target gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7dfb4d5-8510-4904-abb7-834ec65d9bf3",
		Name:         "Treeguard Duo",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters(
				"Treeguard Duo — target creature you control gains vigilance and gets +X/+X until end of turn",
				treeguardDuoPump),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "vigilance"))),
		},
	})
}

// treeguardDuoPump gives the still-legal target vigilance and +X/+X,
// with X the creatures its controller controls now.
func treeguardDuoPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	x := effects.CreaturesControlledBy(g, item.Controller)
	if err := (effects.BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Treeguard Duo — +X/+X"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"}, Label: "Treeguard Duo — vigilance"}.Apply(ctx)
}
