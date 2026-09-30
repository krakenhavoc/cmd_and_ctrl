package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Torgaar, Famine Incarnate — Legendary Creature — Avatar {6}{B}{B},
// 7/6:
//
//	"As an additional cost to cast this spell, you may sacrifice any
//	 number of creatures. This spell costs {2} less to cast for each
//	 creature sacrificed this way.
//	 When Torgaar enters, up to one target player's life total becomes
//	 half their starting life total, rounded down."
//
// The variable sacrifice (ADR 0100 §3) with a per-sacrifice discount:
// the count is announced at CR 601.2b, before 601.2f totals the cost,
// so CostsLessPerSacrificed reads it through the one pricer — the
// auto-tap preview, the bot and CastSpell all charge {B}{B} for a
// Torgaar paid with three creatures. The discount spends generic mana
// only, so a fourth creature buys nothing.
//
// "Becomes" is a life change of whatever size closes the gap (CR
// 119.5): a player below half gains, a player above loses. The starting
// total is the table's (40 in Commander, ADR 0075), so the ETB sets a
// player to 20.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "4229140f-fa5b-4727-a16a-cbbd756d979e",
		Name:           "Torgaar, Famine Incarnate",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeAnyNumberCost("any number of creatures", Creature()),
		SelfCostModifiers: []game.CostModifier{
			CostsLessPerSacrificed("{2}", "This spell costs {2} less to cast for each creature sacrificed this way"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Torgaar, Famine Incarnate — up to one target player's life total becomes half their starting life total", torgaarHalvesALifeTotal),
				TargetPlayer("up to one target player").WithCount(0, 1),
			),
		},
	})
}

// torgaarHalvesALifeTotal sets the chosen player's life total to half
// the table's starting total, rounded down. No target chosen, or one
// that has left the game, does nothing.
func torgaarHalvesALifeTotal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return b31LifeBecomes(ctx, t.ID, startingLifeOf(g)/2)
		}
	}
	return nil
}
