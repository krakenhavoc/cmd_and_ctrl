package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Photon Blast Barrage — {X}{R}{R} Sorcery (Marvel Super Heroes, #4455):
//
//	"When you cast this spell, copy it X times. You may choose new targets
//	 for the copies.
//	 Photon Blast Barrage deals 1 damage to target creature."
//
// Replicate's shape (CR 702.56a) with the announced X as the count. The cast
// trigger goes on the stack above the spell and resolves even if the spell
// was countered in response (CR 603.2, 707.10): it carries the X the spell
// was cast with as data (the spell is read when the trigger is created), and
// copies the spell from last-known information if it has left the stack, as
// storm does. Each copy is offered new targets (CR 707.10c) and resolves for
// its own 1 damage.
//
// No simplification.
func init() {
	cast := WhenYouCastThisSpell("Photon Blast Barrage — copy it X times", replicateEffect)
	build := cast.Build
	cast.Build = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
		item := build(ev, source, lki, g)
		if spell := g.StackItemForEffect(source.InstanceID); spell != nil && spell.XValue > 0 {
			item.XValue = spell.XValue
		}
		return item
	}
	Register(Spec{
		OracleID:     "a5424e33-ff4e-4c82-8aa9-9d25ced7830a",
		Name:         "Photon Blast Barrage",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 1)),
		Triggered:    []game.TriggeredAbility{cast},
		OnResolve:    damageToFirstTarget(1),
	})
}
