package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Neko-Te — Artifact — Equipment {3} (Betrayers of Kamigawa, #5602):
//
//	"Whenever equipped creature deals damage to a creature, tap that creature.
//	 That creature doesn't untap during its controller's untap step for as long
//	 as this Equipment remains on the battlefield.
//	 Whenever equipped creature deals damage to a player, that player loses 1
//	 life.
//	 Equip {2}"
//
// Two triggers on the Equipment, one per damage event the equipped creature is
// the source of, combat or not. The first taps the creature that was dealt the
// damage and holds it for as long as the Neko-Te stays on the battlefield
// (DoesntUntapWhile with the ForAsLongAsOnBattlefield duration, which never
// begins if the Neko-Te has already gone, CR 611.2b); it stays held even if the
// creature is no longer equipped. The second is life LOSS, not damage, so
// prevention doesn't stop it. Equip {2} is the ordinary sorcery-speed equip.
//
// No simplification.
func init() {
	const tapLabel = "Neko-Te — tap that creature; it doesn't untap for as long as this Equipment remains on the battlefield"
	const loseLabel = "Neko-Te — that player loses 1 life"
	effects.Register(effects.Spec{
		OracleID:     "a1d2944e-6668-4157-aac5-54572759c05c",
		Name:         "Neko-Te",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, nekoTeEquippedDealtDamageToACreature, tapLabel,
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					id := item.Trigger.Event.Target
					if !effects.OnBattlefield(g, id) {
						return nil
					}
					if err := g.TapTargetForEffect(id); err != nil {
						return err
					}
					d, ok := effects.DurationWhileSourceRemains(ctx, ctx.Source())
					if !ok {
						return nil
					}
					return effects.DoesntUntapWhile{Targets: []uuid.UUID{id}, Duration: d}.Apply(ctx)
				}),
			effects.On(game.EventDealDamage, nekoTeEquippedDealtDamageToAPlayer, loseLabel,
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Trigger.Event.Target, -1)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// nekoTeEquippedDamage is the damage events the creature this Equipment is
// attached to is the source of.
func nekoTeEquippedDamage(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventDealDamage && ev.Amount > 0 && source.IsAttachedTo(ev.Source)
}

func nekoTeEquippedDealtDamageToACreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !nekoTeEquippedDamage(ev, source) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && effects.OnBattlefield(g, ev.Target) && c.IsCreature()
}

func nekoTeEquippedDealtDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return nekoTeEquippedDamage(ev, source) && g.PlayerByIDForEffect(ev.Target) != nil
}
