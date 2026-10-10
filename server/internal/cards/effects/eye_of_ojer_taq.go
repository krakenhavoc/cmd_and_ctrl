package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eye of Ojer Taq // Apex Observatory — a transforming artifact (#2709,
// ADR 0137 and its 2026-10-10 amendment):
//
//	Eye of Ojer Taq — Artifact {3}
//	  "{T}: Add one mana of any color.
//	   Craft with two that share a card type {6}"
//	Apex Observatory — Artifact
//	  "This artifact enters tapped. As it enters, choose a card type
//	   shared among two exiled cards used to craft it.
//	   {T}: The next spell you cast this turn of the chosen type can be
//	   cast without paying its mana cost."
//
// The craft is a set rule over two materials from either zone: they
// must have a card type in common (CraftWithTwoSharingACardType).
//
// The Observatory's choice is an "as enters" option pick (CR 614.12)
// over the card types its two materials still in exile share, read off
// CR 702.167c's link as it lands; with fewer than two left in exile
// there is no type to choose, and the tap ability then does nothing.
// The tap ability is a one-use "next spell" promise (#1852) whose rider
// is a free alternative cost (game.NextSpellPromise.WithoutPayingManaCost,
// CR 118.9), filtered to the chosen card type. The next spell of that
// type uses it up whichever cost it is cast for.
//
// No simplification.
const eyeOfOjerTaqOracleID = "0b1557e8-3dd6-4276-8b8e-505e7ea944ec"

func init() {
	Register(Spec{
		OracleID:     eyeOfOjerTaqOracleID,
		Name:         "Eye of Ojer Taq",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{
			Craft("Craft with two that share a card type {6}", "{6}", CraftWithTwoSharingACardType()),
		},
	})

	Register(Spec{
		OracleID:     eyeOfOjerTaqOracleID + "#1",
		Name:         "Apex Observatory",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		AsEnters:     apexObservatoryChooseType,
		Activated: []ActivatedAbility{{
			Label:   "{T}: The next spell you cast this turn of the chosen type can be cast without paying its mana cost.",
			Cost:    TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  apexObservatoryPromise,
		}},
	})
}

// apexObservatoryChooseType is the CR 614.12 choice: a card type the
// two materials share, as an option pick. Nothing to choose with fewer
// than two materials in exile.
func apexObservatoryChooseType(card *game.Card, ctx *Context) error {
	mats := craftMaterialsOf(ctx.Game, card)
	if len(mats) < 2 {
		return nil
	}
	var words []string
	for _, t := range game.SharedCardTypes(mats) {
		words = append(words, strings.ToUpper(t[:1])+t[1:])
	}
	if len(words) == 0 {
		return nil
	}
	ctx.Game.QueueChooseOptionAsEntersForEffect(card.Controller, card.InstanceID,
		"Apex Observatory — choose a card type shared among two exiled cards used to craft it", words)
	return nil
}

// apexObservatoryPromise is the tap ability: the next spell of the
// chosen type this turn may be cast for free.
func apexObservatoryPromise(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src, ok := ctx.SourcePermanent()
	if !ok || src.ChosenOption == "" {
		return nil
	}
	chosen := strings.ToLower(src.ChosenOption)
	return GrantNextSpellPromise{From: "Apex Observatory", Promise: game.NextSpellPromise{
		Filter:                game.PermissionFilter{CardType: chosen},
		WithoutPayingManaCost: true,
		Text:                  "The next " + chosen + " spell you cast this turn can be cast without paying its mana cost.",
	}}.Apply(ctx)
}
