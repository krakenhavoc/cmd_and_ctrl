package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Memory Plunder — Instant {U/B}{U/B}{U/B}{U/B}:
//
//	"You may cast target instant or sorcery card from an opponent's
//	 graveyard without paying its mana cost."
//
// Arcane Heist's free cast (ADR 0066) without its exile rider: the
// target is an instant or sorcery card an opponent owns, and as Memory
// Plunder resolves its caster is granted a cast of that card from the
// graveyard for {0}, flash-timed so a stolen sorcery can be cast with
// the resolution still behind it (CR 608.2g). Since #4250 the cast is
// made during the resolution, through the resolution_cast prompt over the
// card. Declining is not casting it. The card is the real card,
// not a copy, so once it resolves or is countered it goes to its
// owner's graveyard, where it came from. A target that left the
// graveyard in response makes Memory Plunder do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36b67014-0507-4976-ad52-a2a36676197d",
		Name:         "Memory Plunder",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target instant or sorcery card from an opponent's graveyard",
			Or(Instant(), Sorcery()), Not(YouOwn())),
		OnResolve: nghaMemoryPlunderOffer,
	})
}

// nghaMemoryPlunderOffer opens the free cast on the still-legal target.
func nghaMemoryPlunderOffer(_ *game.StackItem, ctx *Context) error {
	g := ctx.Game
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		z := g.FindCardZoneForEffect(t.ID)
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:      ctx.Controller(),
			Zone:        game.ZoneGraveyard,
			Duration:    g.UntilEndOfTurnDuration(),
			Cost:        "{0}",
			AltCostKey:  game.AltCostKeyGrantedCost,
			Timing:      game.TimingFlash,
			CastOnly:    true,
			LapseOnPass: game.LapseStaysInHand,
			Source:      ctx.Source(),
			SourceName:  "Memory Plunder",
			Label:       "Memory Plunder — cast " + c.Name + " without paying its mana cost",
		}, []game.Card{c})
		// #4250: cast during the resolution (CR 608.2g).
		offerResolutionCastOver(g, ctx.Controller(), ctx.Source(), "Memory Plunder", game.MayCastKeywordFreeCast, declineLeaveInGraveyard, c.InstanceID)
		return nil
	}
	return nil
}
