package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nulldrifter — Creature — Eldrazi Elemental {7}, 4/4:
//
//	"When you cast this spell, draw two cards.
//	 Flying
//	 Annihilator 1
//	 Evoke {2}{U}"
//
// The draw is a cast trigger, so it resolves above the spell and
// happens whether Nulldrifter is cast for its mana cost or evoked, and
// even if it is countered (the 2024-06-07 ruling). Evoke is the shared
// alternative cost: an evoked Nulldrifter is sacrificed as it enters.
// Flying and annihilator 1 are canonical keyword tokens; the engine
// makes annihilator's attack trigger (ADR 0113 §2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "ef8d9c3a-5860-45bf-8205-14e2df9a6fc8",
		Name:             "Nulldrifter",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying", "annihilator 1"},
		AlternativeCosts: []game.AlternativeCost{Evoke("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Nulldrifter — draw two cards", Do(DrawCards{N: 2})),
		},
	})
}
