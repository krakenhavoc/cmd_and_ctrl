package effects

// Goblin Test Pilot — Creature — Goblin Pilot Wizard {1}{U}{R}, 0/2:
//
//	"Flying
//	 {T}: This creature deals 2 damage to any target chosen at random."
//
// The target is ChosenAtRandom on an activated ability (#4051): it is
// drawn as part of activating (CR 602.2b / 601.2c; the 2013-04-15
// ruling), uniformly from every legal target — players, creatures,
// planeswalkers and battles, the Pilot itself included — through the
// keyed random stream, so an undo replays it. The activator names no
// target and is not asked; everyone sees the target before they can
// respond. The damage is the ordinary any-target damage: a target that
// has become illegal by resolution is not dealt it (CR 608.2b).
//
// No Purpose: the bot cannot aim it, and "deals 2 damage to" a target
// it may not pick is not removal it can count on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4760bb87-56e9-47eb-b5ec-6fa4dd19156e",
		Name:         "Goblin Test Pilot",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 2 damage to any target chosen at random.",
			Cost:    TapCost(),
			Targets: ChosenAtRandom(TargetAny()),
			Effect:  sourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
