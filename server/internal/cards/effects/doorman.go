package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doorman — Legendary Creature — Mutant Hero {1}{G}, 2/3 (Marvel Super
// Heroes Commander, #4455):
//
//	"{T}: Until end of turn, target creature can't be blocked by
//	 creatures with power 2 or less and/or Walls."
//
// Two block restrictions on the one target, both ScopedEffects that end
// with the turn and are read live as blockers are declared (CR 509.1b):
// Rhonas's Stalwart's power bar and Kick in the Door's Wall bar. "And/or"
// is read as both bars at once, so a blocker barred by either one cannot
// block (a Wall with power 5 and a 2-power Bear are both refused).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01ea2130-e212-4532-b68b-2f940d53ef6e",
		Name:         "Doorman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Until end of turn, target creature can't be blocked by creatures with power 2 or less and/or Walls.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (CantBeBlockedThisTurnByPower{Target: t.ID, AtMost: 2, Label: "Doorman — can't be blocked by power 2 or less"}).Apply(ctx); err != nil {
						return err
					}
					return CantBeBlockedThisTurnBySubtypes{Target: t.ID, Subtypes: []string{ctx.TextSubtype("Wall")}, Text: "Walls", Label: "Doorman — can't be blocked by Walls"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
