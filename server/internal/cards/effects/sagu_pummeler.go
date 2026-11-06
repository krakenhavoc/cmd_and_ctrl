package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sagu Pummeler — Creature — Beast {3}{G}, 4/4 (Tarkir: Dragonstorm,
// #4463):
//
//	"Reach
//	 Renew — {4}{G}, Exile this card from your graveyard: Put two +1/+1
//	 counters and a reach counter on target creature. Activate only as a
//	 sorcery."
//
// Renew is renew.go. The reach counter is a keyword counter (CR 122.1b):
// the creature has reach for as long as the counter stays on it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "00f1d167-c143-471d-bd52-da2a65d087c4",
		Name:            "Sagu Pummeler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{
			Renew("{4}{G}", "Put two +1/+1 counters and a reach counter on target creature.",
				TargetCreature("target creature"), saguPummelerRenew),
		},
	})
}

// saguPummelerRenew puts two +1/+1 counters and a reach counter on the
// target, if it is still legal (CR 608.2b).
func saguPummelerRenew(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: id, Kind: game.CounterReach, N: 1}.Apply(ctx)
}
