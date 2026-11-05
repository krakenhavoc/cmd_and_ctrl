package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Locutus, the One Who Speaks — Legendary Artifact Creature — Borg {W}{B},
// 1/1 (Star Trek, #4973):
//
//	"Artifact creatures you control get +1/+1.
//	 {W}{B}, Discard this card: It deals 2 damage to target creature. You
//	 gain 2 life."
//
// The first ability is a layer-7c anthem (b16Anthem) over every artifact
// creature its controller controls, Locutus included, read through the
// layered types. The second is Eiganjo, Seat of the Empire's channel shape
// with no keyword: an activated ability whose Zones is the hand and whose
// cost discards the card itself. "It" is the discarded card, so the damage
// source is the ability's source card. With its one target illegal on
// resolution the ability does nothing (CR 608.2b); otherwise the life gain
// waits for any prompt the damage asks (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46ba09a7-42e4-42cd-911c-71d0ade01180",
		Name:         "Locutus, the One Who Speaks",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(trekB25ArtifactCreaturesYouControl, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}{B}, Discard this card: It deals 2 damage to target creature. You gain 2 life.",
			Cost:    game.AbilityCost{Mana: "{W}{B}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect: Do(
				Instruction(func(ctx *Context) error {
					id := firstLegalCardTarget(ctx)
					if id == uuid.Nil {
						return nil
					}
					return DealDamage{Source: ctx.Item.SourceCardID, Target: id, Amount: 2}.Apply(ctx)
				}),
				GainLife{Amount: 2},
			),
		}},
	})
}
