package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Costume Shop — Artifact — Attraction:
//
//	"Visit — You may put a sticker on a nonland permanent you own."
//
// A visit ability (CR 702.159a, ADR 0207 §5), whose effect is the engine's
// keyed sticker questions (ADR 0208 §4): which permanent, which sticker,
// and for a name sticker where it goes, its ticket cost paid as it goes on
// (CR 123.3c). The lights are the printing's (ADR 0207 §3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24c52448-03f2-43cd-9172-487384baeb18",
		Name:         "Costume Shop",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Visit("Costume Shop", Do(PutASticker{})),
		},
	})
}
