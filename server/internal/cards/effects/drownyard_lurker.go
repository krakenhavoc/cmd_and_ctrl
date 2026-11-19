package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drownyard Lurker — Creature — Eldrazi Trilobite {7}, 7/7 (Modern
// Horizons 3):
//
//	"Vigilance
//	 When you cast or cycle Drownyard Lurker, create a 0/1 colorless
//	 Eldrazi Spawn creature token with "Sacrifice this token: Add {C}."
//	 Cycling {2}{U} ({2}{U}, Discard this card: Draw a card.)"
//
// One printed ability that triggers on two events is two registered
// triggers with the one body: the cast trigger is World Breaker's
// (WhenYouCastThisSpell, which resolves above the spell, so the Spawn
// arrives even if the Lurker is countered) and the cycle trigger is
// Agonasaur Rex's (it fires from the graveyard, CR 702.29c). Cycling is
// the shared cost-and-draw, and the cycle trigger sees the discard.
//
// No simplification.
func init() {
	const label = "Drownyard Lurker — create a 0/1 colorless Eldrazi Spawn"
	body := Do(CreateToken{Template: EldraziSpawnToken(), N: 1})
	Register(Spec{
		OracleID:        "7cf5c3dd-5c07-4eaf-8227-6f4e7f9b134d",
		Name:            "Drownyard Lurker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []ActivatedAbility{Cycling("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell(label+" (cast)", body),
			InGraveyard(On(game.EventCycle, Self, label+" (cycled)", body)),
		},
	})
}
