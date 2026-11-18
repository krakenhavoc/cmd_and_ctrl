package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Okiba Reckoner Raid // Nezumi Road Captain — a transforming Saga
// (Kamigawa: Neon Dynasty, #5493, S82).
//
// Front face, Enchantment — Saga {B}:
//
//	"I, II — Each opponent loses 1 life and you gain 1 life.
//	 III — Exile this Saga, then return it to the battlefield transformed
//	       under your control."
//
// Back face, Enchantment Creature — Rat Rogue, 2/2:
//
//	"Menace
//	 Vehicles you control have menace."
//
// The Vehicle grant is a layer-6 static over Vehicles the Captain's
// controller controls, creature or not.
//
// No simplification.
const okibaReckonerRaidOracleID = "54257132-c67a-4c19-bada-08e7025c2929"

func init() {
	const name = "Okiba Reckoner Raid"
	drain := func(n int) game.TriggeredAbility {
		return TriggerWithPurpose(ChapterTrigger(n, SagaChapterLabel(name, n, "each opponent loses 1 life and you gain 1 life"),
			Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
			game.Purpose{LifeGain: 1})
	}
	Register(Spec{
		OracleID:     okibaReckonerRaidOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			drain(1), drain(2),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})
	Register(Spec{
		OracleID:        okibaReckonerRaidOracleID + "#1",
		Name:            "Nezumi Road Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{{
			Label: "Vehicles you control have menace.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.HasSubtype("Vehicle")
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				appendKeywordsTo(c, []string{"menace"})
			},
		}},
	})
}
