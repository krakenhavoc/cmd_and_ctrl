package effects

// Identity Echo — Enchantment {2}{R} (Reality Fracture):
//
//	"{3}{R}: Exile target creature or planeswalker you control. Reveal
//	 cards from the top of your library until you reveal a creature or
//	 planeswalker card. Put that card onto the battlefield and the rest
//	 on the bottom of your library in a random order. Activate only as a
//	 sorcery."
//
// The same exile-and-reveal sentence as Jace, Multiverse Architect's −3
// (rfExileTargetThenRevealCreatureOrPlaneswalker). A token exiled this
// way ceases to exist (CR 111.7); the reveal still happens.
func init() {
	Register(Spec{
		OracleID:     "4dbf8c43-d7c5-4037-875e-6accd87aaf6f",
		Name:         "Identity Echo",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{3}{R}: Exile target creature or planeswalker you control. Reveal cards from the top of your library until you reveal a creature or planeswalker card. Put that card onto the battlefield and the rest on the bottom of your library in a random order. Activate only as a sorcery.",
			Cost:         ManaCost("{3}{R}"),
			Targets:      TargetPermanent("target creature or planeswalker you control", Or(Creature(), Planeswalker()), YouControl()),
			SorcerySpeed: true,
			Effect:       rfExileTargetThenRevealCreatureOrPlaneswalker,
		}},
	})
}
