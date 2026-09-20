package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thran Power Suit — Artifact — Equipment for {2}:
//
//	"Equipped creature gets +1/+1 for each Aura and Equipment attached
//	 to it and has ward {2}. (Whenever equipped creature becomes the
//	 target of a spell or ability an opponent controls, counter it
//	 unless that player pays {2}.)
//	 Equip {2} ({2}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// The pump is a variable count rather than a constant, so it is a
// hand-written layer 7c static (the shared PumpAttachedPer helper
// counts things the SOURCE's controller controls — Blackblade
// Reforged's lands, Uril's Auras on itself — not things attached to
// the HOST, which is what this card counts). The count reads
// AttachedToSource-shaped state on every recompute the same way
// PumpAttachedPer does, so attaching or detaching a second Equipment
// mid-combat changes the bonus in the same pass with no bookkeeping.
// This Equipment counts itself: the moment it is attached it is one
// of the "Equipment attached to it" it is asking about.
//
// Ward {2} is the same grant Lavaspur Boots and WardAttached already
// model — a triggered ability the Equipment carries, watching for its
// host becoming a target, exactly as CR 702.21a requires and exactly
// as ADR 0038 says ward (unlike protection) needs no new targeting
// vocabulary.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3952c288-5633-4e40-abe6-18940d079644",
		Name:         "Thran Power Suit",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: AttachedToSource,
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				if g == nil || target == nil {
					return
				}
				n := 0
				for i := range g.Battlefield.Cards {
					att := &g.Battlefield.Cards[i]
					if att.IsAttachedTo(target.InstanceID) && (att.HasSubtype("Aura") || att.HasSubtype("Equipment")) {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			WardAttached(WardMana("{2}"), "Thran Power Suit — ward {2}"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
