package effects

// Absolute Virtue — Legendary Creature — Avatar Warrior {6}{W}{U}, 8/8:
//
//	"This spell can't be countered.
//	 Flying
//	 You have protection from each of your opponents. (You can't be
//	 dealt damage, enchanted, or targeted by anything controlled by
//	 your opponents.)"
//
// Aegis of the Gods' shape with a protection token in place of
// hexproof: the grant is a static of the permanent, so it is declared
// and read off the battlefield, and it ends when Absolute Virtue
// leaves. The quality is CR 702.16i over CR 702.16k, tested by the
// source's controller (#2745). The uncounterable rider is
// Spec.CantBeCountered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d4fddf20-6b3a-42c3-a245-002dacbb4725",
		Name:            "Absolute Virtue",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"flying"},
		PlayerKeywords:  []string{ProtectionFromEachOfYourOpponents},
	})
}
