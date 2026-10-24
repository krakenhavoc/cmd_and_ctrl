package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fevered Suspicion — Sorcery {6}{B}{R}:
//
//	"Each opponent exiles cards from the top of their library until
//	 they exile a nonland card. You may cast any number of spells from
//	 among those nonland cards without paying their mana costs.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Each opponent's run is the exile-until primitive (MillToZone with an
// Until), one after another, and the casts are offered once every run
// has finished. Each nonland card exiled this way gets its own free-cast
// permission for you: "{0}" (CR 107.3b locks X at 0), flash timing
// because the cast happens during the resolution (CR 608.2g), and the
// window cascade and rebound use — it closes when you pass priority,
// and an uncast card stays in exile. One grant per card, so casting one
// leaves the others open. Lands exiled on the way stay in exile.
//
// #4250: the casts are made DURING the resolution (CR 608.2g), each its
// own answer to the resolution_cast prompt over the exiled cards, until
// the caster declines or none is left; the decline leaves the rest in
// exile. Nobody receives priority in between.
//
// Rebound is the engine's keyword (game/rebound.go, #1854).
func init() {
	Register(Spec{
		OracleID:        "51545f6e-d0e3-4b4c-bc73-c731f85e26b0",
		Name:            "Fevered Suspicion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return feveredSuspicionExile(ctx, ctx.Opponents(), nil)
		},
	})
}

// feveredSuspicionExile runs the first opponent's exile-until and
// recurses on the rest from its continuation, collecting the nonland
// cards; with nobody left it offers the casts.
func feveredSuspicionExile(ctx *Context, opponents, hits []uuid.UUID) error {
	if len(opponents) == 0 {
		feveredSuspicionGrantCasts(ctx, hits)
		return nil
	}
	return MillToZone{
		Player: opponents[0],
		To:     game.ZoneExile,
		Until:  UntilCard(func(c game.Card) bool { return !c.IsLand() }),
		Then: func(ctx *Context, exiled []uuid.UUID) error {
			for _, id := range exiled {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsLand() {
					hits = append(hits, id)
				}
			}
			return feveredSuspicionExile(ctx, opponents[1:], hits)
		},
	}.Apply(ctx)
}

// feveredSuspicionGrantCasts stamps one pass-closed free cast per
// nonland card still in exile.
func feveredSuspicionGrantCasts(ctx *Context, hits []uuid.UUID) {
	g := ctx.Game
	var granted []uuid.UUID
	defer func() {
		// #4250: "any number of spells", each its own answer during the
		// resolution (CR 608.2g).
		offerResolutionCastOver(g, ctx.Controller(), ctx.Source(), "Fevered Suspicion",
			game.MayCastKeywordFreeCast, declineLeaveInExile, granted...)
	}()
	for _, id := range hits {
		c, ok := g.LookupCardForEffect(id)
		if z := g.FindCardZoneForEffect(id); !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:      ctx.Controller(),
			Zone:        game.ZoneExile,
			Cost:        "{0}",
			Timing:      game.TimingFlash,
			CastOnly:    true,
			Duration:    g.UntilEndOfTurnDuration(),
			LapseOnPass: game.LapseStaysInExile,
			Source:      ctx.Source(),
			SourceName:  "Fevered Suspicion",
			Label:       "Fevered Suspicion — cast it without paying its mana cost",
		}, []game.Card{c})
		granted = append(granted, id)
	}
}
