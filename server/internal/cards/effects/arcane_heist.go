package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcane Heist — Sorcery {2}{U}{U}:
//
//	"You may cast target instant or sorcery card from an opponent's
//	 graveyard without paying its mana cost. If that spell would be put
//	 into their graveyard, exile it instead.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// Torrential Gearhulk's free cast out of a graveyard (ADR 0066), pointed
// at an opponent's: a flash-timed grant closed by the caster's next
// priority pass (the sandbox simplification every "you may cast it"
// during a resolution takes), with #3863's "if that spell would be put
// into a graveyard, exile it instead". A spell goes to its OWNER's
// graveyard, so "their graveyard" is the only one it could reach. Cipher
// is the engine's keyword (game/cipher.go, CR 702.99, #3867).
//
// No simplification beyond the shared one.
func init() {
	Register(Spec{
		OracleID:        "839748e7-ccb4-421e-9626-b6d8be9390ab",
		Name:            "Arcane Heist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets: TargetCardInGraveyard("target instant or sorcery card from an opponent's graveyard",
			Or(Instant(), Sorcery()), Not(YouOwn())),
		OnResolve: arcaneHeistOffer,
	})
}

// arcaneHeistOffer opens the free cast on the still-legal target.
func arcaneHeistOffer(_ *game.StackItem, ctx *Context) error {
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
			SourceName:              "Arcane Heist",
			Label:                   "Arcane Heist — cast " + c.Name + " without paying its mana cost",
		}, []game.Card{c})
		// #4250: cast during the resolution (CR 608.2g).
		offerResolutionCastOver(g, ctx.Controller(), ctx.Source(), "Arcane Heist", game.MayCastKeywordFreeCast, declineLeaveInGraveyard, c.InstanceID)
		return nil
	}
	return nil
}
