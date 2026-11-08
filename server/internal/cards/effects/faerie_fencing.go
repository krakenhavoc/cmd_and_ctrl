package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faerie Fencing — Instant {X}{B} (Wilds of Eldraine, #4471):
//
//	"Target creature gets -X/-X until end of turn. That creature gets an
//	 additional -3/-3 until end of turn if you controlled a Faerie as you
//	 cast this spell."
//
// Slice from the Shadows' -X/-X on the still-legal target (CR 608.2b), X
// the announced X (CR 107.3a). "As you cast this spell" is a cast-time
// value (Spec.CastTimeValues, Lifestream's Blessing's): whether you
// controlled a Faerie is recorded once every cost is paid (CR 601.2i), so
// a Faerie that leaves in response still gives the extra -3/-3 and one
// that arrives in response does not. "Faerie" is a word of the card's
// text (CR 612.2). The two shrinks are applied as one -(X+3)/-(X+3),
// which is the same total until end of turn. A game restored from before
// the value existed reads the Faerie as the spell resolves.
//
// No simplification.
const faerieFencingControlledAFaerie = "controlled_a_faerie"

func init() {
	Register(Spec{
		OracleID:     "450a9b73-9024-4103-b9f9-216a2f395cff",
		Name:         "Faerie Fencing",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      ForTargets(HarmUntilEOT(0, 0, 0)),
		Targets:      TargetCreature("target creature"),
		CastTimeValues: []game.CastTimeValue{{
			Key: faerieFencingControlledAFaerie,
			Value: func(g *game.Game, item *game.StackItem) int {
				return faerieFencingFaeries(NewContext(g, item))
			},
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := ctx.X()
			faeries, ok := ctx.CastValue(faerieFencingControlledAFaerie)
			if !ok {
				faeries = faerieFencingFaeries(ctx)
			}
			if faeries > 0 {
				n += 3
			}
			if n == 0 {
				return nil
			}
			return vsPumpTheTarget(ctx, -n, -n, "Faerie Fencing — -X/-X")
		},
	})
}

// faerieFencingFaeries is how many Faeries the spell's controller
// controls right now.
func faerieFencingFaeries(ctx *Context) int {
	return len(MatchingBattlefield(ctx, And(YouControl(), HasSubtype(ctx.TextSubtype("Faerie")))))
}
