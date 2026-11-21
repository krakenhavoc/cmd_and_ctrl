package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Champion of Dusan — Creature — Human Warrior {2}{G}, 4/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"Trample
//	 Renew — {1}{G}, Exile this card from your graveyard: Put a +1/+1
//	 counter and a trample counter on target creature. Activate only as a
//	 sorcery."
//
// Renew is renew.go, Sagu Pummeler's shape. The trample counter is a
// keyword counter (CR 122.1b): the creature has trample for as long as
// the counter stays on it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0a427652-9745-4802-8c8f-5ae29a29b496",
		Name:            "Champion of Dusan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{
			effects.Renew("{1}{G}", "Put a +1/+1 counter and a trample counter on target creature.",
				effects.TargetCreature("target creature"), championOfDusanRenew),
		},
	})
}

// championOfDusanRenew puts a +1/+1 counter and a trample counter on the
// target, if it is still legal (CR 608.2b).
func championOfDusanRenew(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.AddCounter{Target: id, Kind: game.CounterTrample, N: 1}.Apply(ctx)
}
