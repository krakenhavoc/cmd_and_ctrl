package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teferi, Temporal Archmage — Legendary Planeswalker — Teferi
// {4}{U}{U}, starting loyalty 5:
//
//	"+1: Look at the top two cards of your library. Put one of them
//	 into your hand and the other on the bottom of your library.
//	 −1: Untap up to four target permanents.
//	 −10: You get an emblem with "You may activate loyalty abilities
//	 of planeswalkers you control on any player's turn any time you
//	 could cast an instant."
//	 Teferi, Temporal Archmage can be your commander."
//
// THE EMBLEM is the card #1275 was filed for: Teferi, Master of
// Time's static one zone over. It is `EmblemSpec.ActivationTimings`,
// read by `game.ActivationTimingOpenLocked`'s walk of every seat's
// `Player.Emblems` beside the battlefield — so the engine, the bot
// enumerator, the view's `timing_closed` and the sandbox loyalty verb
// all see it through the one read they already share. Nothing is
// stored: an emblem never leaves (CR 114.2, CR 800.4a), so its
// presence is the duration.
//
// It opens the WINDOW and nothing else. CR 606.3's other half — one
// loyalty ability per planeswalker per turn — still holds on every
// turn it opens, which is the whole difference between this emblem
// and an infinite one.
//
// Teferi's Talent's Aura grants the enchanted planeswalker a −12 that
// makes this same emblem; see teferis_talent.go for why that card
// cannot reach it yet, and why the emblem is built by a function.
//
// THE +1 is LookAtTopThenTakeOneRestOnBottom, Impulse's sentence at
// two cards: a LOOK (only Teferi's controller learns the cards), one
// is mandatory, and "the other" is a pile of one, so the any-order
// prompt the helper would raise for a longer rest is never asked.
// The piece #1275 said this card was waiting on had landed with #952
// by the time it was built.
//
// THE −1 targets up to four permanents of any controller and untaps
// the ones still legal on resolution (CR 608.2b).
//
// "Can be your commander" is deck validation's, read off Scryfall's
// commander legality (deck/validate.go) — not a catalog property.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "07d0b06b-80cb-4518-92c9-84ea87a7e08a",
		Name:         "Teferi, Temporal Archmage",
		Completeness: CompletenessFull,
		// The fallback for tokens, fixtures and the dev spawner; an
		// imported deck reads printed loyalty (ADR 0032 §1).
		StartingLoyalty: 5,
		Emblem:          teferiTemporalArchmageEmblem(),
		Activated: []ActivatedAbility{
			{
				Label: "+1: Look at the top two cards of your library. Put one of them into your hand and the other on the bottom of your library.",
				Cost:  LoyaltyCost(1),
				Effect: LookAtTopThenTakeOneRestOnBottom(2,
					"Teferi, Temporal Archmage — put one into your hand and the other on the bottom"),
			},
			{
				Label:   "−1: Untap up to four target permanents.",
				Cost:    LoyaltyCost(-1),
				Targets: TargetPermanent("up to four target permanents").WithCount(0, 4),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, id := range legalTargetCards(item, g) {
						if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label: "−10: You get an emblem with \"You may activate loyalty abilities of planeswalkers you control on any player's turn any time you could cast an instant.\"",
				Cost:  LoyaltyCost(-10),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// teferiTemporalArchmageEmblem is the emblem, built fresh per caller
// so that Teferi's Talent, the day it can grant its −12, declares the
// same emblem without the two Specs sharing a slice. "On any player's
// turn" is not a clause of its own — see
// ThisSourcesLoyaltyAbilitiesAtInstantSpeed.
func teferiTemporalArchmageEmblem() *EmblemSpec {
	const text = "You may activate loyalty abilities of planeswalkers you control on any player's turn any time you could cast an instant."
	return &EmblemSpec{
		Label: "Teferi, Temporal Archmage emblem",
		Text:  text,
		ActivationTimings: []game.ActivationTiming{
			LoyaltyAbilitiesOfYourPlaneswalkersAtInstantSpeed(text),
		},
	}
}
