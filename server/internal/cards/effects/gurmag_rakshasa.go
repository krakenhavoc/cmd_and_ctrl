package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gurmag Rakshasa — Creature — Demon {4}{B}{B}, 5/5 (Tarkir: Dragonstorm,
// #4463):
//
//	"Menace
//	 When this creature enters, target creature an opponent controls gets
//	 -2/-2 until end of turn and target creature you control gets +2/+2
//	 until end of turn."
//
// Two clauses with different predicates (Gloom Ripper's shape), read back
// by slot. The Rakshasa itself can be the creature you control. Both
// targets are required, so with no creature an opponent controls the
// trigger is removed (CR 603.3d). Each half applies only to a target that
// is still legal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2be43acc-a815-44bc-8803-d702e553fbc1",
		Name:            "Gurmag Rakshasa",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WhenThisEnters("Gurmag Rakshasa — target creature an opponent controls gets -2/-2 and target creature you control gets +2/+2",
					gurmagRakshasaEnters),
				Clauses(
					TargetCreature("target creature an opponent controls", OpponentControls()),
					TargetCreature("target creature you control", YouControl()),
				)),
				ForTargets(HarmUntilEOT(0, -2, -2), BenefitUntilEOT(1, 2, 2))),
		},
	})
}

// gurmagRakshasaEnters shrinks the first target and pumps the second.
func gurmagRakshasaEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
		if err := (BoostUntilEOT{Target: t.ID, Power: -2, Toughness: -2, Label: "Gurmag Rakshasa — -2/-2"}).Apply(ctx); err != nil {
			return err
		}
	}
	if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
		return BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Gurmag Rakshasa — +2/+2"}.Apply(ctx)
	}
	return nil
}
