package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi, Hero of Dominaria — Legendary Planeswalker — Teferi for
// {3}{W}{U}, starting loyalty 4 (EDHREC rank 4205):
//
//	"+1: Draw a card. At the beginning of the next end step, untap up
//	     to two lands.
//	 −3: Put target nonland permanent into its owner's library third
//	     from the top.
//	 −8: You get an emblem with 'Whenever you draw a card, exile
//	     target permanent an opponent controls.'"
//
// The −3 is complete, and it is the reason this card is worth
// writing rather than one more damage-dealing walker: "third from
// the top" is a real printed position, not a rounding of "on top",
// and the difference is two turns of the victim's draw step. It is
// the only tuck-to-a-depth in the catalog, so it brought
// game.Zone.InsertFromTop and zoneRoute.Depth with it — a library
// shorter than three cards takes the permanent on the bottom, which
// is as close to the printed position as the library can get.
//
// Because the −3 goes through the shared exit primitive, a COMMANDER
// tucked by it gets the CR 903.9 prompt and its owner may send it to
// the command zone instead (#529 / #539). Nothing in this file does
// that work; it happens because the tuck uses the engine's one door.
//
// The +1's DRAW is complete. Its untap clause ships with the choice
// made for the player: at the next end step it untaps up to two
// tapped lands its controller controls, taken in battlefield order.
// The engine has no "choose up to two permanents" prompt — the
// sacrifice picker is the only permanent chooser there is — and of
// the three honest options this is the one that is bounded by the
// printed card in every direction: it never untaps more than two, it
// never untaps a land the controller does not control, and the worst
// it can do is untap two lands the player would not have picked,
// which is weaker than the card, never stronger. Omitting the clause
// outright was the alternative and it throws away the reason the +1
// is a plus.
//
// The −8 is REGISTERED since S40 (#623) and is the catalog's first
// TRIGGERED emblem. Its whole declaration is an `Emblem` slot holding
// one `Targeting(WheneverYouDraw(…), …)` — the same two constructors
// a permanent's targeted draw trigger uses — and the emblem object
// then reaches the trigger harvester through the same per-zone walk
// that finds a battlefield permanent's triggers (ADR 0064 Decision 4).
//
// It fires once per card drawn, including the draws Teferi's own +1
// makes and the turn-based draw step, which is the printed card. The
// trigger goes on the stack, picks its target when it is put there
// (CR 603.3d), is dropped when the opponents control no permanent,
// and can be responded to — all of that is the ordinary trigger
// pipeline and none of it is emblem code.
//
// This is also the card behind bug #367, whose reporter described a
// Teferi showing "generic + or −" rows. That is what the client
// renders for a planeswalker with no catalogued abilities, so the
// symptom was this card's absence from the catalog rather than a
// fault in the loyalty pipeline.
func init() {
	Register(Spec{
		OracleID:     "f2f165b6-ef0a-42ad-9352-ba68be8248b0",
		Name:         "Teferi, Hero of Dominaria",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The +1's end-step untap picks two of your tapped lands for you rather than asking.",
		},
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Teferi, Hero of Dominaria emblem",
			Text:  "Whenever you draw a card, exile target permanent an opponent controls.",
			Triggered: []game.TriggeredAbility{
				Targeting(
					WheneverYouDraw(
						"Teferi, Hero of Dominaria emblem — exile target permanent an opponent controls",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							ts := ctx.LegalTargets()
							if len(ts) == 0 {
								return nil
							}
							return ExileTarget{Target: ts[0].ID}.Apply(ctx)
						}),
					TargetPermanent("target permanent an opponent controls", OpponentControls()),
				),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Draw a card. At the beginning of the next end step, untap up to two lands.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return ScheduleDelayedTrigger{
						At:         game.StepEnd,
						Controller: item.Controller,
						Label:      "Teferi, Hero of Dominaria — untap up to two lands",
						Effect:     teferiHeroUntapTwoLands,
					}.Apply(ctx)
				},
			},
			{
				Label:   "−3: Put target nonland permanent into its owner's library third from the top.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetPermanent("target nonland permanent", Nonland()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := g.TuckToLibraryAtDepthForEffect(t.ID, 3); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label: "−8: You get an emblem with \"Whenever you draw a card, exile target permanent an opponent controls.\"",
				Cost:  LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// teferiHeroUntapTwoLands is the body of the +1's delayed trigger.
// Package-level so the scheduled item captures nothing but the
// controller the scheduler already recorded.
//
// "Up to two" is a ceiling, so one tapped land, or none, is a legal
// and silent outcome.
func teferiHeroUntapTwoLands(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var picked []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if len(picked) == 2 {
			break
		}
		if c.Controller == item.Controller && c.IsLand() && c.Tapped {
			picked = append(picked, c.InstanceID)
		}
	}
	for _, id := range picked {
		if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
