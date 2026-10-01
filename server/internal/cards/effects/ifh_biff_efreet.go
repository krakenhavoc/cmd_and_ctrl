package effects

// Ifh-Bíff Efreet — Creature — Efreet {2}{G}{G}, 3/3:
//
//	"Flying
//	 {G}: This creature deals 1 damage to each creature with flying and
//	 each player. Any player may activate this ability."
//
// Flying is a printed keyword (CR 702.9). The activation is an
// any-player row (CR 602.2, 602.1b): whoever activates it pays the {G}
// out of their own pool (CR 602.1a). The Efreet is the source of the
// damage (CR 120.3), read as it last existed if it has left (CR
// 113.7a); it has flying, so it hits itself. "Each creature with
// flying" is read post-layer as the ability resolves, and every player
// is hit, the activator included
// (effects_damage_each_creature_and_player.go).
//
// No Purpose: the effect is symmetric, so the bot does not reach across
// the table for it (ADR 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e503a4f2-a785-4e7a-89a7-a9b24fb98831",
		Name:            "Ifh-Bíff Efreet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:     "{G}: This creature deals 1 damage to each creature with flying and each player. Any player may activate this ability.",
			Cost:      ManaCost("{G}"),
			AnyPlayer: true,
			Effect:    thisDealsDamageToEachCreatureMatchingAndEachPlayer(HasKeyword("flying"), 1),
		}},
	})
}
