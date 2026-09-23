package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi, Who Slows the Sunset — Legendary Planeswalker — Teferi for
// {2}{W}{U}, starting loyalty 4:
//
//	"+1: Choose up to one target artifact, up to one target creature,
//	     and up to one target land. Untap the chosen permanents you
//	     control. Tap the chosen permanents you don't control. You
//	     gain 2 life.
//	 −2: Look at the top three cards of your library. Put one of them
//	     into your hand and the rest on the bottom of your library in
//	     any order.
//	 −7: You get an emblem with 'Untap all permanents you control
//	     during each opponent's untap step' and 'You draw a card
//	     during each opponent's draw step.'"
//
// The −7 EMBLEM is what this file is for (#1315, tracker #884): two
// clauses that modify turn-based actions rather than triggering, so
// they cannot go through Emblem.Triggered. EmblemSpec.UntapStep and
// EmblemSpec.DrawStep are the new slots that let it be declared
// exactly as game.UntapStepPermission / game.DrawStepPermission would
// be on a permanent — see untap.go and draw_step.go for why "during
// each opponent's [step]" is a widened turn-based action and not an
// "at the beginning of" trigger.
//
// The +1 is IN FULL. Its three independent "up to one" clauses are
// Clauses() + WithCount(0, 1) (#937, ADR 0065) — no Distinct, because
// nothing in the printed text says the three targets must differ
// (CR 601.2c lets one object fill more than one of these unless the
// card says otherwise, so an artifact land can legally be both the
// chosen artifact and the chosen land). Each chosen permanent is
// untapped if its controller matches the ability's controller and
// tapped otherwise — read at RESOLUTION, so a permanent that changed
// hands in response is judged by who controls it now, which is what
// "the chosen permanents you control" asks. The life gain is
// unconditional: zero, one, two or three targets, Teferi's controller
// still gains 2 life.
//
// The −2 is a DECLARED SIMPLIFICATION, weaker than printed (matching
// Goblin Ringleader's #259 caveat, not Horn of the Mark's uncaveated
// "random order" clause): the two cards NOT taken go to the bottom of
// the library in a RANDOM order rather than one the player chooses.
// The engine has no ordering prompt for a pile headed to the bottom
// of a library (that primitive is ADR 0088, not yet on this branch).
// LookAtTopOfLibraryForEffect + TakeFromLibraryToHand +
// TakeRestOnBottomInRandomOrder is the same three-piece composition
// Horn of the Mark's May-take clause uses, with Optional:false and
// Max:1 for "put ONE of them" (mandatory, not "you may").
func init() {
	Register(Spec{
		OracleID:     "98c389ed-b960-42e2-9c76-a062f77c9a78",
		Name:         "Teferi, Who Slows the Sunset",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The −2's two cards you don't keep go to the bottom of your library in a random order — you don't get to choose the order.",
		},
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Teferi, Who Slows the Sunset emblem",
			Text: "Untap all permanents you control during each opponent's untap step. " +
				"You draw a card during each opponent's draw step.",
			UntapStep: []game.UntapStepPermission{
				untapDuringEachOtherPlayersUntapStep(
					"Teferi, Who Slows the Sunset emblem — untap all permanents you control during each opponent's untap step",
					nil),
			},
			DrawStep: []game.DrawStepPermission{
				drawDuringEachOtherPlayersDrawStep(
					"Teferi, Who Slows the Sunset emblem — you draw a card during each opponent's draw step", 1),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Choose up to one target artifact, up to one target creature, and up to one target " +
					"land. Untap the chosen permanents you control. Tap the chosen permanents you don't " +
					"control. You gain 2 life.",
				Cost: LoyaltyCost(1),
				Targets: Clauses(
					TargetPermanent("up to one target artifact", Artifact()).WithCount(0, 1),
					TargetCreature("up to one target creature").WithCount(0, 1),
					TargetPermanent("up to one target land", Land()).WithCount(0, 1),
				),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for slot := 0; slot < 3; slot++ {
						t, ok := ctx.ClauseTarget(slot)
						if !ok || t.Kind != game.TargetCard {
							continue
						}
						c, found := g.LookupCardForEffect(t.ID)
						if !found {
							continue
						}
						if c.Controller == item.Controller {
							if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
								return err
							}
						} else if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
				},
			},
			{
				Label: "−2: Look at the top three cards of your library. Put one of them into your hand and " +
					"the rest on the bottom of your library in any order.",
				Cost: LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := item.Controller
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 3),
						Max:    1,
						Label:  "Teferi, Who Slows the Sunset — put one of them into your hand",
						Then:   TakeRestOnBottomInRandomOrder,
					}.Apply(ctx)
				},
			},
			{
				Label: "−7: You get an emblem with \"Untap all permanents you control during each opponent's " +
					"untap step\" and \"You draw a card during each opponent's draw step.\"",
				Cost: LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
