package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quest for Ula's Temple — Enchantment {U} (Worldwake, S82 OLD b487):
//
//	"At the beginning of your upkeep, you may look at the top card of your
//	 library. If it's a creature card, you may reveal it and put a quest counter
//	 on this enchantment.
//	 At the beginning of each end step, if there are three or more quest counters
//	 on this enchantment, you may put a Kraken, Leviathan, Octopus, or Serpent
//	 creature card from your hand onto the battlefield."
//
// The upkeep ability asks its first "you may" as it resolves (CR 603.5): the look is
// private, a top card that is not a creature card is left where it is, and a creature
// card asks the second "you may" before it is revealed and the counter is put (Amareth,
// the Lustrous' look-then-offer). The end-step ability is an intervening "if" (CR
// 603.4) on the enchantment's own counters, checked as it triggers and as it resolves
// (read as they were if it has left), on EACH player's end step; the put is Walking
// Atlas' hand-to-battlefield pick, a put and not a cast, limited to the four creature
// types, read through the card's text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "434807c5-5948-4181-a46c-2c834d0a2f0b",
		Name:         "Quest for Ula's Temple",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Quest for Ula's Temple — look at the top card of your library", questUlaLook),
				"Quest for Ula's Temple — look at the top card of your library?"),
			questUlaEndStep(),
		},
	})
}

// questUlaLook is the look: a creature card on top offers the reveal.
func questUlaLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
	if len(looked) == 0 {
		return nil
	}
	top, ok := g.LookupCardForEffect(looked[0])
	if !ok || !top.IsCreature() {
		return nil
	}
	return effects.MayChoice{
		Player:   item.Controller,
		Question: "Quest for Ula's Temple — reveal " + top.Name + " and put a quest counter on this enchantment?",
		Then:     questUlaReveal,
		Carry:    []uuid.UUID{looked[0]},
	}.Apply(ctx)
}

// questUlaReveal is the "yes": reveal the top card and put a quest counter on the
// enchantment. A card that is no longer on top is not the one that was looked at.
// An on-disk key: never renamed, never reused.
var questUlaReveal = effects.MayYes("confirm/quest-for-ulas-temple-reveal", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1, "Quest for Ula's Temple — reveal the top card of your library")
	return effects.CounterOnThis{Kind: "quest", N: 1}.Apply(ctx)
})

var questUlaTypes = []string{"Kraken", "Leviathan", "Octopus", "Serpent"}

// questUlaEndStep is the second ability: each end step, with three or more quest
// counters, put one of the four creature types from the hand onto the battlefield.
func questUlaEndStep() game.TriggeredAbility {
	holds := func(g *game.Game, source uuid.UUID) bool { return effects.S82OldB487QuestCounters(g, source) >= 3 }
	t := effects.On(game.EventBeginEndStep, func(game.Event, *game.Card, game.Characteristic, *game.Game) bool { return true },
		"Quest for Ula's Temple — you may put a Kraken, Leviathan, Octopus, or Serpent creature card from your hand onto the battlefield",
		effects.Do(effects.PutFromHandOntoBattlefield{
			Match: effects.And(effects.Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				for _, sub := range questUlaTypes {
					if c.HasSubtype(sub) {
						return true
					}
				}
				return false
			}),
			Optional: true,
			Label:    "Quest for Ula's Temple — you may put a Kraken, Leviathan, Octopus, or Serpent creature card from your hand onto the battlefield",
		}))
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && holds(g, source.InstanceID) && (when == nil || when(ev, source, lki, g))
	}
	return effects.InterveningIf(effects.CondBoard{Name: "this enchantment has three or more quest counters", Test: func(ctx *effects.Context) bool {
		return holds(ctx.Game, ctx.Source())
	}}, t)
}
