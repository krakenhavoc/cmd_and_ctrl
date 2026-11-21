package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jaya Ballard, Task Mage — Legendary Creature — Human Spellshaper
// {1}{R}{R}, 2/2:
//
//	"{R}, {T}, Discard a card: Destroy target blue permanent.
//	 {1}{R}, {T}, Discard a card: Jaya Ballard deals 3 damage to any
//	 target. A creature dealt damage this way can't be regenerated this
//	 turn.
//	 {5}{R}{R}, {T}, Discard a card: Jaya Ballard deals 6 damage to each
//	 creature and each player."
//
// Three Spellshaper abilities, each paying mana, the tap and a discard
// chosen at activation (CR 602.2b). The second is Incinerate's body
// (ADR 0108 §2: a creature whose damage was prevented can still be
// regenerated); the third is Pyrohemia's creatures-then-players sweep.
// Jaya is the source of each (CR 120.3), read as she last existed if
// she has left.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbbad3da-695e-4527-9678-0942f3751287",
		Name:         "Jaya Ballard, Task Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}, {T}, Discard a card: Destroy target blue permanent.",
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.DiscardACard()),
				Targets: effects.TargetPermanent("target blue permanent", effects.OfColor("U")),
				Effect:  effects.DestroyChosenPermanent,
			},
			{
				Label:   "{1}{R}, {T}, Discard a card: Jaya Ballard deals 3 damage to any target. A creature dealt damage this way can't be regenerated this turn.",
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost(), effects.DiscardACard()),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
				Effect:  effects.AbilityBody(effects.DamageAnyTargetNoRegenIfDealt(effects.FixedAmount(3))),
			},
			{
				Label:   "{5}{R}{R}, {T}, Discard a card: Jaya Ballard deals 6 damage to each creature and each player.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 6}},
				Cost:    effects.Plus(effects.ManaCost("{5}{R}{R}"), effects.TapCost(), effects.DiscardACard()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.B23DamageEachCreatureAndEachPlayer(effects.NewContext(g, item), 6)
				},
			},
		},
	})
}
