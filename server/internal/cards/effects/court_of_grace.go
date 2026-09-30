package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Court of Grace — Enchantment {2}{W}{W}:
//
//	"When this enchantment enters, you become the monarch.
//	 At the beginning of your upkeep, create a 1/1 white Spirit
//	 creature token with flying. If you're the monarch, create a 4/4
//	 white Angel creature token with flying instead."
//
// The white Court (#1722). The monarch comes from
// WhenThisEntersYouBecomeTheMonarch; the upkeep trigger reads
// YoureTheMonarch as it RESOLVES — "if you're the monarch, … instead"
// is a clause of the effect, not an intervening "if" (CR 603.4), so a
// Court whose controller lost the crown in response still makes the
// Spirit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f63c2438-27d4-449a-828f-f0a2ea86ff16",
		Name:         "Court of Grace",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Court of Grace"),
			AtYourUpkeep("Court of Grace — create a 1/1 Spirit with flying, or a 4/4 Angel with flying if you're the monarch",
				courtOfGraceUpkeep),
		},
	})
}

// courtOfGraceUpkeep makes the Spirit, or the Angel for the monarch.
func courtOfGraceUpkeep(g *game.Game, item *game.StackItem) error {
	token := TokenCard("1/1 white Spirit with flying")
	if YoureTheMonarch(g, item.Controller) {
		token = TokenCard("4/4 white Angel with flying")
	}
	return CreateToken{Template: token, N: 1}.Apply(NewContext(g, item))
}
