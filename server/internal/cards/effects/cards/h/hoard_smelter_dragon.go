package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hoard-Smelter Dragon — Creature — Dragon {4}{R}{R}, 5/5:
//
//	"Flying
//	 {3}{R}: Destroy target artifact. This creature gets +X/+0 until end of
//	 turn, where X is that artifact's mana value."
//
// The artifact's mana value is read as the ability resolves (it is the same
// number before and after the destroy). If the target is no longer legal the whole ability does
// nothing (CR 608.2b). The Dragon's own boost needs the Dragon still on the
// battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cf9db2be-3c95-45a2-ab85-694e3bd3bf1d",
		Name:            "Hoard-Smelter Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}: Destroy target artifact. This creature gets +X/+0 until end of turn, where X is that artifact's mana value.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerPump},
			Cost:    effects.ManaCost("{3}{R}"),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  hoardSmelterDragonSmelt,
		}},
	})
}

func hoardSmelterDragonSmelt(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var target game.TargetRef
	found := false
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			target, found = t, true
			break
		}
	}
	if !found {
		return nil
	}
	card, ok := g.LookupCardForEffect(target.ID)
	if !ok {
		return nil
	}
	x := card.ManaValue()
	// The two sentences are independent: the Dragon grows by the artifact's
	// mana value whether or not it is actually destroyed (an indestructible
	// artifact still gives its X). The mana value is read before either part,
	// so the boost needs no word about the destroy's outcome.
	if x > 0 && g.Battlefield.Contains(item.SourceCardID) {
		if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Label: "Hoard-Smelter Dragon — +X/+0"}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.DestroyTarget{Target: target.ID}.Apply(ctx)
}
