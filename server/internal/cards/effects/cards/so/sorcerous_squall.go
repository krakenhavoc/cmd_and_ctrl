package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sorcerous Squall — Sorcery {6}{U}{U}{U} (Tales of Middle-earth Commander):
//
//	"Delve (Each card you exile from your graveyard while casting this
//	 spell pays for {1}.)
//	 Target opponent mills nine cards, then you may cast an instant or
//	 sorcery spell from that player's graveyard without paying its mana
//	 cost. If that spell would be put into a graveyard, exile it
//	 instead."
//
// Delve is `Delve: true` (CR 702.66). The mill finishes first, so the
// milled cards are castable (CR 608.2c); the cast is made during the
// resolution (CR 608.2g) through the resolution_cast prompt, over every
// instant or sorcery card in that player's graveyard under ONE
// permission, so the first cast spends it: one spell. The permission is
// free ({0}, so X is 0, CR 107.3b), cast-only, and its
// ExileInsteadOfGraveyard is the printed replacement: the spell is
// exiled whether it resolves, is countered or fizzles. A pass closes the
// window and leaves the card where it is. A target that is gone as the
// spell resolves makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Sorcerous Squall"
	effects.Register(effects.Spec{
		OracleID:     "a986d83d-22f1-45c3-bc2d-f5b210c539fb",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Delve:        true,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			return ctx.Game.MillToZoneKeyedMovedForEffect(t.ID, 9, game.ZoneGraveyard, ctx.ItemTail(sorcerousSquallCast))
		},
	})
}

// sorcerousSquallCast is "then you may cast an instant or sorcery spell from
// that player's graveyard", after the mill has finished. The player is the
// spell's target, read from the carried item. An on-disk key: never renamed,
// never reused.
var sorcerousSquallCast = effects.CardTail("then/sorcerous-squall-cast", func(ctx *effects.Context, _ game.TailOutcome) error {
	const name = "Sorcerous Squall"
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	g := ctx.Game
	p := g.PlayerByIDForEffect(t.ID)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		ids = append(ids, c.InstanceID)
	}
	perm := effects.FreeInstantOrSorcery().On(game.CastPermission{
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
		SourceName:              name,
		Label:                   name + " — cast it without paying its mana cost",
	})
	candidates := effects.FreeCastCandidates(g, &perm, ids)
	if len(candidates) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      ctx.Controller(),
		Source:       ctx.Source(),
		Reason:       name,
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: effects.DeclineLeaveInGraveyard,
	})
	return nil
})
