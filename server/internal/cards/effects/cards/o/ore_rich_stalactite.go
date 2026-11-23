package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
// timing, offered during the resolution (#4250). The window closes on the
// caster's next pass only as the fallback, leaving an uncast card in exile,
// as the ruling says ("You can't wait to cast it later in the turn"). The card stays a material if it is not cast.
//
// No simplification.
const OreRichStalactiteOracleID = "e17d8c0e-16e7-477b-830d-0818cedcb354"

func init() {
	effects.Register(effects.Spec{
		OracleID:     OreRichStalactiteOracleID,
		Name:         "Ore-Rich Stalactite",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{R}",
			Label:        "Add {R}. Spend this mana only to cast an instant or sorcery spell",
			Restrictions: []string{effects.ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
		}},
		Activated: []effects.ActivatedAbility{
			effects.Craft("Craft with four or more red instant and/or sorcery cards {3}{R}{R}", "{3}{R}{R}",
				effects.CraftWithCardsOrMore(4, "R", "red", "instant", "sorcery")),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     OreRichStalactiteOracleID + "#1",
		Name:         "Cosmium Catalyst",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, {T}: Choose an exiled card used to craft this artifact at random. You may cast that card without paying its mana cost.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  cosmiumCatalystCast,
		}},
	})
}

// cosmiumCatalystCast is the Catalyst's resolution: one material at
// random, and a free cast of it now.
func cosmiumCatalystCast(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range effects.CraftMaterials(ctx) {
		ids = append(ids, c.InstanceID)
	}
	picked := g.ChooseAtRandomForEffect(effects.RandomDraw(ctx), ids, 1)
	if len(picked) == 0 {
		return nil
	}
	effects.GrantFreeCasts(g, item.Controller, item.SourceCardID, "Cosmium Catalyst", game.TimingFlash, game.LapseStaysInExile, picked)
	return nil
}
