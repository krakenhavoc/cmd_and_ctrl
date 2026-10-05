package effects

// Tajuru Preserver — Creature — Elf Shaman {1}{G}, 2/1:
//
//	"Spells and abilities your opponents control can't cause you to
//	 sacrifice permanents."
//
// The whole card is the static, `OpponentEffectProtections` (#2178):
// an opponent's edict, annihilator trigger or sacrifice-all skips its
// controller, while a sacrifice they pay as a cost, or an effect they
// control themselves, is untouched. No simplification.
func init() {
	Register(Spec{
		OracleID:                  "d12d4b54-a13a-46ba-b176-3aaa453ce3e2",
		Name:                      "Tajuru Preserver",
		Completeness:              CompletenessFull,
		OpponentEffectProtections: CantBeMadeToSacrifice(),
	})
}
