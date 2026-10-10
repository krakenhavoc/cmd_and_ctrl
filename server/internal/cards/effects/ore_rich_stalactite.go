package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ore-Rich Stalactite // Cosmium Catalyst — a transforming artifact
// (#2709, ADR 0137 and its 2026-10-10 amendment):
//
//	Ore-Rich Stalactite — Artifact {1}{R}
//	  "{T}: Add {R}. Spend this mana only to cast an instant or sorcery
//	   spell.
//	   Craft with four or more red instant and/or sorcery cards
//	   {3}{R}{R}"
//	Cosmium Catalyst — Artifact
//	  "{1}{R}, {T}: Choose an exiled card used to craft this artifact at
//	   random. You may cast that card without paying its mana cost."
//
// The mana is restricted to instant and sorcery spells (the cast
// purpose and an any-of type restriction). The craft is graveyard-only
// ("cards", CR 702.167b) with a floor of four (CraftWithCardsOrMore).
// The Catalyst picks one material still in exile at random (ADR 0054's
// keyed stream) and offers a free cast of it as part of the
// resolution: Emergent Ultimatum's grant (grantFreeCasts) at flash
// timing, whose window closes on the caster's next pass and leaves an
// uncast card in exile, as the ruling says ("You can't wait to cast it
// later in the turn"). The card stays a material if it is not cast.
//
// No simplification.
const oreRichStalactiteOracleID = "e17d8c0e-16e7-477b-830d-0818cedcb354"

func init() {
	Register(Spec{
		OracleID:     oreRichStalactiteOracleID,
		Name:         "Ore-Rich Stalactite",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{R}",
			Label:        "Add {R}. Spend this mana only to cast an instant or sorcery spell",
			Restrictions: []string{ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
		}},
		Activated: []ActivatedAbility{
			Craft("Craft with four or more red instant and/or sorcery cards {3}{R}{R}", "{3}{R}{R}",
				CraftWithCardsOrMore(4, "R", "red", "instant", "sorcery")),
		},
	})

	Register(Spec{
		OracleID:     oreRichStalactiteOracleID + "#1",
		Name:         "Cosmium Catalyst",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, {T}: Choose an exiled card used to craft this artifact at random. You may cast that card without paying its mana cost.",
			Cost:    Plus(ManaCost("{1}{R}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  cosmiumCatalystCast,
		}},
	})
}

// cosmiumCatalystCast is the Catalyst's resolution: one material at
// random, and a free cast of it now.
func cosmiumCatalystCast(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range CraftMaterials(ctx) {
		ids = append(ids, c.InstanceID)
	}
	picked := g.ChooseAtRandomForEffect(randomDraw(ctx), ids, 1)
	if len(picked) == 0 {
		return nil
	}
	grantFreeCasts(g, item.Controller, item.SourceCardID, "Cosmium Catalyst", game.TimingFlash, game.LapseStaysInExile, picked)
	return nil
}
