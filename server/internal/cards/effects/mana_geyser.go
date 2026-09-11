package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mana Geyser — Sorcery for {3}{R}{R}:
//
//	"Add {R} for each tapped land your opponents control."
//
// The big multiplayer ritual: at a four-player table in the middle
// of a turn cycle this routinely adds eight or more red mana for
// five. It is a Commander card specifically because the count is
// over OPPONENTS' lands, which only exist in numbers when there are
// three of them.
//
// Two halves, both expressible today:
//
//   - The count is a plain battlefield walk. "Tapped land your
//     opponents control" is Tapped && IsLand() && Controller !=
//     caster, over live seated players — an eliminated player's
//     permanents are already off the battlefield, so no extra
//     filtering is needed.
//   - The mana is AddMana, the spell-side mana primitive that landed
//     with the roadmap's batch 01. The batch-02 triage (#295) filed
//     this card under "mana pipeline — no tracking issue yet", which
//     predates that primitive; the derived COUNT was never the
//     blocker, since it is ordinary Go.
//
// The count is taken on resolution, not on announce — a land that
// untapped in response makes the Geyser smaller, which is the
// printed behaviour and the reason the card is cast in an opponent's
// end step rather than your own upkeep.
//
// Zero tapped opponent lands adds nothing rather than erroring:
// AddMana no-ops on an empty Produced string.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "a8dba58b-2956-492e-ae30-49db2ae68e53",
		Name:     "Mana Geyser",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Tapped && c.IsLand() && c.Controller != item.Controller {
					n++
				}
			}
			if n == 0 {
				return nil
			}
			return AddMana{
				Player:   item.Controller,
				Produced: strings.Repeat("{R}", n),
			}.Apply(ctx)
		},
	})
}
