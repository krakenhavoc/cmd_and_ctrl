package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rootwise Survivor — Creature — Human Survivor {3}{G}{G}, 3/4
// (Duskmourn: House of Horror, #4466):
//
//	"Haste
//	 Survival — At the beginning of your second main phase, if this
//	 creature is tapped, put three +1/+1 counters on up to one target land
//	 you control. That land becomes a 0/0 Elemental creature in addition to
//	 its other types. It gains haste until your next turn."
//
// Survival is Defiant Survivor's trigger (an intervening "if" asked as the
// phase begins and again as the ability resolves, the source read as it
// last existed if it has gone, CR 603.4, CR 608.2h). The clause is "up to
// one target land you control", so it may have no target; a land that left
// or changed hands by resolution does nothing (CR 608.2b).
//
// Three steps in the printed order, each waiting on the one before
// (CR 608.2c, #4155): the counters go on while the land is still not a
// creature, through the CR 614 placement window (so Doubling Season counts
// it as a permanent you control and Hardened Scales does not count it);
// then the land becomes a 0/0 Elemental creature in addition to its other
// types, with no duration (CR 611.2a), pinned to that object so it ends if
// the land leaves (CR 400.7) as awaken's animation does; then it gains
// haste until your next turn (CR 611.2b), a separate effect with its own
// duration, so the haste ends while the land stays a creature, a 3/3 with
// its counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e42a6880-7ca6-498d-abc9-fb7ccd08b111",
		Name:            "Rootwise Survivor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Survival("Rootwise Survivor — three +1/+1 counters on up to one target land; it becomes a 0/0 Elemental creature with haste until your next turn",
					Do(
						Instruction(func(ctx *Context) error {
							if id, ok := b16FirstLegalTargetCard(ctx); ok {
								return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
							}
							return nil
						}),
						Instruction(func(ctx *Context) error {
							id, ok := b16FirstLegalTargetCard(ctx)
							if !ok {
								return nil
							}
							mods := append([]game.Mod{game.AddTypesMod("Creature"), game.AddSubtypesMod(ctx.TextSubtype("Elemental"))},
								game.SetBasePTMods(0, 0)...)
							if err := (ScopedEffectFor{
								Target:   id,
								Mods:     mods,
								Duration: game.IndefiniteDuration(),
								Label:    "Rootwise Survivor — becomes a 0/0 Elemental creature in addition to its other types",
							}).Apply(ctx); err != nil {
								return err
							}
							return ScopedEffectFor{
								Target:   id,
								Mods:     []game.Mod{game.AddKeywordsMod("haste")},
								Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
								Label:    "Rootwise Survivor — haste until your next turn",
							}.Apply(ctx)
						}),
					)),
				TargetPermanent("up to one target land you control", And(Land(), YouControl())).WithCount(0, 1),
			),
		},
	})
}
