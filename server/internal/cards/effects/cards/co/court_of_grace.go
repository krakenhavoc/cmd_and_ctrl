package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f63c2438-27d4-449a-828f-f0a2ea86ff16",
		Name:         "Court of Grace",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Court of Grace"),
			effects.AtYourUpkeep("Court of Grace — create a 1/1 Spirit with flying, or a 4/4 Angel with flying if you're the monarch",
				courtOfGraceUpkeep),
		},
	})
}

// courtOfGraceUpkeep makes the Spirit, or the Angel for the monarch.
func courtOfGraceUpkeep(g *game.Game, item *game.StackItem) error {
	token := effects.TokenCard("1/1 white Spirit with flying")
	if effects.YoureTheMonarch(g, item.Controller) {
		token = effects.TokenCard("4/4 white Angel with flying")
	}
	return effects.CreateToken{Template: token, N: 1}.Apply(effects.NewContext(g, item))
}
