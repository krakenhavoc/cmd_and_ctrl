package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kessig Wolf Run — Land:
//
//	"{T}: Add {C}.
//	 {X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until
//	 end of turn."
//
// X is announced with the activation (CR 107.3k) and read back with
// ctx.X(), the way Deepwood Elder's is. Both halves are until-end-of-turn
// effects on the one target, applied at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6911265-54ef-4c16-bcf2-1ffb24b7d426",
		Name:         "Kessig Wolf Run",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until end of turn",
			Cost:    Plus(ManaCost("{X}{R}{G}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, target := range ctx.LegalTargets() {
					if target.Kind != game.TargetCard {
						continue
					}
					if err := (BoostUntilEOT{
						Target: target.ID,
						Power:  ctx.X(),
						Label:  "Kessig Wolf Run — +X/+0",
					}).Apply(ctx); err != nil {
						return err
					}
					return (GrantKeywordUntilEOT{
						Target:   target.ID,
						Keywords: []string{"trample"},
						Label:    "Kessig Wolf Run — trample",
					}).Apply(ctx)
				}
				return nil
			},
		}},
	})
}
