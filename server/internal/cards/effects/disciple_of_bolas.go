package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disciple of Bolas — Creature — Human Wizard {3}{B}, 2/2 (issue
// #1117):
//
//	"When this creature enters, sacrifice another creature. You gain
//	 X life and draw X cards, where X is that creature's power."
//
// The sacrifice is chosen AT RESOLUTION (the Harrow / Springbloom
// Druid shape, ADR 0013 §5x): the trigger's Effect opens a sacrifice
// prompt over the controller's OTHER creatures and the payout runs as
// its continuation, so the creature really sacrificed — not
// whichever one happened to be biggest when the trigger was built —
// is the one whose power counts. With no other creature to sacrifice
// the prompt never goes up, the continuation still runs (a
// continuation is the rest of a card that is paused mid-resolution,
// never silently skipped), and it gains no life and draws no cards —
// exactly the printed ability doing nothing, not an error.
//
// Sandbox simplification, declared — the same one Jarad, Golgari
// Lich Lord and Greater Good already carry: the sacrificed creature's
// power is its printed power plus its +1/+1 and -1/-1 counters as
// they were when it left. A bonus from another permanent's static
// ability (an anthem) is not in it, because there is no last-known
// characteristic to read for a creature that left as a COST rather
// than dying to an effect.
func init() {
	Register(Spec{
		OracleID:     "8f2c8498-5b10-4ec7-8678-598c90987556",
		Name:         "Disciple of Bolas",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The sacrificed creature's power counts its +1/+1 and -1/-1 counters but not a bonus from another permanent, such as an anthem.",
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Disciple of Bolas — sacrifice another creature; gain life and draw cards equal to its power",
				discipleOfBolasSacrificeThenPayout),
		},
	})
}

// discipleOfBolasSacrificeThenPayout is Disciple of Bolas's whole
// ETB ability: a sacrifice prompt over the controller's other
// creatures, and the life-gain-and-draw as its continuation.
//
// Caller holds g.mu (resolution frame).
func discipleOfBolasSacrificeThenPayout(g *game.Game, item *game.StackItem) error {
	return g.PlayerSacrificesThenForEffect(item.SourceCardID, item.Controller,
		sacrificeSpec("another creature", Creature(), NotSelf(item.SourceCardID)),
		"Disciple of Bolas — sacrifice another creature", 1,
		func(g *game.Game, sacrificed game.PromptedSacrifices) error {
			ids := sacrificed.By(item.Controller)
			if len(ids) == 0 {
				// No other creature to sacrifice — the ability does
				// as much as it can, which is nothing.
				return nil
			}
			power := b17LastKnownPowerOffBattlefield(g, ids[0])
			if power <= 0 {
				return nil
			}
			ctx := NewContext(g, item)
			if err := (GainLife{Player: item.Controller, Amount: power}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: power}.Apply(ctx)
		})
}
