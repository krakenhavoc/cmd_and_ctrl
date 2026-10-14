package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gisa, the Hellraiser — Legendary Creature — Human Warlock {3}{B}{B}, 4/4:
//
//	"Ward—{2}, Pay 2 life.
//	 Skeletons and Zombies you control get +1/+1 and have menace.
//	 Whenever you commit a crime, create two tapped 2/2 blue and black
//	 Zombie Rogue creature tokens. This ability triggers only once each
//	 turn. (Targeting opponents, anything they control, and/or cards in
//	 their graveyards is a crime.)"
//
// TribalAnthem and TribalKeywordGrant both read effective creature
// types, and neither excludes the source: Gisa benefits if another
// effect makes her a Skeleton or Zombie. Her crime trigger uses the
// once-per-turn stack tally, so a countered trigger still spends it.
// The tokens' colors and types are their printed characteristics;
// tapped is a property of this creation event.
//
// No simplification.
const gisaHellraiserCrimeLabel = "Gisa, the Hellraiser — create two tapped Zombie Rogue tokens"

func init() {
	tribe := TribeFilter{Tribes: []string{"Skeleton", "Zombie"}, YoursOnly: true}
	Register(Spec{
		OracleID:     "696363c9-e8a7-4c65-b65c-4a94891ec43d",
		Name:         "Gisa, the Hellraiser",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(tribe, 1, 1),
			TribalKeywordGrant(tribe, "menace"),
		},
		Triggered: []game.TriggeredAbility{
			Ward(WardManaAndLife("{2}", 2), "Gisa, the Hellraiser — ward {2}, pay 2 life"),
			TriggerWithPurpose(On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return YouCommittedCrime(ev, source) &&
					!b11TriggeredThisTurn(g, source.InstanceID, gisaHellraiserCrimeLabel)
			}, gisaHellraiserCrimeLabel, gisaHellraiserCreateTokens), game.Purpose{Tokens: 2}),
		},
	})
}

func gisaHellraiserCreateTokens(g *game.Game, item *game.StackItem) error {
	return g.CreateTokensThenForEffect(game.TokenCreation{
		Controller: item.Controller, Source: item.SourceCardID, Together: true,
		Groups: []game.TokenGroup{{Template: gisaHellraiserZombieRogueToken(), Count: 2, Entry: game.TokenEntryOptions{Tapped: true}}},
	}, nil)
}

func gisaHellraiserZombieRogueToken() game.Card {
	return game.Card{
		Name:      "Zombie Rogue",
		TypeLine:  "Token Creature — Zombie Rogue",
		Power:     2,
		Toughness: 2,
		Colors:    []string{"U", "B"},
	}
}
