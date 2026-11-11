package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flotsam // Jetsam — split card {1}{G/U} // {4}{U/B}{U/B} (Murders at
// Karlov Manor, #4469), oracle b0c28b2b-a2dd-4b76-bb1c-cec55a0a6784:
//
//	Flotsam — Instant: "Mill three cards. Investigate."
//	Jetsam  — Sorcery: "Each opponent mills three cards, then you may
//	          cast a spell from each opponent's graveyard without paying
//	          its mana cost. If a spell cast this way would be put into a
//	          graveyard, exile it instead."
//
// Both halves register as ADR 0034 faces, Flotsam under the bare oracle ID
// and Jetsam under "#1" (ADR 0103), like Wear // Tear.
//
// Jetsam's casts are made during its resolution (CR 608.2g), through the
// resolution_cast prompt, one permission per opponent over every nonland
// card their graveyard holds after the mill: a permission over several
// cards is spent by one cast ("a spell from each opponent's graveyard"),
// the cast pays {0} (X locked at 0), and the card is exiled rather than
// put into a graveyard if the spell would be (ExileInsteadOfGraveyard).
// Timing restrictions are ignored (TimingFlash). The opponents' graveyards
// are read after every opponent has milled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0c28b2b-a2dd-4b76-bb1c-cec55a0a6784",
		Name:         "Flotsam",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (MillCards{N: 3}).Apply(ctx); err != nil {
				return err
			}
			return Investigate{N: 1}.Apply(ctx)
		},
	})
	Register(Spec{
		OracleID:     "b0c28b2b-a2dd-4b76-bb1c-cec55a0a6784#1",
		Name:         "Jetsam",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			g := ctx.Game
			for _, opp := range ctx.Opponents() {
				if err := (MillCards{Player: opp, N: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, opp := range ctx.Opponents() {
				var cards []game.Card
				var ids []uuid.UUID
				if p := g.PlayerByIDForEffect(opp); p != nil && p.Graveyard != nil {
					for _, c := range p.Graveyard.Cards {
						if !c.IsLand() {
							cards = append(cards, c)
							ids = append(ids, c.InstanceID)
						}
					}
				}
				if len(cards) == 0 {
					continue
				}
				g.GrantCastPermissionToCardsForEffect(game.CastPermission{
					Player:                  ctx.Controller(),
					Zone:                    game.ZoneGraveyard,
					Duration:                g.UntilEndOfTurnDuration(),
					Cost:                    "{0}",
					AltCostKey:              game.AltCostKeyGrantedCost,
					Timing:                  game.TimingFlash,
					CastOnly:                true,
					ExileInsteadOfGraveyard: true,
					LapseOnPass:             game.LapseStaysInHand,
					Source:                  ctx.Source(),
					SourceName:              "Jetsam",
					Label:                   "Jetsam — cast a spell from this graveyard without paying its mana cost",
				}, cards)
				offerResolutionCastOver(g, ctx.Controller(), ctx.Source(), "Jetsam", game.MayCastKeywordFreeCast, declineLeaveInGraveyard, ids...)
			}
			return nil
		},
	})
}
