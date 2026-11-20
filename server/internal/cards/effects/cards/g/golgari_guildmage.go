package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golgari Guildmage — Creature — Elf Shaman {B/G}{B/G}, 2/2 (Ravnica:
// Clue Edition, #5929):
//
//	"{4}{B}, Sacrifice a creature: Return target creature card from your
//	 graveyard to your hand.
//	 {4}{G}: Put a +1/+1 counter on target creature."
//
// Hell's Caretaker's sacrifice cost with Cauldron of Essence's graveyard
// target, returning to hand. The sacrificed creature is a cost, so it is
// already in the graveyard and may itself be the card returned (it
// cannot be, as a target is chosen before costs are paid, CR 601.2c:
// targets first, then costs). The counter ability is Amok's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46dad873-11f5-4269-8ec2-f126419fbd7d",
		Name:         "Golgari Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{4}{B}, Sacrifice a creature: Return target creature card from your graveyard to your hand.",
				Cost:    effects.Plus(effects.ManaCost("{4}{B}"), effects.SacrificeACreature()),
				Targets: effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
				Effect:  effects.ReturnFirstLegalGraveyardTargetToHand,
			},
			{
				Label:   "{4}{G}: Put a +1/+1 counter on target creature.",
				Cost:    effects.ManaCost("{4}{G}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.PutPlusOneCounterOnEachLegalTarget,
			},
		},
	})
}
