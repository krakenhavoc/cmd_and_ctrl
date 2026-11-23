package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// ModeDoing), resolved in printed order (CR 608.2c). The first
// bullet reuses the printed Eldrazi Spawn template; the second rides
// Scry's Then the same way Preordain does; the third narrows its
// battlefield clause with WithManaValueAtMostX, the announced-X
// binding Agadeem's Awakening uses for its graveyard clause.
//
// The fourth bullet's "up to X target cards from graveyards" is
// CountFromX with UpToX: the announced X is the ceiling (#1738).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f2c549d-0293-4b9d-b9a8-ff392800f3a5",
		Name:         "Kozilek's Command",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeDoing("Target player creates X 0/1 colorless Eldrazi Spawn creature tokens "+
				"with \"Sacrifice this token: Add {C}.\"",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.CreateToken{Controller: t.ID, Template: effects.EldraziSpawnToken(), N: ctx.X()}.Apply(ctx)
				}),
			effects.ModeDoing("Target player scries X, then draws a card.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					player := t.ID
					draw := ctx.TailN(effects.ThenDrawCards, 1)
					draw.Player = player
					return effects.Scry{Player: player, N: ctx.X(), ThenTail: draw}.Apply(ctx)
				}),
			effects.ModeDoing("Exile target creature with mana value X or less.",
				effects.TargetCreature("target creature with mana value X or less").WithManaValueAtMostX(),
				effects.MurdocksExileTheModesTarget),
			effects.ModeDoing("Exile up to X target cards from graveyards.",
				kozilekGraveyardTargets(),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		),
	})
}

// kozilekGraveyardTargets is the fourth bullet's clause: X cards from
// any graveyard, count capped at the announced X.
func kozilekGraveyardTargets() *game.TargetSpec {
	spec := effects.TargetCardInGraveyard("up to X target cards from graveyards")
	spec.CountFromX = true
	spec.UpToX = true
	return spec
}
