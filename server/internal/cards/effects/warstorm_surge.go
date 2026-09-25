package effects

import (
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
// The entering creature is the trigger's event object, and its power
// is read when the trigger RESOLVES (CR 608.2h), which is the printed
// timing: a Giant Growth in response makes the Surge hit harder. If the
// creature has left the battlefield by then it still deals the damage,
// equal to its last-known power (#1379), and with the lifelink and
// deathtouch it had as it last existed (#1396): a lifelinker killed in
// response still gains its controller the life.
//
// The damage names its source by OBJECT (DealDamage.SourceObject), not
// by card. A creature blinked in response is a new object the trigger
// never named (CR 400.7), so the damage is the departed creature's,
// with the departed creature's keywords — not whatever the new one has.
func init() {
	Register(Spec{
		OracleID:     "42fb1a1c-ab3d-4cdc-a6ff-a591f7481583",
		Name:         "Warstorm Surge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			Key:     "Warstorm Surge — the creature deals damage equal to its power to any target",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			},
			Targets: TargetAny(),
			Effect:  b06WarstormSurgeDamage,
		}},
	})
}

// b06WarstormSurgeDamage is the trigger body: the entered creature —
// live, or by last-known information — deals damage equal to its power.
func b06WarstormSurgeDamage(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	creature, ok := ctx.TriggeringPermanent()
	if !ok {
		return nil
	}
	ref := ctx.Trigger().Object.Ref()
	return DealDamage{
		SourceObject: &ref,
		Target:       item.Targets[0].ID,
		Amount:       creature.Power,
	}.Apply(ctx)
}
