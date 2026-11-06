package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mongoose Lizard — Creature — Mongoose Lizard {4}{R}{R}, 5/6:
//
//	"Menace
//	 When this creature enters, it deals 1 damage to any target.
//	 Mountaincycling {2} ({2}, Discard this card: Search your library
//	 for a Mountain card, reveal it, put it into your hand, then
//	 shuffle.)"
//
// Mountaincycling is typecycling (CR 702.29e) for the Mountain subtype.
func init() {
	Register(Spec{
		OracleID:        "a9bae56c-9666-42e8-94ec-03b0a964b142",
		Name:            "Mongoose Lizard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Mongoose Lizard — 1 damage to any target",
				sourceDealsDamageToEachLegalTarget(1)), TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
		Activated: []ActivatedAbility{
			Typecycling("Mountaincycling", "{2}", "a Mountain card", IsLandWithSubtype("Mountain")),
		},
	})
}
