package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thornbite Staff — Kindred Artifact — Shaman Equipment {2}:
//
//	"Equipped creature has "{2}, {T}: This creature deals 1 damage to
//	 any target" and "Whenever a creature dies, untap this creature."
//	 Whenever a Shaman creature enters, you may attach this Equipment
//	 to it.
//	 Equip {4}"
//
// Two granted abilities (ADR 0093), one activated and one triggered,
// declared as two bundles because the card prints two quoted
// abilities. Both are the EQUIPPED creature's: it deals the damage,
// it taps, its controller activates the ping and controls the untap
// trigger, and CR 302.6 keeps a creature that entered this turn from
// pinging. Moving the Staff moves both.
//
// "Whenever a creature dies" includes the equipped creature's own
// death — a leaves-the-battlefield ability looks back (CR 603.10a) —
// and untapping a creature in a graveyard does nothing, so the
// resolution checks it is still on the battlefield.
//
// "A Shaman creature" is ANY player's: CR 301.5c lets an Equipment be
// attached to a creature its controller does not control, and only
// the equip ACTIVATION is limited to your own creatures. The trigger is
// a "you may", so nobody is made to.
//
// No simplification.
const (
	thornbiteStaffPing  = "thornbite-staff/ping"
	thornbiteStaffUntap = "thornbite-staff/untap"
)

func init() {
	Register(Spec{
		OracleID:     "dae4815e-9025-4993-ab46-52a3f1a7219e",
		Name:         "Thornbite Staff",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: thornbiteStaffPing,
				Activated: []ActivatedAbility{{
					Label:   "{2}, {T}: This creature deals 1 damage to any target",
					Cost:    Plus(ManaCost("{2}"), TapCost()),
					Targets: TargetAny(),
					Effect:  sourceDealsOneToFirstTarget,
				}},
				Text: "{2}, {T}: This creature deals 1 damage to any target.",
			},
			{
				Key: thornbiteStaffUntap,
				Triggered: []game.TriggeredAbility{
					On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
						_, died := diedCreature(ev, g)
						return died
					}, "Thornbite Staff — untap this creature", b31UntapSelf),
				},
				Text: "Whenever a creature dies, untap this creature.",
			},
		},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(thornbiteStaffPing, thornbiteStaffUntap)},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: shamanCreatureEntered,
			Key:       "Thornbite Staff — attach to the Shaman",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return attachSourceTo(g, item, item.Trigger.Event.CardID)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Thornbite Staff — attach it to the Shaman?"},
		}},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}

// shamanCreatureEntered is "whenever a Shaman creature enters" — any
// player's.
func shamanCreatureEntered(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype("Shaman")
}

// attachSourceTo attaches the ability's source Equipment to `host`.
func attachSourceTo(g *game.Game, item *game.StackItem, host uuid.UUID) error {
	return g.AttachSourceForEffect(item, game.TargetRef{Kind: game.TargetCard, ID: host})
}
