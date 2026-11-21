package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heliophial — Artifact {5}:
//
//	"Sunburst (This artifact enters with a charge counter on it for
//	 each color of mana spent to cast it.)
//	 {2}, Sacrifice this artifact: It deals damage equal to the number
//	 of charge counters on it to any target."
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552). The sacrifice is a cost, so the Heliophial is gone
// when the ability resolves: the count and the damage source are its
// last-known information (CR 608.2h, 113.7a).
func init() {
	effects.Register(effects.Spec{
		OracleID:            "6c3be36d-2461-4e38-9ab4-4394cd4996e7",
		Name:                "Heliophial",
		ReadsManaSpent:      game.ReadsOwnPayment,
		Completeness:        effects.CompletenessFull,
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice this artifact: It deals damage equal to the number of charge counters on it to any target.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Effect:  heliophialDamage,
		}},
	})
}

// heliophialDamage deals the sacrificed Heliophial's last-known charge
// counters as damage, from it, to the target.
func heliophialDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	n := info.Counters[game.CounterCharge]
	ts := ctx.LegalTargets()
	if n <= 0 || len(ts) == 0 {
		return nil
	}
	ref, _ := ctx.SourceRef()
	return effects.DealDamage{SourceObject: &ref, Target: ts[0].ID, Amount: n}.Apply(ctx)
}
