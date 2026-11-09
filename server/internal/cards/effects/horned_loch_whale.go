package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        hornedLochWhaleOracleID,
		Name:            "Horned Loch-Whale",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements:    []game.ReplacementEffect{SelfEntersTappedUnless(isActivePlayer)},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Horned Loch-Whale — ward {2}"),
		},
	})
	Register(Spec{
		OracleID:     hornedLochWhaleOracleID + "#1",
		Name:         "Lagoon Breach",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target attacking creature you don't control", attackingCreature(), OpponentControls()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ownerTucksFirstTargetTopOrBottom(ctx, "Lagoon Breach — put it on your choice of the top or bottom of your library")
		},
	})
}
