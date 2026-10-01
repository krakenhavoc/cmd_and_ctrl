package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Simic Fluxmage — Creature — Merfolk Wizard {2}{U}, 1/2:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 {1}{U}, {T}: Move a +1/+1 counter from this creature onto target
//	 creature."
//
// CR 122.5: moving a counter is removing it from the first object and
// putting it on the second, and if either can't be done neither is. So
// the Fluxmage must still be this object (CR 400.7) and still have a
// +1/+1 counter when the ability resolves, and the counter is removed
// only once the placement on the target has actually landed — a target
// that left, or that can't have counters put on it, keeps the counter
// on the Fluxmage. The placement goes through the ordinary counter
// window, so a Hardened Scales on the target's side adds one. Evolve
// is the engine's keyword trigger (game/evolve.go, #1805).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "68af451d-82f9-4c78-8bb1-36503e3f2e34",
		Name:            "Simic Fluxmage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Move a +1/+1 counter from this creature onto target creature.",
			Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  simicFluxmageMove,
		}},
	})
}

// simicFluxmageMove is the move: place one +1/+1 counter on the target,
// then, if it landed, take one off the Fluxmage.
func simicFluxmageMove(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || sourceIsNewObject(g, item) {
		return nil
	}
	self, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, item.SourceCardID) || self.Counters[game.CounterPlusOne] <= 0 {
		return nil
	}
	source := item.SourceCardID
	return g.AddCounterByThenForEffect(item.Controller, targets[0].ID, game.CounterPlusOne, 1, func(g *game.Game, placed int) error {
		if placed <= 0 {
			return nil
		}
		return g.AddCounterForEffect(source, game.CounterPlusOne, -1)
	})
}
