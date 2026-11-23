package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// slimyPiperAttack is Slimy Piper's attack trigger body: +1/+1 until end of
// turn, or, if the controller controls four or more creatures as it
// resolves, +2/+2 and indestructible instead.
func SlimyPiperAttack(g *game.Game, item *game.StackItem) error {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == item.Controller && c.IsCreature() {
			n++
		}
	}
	ctx := NewContext(g, item)
	if n < 4 {
		return BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1, Label: "Slimy Piper — +1/+1 until end of turn"}.Apply(ctx)
	}
	if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 2, Label: "Slimy Piper — +2/+2 until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	return ThisGainsKeywordUntilEndOfTurn("indestructible", "Slimy Piper — indestructible until end of turn")(g, item)
}

// spinedTyrranaxCounterAndTrample is the reflexive body of Spined Tyrranax:
// a +1/+1 counter on the target creature, which gains trample until end of
// turn (the first target still legal as it resolves, CR 608.2b).
func SpinedTyrranaxCounterAndTrample(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Spined Tyrranax — trample until end of turn"}.Apply(ctx)
	}
	return nil
}
