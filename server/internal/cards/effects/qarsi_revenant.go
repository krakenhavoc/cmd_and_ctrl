package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Qarsi Revenant — Creature — Vampire {1}{B}{B}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying, deathtouch, lifelink
//	 Renew — {2}{B}, Exile this card from your graveyard: Put a flying
//	 counter, a deathtouch counter, and a lifelink counter on target
//	 creature. Activate only as a sorcery."
//
// Renew is renew.go, Kheru Goldkeeper's shape: each counter is a keyword
// counter (CR 122.1b). A target that has left by resolution is illegal and
// nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bbea79bc-71d4-4100-92f4-783c2c4b46da",
		Name:            "Qarsi Revenant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch", "lifelink"},
		Activated: []ActivatedAbility{
			Renew("{2}{B}", "Put a flying counter, a deathtouch counter, and a lifelink counter on target creature.",
				TargetCreature("target creature"), qarsiRevenantRenew),
		},
	})
}

// qarsiRevenantRenew puts a flying, a deathtouch and a lifelink counter on
// the target, if it is still legal (CR 608.2b).
func qarsiRevenantRenew(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	for _, kind := range []string{game.CounterFlying, game.CounterDeathtouch, game.CounterLifelink} {
		if err := (AddCounter{Target: id, Kind: kind, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
