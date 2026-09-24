package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirri, Weatherlight Duelist — Legendary Creature — Cat Warrior, 3/2
// for {1}{G}{W}:
//
//	"First strike
//	 Whenever Mirri attacks, each opponent can't block with more than
//	 one creature this combat.
//	 As long as Mirri is tapped, no more than one creature can attack
//	 you each combat."
//
// The proof card for #1534's two engine shapes (ADR 0045 amendment of
// 2026-09-24, Decision 46), one per line:
//
//   - The attack trigger registers a per-defender block limit
//     (EachOpponentCantBlockWithMoreThanN, BlockRule.LimitPerDefender):
//     every opponent may block with one creature, each counted on their
//     own. It outlives Mirri, and binds creatures an opponent flashes
//     in after it resolves — a rule about players, which CR 611.2c
//     does not lock. "This combat" is the turn's combat, because the
//     engine has no extra combats (#753); the primitive's comment says
//     what has to change when it does.
//   - The static is Crawlspace's line behind a condition
//     (AttackLimit.While via AsLongAs(ThisIsTapped, …)). It is read
//     live, so it switches on the moment Mirri taps to attack and stays
//     on through the opponents' turns until she untaps. Tapping Mirri
//     after attackers are declared unmakes none of them — the same
//     raise-the-count rule as any limit that arrives late.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "60bc9cd2-2e03-4889-b1ef-fdab9a9a6d08",
		Name:            "Mirri, Weatherlight Duelist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Mirri, Weatherlight Duelist — each opponent can't block with more than one creature this combat",
				Do(EachOpponentCantBlockWithMoreThanN{
					N:     1,
					Label: "each opponent can't block with more than one creature this combat (Mirri, Weatherlight Duelist)",
				})),
		},
		AttackLimits: []game.AttackLimit{
			AsLongAs(ThisIsTapped, NoMoreThanNCanAttackYouEachCombat(1)),
		},
	})
}
