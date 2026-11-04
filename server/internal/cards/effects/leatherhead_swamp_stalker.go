package effects

import (
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leatherhead, Swamp Stalker — {2}{G}{G} Legendary Creature — Crocodile
// Mutant Rogue 5/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Trample
//	 Leatherhead enters with a hexproof counter on her.
//	 Whenever Leatherhead deals combat damage to a player, you may remove a
//	 counter from her. When you do, destroy target artifact or enchantment
//	 that player controls."
//
// The hexproof counter is Arwen, Mortal Queen's entry counter, a keyword
// counter the engine reads itself (CR 122.1b, ADR 0101).
//
// The combat-damage trigger's "you may" is its own (Optional). Removing
// "a counter" leaves the kind to you, as Spinal Parasite's does: with one
// kind on her it is removed without a question, with several you pick
// which. Only once a counter has really come off her is the CR 603.12
// reflexive trigger created, so a Leatherhead with no counters (or one
// that left the battlefield) destroys nothing. The damaged player rides
// the reflexive trigger (Auron, Venerated Guardian's ControlledBy on
// Params.Player), and its target is chosen as it goes on the stack.
// Removing the hexproof counter is a real cost: she loses hexproof.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ee9b606a-a1fe-43c8-84d0-9bf7b012cc75",
		Name:            "Leatherhead, Swamp Stalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterHexproof, 1, "Leatherhead, Swamp Stalker: enters with a hexproof counter on her"),
		},
		Triggered: []game.TriggeredAbility{
			Optional(
				WheneverThisDealsCombatDamageToAPlayer("Leatherhead, Swamp Stalker — you may remove a counter from her",
					func(g *game.Game, item *game.StackItem) error {
						if !sourceIsStillThisPermanent(g, item) {
							return nil
						}
						return leatherheadRemoveACounter(NewContext(g, item), item.SourceCardID, item.Trigger.Event.Target)
					}),
				"Leatherhead, Swamp Stalker — remove a counter from her to destroy an artifact or enchantment that player controls?"),
		},
	})
}

// leatherheadRemoveACounter removes one counter of a kind you pick from
// Leatherhead and then creates the reflexive destroy aimed at `player`.
func leatherheadRemoveACounter(ctx *Context, self, player uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(self)
	if !ok {
		return nil
	}
	var kinds []string
	for kind, n := range c.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	switch len(kinds) {
	case 0:
		return nil
	case 1:
		return leatherheadRemoveThenDestroy(ctx, self, player, kinds[0])
	}
	options := make([]game.ChoiceOption, len(kinds))
	for i, kind := range kinds {
		options[i] = game.ChoiceOption{Label: removeACounterPrefix + kind + removeACounterSuffix}
	}
	return PickOption{
		Question: "Leatherhead, Swamp Stalker — remove which counter from her?",
		Options:  options,
		ThenKey:  leatherheadCounterPicked,
		Carry:    []uuid.UUID{self, player},
		WithItem: true,
	}.Apply(ctx)
}

// leatherheadCounterPicked is the kind chosen. An on-disk key.
var leatherheadCounterPicked = OptionPickThen("option-pick/leatherhead-remove-a-counter", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) != 2 {
		return nil
	}
	kind, ok := strings.CutPrefix(r.Option.Label, removeACounterPrefix)
	if !ok {
		return nil
	}
	return leatherheadRemoveThenDestroy(ctx, r.Carry[0], r.Carry[1], strings.TrimSuffix(kind, removeACounterSuffix))
})

// leatherheadRemoveThenDestroy removes the counter and, when it really came
// off, creates "when you do, destroy target artifact or enchantment that
// player controls".
func leatherheadRemoveThenDestroy(ctx *Context, self, player uuid.UUID, kind string) error {
	if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(self)
	before := c.Counters[kind]
	if !ok || before <= 0 {
		return nil
	}
	if err := (AddCounter{Target: self, Kind: kind, N: -1}).Apply(ctx); err != nil {
		return err
	}
	if after, ok := ctx.Game.LookupCardForEffect(self); !ok || after.Counters[kind] >= before {
		return nil
	}
	return ReflexiveTrigger{
		Label:  "Leatherhead, Swamp Stalker — destroy target artifact or enchantment that player controls",
		Body:   leatherheadDestroyBody,
		Params: game.EffectParams{Player: player},
	}.Apply(ctx)
}

// leatherheadDestroyBody is the reflexive half. An on-disk key.
var leatherheadDestroyBody = game.ReflexiveBody("leatherhead-swamp-stalker/destroy", simpleBody(destroyFirstLegalCardTarget),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetPermanent("target artifact or enchantment that player controls", Or(Artifact(), Enchantment()), ControlledBy(p.Player))
	})
