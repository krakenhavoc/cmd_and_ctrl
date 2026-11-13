package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kangee, Sky Warden — Legendary Creature — Bird Wizard {3}{W}{U}, 3/3
// (Commander Legends, #5507):
//
//	"Flying, vigilance
//	 Whenever Kangee attacks, attacking creatures with flying get +2/+0
//	 until end of turn.
//	 Whenever Kangee blocks, blocking creatures with flying get +0/+2
//	 until end of turn."
//
// Two combat triggers with Inniaz's mass pump: the set is locked as each
// trigger resolves (CR 611.2c), so only the creatures that have flying and
// are attacking (or blocking) then get the bonus, and flying gained or
// lost afterwards changes nothing (the ruling). Kangee is one of them
// while it has flying. "Whenever Kangee blocks" triggers once however
// many creatures it blocks (CR 509.3a, selfBlocksOnce), and "blocking
// creatures" is every blocker, not only yours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6a344d2e-fbd7-4ff9-9e2a-97ae69d60c66",
		Name:            "Kangee, Sky Warden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Kangee, Sky Warden — attacking creatures with flying get +2/+0 until end of turn",
				Do(BoostUntilEOT{Match: And(AttackingCreature(), HasKeyword("flying")), Power: 2,
					Label: "Kangee, Sky Warden — attacking creatures with flying get +2/+0"})),
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return selfBlocksOnce(ev, source)
			}, "Kangee, Sky Warden — blocking creatures with flying get +0/+2 until end of turn",
				Do(BoostUntilEOT{Match: And(BlockingCreature(), HasKeyword("flying")), Toughness: 2,
					Label: "Kangee, Sky Warden — blocking creatures with flying get +0/+2"})),
		},
	})
}
