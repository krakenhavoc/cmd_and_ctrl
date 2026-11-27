package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arashi, the Sky Asunder — Legendary Creature — Spirit {3}{G}{G}, 5/5
// (Commander 2021, S82 OLD b058, #5504):
//
//	"{X}{G}, {T}: Arashi deals X damage to target creature with flying.
//	 Channel — {X}{G}{G}, Discard this card: It deals X damage to each
//	 creature with flying."
//
// The tap ability is Crimson Hellkite's X damage with a flying-only target,
// judged at announce and again at resolution (CR 608.2b). Channel works from
// the hand (Zones, DiscardThis) and is Corrosive Gale's flying sweep at X; X
// is announced in the hand-zone cost (CR 602.2b) and the damage is dealt by
// the discarded card as last known.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9277b680-a146-4d4f-9649-39cddafdb585",
		Name:         "Arashi, the Sky Asunder",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{X}{G}, {T}: Arashi deals X damage to target creature with flying.",
				Cost:    effects.Plus(effects.ManaCost("{X}{G}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
				Effect:  effects.OldB584SourceDealsXToTarget(),
			},
			{
				Label:   "Channel — {X}{G}{G}, Discard this card: It deals X damage to each creature with flying.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
				Cost:    effects.Plus(effects.ManaCost("{X}{G}{G}"), effects.DiscardThis()),
				Zones:   []game.ZoneKind{game.ZoneHand},
				Effect:  effects.OldB584SourceDealsXToEach(effects.HasKeyword("flying")),
			},
		},
	})
}
