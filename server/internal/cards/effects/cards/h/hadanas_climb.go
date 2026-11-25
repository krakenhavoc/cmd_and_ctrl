package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hadana's Climb // Winged Temple of Orazca (Rivals of Ixalan, S82 OLD batch
// 218, #5531).
//
// Front face, Legendary Enchantment {1}{G}{U}:
//
//	"At the beginning of combat on your turn, put a +1/+1 counter on target
//	 creature you control. Then if that creature has three or more +1/+1
//	 counters on it, transform Hadana's Climb."
//
// Back face, Legendary Land:
//
//	"(Transforms from Hadana's Climb.)
//	 {T}: Add one mana of any color.
//	 {1}{G}{U}, {T}: Target creature you control gains flying and gets +X/+X
//	 until end of turn, where X is its power."
//
// The trigger's target is chosen as it goes on the stack and re-checked as it
// resolves (CR 608.2b). The counter is put on through the counter pipeline
// (replacements may add or pause on it), and the "then if" reads the creature's
// count once the counter has landed, in a registered tail: three or more
// +1/+1 counters transforms the enchantment, which ignores a second instruction
// once it has transformed (CR 701.27f). The land's pump reads the creature's
// power as the ability resolves, before it is raised, and grants flying until
// end of turn with it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     hadanasClimbOracleID,
		Name:         "Hadana's Climb",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat(
				"Hadana's Climb — put a +1/+1 counter on target creature you control; then if it has three or more, transform Hadana's Climb",
				hadanasClimbCounter),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     hadanasClimbOracleID + "#1",
		Name:         "Winged Temple of Orazca",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}{U}, {T}: Target creature you control gains flying and gets +X/+X until end of turn, where X is its power.",
			Cost:    effects.Plus(effects.ManaCost("{1}{G}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
			Effect:  wingedTempleOfOrazcaPump,
		}},
	})
}

const hadanasClimbOracleID = "93b91d18-6acf-42e5-9a31-bc6e01f90c1f"

func hadanasClimbCounter(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	then := ctx.ItemTail(hadanasClimbThen)
	then.Carry = []uuid.UUID{id}
	return g.AddCounterByKeyedForEffect(uuid.Nil, id, game.CounterPlusOne, 1, then)
}

// hadanasClimbThen is "Then if that creature has three or more +1/+1 counters
// on it, transform Hadana's Climb". An on-disk key: never renamed, never reused.
var hadanasClimbThen = effects.CardTail("then/hadanas-climb-transform", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.Tail.Carry[0])
	if !ok || c.Counters[game.CounterPlusOne] < 3 {
		return nil
	}
	return effects.TransformThis{}.Apply(ctx)
})

func wingedTempleOfOrazcaPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	c, found := g.LookupCardForEffect(id)
	if !found {
		return nil
	}
	x := c.CurrentPower()
	const label = "Winged Temple of Orazca — flying and +X/+X until end of turn"
	if err := effects.UntilEndOfTurn(ctx, id, nil, label, game.AddKeywordsMod("flying")); err != nil {
		return err
	}
	return effects.BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: label}.Apply(ctx)
}
