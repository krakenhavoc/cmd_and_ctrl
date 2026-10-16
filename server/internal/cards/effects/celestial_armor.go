package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Celestial Armor — Artifact — Equipment {2}{W} (the Squall deck
// request, #3918):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you
//	 control. That creature gains hexproof and indestructible until end
//	 of turn.
//	 Equipped creature gets +2/+0 and has flying.
//	 Equip {3}{W}"
//
// Maul of the Skyclaves with flash and a protective rider. The entry
// trigger targets on the stack, so an opponent can answer the creature
// in response; a target that has gone gets nothing (CR 608.2b).
//
// "That creature" is the target, not the equipped creature: the grant
// does not depend on the attach. If the Armor is destroyed in response,
// the attach does nothing (CR 701.3b) and the still-legal creature
// still gains hexproof and indestructible, which is what the rulings
// say. The grant happens after the attach, so hexproof given by the
// trigger never stops its own attach (the attach does not target).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4f1b68d7-7d16-4bb1-af5f-f1708b23ca73",
		Name:            "Celestial Armor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Celestial Armor — attach it to target creature you control; it gains hexproof and indestructible",
					celestialArmorAttachAndProtect),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}{W}"),
		},
	})
}

// celestialArmorAttachAndProtect attaches the Armor to the still-legal
// target, then gives that creature hexproof and indestructible until end
// of turn whether or not the attach happened.
func celestialArmorAttachAndProtect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := AttachSourceToTarget(g, item); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"hexproof", "indestructible"},
		Label:    "Celestial Armor — hexproof and indestructible",
	}.Apply(NewContext(g, item))
}
