package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illusionist's Bracers — Artifact — Equipment {2}:
//
//	"Whenever an ability of equipped creature is activated, if it
//	 isn't a mana ability, copy that ability. You may choose new
//	 targets for the copy.
//	 Equip {3}"
//
// The ability-copy family (#1223) already has this exact shape:
// Rings of Brighthearth watches "whenever YOU activate an ability"
// and copies it for {2}; this watches "whenever an ability of
// EQUIPPED CREATURE is activated" and copies it for free. Both read
// game.EventActivateAbility, which only a CR 602 activation emits —
// a mana ability announces the different EventManaAbilityActivated
// kind and never reaches this trigger, which is what "if it isn't a
// mana ability" turns out to mean in practice (Rings' own comment).
//
// The activation's stack item id is read back at RESOLUTION off
// ctx.Trigger().Event.StackItemID, not captured in Build, mirroring
// ringsCopyActivatedAbility exactly — by the time this trigger
// resolves the copied ability is still on the stack underneath it
// (an open trigger stops priority from passing), so CopyAbility has
// something to copy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d1d78af-7982-419d-b9be-2bf4c149d97d",
		Name:         "Illusionist's Bracers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventActivateAbility, illusionistsBracersWatchesEquippedCreature,
				"Illusionist's Bracers — copy that ability", illusionistsBracersCopyActivatedAbility),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}

func illusionistsBracersWatchesEquippedCreature(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source.IsAttachedTo(ev.CardID)
}

func illusionistsBracersCopyActivatedAbility(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ability := ctx.Trigger().Event.StackItemID
	if ability == uuid.Nil {
		return nil
	}
	return CopyAbility{
		ItemID:           ability,
		Controller:       item.Controller,
		ChooseNewTargets: true,
	}.Apply(ctx)
}
