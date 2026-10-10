package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Altar of the Wretched // Wretched Bonemass — a transforming artifact
// (#2709, ADR 0137 and its 2026-10-10 amendment):
//
//	Altar of the Wretched — Artifact {2}{B}
//	  "When this artifact enters, you may sacrifice a nontoken
//	   creature. If you do, draw X cards, then mill X cards, where X is
//	   that creature's power.
//	   Craft with one or more creatures {2}{B}{B}
//	   {2}{B}: Return this card from your graveyard to your hand."
//	Wretched Bonemass — Creature — Skeleton Horror, */*
//	  "Wretched Bonemass's power and toughness are each equal to the
//	   total power of the exiled cards used to craft it.
//	   This creature has flying as long as an exiled card used to craft
//	   it has flying. The same is true for first strike, double strike,
//	   deathtouch, haste, hexproof, indestructible, lifelink, menace,
//	   protection, reach, trample, and vigilance."
//
// The enters trigger is a "you may" and then the sacrifice prompt, and
// the draw and mill are its continuation, so they happen only once a
// creature has really gone. X is the creature's power as it last
// existed on the battlefield (CR 608.2h), counters included. The
// graveyard ability is Blessed Ghoul's.
//
// The craft is an open count of creatures, mixed from both zones. The
// Bonemass reads CR 702.167c's link twice: a characteristic-defining
// ability for its power and toughness (the materials' total power,
// each at its power in exile with its own characteristic-defining
// ability applied, per the ruling), and a layer-6 static that gives it
// each listed keyword some material still in exile has. "Protection"
// is every protection ability a material has, with its quality.
//
// No simplification.
const altarOfTheWretchedOracleID = "c551f007-9740-4a2a-8ef5-7be1a7afe14b"

func init() {
	Register(Spec{
		OracleID:     altarOfTheWretchedOracleID,
		Name:         "Altar of the Wretched",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Altar of the Wretched — you may sacrifice a nontoken creature, then draw and mill X", altarOfTheWretchedEnters),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with one or more creatures {2}{B}{B}", "{2}{B}{B}", CraftWithOneOrMore("creature")),
			{
				Label:   "{2}{B}: Return this card from your graveyard to your hand.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{2}{B}"),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Effect:  returnThisCardFromYourGraveyardToYourHand,
			},
		},
	})

	Register(Spec{
		OracleID:     altarOfTheWretchedOracleID + "#1",
		Name:         "Wretched Bonemass",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7A_CDA,
				AppliesTo: selfOnly,
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := craftMaterialsTotalPower(g, source)
					c.Power, c.Toughness = n, n
				},
			},
			{
				Layer:     game.Layer6Ability,
				AppliesTo: selfOnly,
				Label:     "Has each listed keyword an exiled card used to craft it has",
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					for _, kw := range wretchedBonemassKeywords(craftMaterialsOf(g, source)) {
						c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
					}
				},
			},
		},
	})
}

// wretchedBonemassSharedKeywords is the Bonemass's printed list, in
// its order.
var wretchedBonemassSharedKeywords = []string{
	"flying", "first strike", "double strike", "deathtouch", "haste", "hexproof",
	"indestructible", "lifelink", "menace", "reach", "trample", "vigilance",
}

// wretchedBonemassKeywords is the keywords of the list the materials
// have, each once, plus every protection ability any of them has.
func wretchedBonemassKeywords(materials []game.Card) []string {
	have := map[string]bool{}
	var protections []string
	for _, m := range materials {
		for _, a := range m.Effective().Abilities {
			have[a] = true
			if strings.HasPrefix(a, game.KeywordProtection+" from ") && !keywordSliceContains(protections, a) {
				protections = append(protections, a)
			}
		}
	}
	var out []string
	for _, kw := range wretchedBonemassSharedKeywords {
		if have[kw] {
			out = append(out, kw)
		}
	}
	return append(out, protections...)
}

// altarOfTheWretchedEnters is the enters trigger: the optional
// sacrifice, then draw X and mill X for that creature's power.
func altarOfTheWretchedEnters(g *game.Game, item *game.StackItem) error {
	nontokenCreature := And(Creature(), Not(IsTokenPredicate()))
	if countControlled(g, item.Controller, func(c game.Card) bool { return c.IsCreature() && !IsToken(c) }) == 0 {
		return nil
	}
	return MayChoice{
		Question: "Altar of the Wretched — sacrifice a nontoken creature to draw and mill cards equal to its power?",
		YesLabel: "Sacrifice a creature",
		NoLabel:  "Don't",
		OnYes: func(ctx *Context) error {
			return ctx.Game.PlayerSacrificesThenForEffect(
				item.SourceCardID, item.Controller,
				sacrificeSpec("a nontoken creature", nontokenCreature),
				"Altar of the Wretched — sacrifice a nontoken creature",
				1,
				func(g *game.Game, sacrificed game.PromptedSacrifices) error {
					ids := sacrificed.Cards()
					if len(ids) == 0 {
						return nil
					}
					info, ok := g.LastKnownPermanentForEffect(ids[0])
					if !ok || info.Power <= 0 {
						return nil
					}
					ctx := NewContext(g, item)
					if err := (DrawCards{Player: item.Controller, N: info.Power}).Apply(ctx); err != nil {
						return err
					}
					return MillCards{Player: item.Controller, N: info.Power}.Apply(ctx)
				})
		},
	}.Apply(NewContext(g, item))
}
