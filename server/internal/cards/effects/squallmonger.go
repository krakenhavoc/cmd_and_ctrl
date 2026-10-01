package effects

// Squallmonger — Creature — Monger {3}{G}, 3/3:
//
//	"{2}: This creature deals 1 damage to each creature with flying and
//	 each player. Any player may activate this ability."
//
// An any-player row (CR 602.2, 602.1b): whoever activates it pays the
// {2} out of their own pool (CR 602.1a). Squallmonger is the source of
// the damage (CR 120.3), read as it last existed if it has left (CR
// 113.7a). "Each creature with flying" is read post-layer as the
// ability resolves, and every player is hit, the activator included
// (effects_damage_each_creature_and_player.go).
//
// No Purpose: the effect is symmetric, so the bot does not reach across
// the table for it (ADR 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50bf5af6-50d7-4dea-936e-508e24db0a03",
		Name:         "Squallmonger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{2}: This creature deals 1 damage to each creature with flying and each player. Any player may activate this ability.",
			Cost:      ManaCost("{2}"),
			AnyPlayer: true,
			Effect:    thisDealsDamageToEachCreatureMatchingAndEachPlayer(HasKeyword("flying"), 1),
		}},
	})
}
