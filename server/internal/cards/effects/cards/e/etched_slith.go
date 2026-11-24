package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Etched Slith — Artifact Creature — Phyrexian Slith {1}{B}, 1/1
// (March of the Machine Commander; Mystery Booster 2, #6100):
//
//	"Menace
//	 Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it. When you do, you may remove a counter from another
//	 target permanent or opponent."
//
// The combat-damage trigger puts the counter on the Slith, and only when a
// counter really landed (a Slith that left, or one that can't have
// counters put on it, puts none) does it create the CR 603.12 reflexive
// trigger. That trigger targets another permanent, or an opponent, as it
// goes on the stack (CR 603.3d). A target need not have a counter: it is
// still targeted, and then nothing is removed.
//
// On resolution you choose which counter, if any, to remove: one question
// per kind of counter the target has, in a fixed order, each 0 or 1 while
// none has been chosen yet (Glissa Sunslayer's counter choice, extended to
// a player's counters: poison, energy, experience, rad and the rest).
// Answering 0 to every question is the "you may" declined.
//
// No simplification.
const etchedSlithName = "Etched Slith"

var etchedSlithRemoveBody = game.ReflexiveBody("etched-slith/you-may-remove-a-counter",
	effects.SimpleBody(etchedSlithRemove),
	effects.ConstTargets(etchedSlithTargets))

// etchedSlithRemoves is the counter choice (#4294's registered shape). The
// bot's goal: from an opponent it removes any counter but poison; from a
// permanent it removes a -1/-1 counter only from its own and anything but a
// -1/-1 counter from someone else's. The count chosen comes off the
// permanent or the player.
var etchedSlithRemoves = effects.RegisterCounterChoice("number/etched-slith-remove-a-counter",
	func(ctx *effects.Context, target uuid.UUID, kind string, most int) int {
		if ctx.Game.PlayerByIDForEffect(target) != nil {
			if kind == game.CounterPoison {
				return 0
			}
			return most
		}
		mine := false
		if c, ok := ctx.Game.LookupCardForEffect(target); ok {
			mine = c.Controller == ctx.Controller()
		}
		if (kind == game.CounterMinusOne) == mine {
			return most
		}
		return 0
	},
	func(ctx *effects.Context, target uuid.UUID, _ []uuid.UUID, picked map[string]int) error {
		for kind, n := range picked {
			if n <= 0 {
				continue
			}
			if ctx.Game.PlayerByIDForEffect(target) != nil {
				return ctx.Game.AddPlayerCounterByForEffect(ctx.Controller(), target, kind, -n)
			}
			return effects.AddCounter{Target: target, Kind: kind, N: -n}.Apply(ctx)
		}
		return nil
	})

func init() {
	effects.Register(effects.Spec{
		OracleID:        "f1bd0d32-7513-4d1d-94aa-d06101a940c0",
		Name:            etchedSlithName,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				etchedSlithName+" — put a +1/+1 counter on it; when you do, you may remove a counter from another target permanent or opponent",
				etchedSlithGrow),
		},
	})
}

// etchedSlithTargets is "another target permanent or opponent".
func etchedSlithTargets() *game.TargetSpec {
	opponent := effects.Opponent()
	return &game.TargetSpec{
		Mode:          "any",
		Label:         "another target permanent or opponent",
		Players:       true,
		PlayerOK:      func(g *game.Game, caster uuid.UUID, p *game.Player) bool { return opponent(g, caster, p) },
		Zones:         []game.ZoneKind{game.ZoneBattlefield},
		ExcludeSource: true,
		Min:           1, Max: 1,
	}
}

// etchedSlithGrow puts the counter on the Slith and, when one landed,
// creates the reflexive trigger.
func etchedSlithGrow(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	src := item.SourceCardID
	if !effects.OnBattlefield(g, src) || ctx.IsNewSourceObject(src) {
		return nil
	}
	before := 0
	if c, ok := g.LookupCardForEffect(src); ok {
		before = c.Counters[game.CounterPlusOne]
	}
	if err := (effects.AddCounter{Target: src, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if c, ok := g.LookupCardForEffect(src); !ok || c.Counters[game.CounterPlusOne] <= before {
		return nil
	}
	return effects.WhenYouDo(etchedSlithName+" — you may remove a counter from another target permanent or opponent", etchedSlithRemoveBody).Apply(ctx)
}

// etchedSlithRemove is the reflexive trigger's resolution: the choice of
// which counter, if any, comes off the target.
func etchedSlithRemove(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard || t.Kind == game.TargetPlayer {
			return etchedSlithRemoves.Choose(ctx, t.ID, nil, 1, etchedSlithName+" — you may remove a counter")
		}
	}
	return nil
}
