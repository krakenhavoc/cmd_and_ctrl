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
// equal to its last-known power (#1379).
//
// Declared simplification: a creature that has LEFT deals that damage
// as a plain source. The damage tail reads lifelink and deathtouch off
// the battlefield only (ADR 0056 Decision 2, step 4: durable LKI for
// damage sources is out of scope there; #1396), so a lifelinker killed in
// response gains its controller nothing. Weaker than printed, and only
// in that corner.
func init() {
	Register(Spec{
		OracleID:     "42fb1a1c-ab3d-4cdc-a6ff-a591f7481583",
		Name:         "Warstorm Surge",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If the creature has left the battlefield before the trigger resolves, it still deals damage equal to its last power, but without any lifelink or deathtouch it had."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			},
			Targets: TargetAny(),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Warstorm Surge — the creature deals damage equal to its power to any target",
					b06WarstormSurgeDamage)
			},
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
	return DealDamage{
		Source: ctx.Trigger().Object.ID,
		Target: item.Targets[0].ID,
		Amount: creature.Power,
	}.Apply(ctx)
}
