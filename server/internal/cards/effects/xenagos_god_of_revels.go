package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Xenagos, God of Revels — Legendary Enchantment Creature — God
// {3}{R}{G}, 6/5:
//
//	"Indestructible
//	 As long as your devotion to red and green is less than seven,
//	 Xenagos isn't a creature.
//	 At the beginning of combat on your turn, another target creature
//	 you control gains haste and gets +X/+X until end of turn, where X
//	 is that creature's power."
//
// Indestructible is a printed keyword that holds whether or not
// Xenagos is a creature (a God that isn't a creature is still an
// indestructible enchantment). The God clause is the shared layer-4
// self-static (theros_gods.go) over devotion to TWO colours: CR 700.5
// counts each symbol that is red, green or both once, so a {R/G} hybrid
// is one, not two (devotionToColors). Xenagos's own {R}{G} count while
// it is on the battlefield, and the static is re-read on every layer
// pass, so it turns on and off with the board. It still works while
// Xenagos has lost all abilities, because notACreature runs in layer 4,
// before layer 6 can remove anything (the 2020-11-10 ruling).
//
// The trigger targets as it goes on the stack (CR 603.3d) with
// "another" as object identity (Another). X is read once, as it
// resolves (the 2014-02-01 ruling), unclamped, so a creature with
// negative power gets -X/-X; haste and the bonus are one effect at one
// timestamp.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb15a8dd-57fe-466f-847e-66476b690a1f",
		Name:            "Xenagos, God of Revels",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{godUnlessDevotionTo(7, "R", "G")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Xenagos, God of Revels — another creature you control gains haste and gets +X/+X",
					xenagosRevel),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}

// xenagosRevel is the combat trigger's body: +X/+X and haste on the
// target, X its power now.
func xenagosRevel(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	x := c.PowerForComparison()
	return untilEndOfTurn(ctx, id, nil, "Xenagos, God of Revels — haste and +X/+X",
		game.ModifyPTMod(x, x), game.AddKeywordsMod("haste"))
}
