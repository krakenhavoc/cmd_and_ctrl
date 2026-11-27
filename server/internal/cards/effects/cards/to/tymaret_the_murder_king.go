package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tymaret, the Murder King — Legendary Creature — Zombie Warrior {B}{R},
// 2/2 (Theros, S82 old sets):
//
//	"{1}{R}, Sacrifice another creature: Tymaret deals 2 damage to target
//	 player or planeswalker.
//	 {1}{B}, Sacrifice a creature: Return this card from your graveyard to
//	 your hand."
//
// The first ability is Barrage of Expendables' sacrifice outlet with Barbed
// Lightning's player-or-planeswalker target. The second works only from the
// graveyard (the ability's zone is the graveyard), where the sacrificed
// creature is necessarily another one; the return is Abzan Devotee's and does
// nothing if the card has left the graveyard by resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a029c61-b1de-427b-a8e8-36f3a23f001b",
		Name:         "Tymaret, the Murder King",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{R}, Sacrifice another creature: Tymaret deals 2 damage to target player or planeswalker.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Effect:  effects.DamageToFirstTarget(2),
			},
			{
				Label:   "{1}{B}, Sacrifice a creature: Return this card from your graveyard to your hand.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeACreature()),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
			},
		},
	})
}
