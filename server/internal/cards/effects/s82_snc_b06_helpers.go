package effects

import (
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// snc6EntersWithShield is "this creature enters with a shield counter
// on it" (CR 122.1c, CR 614.1c): a self-entry replacement, so the
// counter is there before any enters trigger and before the creature
// can be damaged.
func Snc6EntersWithShield(name string, n int) game.ReplacementEffect {
	rowLabel := name + ": enters with a shield counter"
	if n != 1 {
		rowLabel = name + ": enters with shield counters"
	}
	return B10EntersWithCounters(game.CounterShield, n, rowLabel)
}

// snc6AttacksPlayerOrPlaneswalker is the defender of an attack event
// being a player or a planeswalker, never a battle (CR 506.2).
func Snc6AttacksPlayerOrPlaneswalker(g *game.Game, ev game.Event) bool {
	c, ok := g.LookupCardForEffect(ev.Target)
	return !ok || !c.IsBattle()
}

const (
	snc6RemovePrefix = "Remove a "
	snc6RemoveInfix  = " counter from "
)

// snc6WardenRemove is Sanctuary Warden's "you may remove a counter from
// a creature or planeswalker you control. If you do, draw a card and
// create a 1/1 green and white Citizen creature token." One option per
// (permanent, counter kind) pair, after a first "don't" option the
// chooser can always take. The permanent is chosen, not targeted.
func Snc6WardenRemove(ctx *Context) error {
	options := []game.ChoiceOption{{Label: "Don't remove a counter"}}
	carry := []uuid.UUID{uuid.Nil}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller != ctx.Controller() || !(c.IsCreature() || c.HasCardType("planeswalker")) {
			continue
		}
		var kinds []string
		for kind, n := range c.Counters {
			if n > 0 {
				kinds = append(kinds, kind)
			}
		}
		sort.Strings(kinds)
		for _, kind := range kinds {
			options = append(options, game.ChoiceOption{Label: snc6RemovePrefix + kind + snc6RemoveInfix + c.Name})
			carry = append(carry, c.InstanceID)
		}
	}
	if len(options) == 1 {
		return nil
	}
	return PickOption{
		Question: "Sanctuary Warden — remove a counter?",
		Options:  options,
		ThenKey:  snc6WardenRemoveAnswer,
		Carry:    carry,
	}.Apply(ctx)
}

var snc6WardenRemoveAnswer = OptionPickThen("option-pick/sanctuary-warden-remove-a-counter", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || r.Index <= 0 || r.Index >= len(r.Carry) {
		return nil
	}
	rest, ok := strings.CutPrefix(r.Option.Label, snc6RemovePrefix)
	if !ok {
		return nil
	}
	kind, _, ok := strings.Cut(rest, snc6RemoveInfix)
	if !ok {
		return nil
	}
	id := r.Carry[r.Index]
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || !OnBattlefield(ctx.Game, id) || c.Counters[kind] <= 0 {
		return nil
	}
	if err := (AddCounter{Target: id, Kind: kind, N: -1}).Apply(ctx); err != nil {
		return err
	}
	if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	return CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 1}.Apply(ctx)
})

// snc6Entry is the one-shield-counter self-entry replacement as the
// Replacements slice a Spec wants.
func Snc6Entry(name string) []game.ReplacementEffect {
	return []game.ReplacementEffect{Snc6EntersWithShield(name, 1)}
}
