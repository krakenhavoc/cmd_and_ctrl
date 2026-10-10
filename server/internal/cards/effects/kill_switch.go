package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kill Switch — Artifact for {3}:
//
//	"{2}, {T}: Tap all other artifacts. They don't untap during their controllers' untap steps for as long as this artifact remains tapped."
//
// ADR 0109 owner decision 4: Rust Tick's untap hold over a set. The
// artifacts tapped are the other artifacts on the battlefield as the
// ability resolves; each is held for as long as Kill Switch remains
// tapped (CR 611.2b). Kill Switch prints no untap opt-out, so it
// untaps in its controller's next untap step and the holds end there.
// In that same step the held artifacts its controller controls stay
// tapped, because the set of permanents that untap is fixed before
// anything untaps (CR 502.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0cd5eef6-6830-4c6d-8578-2726900c39a5",
		Name:         "Kill Switch",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Tap all other artifacts. They don't untap during their controllers' untap steps for as long as this artifact remains tapped.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				var others []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.InstanceID != item.SourceCardID && c.IsArtifact() {
						others = append(others, c.InstanceID)
					}
				}
				return TapAndHoldWhileThisRemainsTapped(NewContext(g, item), others)
			},
		}},
	})
}
