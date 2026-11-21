package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Enigma Jewel // Locus of Enlightenment — a transforming legendary
// artifact (#3201, ADR 0155 PR 3):
//
//	The Enigma Jewel — Legendary Artifact {U}
//	  "The Enigma Jewel enters tapped.
//	   {T}: Add {C}{C}. Spend this mana only to activate abilities.
//	   Craft with four or more nonlands with activated abilities {8}{U}"
//	Locus of Enlightenment — Legendary Artifact
//	  "Locus of Enlightenment has each activated ability of the exiled
//	   cards used to craft it. You may activate each of those abilities
//	   only once each turn.
//	   Whenever you activate an ability that isn't a mana ability, copy
//	   it. You may choose new targets for the copy."
//
// The craft is a set over four or more other permanents you control
// and/or cards in your graveyard, each a nonland with an activated ability
// (CraftWithNonlandsWithActivatedAbilities). The materials are CR
// 702.167c's craft link (Card.CraftedWith), and the Locus borrows from
// the ones still in exile (CraftedWithLenders) with the once-each-turn
// tally (AbilityBorrow.EachOncePerTurn). A material that leaves exile
// stops lending (DependsOnExile).
//
// The copy trigger is Rings of Brighthearth's without the {2}: it reads
// the activated ability off the triggering event and copies it with the
// CR 707.10c new-targets offer. A mana ability announces another event
// kind and never uses the stack (CR 605.3b), so it is never copied, and
// the copy is created rather than activated, so it does not trigger the
// Locus again (CR 707.10a).
//
// No simplification.
const theEnigmaJewelOracleID = "c0d2bdd9-1e68-4d3a-83f8-ec2ba440a494"

func init() {
	effects.Register(effects.Spec{
		OracleID:     theEnigmaJewelOracleID,
		Name:         "The Enigma Jewel",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{C}{C}",
			Label:        "{T}: Add {C}{C}. Spend this mana only to activate abilities.",
			Restrictions: []string{game.ManaRestrictActivate},
		}},
		Activated: []effects.ActivatedAbility{
			effects.Craft("Craft with four or more nonlands with activated abilities {8}{U}", "{8}{U}", effects.CraftWithNonlandsWithActivatedAbilities(4)),
		},
	})

	const clause = "Locus of Enlightenment has each activated ability of the exiled cards used to craft it. You may activate each of those abilities only once each turn."
	borrow := effects.HasActivatedAbilitiesOf(game.AbilityBorrow{
		Label:           clause,
		Lenders:         effects.CraftedWithLenders(),
		EachOncePerTurn: true,
	})
	borrow.DependsOnExile = true
	effects.Register(effects.Spec{
		OracleID:     theEnigmaJewelOracleID + "#1",
		Name:         "Locus of Enlightenment",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{borrow},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventActivateAbility, effects.ByYou, "Locus of Enlightenment — copy that ability", locusCopyActivatedAbility),
		},
	})
}

// locusCopyActivatedAbility copies the ability the trigger fired on, with
// the new-targets offer. Nothing to copy when the event names no stack
// item, or the ability has already left the stack.
func locusCopyActivatedAbility(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ability := ctx.Trigger().Event.StackItemID
	if ability == uuid.Nil {
		return nil
	}
	return effects.CopyAbility{ItemID: ability, Controller: item.Controller, ChooseNewTargets: true}.Apply(ctx)
}
