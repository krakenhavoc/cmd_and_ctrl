package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archfiend of Spite — Creature — Demon {5}{B}{B}, 6/6:
//
//	"Flying
//	 Whenever a source an opponent controls deals damage to this
//	 creature, that source's controller loses that much life unless
//	 they sacrifice that many permanents of their choice.
//	 Madness {3}{B}{B}"
//
// Phyrexian Obliterator's trigger with an "unless": one trigger per
// damage event, the source's controller and the amount captured as it
// fires (b17DamageSourceController), and only for a source an opponent
// of the Archfiend's controller controls. As it resolves, that player
// picks: lose the life, or sacrifice that many permanents of their
// choice. A player who controls fewer permanents than that can't pay
// the sacrifice (CR 118.3) and is not offered it; they lose the life.
// Madness is the shared keyword (Spec.Madness).
//
// No simplification.
func init() {
	const label = "Archfiend of Spite — that source's controller loses that much life unless they sacrifice that many permanents"
	Register(Spec{
		OracleID:        "6beec974-891f-4354-8621-ac403c07509a",
		Name:            "Archfiend of Spite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Madness:         "{3}{B}{B}",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !b17SourceDealtDamageToSelf(ev, source) {
					return false
				}
				dealer := b17DamageSourceController(ev, g)
				return dealer != uuid.Nil && dealer != source.Controller
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Player = b17DamageSourceController(ev, g)
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: archfiendOfSpitePunish,
		}},
	})
}

// archfiendOfSpitePunish is the trigger's resolution: the source's
// controller chooses between the life and the permanents.
func archfiendOfSpitePunish(g *game.Game, item *game.StackItem) error {
	victim, n := item.Params.Player, item.Params.Amount
	if n <= 0 {
		return nil
	}
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	if len(PermanentsControlledBy(g, victim)) < n {
		return g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -n)
	}
	return PickOption{
		Player:   victim,
		Question: "Archfiend of Spite — lose " + strconv.Itoa(n) + " life, or sacrifice " + strconv.Itoa(n) + " permanents?",
		Options: []game.ChoiceOption{
			{Label: "Lose " + strconv.Itoa(n) + " life"},
			{Label: "Sacrifice " + strconv.Itoa(n) + " permanents"},
		},
		ThenKey: archfiendOfSpiteAnswered,
		Carry:   []uuid.UUID{victim},
		Nums:    []int{n},
	}.Apply(NewContext(g, item))
}

// archfiendOfSpiteAnswered is the punished player's answer, keyed (#4253):
// the player rides the prompt's Carry and the amount its Nums, and the
// Archfiend is the prompt's source. An on-disk key: never renamed, never
// reused.
var archfiendOfSpiteAnswered = OptionPickThen("option-pick/archfiend-of-spite-life-or-sacrifice", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || len(r.Nums) == 0 {
		return nil
	}
	victim, n, source := r.Carry[0], r.Nums[0], ctx.Source()
	if r.Index == 1 && len(PermanentsControlledBy(ctx.Game, victim)) >= n {
		ctx.Game.PlayerSacrificesNForEffect(source, victim, nil, "Archfiend of Spite — sacrifice a permanent", n)
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(source, victim, -n)
})
