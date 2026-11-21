package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horned Loch-Whale // Lagoon Breach — Creature — Whale {4}{U}{U}, 6/6 //
// Instant — Adventure {1}{U} (Wilds of Eldraine, #4471):
//
//	Horned Loch-Whale — "Flash
//	 Ward {2}
//	 This creature enters tapped unless it's your turn."
//	Lagoon Breach — "The owner of target attacking creature you don't
//	 control puts it on their choice of the top or bottom of their library.
//	 (Then exile this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID and
// "#1"). The enters-tapped is a replacement read as the creature enters
// (the entering player's turn, not a later tap), the way the slowlands
// read their condition. Ward is the shared trigger. Lagoon Breach is Dire
// Downdraft's placement, the OWNER's choice, over a target attacking
// creature an opponent controls.
//
// No simplification.
const hornedLochWhaleOracleID = "adca3929-4ecc-45bc-932e-604bcc32550a"

func init() {
	effects.Register(effects.Spec{
		OracleID:        hornedLochWhaleOracleID,
		Name:            "Horned Loch-Whale",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTappedUnless(effects.IsActivePlayer)},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Horned Loch-Whale — ward {2}"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     hornedLochWhaleOracleID + "#1",
		Name:         "Lagoon Breach",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking creature you don't control", effects.IsAttackingCreature(), effects.OpponentControls()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OwnerTucksFirstTargetTopOrBottom(ctx, "Lagoon Breach — put it on your choice of the top or bottom of your library")
		},
	})
}
