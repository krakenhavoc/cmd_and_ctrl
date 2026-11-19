package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wild Hunger — Instant {2}{G} (Innistrad Remastered):
//
//	"Target creature gets +3/+1 and gains trample until end of turn.
//	 Flashback {3}{R}"
func init() {
	Register(Spec{
		OracleID:         "9d250088-d5f2-40ed-bdb7-a6decce0fb67",
		Name:             "Wild Hunger",
		Purpose:          ForTargets(BenefitUntilEOT(0, 3, 1, "trample")),
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 3, Toughness: 1, Label: "Wild Hunger — +3/+1"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Wild Hunger — trample"}.Apply(ctx)
		},
	})
}
