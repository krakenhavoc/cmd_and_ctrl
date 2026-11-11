package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rose Room Treasurer — Creature — Ogre Warrior {3}{R}, 4/3 (#5368):
//
//	"Alliance — Whenever another creature you control enters, create a
//	 Treasure token if this is the first or second time this ability has
//	 resolved this turn. Otherwise, you may pay {X}. When you do, this
//	 creature deals X damage to any target."
//
// Which sentence applies is Vito's read: the engine's per-object tally of
// how many times this ability has resolved this turn, counting the
// resolution in progress. The third and later resolutions are Electro's
// "you may pay {X}. When you do" (MayPayX): X is chosen as the trigger
// resolves (CR 608.2d), paying nothing declines, and paying makes a CR
// 603.12 reflexive trigger that picks any target as it goes on the stack
// and carries X.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d73f02e0-53fa-4437-ab3a-89a9a1db27c5",
		Name:         "Rose Room Treasurer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b13AnotherCreatureYouControlEntered(ev, source, g)
			}, roseRoomTreasurerLabel, roseRoomTreasurerResolves),
		},
	})
}

const roseRoomTreasurerLabel = "Rose Room Treasurer — create a Treasure token, or you may pay {X} to deal X damage to any target"

func roseRoomTreasurerResolves(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if g.ResolvedThisTurn(item.SourceCardID, roseRoomTreasurerLabel) <= 2 {
		return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
	}
	return MayPayX{
		Cost:  "{X}",
		Label: "Rose Room Treasurer",
		Buys:  "deal X damage to any target",
		Goal:  XAsHighAsYouCan,
		Unit:  game.PayAmountDamage,
		Then:  roseRoomTreasurerPaid,
	}.Apply(ctx)
}

// roseRoomTreasurerPaid is the paid half, keyed: the X paid rides the prompt.
var roseRoomTreasurerPaid = PayUnlessBranches("pay-unless/rose-room-treasurer-damage",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Rose Room Treasurer — this creature deals X damage to any target",
			Body:   roseRoomTreasurerDamageBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

// roseRoomTreasurerDamageBody is the reflexive trigger: Params.Amount
// damage from this creature to the chosen target.
var roseRoomTreasurerDamageBody = game.ReflexiveBody("rose-room-treasurer/damage", breechesBlastEffect,
	constTargets(func() *game.TargetSpec { return TargetAny() }))
