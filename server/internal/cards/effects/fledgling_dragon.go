package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fledgling Dragon — Creature — Dragon {2}{R}{R}, 2/2 (Modern Horizons 3):
//
//	"Flying
//	 Threshold — As long as there are seven or more cards in your graveyard,
//	 this creature gets +3/+3 and has "{R}: This creature gets +1/+0 until end of
//	 turn.""
//
// Both halves read the graveyard live on every pass: the pump is a layer 7c
// self-modifier, the ability a bundle granted to this creature only while
// threshold holds.
func init() {
	const (
		grant = "fledgling-dragon/firebreathing"
		text  = "{R}: This creature gets +1/+0 until end of turn."
	)
	Register(Spec{
		OracleID:        "be007706-3da7-4dd3-84db-dbb5d8831bba",
		Name:            "Fledgling Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []AbilityGrant{{
			Key: grant,
			Activated: []ActivatedAbility{{
				Label:   text,
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
				Cost:    ManaCost("{R}"),
				Effect:  thisGetsUntilEndOfTurn(1, 0, "Fledgling Dragon — +1/+0 until end of turn"),
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{
			fraThresholdSelfPT(3, 3),
			Labelled(GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && fraThresholdMet(g, source.Controller)
			}, grant), "Threshold — As long as there are seven or more cards in your graveyard, this creature has \""+text+"\""),
		},
	})
}
