package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mjölnir, Storm Hammer — Legendary Artifact — Equipment {4} (Assassin's
// Creed, #5948):
//
//	"When Mjölnir enters, attach it to target legendary creature you control.
//	 Whenever equipped creature attacks, tap target creature defending player
//	 controls and put a stun counter on it. Then Mjölnir deals damage to each
//	 opponent equal to the number of tapped creatures that opponent controls.
//	 Equip {4}"
//
// The entry trigger is Baseball Bat's attach over a legendary creature. The
// attack trigger is on the Equipment, keyed on the creature it is attached to
// and targeting a creature the defending player controls (Icingdeath's
// clause). As it resolves the creature is tapped and gets a stun counter (the
// engine's untap rule, CR 122.1d), then each opponent, the defender and any
// other, takes damage from Mjölnir equal to the tapped creatures they control
// at that moment, the one just tapped included; an opponent with none takes
// none. If the target is gone the tap and counter do nothing, but the damage
// is still dealt (CR 608.2b: the instruction is not about the target).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f497e9a5-0ad3-4aec-8422-14be8e0dba4c",
		Name:         "Mjölnir, Storm Hammer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Mjölnir — attach it to target legendary creature you control", AttachSourceToTarget),
				PermanentYouControl("target legendary creature you control", Creature(), Legendary()),
			),
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attachedCreatureAttacked(ev, source)
				},
				TargetsFrom: mshB2TargetCreatureTheAttackerDefends,
				Key:         "Mjölnir — tap target creature defending player controls and put a stun counter on it; damage to each opponent",
				Effect:      mjolnirAttack,
			},
		},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}

func mjolnirAttack(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if id, ok := b16FirstLegalTargetCard(ctx); ok {
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		if err := (AddCounter{Target: id, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	var opponents []uuid.UUID
	opponents = append(opponents, ctx.Opponents()...)
	// One printed instruction, so one instance of damage (CR 615.8).
	return g.DamageInstanceForEffect(func() error {
		for _, opp := range opponents {
			n := countControlled(g, opp, func(c game.Card) bool { return c.IsCreature() && c.Tapped })
			if n <= 0 {
				continue
			}
			if err := (DealDamage{Source: ctx.Source(), Target: opp, Amount: n}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
