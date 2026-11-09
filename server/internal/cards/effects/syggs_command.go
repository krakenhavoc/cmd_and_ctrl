package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sygg's Command — Kindred Sorcery — Merfolk {1}{W}{U} (Lorwyn Eclipsed, #4458):
//
//	"Choose two —
//	 • Create a token that's a copy of target Merfolk you control.
//	 • Creatures target player controls gain lifelink until end of turn.
//	 • Target player draws a card.
//	 • Tap target creature. Put a stun counter on it."
//
// Each bullet carries its own target group, and with at least one still
// legal the spell does what it can (the ruling). The Merfolk is any permanent
// you control with the type, a Kindred Merfolk included; X in its cost is 0.
func init() {
	Register(Spec{
		OracleID:     "7fe5baab-b182-4cd9-bea5-912ce9b1464a",
		Name:         "Sygg's Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeDoing("Create a token that's a copy of target Merfolk you control.",
				TargetPermanent("target Merfolk you control", YouControl(), HasSubtype("Merfolk")),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return CreateTokenCopy{Copy: t.ID, N: 1}.Apply(ctx)
				}),
			ModeDoing("Creatures target player controls gain lifelink until end of turn.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return GrantKeywordUntilEOT{
						Match:    And(Creature(), ControlledBy(t.ID)),
						Keywords: []string{"lifelink"},
						Label:    "Sygg's Command — lifelink until end of turn",
					}.Apply(ctx)
				}),
			ModeDoing("Target player draws a card.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return DrawCards{Player: t.ID, N: 1}.Apply(ctx)
				}),
			ModeDoing("Tap target creature. Put a stun counter on it.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}.Apply(ctx)
				}),
		),
	})
}
