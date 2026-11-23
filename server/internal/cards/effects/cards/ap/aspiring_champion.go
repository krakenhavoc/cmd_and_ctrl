package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aspiring Champion — Creature — Astartes Warrior {3}{R}, 3/3 (Warhammer
// 40,000 Commander, S82 GAP batch 193, #5948):
//
//	"Menace
//	 Ruinous Ascension — When this creature deals combat damage to a player,
//	 sacrifice it. If you do, reveal cards from the top of your library until
//	 you reveal a creature card. Put that card onto the battlefield, then shuffle
//	 the rest into your library. If that creature is a Demon, it deals damage
//	 equal to its power to each opponent."
//
// Sacrifice-then (CR 701.21): the reveal runs only when the Champion really
// left the battlefield. The reveal is Shifting Shadow's, the rest is shuffled
// into the library, and the damage is dealt by the creature that arrived, as
// many points as its power as it deals them; a library with no creature card
// is revealed whole and shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d343c116-b33a-4ed6-ad15-d7270e739b04",
		Name:            "Aspiring Champion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Aspiring Champion — sacrifice it; if you do, reveal until a creature card and put it onto the battlefield; a Demon deals damage equal to its power to each opponent",
				effects.SacrificeThisThen("if-it-went/aspiring-champion-ruinous-ascension", effects.GapB193AspiringChampionAscend)),
		},
	})
}
