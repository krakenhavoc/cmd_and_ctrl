package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Warstorm Surge — Enchantment {5}{R} (EDHREC rank 721):
//
//	"Whenever a creature you control enters, it deals damage equal to
//	 its power to any target."
//
// Impact Tremors that scales with the body and aims. A targeted
// trigger: the controller picks any target when the trigger fires,
// and the entering creature — not the enchantment — is the damage
// source, so lifelink on it gains life and a prevention effect
// naming it applies.
//
// The entering creature's ID is captured in Build by value (a copied
// uuid, not a pointer — the same posture Arcane Denial takes), and its
// power is read when the trigger RESOLVES, which is the printed
// timing: a Giant Growth in response makes the Surge hit harder.
//
// Sandbox simplification: if the creature has left the battlefield by
// resolution, the trigger does nothing. Printed, it uses last known
// information; the harvester's LKI reaches only the trigger's own
// source, not another card. Weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "42fb1a1c-ab3d-4cdc-a6ff-a591f7481583",
		Name:         "Warstorm Surge",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If the creature has left the battlefield before the trigger resolves, no damage is dealt."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			},
			Targets: TargetAny(),
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Warstorm Surge — the creature deals damage equal to its power to any target",
					b06WarstormSurgeDamage(ev.CardID))
			},
		}},
	})
}

// b06WarstormSurgeDamage builds the trigger body for one entering
// creature.
func b06WarstormSurgeDamage(creature uuid.UUID) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if len(item.Targets) == 0 {
			return nil
		}
		c, ok := g.LookupCardForEffect(creature)
		if !ok {
			return nil
		}
		if z := g.FindCardZoneForEffect(creature); z == nil || z.Kind != game.ZoneBattlefield {
			return nil
		}
		return DealDamage{
			Source: creature,
			Target: item.Targets[0].ID,
			Amount: c.CurrentPower(),
		}.Apply(NewContext(g, item))
	}
}
