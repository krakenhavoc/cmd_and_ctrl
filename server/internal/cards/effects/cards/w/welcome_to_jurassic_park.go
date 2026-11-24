package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Welcome to . . . // Jurassic Park — Enchantment — Saga {1}{G}{G} //
// Legendary Land (Jurassic World Collection):
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I — For each opponent, up to one target noncreature artifact they
//	     control becomes a 0/4 Wall artifact creature with defender for as
//	     long as you control this Saga.
//	 II — Create a 3/3 green Dinosaur creature token with trample. It gains
//	      haste until end of turn.
//	 III — Destroy all Walls. Exile this Saga, then return it to the
//	       battlefield transformed under your control.
//	 //
//	 Jurassic Park — Each Dinosaur card in your graveyard has escape. The
//	 escape cost is equal to the card's mana cost plus exile three other
//	 cards from your graveyard. (You may cast cards from your graveyard for
//	 their escape cost.)
//	 {T}: Add {G} for each Dinosaur you control."
//
// Chapter I is Riptide Gearhulk's one clause per opponent and Vault 87's
// "for as long as you control this Saga" (CR 611.2b): a layer 4 type, a
// layer 7b base P/T and a layer 6 keyword, at one timestamp, ending when
// the Saga leaves or changes control. Chapter III exiles and returns the
// Saga transformed (CR 714.3c, 712.14). The land's escape is Underworld
// Breach's standing permission, narrowed to Dinosaur cards (CR 702.138a).
//
// No simplification.
func init() {
	const name = "Welcome to . . ."
	chapterOne := effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1,
		"for each opponent, up to one target noncreature artifact they control becomes a 0/4 Wall artifact creature with defender for as long as you control this Saga"),
		effects.PerOpponentClauses(game.MaxPlayers-1, "noncreature artifact", effects.Artifact(), effects.Noncreature()),
		welcomeToWallUp)
	effects.Register(effects.Spec{
		OracleID:     WelcomeToJurassicParkOracleID,
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chapterOne,
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "create a 3/3 green Dinosaur creature token with trample; it gains haste until end of turn"),
				welcomeToDinosaur),
			effects.TriggerWithPurpose(effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "destroy all Walls; exile this Saga, then return it transformed"),
				welcomeToDestroyWallsThenTransform),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     WelcomeToJurassicParkOracleID + "#1",
		Name:         "Jurassic Park",
		Completeness: effects.CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:                    game.ZoneGraveyard,
			Filter:                  game.PermissionFilter{CardInZone: true, Subtype: "Dinosaur"},
			AltCostKey:              "escape",
			ExileOtherFromGraveyard: 3,
			Label:                   "Escape — its mana cost, Exile three other cards from your graveyard (Jurassic Park)",
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedPerPermanent("G", effects.MatchCreatureSubtype("Dinosaur")),
			Label:          "Add {G} for each Dinosaur you control",
		}},
	})
}

// WelcomeToJurassicParkOracleID is the card's oracle ID; the back face
// registers under it plus "#1" (game.CatalogKey).
const WelcomeToJurassicParkOracleID = "a4b37d16-95b3-4143-a0b2-ad9f2aba91f8"

func welcomeToWallUp(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	d, ok := effects.DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.ScopedEffectFor{
			Target: t.ID,
			Mods: []game.Mod{
				game.AddTypesMod("Artifact", "Creature"),
				game.AddSubtypesMod(ctx.TextSubtypes("Wall")...),
				game.SetBasePowerMod(0),
				game.SetBaseToughnessMod(4),
				game.AddKeywordsMod("defender"),
			},
			Duration: d,
			Label:    "Welcome to . . . — a 0/4 Wall artifact creature with defender for as long as you control this Saga",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

func welcomeToDinosaur(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ids, err := g.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(effects.TokenCard("3/3 green Dinosaur with trample")), 1, game.TokenEntryOptions{})
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"},
			Label: "Welcome to . . . — the Dinosaur gains haste until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

func welcomeToDestroyWallsThenTransform(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.DestroyAllMatching{Match: effects.HasSubtype(ctx.TextSubtype("Wall"))}).Apply(ctx); err != nil {
		return err
	}
	return effects.ChapterExileAndReturnTransformed(g, item)
}
