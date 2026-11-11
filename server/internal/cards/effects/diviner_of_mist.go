package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diviner of Mist — Creature — Dragon {4}{U}, 4/5 (Tarkir: Dragonstorm
// Commander):
//
//	"Flying
//	 Whenever this creature attacks, mill four cards. You may cast an instant
//	 or sorcery spell from your graveyard with mana value 4 or less without
//	 paying its mana cost. If that spell would be put into your graveyard,
//	 exile it instead."
//
// Bilbo, Thief in the Night's one cast from the graveyard (a ONE permission
// over every card that could be cast, spent whole by the first cast: "an ...
// spell", not each), made free ({0}, so X is 0, no alternative cost) and
// capped on the mana value of the spell being cast
// (CastPermission.MaxSpellManaValue, judged face by face). The offer is made
// as the ability resolves (CR 608.2g), after the mill has settled, so the
// four milled cards are on offer, and the spell is exiled whether it
// resolves, is countered or fizzles (ExileInsteadOfGraveyard). Declining
// leaves the cards where they are.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "abf4dcd8-5176-4f4a-b7c1-5ce24ff181bf",
		Name:            "Diviner of Mist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Diviner of Mist — mill four cards, then you may cast an instant or sorcery spell with mana value 4 or less from your graveyard without paying its mana cost",
				Do(
					MillCards{N: 4},
					Instruction(divinerOfMistOffer),
				)),
		},
	})
}

// divinerOfMistOffer is the second and third sentences.
func divinerOfMistOffer(ctx *Context) error {
	g := ctx.Game
	you := ctx.Controller()
	limit := 4
	perm := game.CastPermission{
		Player:                  you,
		Zone:                    game.ZoneGraveyard,
		Duration:                g.UntilEndOfTurnDuration(),
		Cost:                    "{0}",
		AltCostKey:              game.AltCostKeyGrantedCost,
		Filter:                  game.PermissionFilter{InstantOrSorceryOnly: true},
		MaxSpellManaValue:       &limit,
		Timing:                  game.TimingFlash,
		CastOnly:                true,
		ExileInsteadOfGraveyard: true,
		// A pass closes the window; the cards stay in the graveyard.
		LapseOnPass: game.LapseStaysInHand,
		Source:      ctx.Source(),
		SourceName:  "Diviner of Mist",
		Label:       "Diviner of Mist — cast an instant or sorcery spell with mana value 4 or less from your graveyard without paying its mana cost",
	}
	cards := graveyardCardIDs(ctx, you, perm.OffersASpell)
	if len(cards) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       ctx.Source(),
		Reason:       "Diviner of Mist",
		Cards:        cards,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: declineLeaveInGraveyard,
	})
	return nil
}
