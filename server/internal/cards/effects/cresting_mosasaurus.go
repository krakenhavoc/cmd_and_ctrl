package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cresting Mosasaurus — Creature — Dinosaur {6}{U}{U}, 4/8:
//
//	"Emerge {6}{U} (You may cast this spell by sacrificing a creature and
//	 paying the emerge cost reduced by that creature's mana value.)
//	 When this creature enters, if you cast it, return each non-Dinosaur
//	 creature to its owner's hand."
//
// Emerge is the shared alternative cost (ADR 0135 §4). "If you cast it"
// is a CR 603.4 intervening if: the permanent came from the stack, and its
// cast record (CR 400.7d) names the player who controls it now as its
// caster (ADR 0104), so a Mosasaurus put onto the battlefield any other
// way, or one whose spell was stolen on the stack, bounces nothing. The
// bounce is one simultaneous return of every creature that is not a
// Dinosaur, the Mosasaurus itself staying; its Purpose says so, a
// partial bounce sweep (ADR 0126 §6), so the bot prices it as one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "bf245433-8e3b-47d7-b628-3cdae93c384d",
		Name:             "Cresting Mosasaurus",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Emerge("{6}{U}")},
		// ADR 0126 §6: the enters trigger is a bounce sweep that spares
		// Dinosaurs.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, crestingMosasaurusYouCastIt),
				"Cresting Mosasaurus — return each non-Dinosaur creature to its owner's hand",
				Do(BounceAllMatching{Match: And(Creature(), Not(HasSubtype("Dinosaur")))})),
		},
	})
}

// crestingMosasaurusYouCastIt is the enter trigger's "if you cast it",
// the catalog's one reading of it (b16EnteredFromStack).
func crestingMosasaurusYouCastIt(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && b16EnteredFromStack(g, source.InstanceID)
}
