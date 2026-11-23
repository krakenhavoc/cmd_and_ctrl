package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assault Suit — Artifact — Equipment {4} (Commander Masters):
//
//	"Equipped creature gets +2/+2, has haste, can't attack you or
//	 planeswalkers you control, and can't be sacrificed.
//	 At the beginning of each opponent's upkeep, you may have that player
//	 gain control of equipped creature until end of turn. If you do, untap
//	 it.
//	 Equip {3}"
//
// The pump and haste are the usual attached statics. "Can't attack you or
// planeswalkers you control" is the Vows' layer-6 bar naming the Equipment's
// controller (cmdr2B19BarAttackingSourcesController, CR 508.1c), and "can't
// be sacrificed" is the CantBeSacrificed flag (#3275) written onto the same
// creature in the same layer-6 pass. The upkeep trigger is the Equipment's
// own: the controller may have the upkeep's player (read off the event's
// actor) gain control of the creature until end of turn (CR 611.2b), and if
// they do it untaps. Nothing happens if the Equipment is no longer attached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3b6e114-7bc8-4890-97d9-edbeea1f425f",
		Name:         "Assault Suit",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("haste"),
			{
				Label:     "Equipped creature can't attack you or planeswalkers you control and can't be sacrificed.",
				Layer:     game.Layer6Ability,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					effects.Cmdr2B19BarAttackingSourcesController(c, source)
					c.CantBeSacrificed = true
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.AtEachOpponentsUpkeep("Assault Suit — that player gains control of equipped creature until end of turn; untap it", assaultSuitLend),
				"Assault Suit — have that player gain control of equipped creature until end of turn?"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}

func assaultSuitLend(g *game.Game, item *game.StackItem) error {
	host := effects.AttachedHostFor(g, item.SourceCardID)
	if host == nil || item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if err := (effects.GainControl{
		Target:     host.InstanceID,
		Controller: item.Trigger.Event.Actor,
		Duration:   g.UntilEndOfTurnDuration(),
		Label:      "Assault Suit — gain control of equipped creature until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	return effects.UntapTarget{Target: host.InstanceID}.Apply(ctx)
}
