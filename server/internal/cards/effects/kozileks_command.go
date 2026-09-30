package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kozilek's Command — Kindred Instant — Eldrazi {X}{C}{C} (EDHREC
// rank 2544):
//
//	"Choose two —
//	 • Target player creates X 0/1 colorless Eldrazi Spawn creature
//	   tokens with "Sacrifice this token: Add {C}."
//	 • Target player scries X, then draws a card.
//	 • Exile target creature with mana value X or less.
//	 • Exile up to X target cards from graveyards."
//
// Mishra's Command's exact shape one mana symbol over: choose two of
// four X-scaled bullets, each with its own target group (#764,
// ModeDoing), resolved in announce order (CR 608.2c). The first
// bullet reuses the printed Eldrazi Spawn template; the second rides
// Scry's Then the same way Preordain does; the third narrows its
// battlefield clause with WithManaValueAtMostX, the announced-X
// binding Agadeem's Awakening uses for its graveyard clause.
//
// Declared simplification on the fourth bullet only: "up to X target
// cards from graveyards" is CountFromX, which is EXACTLY X rather than
// a real "up to" — the same declared gap as Crackle with Power, for
// the same reason (no "up to X" target-count shape exists yet). A
// caster who wants fewer than X exiled announces a smaller X —
// weaker than printed, never stronger. The other three bullets have
// no such gap.
func init() {
	Register(Spec{
		OracleID:     "2f2c549d-0293-4b9d-b9a8-ff392800f3a5",
		Name:         "Kozilek's Command",
		XMatters:     true,
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The graveyard-exile mode exiles exactly X cards rather than up to X."},
		Modes: ChooseN("Choose two", 2, 2,
			ModeDoing("Target player creates X 0/1 colorless Eldrazi Spawn creature tokens "+
				"with \"Sacrifice this token: Add {C}.\"",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return CreateToken{Controller: t.ID, Template: EldraziSpawnToken(), N: ctx.X()}.Apply(ctx)
				}),
			ModeDoing("Target player scries X, then draws a card.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					player := t.ID
					return Scry{Player: player, N: ctx.X(), Then: func(g *game.Game) error {
						return g.DrawNForEffect(player, 1)
					}}.Apply(ctx)
				}),
			ModeDoing("Exile target creature with mana value X or less.",
				TargetCreature("target creature with mana value X or less").WithManaValueAtMostX(),
				exileTheModesTarget),
			ModeDoing("Exile up to X target cards from graveyards.",
				kozilekGraveyardTargets(),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		),
	})
}

// kozilekGraveyardTargets is the fourth bullet's clause: X cards from
// any graveyard, count bound to the announced X (see the file
// comment's declared caveat on "up to").
func kozilekGraveyardTargets() *game.TargetSpec {
	spec := TargetCardInGraveyard("up to X target cards from graveyards")
	spec.CountFromX = true
	return spec
}
