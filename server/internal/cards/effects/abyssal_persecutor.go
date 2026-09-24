package effects

// Abyssal Persecutor — Creature — Demon {2}{B}{B}, 6/6:
//
//	"Flying, trample
//	 You can't win the game and your opponents can't lose the game."
//
// Platinum Angel's mirror image (ADR 0057 Decision 4, #749), and the
// proof that the gates are two independent halves: YouCantWin on its
// controller, CantLose on every opponent. A 6/6 flying trampler for
// four whose drawback is that nothing it does can finish anybody —
// opponents sit at 0 or less life, at ten poison, behind an empty
// library, and stay in.
//
// Its 2017 ruling is what the derived read gets right for free: if the
// Persecutor leaves while an opponent is at 0 or less life, "that
// opponent will lose the game as a state-based action. No player can
// respond" — the next check reads the board without it. And "can't
// win" stops only an EFFECT win: when the last opponent concedes, its
// controller wins anyway (CR 104.2a overrides every "can't win").
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "7282f643-191b-41be-9f6f-82360c915d6d",
		Name:            "Abyssal Persecutor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		GameEndGates:    YouCantWinOpponentsCantLose(),
	})
}
