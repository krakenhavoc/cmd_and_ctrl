package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// T-45 Power Armor — Artifact — Equipment {2}:
//
//	"When this Equipment enters, you get {E}{E} (two energy counters).
//	 Equipped creature gets +3/+3 and doesn't untap during its
//	 controller's untap step.
//	 At the beginning of your upkeep, you may pay {E}. If you do, untap
//	 equipped creature, then put your choice of a menace, trample, or
//	 lifelink counter on it.
//	 Equip {3}"
//
// ADR 0129 §3 (#1995): the upkeep payment is made as the trigger
// resolves (CR 118.12). "Equipped creature" is read then: the creature
// the Armor is attached to as the energy is paid, untapped, then given
// the chosen keyword counter. Unattached, paying does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "370aced9-d8bc-4abe-b648-4c62b5aee5ba",
		Name:         "T-45 Power Armor",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Static:       []game.StaticAbility{PumpAttached(3, 3)},
		UntapStepRestrictions: []game.UntapStepRestriction{
			{Label: "equipped creature doesn't untap", Restricts: AttachedToSource},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("T-45 Power Armor", 2),
			AtYourUpkeep("T-45 Power Armor — you may pay {E}",
				mayPayEnergyThen("T-45 Power Armor", 1, "untap equipped creature and give it a counter",
					func(g *game.Game, item *game.StackItem) error {
						equipped := equippedCreatureOf(g, item.SourceCardID)
						if equipped == uuid.Nil {
							return nil
						}
						ctx := NewContext(g, item)
						if err := (UntapTarget{Target: equipped}).Apply(ctx); err != nil {
							return err
						}
						return putChosenCounterOn(ctx, item.Controller, equipped,
							"T-45 Power Armor — choose a counter for the equipped creature",
							[]string{game.CounterMenace, game.CounterTrample, game.CounterLifelink})
					})),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

// equippedCreatureOf is the creature the Equipment `source` is attached
// to now, or uuid.Nil when it is unattached or gone.
func equippedCreatureOf(g *game.Game, source uuid.UUID) uuid.UUID {
	eq, ok := g.LookupCardForEffect(source)
	if !ok || !onBattlefield(g, source) || eq.AttachedTo.Kind != game.TargetCard {
		return uuid.Nil
	}
	return eq.AttachedTo.ID
}
