package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	coven := CondBoard{
		Name: "you control three or more creatures with different powers",
		Test: func(ctx *Context) bool { return trekB48Coven(ctx.Game, ctx.Controller(), ctx.Source()) },
	}
	Register(Spec{
		OracleID:        "90c62b27-027a-44a3-8d21-3c62701b18a8",
		Name:            "Redemption Choir",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(coven, Targeting(
				OnAny([]game.EventKind{game.EventETB, game.EventAttack},
					func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
						return Self(ev, source, ch, g) && trekB48Coven(g, source.Controller, source.InstanceID)
					},
					"Redemption Choir — Coven: return target permanent card with mana value 3 or less from your graveyard to the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneBattlefield}.Apply(NewContext(g, item))
					}),
				TargetCardInGraveyard("target permanent card with mana value 3 or less from your graveyard",
					YouOwn(), Permanent(), ManaValueLE(3)))),
		},
	})
}
