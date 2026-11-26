package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nissa, Vastwood Seer // Nissa, Sage Animist — Magic Origins (S82 older sets),
// a transforming legendary creature whose back face is a planeswalker.
//
// Front face, Legendary Creature — Elf Scout {2}{G}, 2/2:
//
//	"When Nissa enters, you may search your library for a basic Forest card,
//	 reveal it, put it into your hand, then shuffle.
//	 Whenever a land you control enters, if you control seven or more lands,
//	 exile Nissa, then return her to the battlefield transformed under her
//	 owner's control."
//
// Back face, Legendary Planeswalker — Nissa, loyalty 3:
//
//	"+1: Reveal the top card of your library. If it's a land card, put it onto
//	     the battlefield. Otherwise, put it into your hand.
//	 −2: Create Ashaya, the Awoken World, a legendary 4/4 green Elemental
//	     creature token.
//	 −7: Untap up to six target lands. They become 6/6 Elemental creatures.
//	     They're still lands."
//
// The back face registers under "<oracle_id>#1". The search is optional and
// the type a word of the source's text (CR 612.2). The flip is an intervening
// "if" (CR 603.4) checked as a land enters and again as the trigger resolves;
// Nissa returns as a new object under her owner's control (CR 712.14a). The +1
// is Coiling Oracle's reveal. The −7 untaps the lands and makes each a 6/6
// Elemental creature in addition to being a land, with no duration (CR 611.2a), the
// set fixed as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	const oracle = "35754a21-9fba-4370-a254-292918a777ba"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Nissa, Vastwood Seer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Nissa, Vastwood Seer — you may search your library for a basic Forest card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLandOfAnySubtype(ctx.TextSubtype("Forest")),
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Nissa, Vastwood Seer — a basic Forest card",
					}.Apply(ctx)
				}),
			effects.InterveningIfYou("you control seven or more lands",
				func(g *game.Game, you uuid.UUID) bool { return effects.B02CountLandsControlledBy(g, you) >= 7 },
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && c.IsLand()
				}, "Nissa, Vastwood Seer — exile Nissa, then return her to the battlefield transformed",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
					})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Nissa, Sage Animist",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Reveal the top card of your library. If it's a land card, put it onto the battlefield. Otherwise, put it into your hand.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.OldB322RevealTopLandElseHand,
			},
			{
				Label:   "−2: Create Ashaya, the Awoken World, a legendary 4/4 green Elemental creature token.",
				Cost:    effects.LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 green Legendary Elemental named Ashaya, the Awoken World"), N: 1}),
			},
			{
				Label:   "−7: Untap up to six target lands. They become 6/6 Elemental creatures. They're still lands.",
				Cost:    effects.LoyaltyCost(-7),
				Targets: effects.TargetPermanent("up to six target lands", effects.Land()).WithCount(0, 6),
				Effect:  nissaSageAnimistAnimate,
			},
		},
	})
}

func nissaSageAnimistAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var lands []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			lands = append(lands, t.ID)
		}
	}
	for _, id := range lands {
		if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, id := range lands {
		if err := (effects.BecomeCreatureUntilEOT{
			Target:       id,
			Types:        []string{"Creature"},
			Subtypes:     []string{"Elemental"},
			SetPower:     6,
			SetToughness: 6,
			Label:        "Nissa, Sage Animist — becomes a 6/6 Elemental creature",
		}).ApplyFor(ctx, game.IndefiniteDuration()); err != nil {
			return err
		}
	}
	return nil
}
