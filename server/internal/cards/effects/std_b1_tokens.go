package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b1_tokens.go — token templates for the Edge of Eternities
// leftovers, batches 1 and 2 (#4461).

// printedDroneToken is the Edge of Eternities Drone: a 1/1 colourless
// Drone artifact creature with flying and "This token can block only
// creatures with flying." (Desculpting Blast). Its slug is an on-disk
// identity: never renamed, never reused.
func printedDroneToken() tokenTemplate {
	return tokenTemplate{
		Slug: "drone-blocks-only-fliers",
		Card: game.Card{
			Name:      "Drone",
			TypeLine:  "Token Artifact Creature — Drone",
			Power:     1,
			Toughness: 1,
			Keywords:  []string{"flying"},
		},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
		Text: "Flying\nThis token can block only creatures with flying.",
	}
}

// DroneToken is the Drone's card value for CreateToken.
func DroneToken() game.Card { return tokenFromCatalog(printedDroneToken) }
