package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voice of Victory — Creature — Human Bard {1}{W}, 1/3:
//
//	"Mobilize 2 (Whenever this creature attacks, create two tapped and
//	 attacking 1/1 red Warrior creature tokens. Sacrifice them at the
//	 beginning of the next end step.)
//	 Your opponents can't cast spells during your turn."
//
// Mobilize is mobilize.go. The restriction is Grand Abolisher's spell
// half (OpponentsCantCastDuringYourTurn), derived from Voice being on
// the battlefield, so it lifts the moment it leaves. Abilities are
// untouched: an opponent can still activate and play lands.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	Register(Spec{
		OracleID:     "b12cbda1-ac09-49eb-96b3-a4df27bc37ea",
		Name:         "Voice of Victory",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Mobilize("Voice of Victory", 2)},
		CastRestrictions: []game.CastRestriction{
			OpponentsCantCastDuringYourTurn(
				"Voice of Victory — your opponents can't cast spells during your turn.", nil),
		},
	})
}
