package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "01ea2130-e212-4532-b68b-2f940d53ef6e",
		Name:         "Doorman",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Until end of turn, target creature can't be blocked by creatures with power 2 or less and/or Walls.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.CantBeBlockedThisTurnByPower{Target: t.ID, AtMost: 2, Label: "Doorman — can't be blocked by power 2 or less"}).Apply(ctx); err != nil {
						return err
					}
					return effects.CantBeBlockedThisTurnBySubtypes{Target: t.ID, Subtypes: []string{ctx.TextSubtype("Wall")}, Text: "Walls", Label: "Doorman — can't be blocked by Walls"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
