package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jiwari, the Earth Aflame — Legendary Creature — Spirit {3}{R}{R}, 3/3
// (Saviors of Kamigawa, S82 OLD b584, #5601):
//
//	"{X}{R}, {T}: Jiwari deals X damage to target creature without flying.
//	 Channel — {X}{R}{R}{R}, Discard this card: It deals X damage to each
//	 creature without flying."
//
// Arashi, the Sky Asunder with the flying test inverted: the target is
// judged at announce and again at resolution (CR 608.2b), X is announced in
// the hand-zone cost (CR 602.2b) and the channel damage is dealt by the
// discarded card as last known.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03199a9e-3a52-42af-8027-50b455a054d4",
		Name:         "Jiwari, the Earth Aflame",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{X}{R}, {T}: Jiwari deals X damage to target creature without flying.",
				Cost:    effects.Plus(effects.ManaCost("{X}{R}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature without flying", effects.WithoutKeyword("flying")),
				Effect:  effects.OldB584SourceDealsXToTarget(),
			},
			{
				Label:   "Channel — {X}{R}{R}{R}, Discard this card: It deals X damage to each creature without flying.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
				Cost:    effects.Plus(effects.ManaCost("{X}{R}{R}{R}"), effects.DiscardThis()),
				Zones:   []game.ZoneKind{game.ZoneHand},
				Effect:  effects.OldB584SourceDealsXToEach(effects.WithoutKeyword("flying")),
			},
		},
	})
}
