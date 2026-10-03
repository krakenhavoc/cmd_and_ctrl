package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Narset Transcendent — Legendary Planeswalker — Narset {2}{W}{U},
// starting loyalty 6:
//
//	"+1: Look at the top card of your library. If it's a noncreature,
//	 nonland card, you may reveal it and put it into your hand.
//	 −2: When you next cast an instant or sorcery spell from your hand
//	 this turn, it gains rebound. (Exile the spell as it resolves. At
//	 the beginning of your next upkeep, you may cast that card from
//	 exile without paying its mana cost.)
//	 −9: You get an emblem with "Your opponents can't cast noncreature
//	 spells.""
//
// THE EMBLEM is the card #1899 was filed for (ADR 0109 §5). It is
// EmblemSpec.CastRestrictions, read by the cast gate's walk of every
// seat's emblems beside the battlefield, with the emblem as the source,
// so the engine, the bot's move list and the view's `cant_cast` all
// refuse an opponent's noncreature spell through the one read they
// share. It is not a CastBanRule written onto each opponent: the
// emblem is the object CR 114.4 says the ability is on, the board shows
// it, and it leaves with its owner (CR 800.4a). A creature spell, and
// Narset's controller's own spells, are not refused; a land is played,
// not cast, so it is not either.
//
// THE +1 is a private look (only Narset's controller learns the card),
// then a "you may" over a card that qualifies: TakeFromLibraryToHand
// with Reveal, so the card is shown to the table only if it is taken.
// A card that does not qualify stays on top, as printed.
//
// THE −2 is a CR 603.7 delayed trigger watching for the next instant or
// sorcery its controller casts FROM HAND this turn. A cast from
// anywhere else neither fires it nor uses it up (CR 603.7b: it triggers
// the next time its trigger EVENT occurs). It ends with the turn
// (CR 514.2). The trigger goes on the stack above the spell and resolves
// first, giving the spell rebound (CR 702.88a), so the spell, cast from
// hand, is exiled as it resolves and comes back at the next upkeep. A
// spell countered before the trigger resolves gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e1de0c94-0ecc-425a-9b92-27c2745d07e7",
		Name:         "Narset Transcendent",
		Completeness: CompletenessFull,
		// The fallback for tokens, fixtures and the dev spawner; an
		// imported deck reads printed loyalty (ADR 0032 §1).
		StartingLoyalty: 6,
		Emblem: &EmblemSpec{
			Label: "Narset Transcendent emblem",
			Text:  "Your opponents can't cast noncreature spells.",
			CastRestrictions: []game.CastRestriction{
				OpponentsCantCast("Your opponents can't cast noncreature spells.", Noncreature()),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:  "+1: Look at the top card of your library. If it's a noncreature, nonland card, you may reveal it and put it into your hand.",
				Cost:   LoyaltyCost(1),
				Effect: narsetLookAtTheTop,
			},
			{
				Label: "−2: When you next cast an instant or sorcery spell from your hand this turn, it gains rebound.",
				Cost:  LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return WhenYouNextCastFromHand(
						"Narset Transcendent — it gains rebound",
						game.CastFilter{Types: []string{"Instant", "Sorcery"}},
						reboundTheSpellBody,
					).Apply(NewContext(g, item))
				},
			},
			{
				Label: "−9: You get an emblem with \"Your opponents can't cast noncreature spells.\"",
				Cost:  LoyaltyCost(-9),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// narsetLookAtTheTop is the +1: look at the top card, and offer to
// reveal it and take it if it is a noncreature, nonland card.
func narsetLookAtTheTop(g *game.Game, item *game.StackItem) error {
	return TakeFromLibraryToHand{
		Player:   item.Controller,
		Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 1),
		Match:    And(Noncreature(), Nonland()),
		Max:      1,
		Optional: true,
		Reveal:   true,
		Label:    "Narset Transcendent — you may reveal it and put it into your hand",
	}.Apply(NewContext(g, item))
}
