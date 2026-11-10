package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Assimilation Aegis — Artifact — Equipment {1}{W}{U} (Outlaws of Thunder
// Junction, #4468; engine seam #5399):
//
//	"When this Equipment enters, exile up to one target creature until
//	 this Equipment leaves the battlefield.
//	 Whenever this Equipment becomes attached to a creature, for as long
//	 as this Equipment remains attached to it, that creature becomes a
//	 copy of a creature card exiled with this Equipment.
//	 Equip {2}"
//
// The enters trigger is ExileUntil{ThisLeaves} (CR 610.3) with the
// CR 607.2a link stamped (Linked), so the card comes back when this
// Aegis leaves, and until then is "exiled with" this object by that row.
// An Aegis gone before the trigger resolves exiles nothing (CR 610.3b).
//
// The attach trigger (ThisBecameAttachedToACreature) reads that row's
// cards still in exile, keeps the creature cards (the ruling: a
// noncreature card exiled with it can't be chosen), and makes the
// creature it became attached to a copy of the one there is, or of the
// one its controller picks among several (the ruling: not necessarily
// the same one each time). None is no copy. The copy is a layer-1 copy
// effect (CR 707.2, CR 613.2a) of the card as printed, for as long as
// this Aegis stays attached to that creature (CopyWhileSourceAttached,
// CR 611.2b): equipping another creature ends it on the first and the
// new attach triggers again for the second; an Aegis moved before the
// trigger resolves copies nothing. Becoming a copy is not entering or
// leaving, so nothing triggers on it; counters and other effects on the
// creature still apply.
//
// No simplification.
const assimilationAegisOracleID = "6d39d218-254a-4886-907a-89df605d7ce4"

const assimilationAegisExileKey = "Assimilation Aegis — exile up to one target creature until it leaves the battlefield"

func init() {
	Register(Spec{
		OracleID:     assimilationAegisOracleID,
		Name:         "Assimilation Aegis",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters(assimilationAegisExileKey, assimilationAegisExile),
				UpToOneTargetCreature("up to one target creature"),
			),
			On(game.EventAttach, ThisBecameAttachedToACreature,
				"Assimilation Aegis — the creature becomes a copy of a creature card exiled with it",
				assimilationAegisCopy),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// assimilationAegisExile is the enters trigger: the chosen creature, if
// any is still legal, exiled until this Aegis leaves and linked to it.
func assimilationAegisExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return ExileUntil{
				Target:     t.ID,
				ThisLeaves: true,
				Linked:     true,
				Label:      "Assimilation Aegis — the exiled card returns when Assimilation Aegis leaves the battlefield",
			}.Apply(ctx)
		}
	}
	return nil
}

// assimilationAegisCopy is the attach trigger: "that creature" is the
// host the trigger saw.
func assimilationAegisCopy(g *game.Game, item *game.StackItem) error {
	return BecomeCopyOfExiledWith{
		Target:   item.Trigger.Event.Target,
		Row:      assimilationAegisExileKey,
		Match:    func(c game.Card) bool { return c.IsCreature() },
		Duration: CopyWhileSourceAttached,
		Question: "Assimilation Aegis — choose a creature card exiled with it; the equipped creature becomes a copy of it",
		Label:    "Assimilation Aegis",
	}.Apply(NewContext(g, item))
}
