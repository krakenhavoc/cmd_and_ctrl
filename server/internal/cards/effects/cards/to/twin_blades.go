package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twin Blades — Artifact — Equipment {2}{R} (Avatar: The Last Airbender,
// #4459):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you
//	 control. That creature gains double strike until end of turn.
//	 Equipped creature gets +1/+1.
//	 Equip {2}"
//
// The entry trigger is Novel Nunchaku's attach. The double strike goes to
// the target creature whether or not the attach happened: an Equipment
// that left in response is a new object and attaches to nothing (CR
// 400.7), but "that creature" is still the trigger's legal target and
// still gains double strike until end of turn. A target gone by resolution gets neither (CR
// 608.2b). The +1/+1 and Equip {2} are the shared Equipment statics and
// ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "88b076a4-46ff-4a41-95a5-2821f4d8944d",
		Name:            "Twin Blades",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Twin Blades — attach it to target creature you control; it gains double strike until end of turn", twinBladesAttach),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

func twinBladesAttach(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := g.AttachSourceForEffect(item, t); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"double strike"},
			Label: "Twin Blades — double strike until end of turn"}.Apply(ctx)
	}
	return nil
}
