package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Katilda, Dawnhart Prime — Legendary Creature — Human Warlock {G}{W},
// 1/1 (Midnight Hunt Commander, #5497):
//
//	"Protection from Werewolves
//	 Human creatures you control have '{T}: Add one mana of any of this
//	 creature's colors.'
//	 {4}{G}{W}, {T}: Put a +1/+1 counter on each creature you control."
//
// The second line is Tazri's grant of a colour-reading mana ability
// (ADR 0093; tazriProduced reads the colours of the creature that has the
// ability as it is activated, through the layers), without Tazri's spend
// restriction or activation condition. It is given to Human creatures
// you control by effective subtype, Katilda included, and a colourless
// Human produces nothing. Protection from Werewolves is printed data
// (Katilda, Dawnhart Martyr's shape for a creature type). The last line
// is Action News Crew's counter on each creature, read as it resolves.
//
// No simplification.
const katildaPrimeGrant = "katilda-dawnhart-prime/tap-for-colors"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "6c728edd-ab67-44ce-b19d-aa98dfce68bb",
		Name:            "Katilda, Dawnhart Prime",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from Werewolves"},
		Grants: []effects.AbilityGrant{{
			Key: katildaPrimeGrant,
			Mana: []effects.ManaAbility{{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Label:        "{T}: Add one mana of any of this creature's colors.",
				ProducedFunc: effects.TazriProduced,
			}},
			Text: "{T}: Add one mana of any of this creature's colors.",
		}},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Human"))
			}, katildaPrimeGrant),
				"Human creatures you control have \"{T}: Add one mana of any of this creature's colors.\""),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}{W}, {T}: Put a +1/+1 counter on each creature you control.",
			Cost:    game.AbilityCost{Tap: true, Mana: "{4}{G}{W}"},
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect:  effects.B11PutCounterOnEachCreatureYouControl,
		}},
	})
}
