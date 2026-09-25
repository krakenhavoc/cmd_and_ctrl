package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hero's Blade — Artifact — Equipment for {2}:
//
//	"Equipped creature gets +3/+2.
//	 Whenever a legendary creature you control enters, you may attach
//	 this Equipment to it.
//	 Equip {4} ({4}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// The pump and the equip ability are the ordinary shapes. The extra
// clause is an optional triggered ability, not a second equip — it
// attaches for free, off the stack resolution of the trigger rather
// than an activation, so it costs nothing and does not require
// sorcery speed (the trigger can go on the stack, and its Effect can
// resolve, at any time a legendary creature enters under the
// controller's control).
//
// The entering creature's ID is read at resolution off the item's
// carried trigger context (item.Trigger.Event.CardID, #1223) rather
// than captured at trigger time: undo restores a cloned game, so the
// Effect closure reads everything off the item it is handed (ADR
// 0018). AttachSourceForEffect does the rest, including the two quiet
// ways this can do nothing instead of erroring — the entering creature
// left the battlefield before the trigger resolved, or this Equipment
// did — exactly as the equip ability's own resolution does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e6bcd25d-39b6-4619-8a27-9f3d87f4a17b",
		Name:         "Hero's Blade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(3, 2)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.HasSupertype("Legendary")
			},
			Key: "Hero's Blade — attach to the legendary creature",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.AttachSourceForEffect(item, game.TargetRef{Kind: game.TargetCard, ID: item.Trigger.Event.CardID})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Hero's Blade — attach to it?"},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{4}"),
		},
	})
}
