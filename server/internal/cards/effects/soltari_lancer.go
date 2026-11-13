package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soltari Lancer — Creature — Soltari Knight {2}{W}, 2/2
// (Tempest, #5628):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 This creature has first strike as long as it's attacking."
//
// Shadow is a printed keyword (CR 702.28b). The first strike is Kitesail
// Corsair's "as long as it's attacking" grant with first strike in
// place of flying: a layer-6 static on itself (CR 613.1f) that reads its
// attacking status live, so it has first strike from the moment it is
// declared as an attacker until it leaves combat, and never while it
// blocks. It has it as the combat damage step begins, which is when
// first strike is read (CR 702.7b).
//
// No simplification.
func init() {
	firstStrike := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
	}, "first strike")
	firstStrike.Label = "This creature has first strike as long as it's attacking."
	firstStrike.DependsOnAttackingStatus = true
	Register(Spec{
		OracleID:        "d2cbd71a-841b-4ff2-a794-247a030e0233",
		Name:            "Soltari Lancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Static:          []game.StaticAbility{firstStrike},
	})
}
