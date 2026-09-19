package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mishra's Command — Sorcery {X}{R}:
//
//	"Choose two —
//	 • Choose target player. They may discard up to X cards. Then
//	   they draw a card for each card discarded this way.
//	 • This spell deals X damage to target creature.
//	 • This spell deals X damage to target planeswalker.
//	 • Target creature gets +X/+0 and gains haste until end of turn."
//
// Four X-scaled bullets, choose two, three of which target — the
// per-mode target slots #764 shipped (ADR 0065 §3, Kolaghan's
// Command's pattern): each chosen bullet gets its own target group,
// picked in the order chosen, and resolved in that same order (CR
// 700.2c). Choosing "deal X damage to target creature" and "target
// creature gets +X/+0 and gains haste" announces two creature
// targets that may be the same creature or different ones, and each
// bullet re-checks ITS OWN target at resolution (CR 608.2b).
//
// The first bullet is a "discard, THEN count what was discarded"
// wheel, same shape as Syphon Mind's fan-out (#1027,
// PlayerDiscardsThenForEffect): the discard is a real up-to-X choice
// by the targeted player, and the draw is the RUN's continuation, so
// it happens once, after the discard has actually landed, and for
// exactly as many cards as were really discarded (CR 701.8a) — never
// before the player has chosen, which would let the draw happen
// before there was anything to count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "25434ea3-bcfa-4dae-a16e-10ab87ce32af",
		Name:         "Mishra's Command",
		XMatters:     true,
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeDoing("Choose target player. They may discard up to X cards. "+
				"Then they draw a card for each card discarded this way.",
				TargetPlayer("target player"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					target := t.ID
					return ctx.Game.PlayerDiscardsThenForEffect(
						game.DiscardPrompt{
							Player:   target,
							Source:   item.SourceCardID,
							N:        ctx.X(),
							UpTo:     true,
							Question: "Mishra's Command — discard up to X cards",
						},
						func(g *game.Game, discarded game.PromptedDiscards) error {
							return g.DrawNForEffect(target, discarded.Count())
						})
				}),
			ModeDoing("This spell deals X damage to target creature.",
				TargetCreature("target creature"),
				DealXDamageToModesTarget),
			ModeDoing("This spell deals X damage to target planeswalker.",
				TargetPermanent("target planeswalker", Planeswalker()),
				DealXDamageToModesTarget),
			ModeDoing("Target creature gets +X/+0 and gains haste until end of turn.",
				TargetCreature("target creature"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (BoostUntilEOT{Target: t.ID, Power: ctx.X(), Label: "Mishra's Command"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"}, Label: "Mishra's Command"}.Apply(ctx)
				}),
		),
	})
}
