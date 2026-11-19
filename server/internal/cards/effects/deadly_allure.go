package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadly Allure — Sorcery {B} (Innistrad Remastered):
//
//	"Target creature gains deathtouch until end of turn and must be blocked
//	 this turn if able.
//	 Flashback {G}"
func init() {
	Register(Spec{
		OracleID:         "f9cdaecd-9621-4a9a-888e-748ecfe551fd",
		Name:             "Deadly Allure",
		Purpose:          ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch")),
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{G}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Deadly Allure — deathtouch"}).Apply(ctx); err != nil {
				return err
			}
			return BlockRequirementUntilEOT{
				Target: id,
				Kind:   game.BlockRequirementMustBeBlocked,
				Label:  "Deadly Allure — must be blocked this turn",
			}.Apply(ctx)
		},
	})
}
