package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dora Milaje Elite — Creature — Human Warrior {1}{W}, 2/2 (Marvel Super
// Heroes Commander, #4455):
//
//	"First strike
//	 When this creature enters, if an opponent controls more lands than
//	 you, create a tapped Vibranium token. (It's an artifact with
//	 indestructible and "{T}: Add {C}. This mana can't be spent to cast a
//	 nonartifact spell.")
//	 Sacrifice this creature: Legendary permanents you control gain
//	 indestructible until end of turn."
//
// The ETB is Knight of the White Orchid's intervening if (CR 603.4: read
// when the creature enters and again on resolution) over the shared
// b05OpponentControlsMoreLands count. The Vibranium token is the printed
// one (Powerstone's restricted colorless mana: ManaRestrictNotNonartifactSpell,
// which still pays for abilities). The sacrifice ability is Dauntless
// Escort's, over legendary permanents of any type; the set is fixed on
// resolution (CR 611.2c), so the sacrificed Elite is never in it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "406f1982-5d0c-4dba-9010-28b1c6a3602c",
		Name:            "Dora Milaje Elite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "an opponent controls more lands than you",
				Test: func(ctx *Context) bool { return b05OpponentControlsMoreLands(ctx.Game, ctx.Controller()) },
			}, On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b05OpponentControlsMoreLands(g, source.Controller)
			}, "Dora Milaje Elite — create a tapped Vibranium token",
				Do(CreateTokenAdvanced{Spec: Token(VibraniumToken()).EntersTapped(), N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Legendary permanents you control gain indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    game.AbilityCost{SacrificeSelf: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Match:    And(Legendary(), YouControl()),
					Keywords: []string{"indestructible"},
					Label:    "Dora Milaje Elite — indestructible",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// VibraniumToken is the Vibranium token as PRINTED.
func VibraniumToken() game.Card { return tokenFromCatalog(printedVibraniumToken) }

// printedVibraniumToken is the catalog's entry for Vibranium: a colorless
// artifact token with indestructible and a restricted colorless mana
// ability. Its slug is an on-disk identity and is never renamed.
func printedVibraniumToken() tokenTemplate {
	return tokenTemplate{
		Slug: "vibranium",
		Card: game.Card{
			Name:     "Vibranium",
			TypeLine: "Token Artifact — Vibranium",
			Keywords: []string{"indestructible"},
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:      true,
			Produced:     "{C}",
			Restrictions: []string{ManaRestrictNotNonartifactSpell},
			Label:        "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell",
		}},
		Text: "Indestructible\n{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.",
	}
}
