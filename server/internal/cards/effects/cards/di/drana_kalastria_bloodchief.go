package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drana, Kalastria Bloodchief — Legendary Creature — Vampire Shaman
// {3}{B}{B}, 4/4 (Commander 2017, #5533):
//
//	"Flying
//	 {X}{B}{B}: Target creature gets -0/-X until end of turn and Drana gets
//	 +X/+0 until end of turn."
//
// One X, announced with the activation (CR 107.3k) and read as it
// resolves. The target takes -0/-X as long as it is still a legal target
// (CR 608.2b); Drana gets +X/+0 whether or not it is (she is not a target),
// as long as she is still on the battlefield. Drana may target herself: she
// then gets both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "67391f99-739a-490a-ba16-7fa1d758ec34",
		Name:            "Drana, Kalastria Bloodchief",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		XMatters:        true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{B}{B}: Target creature gets -0/-X until end of turn and Drana gets +X/+0 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, -1)),
			Cost:    effects.ManaCost("{X}{B}{B}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				if t, ok := effects.FirstLegalTarget(ctx); ok && t.Kind == game.TargetCard {
					if err := (effects.BoostUntilEOT{Target: t.ID, Toughness: -x, Label: "Drana — -0/-X"}).Apply(ctx); err != nil {
						return err
					}
				}
				if !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Label: "Drana — +X/+0"}.Apply(ctx)
			},
		}},
	})
}
