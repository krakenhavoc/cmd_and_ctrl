package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wasp, Shrinking Savior — Legendary Creature — Human Hero {2}{U}, 1/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever Wasp attacks, up to one other target creature gets -3/-0
//	 until your next turn. Then draw a card for each creature with power
//	 less than 0 on the battlefield."
//
// The -3/-0 lasts until the controller's next turn begins
// (DurationUntilYourNextTurn, CR 611.2b). "Then" is the second sentence,
// run after the shrink with the layers recomputed, so the target itself
// (a 2/2 made -1/2) and any other creature already below zero power count.
// With no target chosen nothing shrinks, but the draw still counts the
// board; a chosen target that is gone makes the whole ability do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "002cd9b4-efce-429f-82cd-d37352d2e932",
		Name:            "Wasp, Shrinking Savior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Wasp, Shrinking Savior — up to one other target creature gets -3/-0 until your next turn, then draw a card for each creature with power less than 0", waspShrinks),
				Another(UpToOneTargetCreature("up to one other target creature"))),
		},
	})
}

func waspShrinks(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok && len(item.Targets) > 0 {
		// Its only target is illegal: the ability does nothing (CR 608.2b).
		return nil
	}
	if ok {
		if err := (ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.ModifyPTMod(-3, 0)},
			Duration: DurationUntilYourNextTurn(ctx, item.Controller),
			Label:    "Wasp, Shrinking Savior — -3/-0 until your next turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	g.RecomputeLayersIfStaleLocked()
	negative := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.PowerForComparison() < 0 {
			negative++
		}
	}
	if negative == 0 {
		return nil
	}
	return DrawCards{Player: item.Controller, N: negative}.Apply(ctx)
}
