package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skystinger — Creature — Insect Warrior {2}{G}, 3/3 (Edge of Eternities,
// #4461):
//
//	"Reach
//	 Whenever this creature blocks a creature with flying, this creature
//	 gets +5/+0 until end of turn."
//
// The block trigger is Giant Shark's EventBlock watch with this creature
// as the blocker: one event per (blocker, attacker) pair, so it is
// CR 509.3b's "blocks a creature", judged per attacker. The attacker's
// flying is read through the layers as the block is declared, so a
// creature that lost flying is not one. The pump is the source's own
// (thisGetsUntilEndOfTurn) and ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6bb7419b-e085-4616-8e5c-4ed08cb2b4c1",
		Name:            "Skystinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, skystingerBlockedAFlier,
				"Skystinger — +5/+0 until end of turn",
				thisGetsUntilEndOfTurn(5, 0, "Skystinger — +5/+0 until end of turn")),
		},
	})
}

// skystingerBlockedAFlier is "this creature blocks a creature with
// flying": the block pair names this creature as the blocker and an
// attacker that has flying.
func skystingerBlockedAFlier(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID != source.InstanceID {
		return false
	}
	attacker, ok := g.LookupCardForEffect(ev.Target)
	return ok && game.HasKeyword(&attacker, "flying")
}
