package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	body := effects.Do(effects.CreateToken{Template: effects.EldraziSpawnToken(), N: 1})
	effects.Register(effects.Spec{
		OracleID:        "7cf5c3dd-5c07-4eaf-8227-6f4e7f9b134d",
		Name:            "Drownyard Lurker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell(label+" (cast)", body),
			effects.InGraveyard(effects.On(game.EventCycle, effects.Self, label+" (cycled)", body)),
		},
	})
}
