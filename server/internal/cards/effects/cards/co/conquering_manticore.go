package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conquering Manticore — Creature — Manticore {4}{R}{R}, 5/5 (Duel Decks:
// Heroes vs. Monsters, S82 old sets):
//
//	"Flying
//	 When this creature enters, gain control of target creature an opponent
//	 controls until end of turn. Untap that creature. It gains haste until end
//	 of turn."
//
// Zealous Conscripts' three steps in printed order, narrowed to a creature an
// opponent controls: a layer-2 until-end-of-turn control effect, an untap and a
// layer-6 haste grant, all on the target if it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "140b58b9-7e82-4feb-99d7-5e8b707b845c",
		Name:            "Conquering Manticore",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Conquering Manticore — gain control of target creature an opponent controls until end of turn, untap it, it gains haste",
				effects.S82OldB398ThreatenTarget("Conquering Manticore")),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
