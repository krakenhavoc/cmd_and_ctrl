package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treacherous Werewolf — Creature — Werewolf Minion {2}{B}, 2/2 (Judgment,
// #5611):
//
//	"Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +2/+2 and has "When this creature dies,
//	 you lose 4 life.""
//
// Fledgling Dragon's threshold pair: a layer 7c self-modifier and a layer 6
// self-grant of the dies trigger, both read live.
//
// No simplification.
func init() {
	grant, static := effects.OldB622ThresholdSelfDiesLoseLife("Treacherous Werewolf", "treacherous-werewolf/dies-lose-4", 4)
	effects.Register(effects.Spec{
		OracleID:     "59681998-c395-4d37-a419-2a78cf894506",
		Name:         "Treacherous Werewolf",
		Completeness: effects.CompletenessFull,
		Grants:       []effects.AbilityGrant{grant},
		Static:       []game.StaticAbility{effects.FraThresholdSelfPT(2, 2), static},
	})
}
