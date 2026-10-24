package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mobilize.go — mobilize (CR 702.181): "Mobilize N (Whenever this
// creature attacks, create N tapped and attacking 1/1 red Warrior
// creature tokens. Sacrifice them at the beginning of the next end
// step.)"
//
// A triggered ability on the attack declaration. The Warriors were
// never declared as attackers (CR 508.4), so they trigger nothing that
// reads "whenever a creature attacks"; their controller chooses which
// player, planeswalker or battle each attacks, the way Hero of
// Bladehold's Soldiers do, with the player the source attacked offered
// first. One delayed trigger sacrifices exactly the tokens this
// resolution made, at the beginning of the next end step.
//
//	Triggered: []game.TriggeredAbility{Mobilize("Voice of Victory", 2)},
//
// Append-only: a new mobilize card calls this.
func Mobilize(name string, n int) game.TriggeredAbility {
	label := name + " — mobilize " + strconv.Itoa(n)
	return WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
		return CreateTokensAttackingYourChoice{
			Template:    TokenCard("1/1 red Warrior"),
			RuleDefined: true,
			N:           n,
			Tapped:      true,
			Prefer:      triggerAttackTarget(item),
			Label:       name,
			Tail:        game.KeyedTail{Ref: mobilizeSacrificeAtEndStep},
		}.Apply(NewContext(g, item))
	})
}

// mobilizeSacrificeAtEndStep is the Warriors' tail, keyed (#4294): the
// end step trigger that sacrifices them.
var mobilizeSacrificeAtEndStep = AttackingTokensThen("attacking-tokens/mobilize-sacrifice-the-warriors",
	func(ctx *Context, created []uuid.UUID, _ game.KeyedTail) error {
		if len(created) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label: "Mobilize — sacrifice the Warriors",
			Cards: created,
			Body:  sacrificeListedCardsBody,
		}.Apply(ctx)
	})
