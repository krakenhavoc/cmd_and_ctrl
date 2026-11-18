package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dormant Grove // Gnarled Grovestrider — {3}{G} Enchantment // Creature —
// Treefolk 3/6 (Innistrad: Crimson Vow):
//
//	Front: "At the beginning of combat on your turn, put a +1/+1 counter on
//	        target creature you control. Then if that creature has toughness
//	        6 or greater, transform this enchantment."
//	Back:  "Vigilance
//	        Other creatures you control have vigilance."
//
// The toughness is read through the layers after the counter lands, so the
// counter itself can be what takes the creature to 6. If the target is
// illegal as the trigger resolves it does nothing (CR 608.2b).
//
// No simplification.
func init() {
	const oracle = "1be58157-186e-48e4-b375-dc5f10d785ac"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Dormant Grove",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Dormant Grove — put a +1/+1 counter on target creature you control; then if it has toughness 6 or greater, transform this enchantment",
					dormantGroveBody),
				TargetCreature("target creature you control", YouControl())),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Gnarled Grovestrider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Others: true, YoursOnly: true}, "vigilance"),
		},
	})
}

func dormantGroveBody(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		toughness := 0
		g.RecomputeLayersIfStaleLocked() // the item resolves under the game lock
		if c, ok := g.LookupCardForEffect(t.ID); ok && onBattlefield(g, t.ID) {
			toughness = c.CurrentToughness()
		}
		if toughness >= 6 {
			return TransformThis{}.Apply(ctx)
		}
		return nil
	}
	return nil
}
