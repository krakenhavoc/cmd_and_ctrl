package effects

// Sidisi, Regent of the Mire — Legendary Creature — Zombie Snake Warlock
// {1}{B}, 1/3 (Tarkir: Dragonstorm, #4463):
//
//	"{T}, Sacrifice a creature you control with mana value X other than
//	 Sidisi: Return target creature card with mana value X plus 1 from
//	 your graveyard to the battlefield. Activate only as a sorcery."
//
// X is announced with the activation (CR 602.2b), exactly as Lazav's and
// Chthonian Nightmare's "mana value X" are. It is the one number both
// clauses read: the sacrifice cost must name a creature with mana value X
// (SacrificeAnotherWithManaValueX, game.SacrificeManaValueX), and the
// target is a creature card in your graveyard with mana value X + 1
// (TargetSpec.ManaValueEqualsXPlus). The target is judged at announce and
// again as the ability resolves (CR 608.2b); a creature with {X} in its
// cost counts it as 0 (the ruling). A token has mana value 0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72074972-418d-4dce-83cf-ce8cf310fc42",
		Name:         "Sidisi, Regent of the Mire",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{T}, Sacrifice a creature you control with mana value X other than Sidisi: Return target creature card with mana value X plus 1 from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:         Plus(TapCost(), SacrificeAnotherWithManaValueX("a creature you control with mana value X other than Sidisi", Creature(), YouControl())),
			SorcerySpeed: true,
			Targets: TargetCardInGraveyard("target creature card with mana value X plus 1 from your graveyard",
				YouOwn(), Creature()).WithManaValueEqualsXPlus(1),
			Effect: returnFirstLegalGraveyardTargetToBattlefield,
		}},
	})
}
