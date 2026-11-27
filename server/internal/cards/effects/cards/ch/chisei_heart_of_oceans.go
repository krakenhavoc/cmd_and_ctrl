package cardsch

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chisei, Heart of Oceans — Legendary Creature — Spirit {2}{U}{U}, 4/4
// (Betrayers of Kamigawa, #5602):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice Chisei unless you remove a
//	 counter from a permanent you control."
//
// A pay-unless made as Quickling's sacrifice-unless-return: as the trigger
// resolves, a choose_cards prompt names one permanent you control that holds a
// counter, or none, and naming none (or controlling no counter at all, with
// nothing to pay with, CR 118.3) sacrifices Chisei. The counter is of any kind:
// with one kind on the permanent it comes off at once, and with several a
// second prompt names the kind. Chisei herself and the counters on her may be
// the ones chosen. A Chisei that has left, or come back as a new object (CR
// 400.7), is not sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f91407fe-3c3b-4ffc-81de-cb8c5b0fe526",
		Name:            "Chisei, Heart of Oceans",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Chisei, Heart of Oceans — sacrifice it unless you remove a counter from a permanent you control", chiseiUpkeep),
		},
	})
}

// chiseiCounterKinds are the counter kinds a permanent holds, in a fixed order.
func chiseiCounterKinds(c game.Card) []string {
	var kinds []string
	for k, n := range c.Counters {
		if n > 0 {
			kinds = append(kinds, k)
		}
	}
	sort.Strings(kinds)
	return kinds
}

func chiseiUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	var holders []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && len(chiseiCounterKinds(c)) > 0 {
			holders = append(holders, c.InstanceID)
		}
	}
	if len(holders) == 0 {
		return effects.SacrificeThisIfStillThere(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     item.SourceCardID,
		Question:   "Chisei, Heart of Oceans — remove a counter from a permanent you control (or sacrifice Chisei)",
		Cards:      holders,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    chiseiPicked,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// chiseiPicked takes the counter off the permanent named, or sacrifices Chisei
// when none was named or the one named has lost its counters. An on-disk key.
var chiseiPicked = effects.AskedPick("choose-cards/s82-old-b599-chisei-remove-counter", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return effects.SacrificeThisIfStillThere(ctx)
	}
	id := r.Picked[0]
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || !effects.OnBattlefield(ctx.Game, id) {
		return effects.SacrificeThisIfStillThere(ctx)
	}
	kinds := chiseiCounterKinds(c)
	switch len(kinds) {
	case 0:
		return effects.SacrificeThisIfStillThere(ctx)
	case 1:
		return effects.AddCounter{Target: id, Kind: kinds[0], N: -1}.Apply(ctx)
	}
	options := make([]game.ChoiceOption, 0, len(kinds))
	for _, k := range kinds {
		options = append(options, game.ChoiceOption{Label: k})
	}
	return effects.PickOption{
		Question: "Chisei, Heart of Oceans — which kind of counter do you remove?",
		Options:  options,
		ThenKey:  chiseiKindPicked,
		Carry:    []uuid.UUID{id},
		WithItem: true,
	}.Apply(ctx)
})

// chiseiKindPicked removes one counter of the kind named from the permanent
// the first prompt chose. An on-disk key.
var chiseiKindPicked = effects.OptionPickThen("option-pick/s82-old-b599-chisei-counter-kind", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 {
		return effects.SacrificeThisIfStillThere(ctx)
	}
	return effects.AddCounter{Target: r.Carry[0], Kind: r.Option.Label, N: -1}.Apply(ctx)
})
