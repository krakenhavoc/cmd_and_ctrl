package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d7dfb4d5-8510-4904-abb7-834ec65d9bf3",
		Name:         "Treeguard Duo",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters(
				"Treeguard Duo — target creature you control gains vigilance and gets +X/+X until end of turn",
				treeguardDuoPump),
				TargetCreature("target creature you control", YouControl())),
				ForTargets(BenefitUntilEOT(0, 0, 0, "vigilance"))),
		},
	})
}

// treeguardDuoPump gives the still-legal target vigilance and +X/+X,
// with X the creatures its controller controls now.
func treeguardDuoPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	x := creaturesControlledBy(g, item.Controller)
	if err := (BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Treeguard Duo — +X/+X"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"}, Label: "Treeguard Duo — vigilance"}.Apply(ctx)
}
