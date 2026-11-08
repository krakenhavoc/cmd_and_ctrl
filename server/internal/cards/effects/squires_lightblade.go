package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squire's Lightblade — Artifact — Equipment {W} (Edge of Eternities,
// #4461):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 That creature gains first strike until end of turn.
//	 Equipped creature gets +1/+0.
//	 Equip {3}"
//
// Twin Blades with first strike. The first strike goes to the target
// whether or not the attach happened: a Lightblade that left in response
// is a new object and attaches to nothing (CR 400.7), but "that
// creature" is still the trigger's legal target and still gains first
// strike until end of turn. A target gone by resolution gets neither
// (CR 608.2b). The +1/+0 and Equip {3} are the shared Equipment static
// and ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0ea2a099-7c81-4012-8148-4d377f2d64e1",
		Name:            "Squire's Lightblade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Squire's Lightblade — attach it to target creature you control; it gains first strike until end of turn", squiresLightbladeAttach),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

func squiresLightbladeAttach(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := g.AttachSourceForEffect(item, t); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike"},
			Label: "Squire's Lightblade — first strike until end of turn"}.Apply(ctx)
	}
	return nil
}
