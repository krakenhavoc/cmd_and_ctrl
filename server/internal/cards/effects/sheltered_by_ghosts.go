package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sheltered by Ghosts — Enchantment — Aura {1}{W}:
//
//	"Enchant creature you control
//	 When this Aura enters, exile target nonland permanent an opponent
//	 controls until this Aura leaves the battlefield.
//	 Enchanted creature gets +1/+0 and has lifelink and ward {2}."
//
// Ossification's two halves — an entry trigger that exiles and a leave
// trigger that returns what that exile took, keyed on one shared label
// (CR 610.3) — with the ward as the Aura's own trigger watching its
// host becoming a target (WardAttached).
//
// One guard Ossification does not need: the exile is refused unless
// the Aura is still the same permanent on the battlefield as the
// trigger resolves (CR 610.3c — an "until" effect whose duration has
// already ended never starts). Without it an Aura removed in response
// would exile the permanent for good, because the leave trigger it
// would have used has already gone by. The object is named on the
// item's Params, stamped when the trigger is put on the stack.
//
// The permanent returns under its OWNER's control (CR 610.3), as a new
// object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d13fc657-c6fc-4394-bc70-691050550226",
		Name:         "Sheltered by Ghosts",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("lifelink"),
		},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Targets: TargetPermanent("target nonland permanent an opponent controls",
					Nonland(), OpponentControls()),
				Key: shelteredByGhostsExileLabel,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, shelteredByGhostsExileLabel)
					item.Params.Object = game.ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}
					return item
				},
				Effect: shelteredByGhostsExile,
			},
			On(game.EventLTB, Self, "Sheltered by Ghosts — return the exiled card",
				b41ReturnCardsExiledWithToTheBattlefield(shelteredByGhostsExileLabel)),
			WardAttached(WardMana("{2}"), "Sheltered by Ghosts — ward {2}"),
		},
	})
}

// shelteredByGhostsExileLabel is the exile trigger's stack label; the
// "exiled with" record keys on it, so both halves share the constant.
const shelteredByGhostsExileLabel = "Sheltered by Ghosts — exile target nonland permanent an opponent controls"

// shelteredByGhostsExile exiles the chosen permanent unless the Aura
// that asked has already left the battlefield.
func shelteredByGhostsExile(g *game.Game, item *game.StackItem) error {
	src, ok := g.LookupCardForEffect(item.Params.Object.ID)
	if !ok || src.ObjectEpoch != item.Params.Object.Epoch || !onBattlefield(g, src.InstanceID) {
		return nil
	}
	return b27ExileChosenTarget(g, item)
}
