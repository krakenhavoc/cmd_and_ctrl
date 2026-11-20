package effects

// Curse of Disturbance — Enchantment — Aura Curse {2}{B} (Commander 2017,
// #5533):
//
//	"Enchant player
//	 Whenever enchanted player is attacked, create a 2/2 black Zombie
//	 creature token. Each opponent attacking that player does the same."
//
// Curse of Opulence's shape with a Zombie in place of a Gold token
// (c17CurseWhenAttacked): one declaration is one "is attacked", and each
// attacking opponent creates one Zombie under their own control however
// many creatures they send.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6cbd36d9-de47-41b0-9ef4-a72ca01adccd",
		Name:         "Curse of Disturbance",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(),
		Triggered: c17CurseWhenAttacked("Curse of Disturbance — create a 2/2 black Zombie",
			"Curse of Disturbance — the attacking opponent creates a 2/2 black Zombie",
			Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1})),
	})
}
