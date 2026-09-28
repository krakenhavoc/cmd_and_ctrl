package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dwarven Armorer — Creature — Dwarf {R}, 0/2 (#1664):
//
//	"{R}, {T}, Discard a card: Put a +0/+1 counter or a +1/+0 counter
//	 on target creature."
//
// Waited on #1664: a +1/+0 or +0/+1 counter used to be stored and
// change nothing. Every P/T counter kind now counts
// (game.PTCounterDelta, CR 122.1a).
//
// The "or" is the controller's choice as the ability RESOLVES (CR
// 608.2), not a mode picked at activation — the ability is not modal —
// so it is a PickOption asked once the target is known to still be
// legal. A target that has gone takes the whole ability with it (CR
// 608.2b) and no question is asked. The {T} is a creature's {T}, so
// the Armorer waits out summoning sickness (CR 302.6).
func init() {
	Register(Spec{
		OracleID:     "5bbd27b1-0afd-4d98-a73c-c348c8f08625",
		Name:         "Dwarven Armorer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}, Discard a card: Put a +0/+1 counter or a +1/+0 counter on target creature.",
			Cost:    Plus(ManaCost("{R}"), TapCost(), DiscardACard()),
			Targets: TargetCreature("target creature"),
			Effect:  dwarvenArmorerCounter,
		}},
	})
}

// dwarvenArmorerKinds are the two printed choices, in printed order.
var dwarvenArmorerKinds = [...]string{"+0/+1", "+1/+0"}

func dwarvenArmorerCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return PickOption{
		Question: "Dwarven Armorer — which counter?",
		Options: []game.ChoiceOption{
			{Label: "Put a " + dwarvenArmorerKinds[0] + " counter on it"},
			{Label: "Put a " + dwarvenArmorerKinds[1] + " counter on it"},
		},
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(dwarvenArmorerKinds) {
				return nil
			}
			return AddCounter{Target: target, Kind: dwarvenArmorerKinds[index], N: 1}.Apply(ctx)
		},
	}.Apply(ctx)
}
