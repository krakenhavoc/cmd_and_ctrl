package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cbbad3da-695e-4527-9678-0942f3751287",
		Name:         "Jaya Ballard, Task Mage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{R}, {T}, Discard a card: Destroy target blue permanent.",
				Cost:    Plus(ManaCost("{R}"), TapCost(), DiscardACard()),
				Targets: TargetPermanent("target blue permanent", OfColor("U")),
				Effect:  destroyChosenPermanent,
			},
			{
				Label:   "{1}{R}, {T}, Discard a card: Jaya Ballard deals 3 damage to any target. A creature dealt damage this way can't be regenerated this turn.",
				Cost:    Plus(ManaCost("{1}{R}"), TapCost(), DiscardACard()),
				Targets: TargetAny(),
				Effect:  abilityBody(damageAnyTargetNoRegenIfDealt(fixedAmount(3))),
			},
			{
				Label:   "{5}{R}{R}, {T}, Discard a card: Jaya Ballard deals 6 damage to each creature and each player.",
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 6}},
				Cost:    Plus(ManaCost("{5}{R}{R}"), TapCost(), DiscardACard()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b23DamageEachCreatureAndEachPlayer(NewContext(g, item), 6)
				},
			},
		},
	})
}
