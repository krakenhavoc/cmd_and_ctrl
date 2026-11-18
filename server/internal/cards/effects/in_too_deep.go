package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const inTooDeepGrant = "in-too-deep/clue-draw"

// In Too Deep — Enchantment — Aura {U}{U}:
//
//	"Split second
//	 Enchant creature, planeswalker, or Clue
//	 Enchanted permanent is a colorless Clue artifact with '{2}, Sacrifice
//	 this artifact: Draw a card' and loses all other abilities. (It's no
//	 longer a creature or planeswalker.)"
//
// Imprisoned in the Moon's three layers: layer 4 makes it an artifact and
// nothing else (a Clue by subtype, losing the creature or planeswalker
// type, CR 205.1a), layer 5 colorless, layer 6 removes every ability and
// grants the Clue's draw ability as one effect (ADR 0093 Decision 3). The
// permanent keeps its counters, so a planeswalker's loyalty is inert. The
// Clue's ability is the enchanted permanent's, so its controller pays {2}
// and sacrifices it, and the Aura falls off with it. Split second is the
// printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2829ae08-9964-4679-80e8-46d5d33d10b9",
		Name:            "In Too Deep",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Targets: TargetPermanent("enchant creature, planeswalker, or Clue",
			Or(Creature(), Planeswalker(), HasSubtype("Clue"))),
		Grants: []AbilityGrant{{
			Key: inTooDeepGrant,
			Activated: []ActivatedAbility{{
				Label:   "{2}, Sacrifice this artifact: Draw a card",
				Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			}},
			Text: "{2}, Sacrifice this artifact: Draw a card",
		}},
		Static: []game.StaticAbility{
			SetAttachedTypes([]string{"Artifact"}, []string{"Clue"}),
			SetAttachedColors(),
			inTooDeepAbilities(),
		},
	})
}

// inTooDeepAbilities is "loses all other abilities" and "has '{2},
// Sacrifice this artifact: Draw a card'" as the ONE layer-6 effect.
func inTooDeepAbilities() game.StaticAbility {
	s := LoseAllAbilities()
	s.GrantAbilities = []string{inTooDeepGrant}
	return s
}
