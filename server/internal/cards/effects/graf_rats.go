package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Graf Rats — Creature — Rat {1}{B}, 2/1 (Innistrad Remastered, S82 GAP,
// #5948):
//
//	"At the beginning of combat on your turn, if you both own and control
//	 this creature and a creature named Midnight Scavengers, exile them,
//	 then meld them into Chittering Host."
//
// The printed meld ability (CR 701.42, ADR 0145): Titania's upkeep meld
// at the beginning of combat. The condition is an intervening "if"
// (CR 603.4), checked as the trigger fires and again as it resolves;
// MeldWith rechecks that both are still owned and controlled. The
// combined back face is Chittering Host.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "852edc00-9d98-427c-9869-7d0c8c121118",
		Name:         "Graf Rats",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(condYouOwnAndControlThisAndANamed{partner: "Midnight Scavengers", partnerType: "creature"},
				AtBeginningOfYourCombat("Graf Rats — exile this creature and Midnight Scavengers, then meld them into Chittering Host",
					Do(MeldWith{Partner: "Midnight Scavengers", PartnerType: "creature"}))),
		},
	})
}
