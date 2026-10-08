package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Volatile Arsonist // Dire-Strain Anarchist — {3}{R}{R} Creature — Human
// Werewolf 4/4 // Creature — Werewolf 5/5 (#2586, ADR 0132):
//
//	Front: "Menace, haste
//	        Whenever this creature attacks, it deals 1 damage to each of
//	        up to one target creature, up to one target player, and/or up
//	        to one target planeswalker.
//	        Daybound"
//	Back:  "Menace, haste
//	        Whenever this creature attacks, it deals 2 damage to each of
//	        up to one target creature, up to one target player, and/or up
//	        to one target planeswalker.
//	        Nightbound"
//
// Three independent "up to one" clauses, read by slot: each slot that was
// filled and is still legal on resolution takes the damage (CR 608.2b),
// and an empty or illegal slot is simply skipped, so the ability can be
// aimed at any one, two or all three kinds.
//
// No simplification.
func init() {
	const oracle = "215dfa88-b130-44df-9cfc-f1f0e4a36f4d"
	attack := func(name string, n int) game.TriggeredAbility {
		return Targeting(
			WheneverThisAttacks(name+" — deals damage to up to one target creature, player and planeswalker",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					// One printed instruction is one damage instance (CR 615.8).
					return g.DamageInstanceForEffect(func() error {
						for slot := 0; slot < 3; slot++ {
							t, ok := ctx.ClauseTarget(slot)
							if !ok {
								continue
							}
							if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
				}),
			Clauses(
				TargetCreature("up to one target creature").WithCount(0, 1),
				TargetPlayer("up to one target player").WithCount(0, 1),
				TargetPermanent("up to one target planeswalker", Planeswalker()).WithCount(0, 1),
			))
	}
	Register(Spec{
		OracleID:        oracle,
		Name:            "Volatile Arsonist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "haste", "daybound"},
		Triggered:       []game.TriggeredAbility{attack("Volatile Arsonist", 1)},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Dire-Strain Anarchist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "haste", "nightbound"},
		Triggered:       []game.TriggeredAbility{attack("Dire-Strain Anarchist", 2)},
	})
}
