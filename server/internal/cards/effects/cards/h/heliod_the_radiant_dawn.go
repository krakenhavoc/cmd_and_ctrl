package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heliod, the Radiant Dawn // Heliod, the Warped Eclipse — {2}{W}{W}
// Legendary Enchantment Creature — God 4/4 // Legendary Enchantment
// Creature — Phyrexian God 4/6 (March of the Machine):
//
//	Front: "When Heliod enters, return target enchantment card that isn't a
//	        God from your graveyard to your hand.
//	        {3}{U/P}: Transform Heliod. Activate only as a sorcery."
//	Back:  "You may cast spells as though they had flash.
//	        Spells you cast cost {1} less to cast for each card your
//	        opponents have drawn this turn."
//
// The reduction counts every card an opponent of Heliod's controller has
// drawn this turn (Game.DrawnThisTurn, appended only by an actual draw and
// never pruned, so a card drawn and then discarded still counts, and a
// card put into a hand without being drawn does not). It lowers generic
// mana only, as every cost reduction does (CR 601.2f).
//
// No simplification.
func init() {
	const oracle = "c27c6ba9-8d72-4101-b3a5-1e8d81755568"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Heliod, the Radiant Dawn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Heliod, the Radiant Dawn — return target enchantment card that isn't a God from your graveyard to your hand",
				effects.S82Mom16ReturnTargetToHand),
				effects.TargetCardInGraveyard("target enchantment card that isn't a God from your graveyard", effects.YouOwn(), effects.Enchantment(), effects.Not(effects.HasSubtype("God")))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{U/P}: Transform Heliod. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{U/P}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Heliod, the Warped Eclipse",
		Completeness: effects.CompletenessFull,
		CastTimings:  []game.CastTimingRule{effects.CastAsThoughFlash()},
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.S82Mom16CardsOpponentsDrewThisTurn,
				"Spells you cast cost {1} less to cast for each card your opponents have drawn this turn.",
				effects.YourSpell()),
		},
	})
}
