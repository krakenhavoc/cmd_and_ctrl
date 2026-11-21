package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Redemption Choir — Creature — Vampire Cleric {2}{W}{W}, 3/3 (Commander,
// #5368):
//
//	"Lifelink
//	 Coven — Whenever this creature enters or attacks, if you control
//	 three or more creatures with different powers, return target
//	 permanent card with mana value 3 or less from your graveyard to the
//	 battlefield."
//
// One ability watching two events (Sun Titan's shape). Coven is an
// intervening "if" (CR 603.4): three distinct powers among the creatures
// you control, read through the layers (trekB48Coven), checked as it
// triggers and again as it resolves. The target is chosen as the trigger
// goes on the stack (CR 603.3d), so with no legal card the trigger is
// removed; the return is mandatory.
//
// No simplification.
func init() {
	coven := effects.CondBoard{
		Name: "you control three or more creatures with different powers",
		Test: func(ctx *effects.Context) bool { return effects.TrekB48Coven(ctx.Game, ctx.Controller(), ctx.Source()) },
	}
	effects.Register(effects.Spec{
		OracleID:        "90c62b27-027a-44a3-8d21-3c62701b18a8",
		Name:            "Redemption Choir",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(coven, effects.Targeting(
				effects.OnAny([]game.EventKind{game.EventETB, game.EventAttack},
					func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
						return effects.Self(ev, source, ch, g) && effects.TrekB48Coven(g, source.Controller, source.InstanceID)
					},
					"Redemption Choir — Coven: return target permanent card with mana value 3 or less from your graveyard to the battlefield",
					effects.HobB8ReturnTargetGraveyardCardToBattlefield),
				effects.TargetCardInGraveyard("target permanent card with mana value 3 or less from your graveyard",
					effects.YouOwn(), effects.Permanent(), effects.ManaValueLE(3)))),
		},
	})
}
