package effects

// Opposition Agent — Creature — Human Rogue {2}{B}, 3/2:
//
//	"Flash
//	 You control your opponents while they're searching their
//	 libraries.
//	 While an opponent is searching their library, they exile each card
//	 they find. You may play those cards for as long as they remain
//	 exiled, and you may spend mana as though it were mana of any color
//	 to cast them."
//
// Both statics are Spec.ControlsOpponentsSearches, done by the engine's
// one search primitive (game/search_control.go, #3998, ADR 0166's
// 2026-10-16 amendment), so every tutor and land fetch in the catalog is
// covered whatever its destination or count:
//
//   - CR 723.2's limited control: while an opponent's search prompt over
//     their own library is open, the Agent's controller answers it as
//     them (CR 723.5) and sees what they see (CR 723.4: the cards on
//     offer and their hand), and nothing more once it closes.
//   - The found cards are exiled face up instead of going where the
//     effect put them, and the Agent's controller may play them for as
//     long as they stay exiled, spending mana as though it were mana of
//     any color. The permission outlives the Agent.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "1f438b8f-fe23-4f3b-ab2e-f6c33676c462",
		Name:            "Opposition Agent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},

		ControlsOpponentsSearches: true,
	})
}
