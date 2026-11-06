package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magic Pot — Artifact Creature — Goblin Construct {3}, 1/4 (Final
// Fantasy, #4462):
//
//	"When this creature dies, create a Treasure token.
//	 {2}, {T}: Exile target card from a graveyard."
//
// Common Crook's death trigger and Soul-Guide Lantern's graveyard exile
// (ExileFirstTarget): any graveyard, any card, and nothing if the card
// left in response (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "caed9999-74b8-43e3-a319-fa505c04c71c",
		Name:         "Magic Pot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Magic Pot — create a Treasure token", Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Exile target card from a graveyard.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect:  ExileFirstTarget,
		}},
	})
}
