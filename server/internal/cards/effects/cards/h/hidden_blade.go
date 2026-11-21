package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hidden Blade — Artifact — Equipment {2} (Assassin's Creed, #5948):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 If that creature is an Assassin, it gains deathtouch until end of turn.
//	 Equipped creature gets +1/+0 and has first strike.
//	 Equip {2}"
//
// Celestial Armor's shape: the entry trigger targets on the stack, and "that
// creature" is the target, not the equipped creature, so the deathtouch does
// not depend on the attach. The Assassin check is read as the trigger
// resolves, through the layered view of the creature (CR 608.2c); a target
// that has gone gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "37ec9b95-8c9a-4dd5-9ff2-3d6527984126",
		Name:            "Hidden Blade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("first strike"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Hidden Blade — attach it to target creature you control; an Assassin gains deathtouch", hiddenBladeAttach),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// hiddenBladeAttach attaches the Equipment to the still-legal target, then
// gives that creature deathtouch until end of turn if it is an Assassin,
// whether or not the attach happened.
func hiddenBladeAttach(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := effects.AttachSourceToTarget(g, item); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || !c.HasSubtype(ctx.TextSubtype("Assassin")) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Hidden Blade — deathtouch"}.Apply(effects.NewContext(g, item))
}
