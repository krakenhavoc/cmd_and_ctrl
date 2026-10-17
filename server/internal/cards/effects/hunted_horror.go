package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunted Horror — Creature — Horror {B}{B}, 7/7 (#4092):
//
//	"Trample
//	 When this creature enters, target opponent creates two 3/3 green
//	 Centaur creature tokens with protection from black."
//
// A targeted ETB trigger whose tokens go to the target: the opponent
// creates them, so the opponent controls and owns them (CR 111.2). The
// protection is the enforced "protection from black" token (ADR 0072),
// so the Centaurs can block the Horror and take no damage from it, and
// can't be targeted by black removal. An opponent who has left by the
// time the trigger resolves gets nothing, and the trigger does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3469463e-ed4d-44bb-81f3-42bbe09d34ad",
		Name:            "Hunted Horror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Hunted Horror — target opponent creates two 3/3 green Centaurs with protection from black", huntedHorrorCentaurs),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

func huntedHorrorCentaurs(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return CreateToken{
			Controller: t.ID,
			Template:   TokenCard("3/3 green Centaur with protection from black"),
			N:          2,
		}.Apply(ctx)
	}
	return nil
}
