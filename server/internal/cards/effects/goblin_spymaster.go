package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Spymaster — Creature — Goblin Rogue {2}{R}, 2/1 (Commander 2016,
// #5537):
//
//	"First strike
//	 At the beginning of each opponent's end step, that player creates a
//	 1/1 red Goblin creature token with "Creatures you control attack
//	 each combat if able.""
//
// "That player" is the opponent whose end step it is (ByAnOpponent on
// the end-step event), and the token is theirs. Its ability is the
// token's own static, Goblin Rabblemaster's attack requirement over the
// creatures the token's controller controls, the token included; it
// works only while that player controls it. A creature that is tapped,
// or can't attack, or would have a cost to pay to attack, is not forced
// (CR 508.1d), as the ruling says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ec08df88-3c13-4940-a090-0c502b9b12f1",
		Name:            "Goblin Spymaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, ByAnOpponent,
				"Goblin Spymaster — that player creates a 1/1 red Goblin with \"Creatures you control attack each combat if able.\"",
				goblinSpymasterToken),
		},
	})
}

func goblinSpymasterToken(g *game.Game, item *game.StackItem) error {
	player := triggeringActor(item)
	if player == uuid.Nil {
		return nil
	}
	return CreateToken{Controller: player, Template: SpymasterGoblinToken(), N: 1}.Apply(NewContext(g, item))
}

// SpymasterGoblinToken is Goblin Spymaster's 1/1 red Goblin with
// "Creatures you control attack each combat if able."
func SpymasterGoblinToken() game.Card { return tokenFromCatalog(printedSpymasterGoblinToken) }

func printedSpymasterGoblinToken() tokenTemplate {
	return tokenTemplate{
		Slug: "goblin-creatures-attack-each-combat",
		Card: game.Card{
			Name:      "Goblin",
			TypeLine:  "Token Creature — Goblin",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"R"},
		},
		Static: []game.StaticAbility{
			Labelled(AttacksEachCombatWhere(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller
			}), "Creatures you control attack each combat if able."),
		},
		Text: "Creatures you control attack each combat if able.",
	}
}
