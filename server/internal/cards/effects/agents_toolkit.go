package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Agent's Toolkit — Artifact — Clue {1}{G}{U} (New Capenna Commander,
// #5491):
//
//	"This artifact enters with a +1/+1 counter, a flying counter, a
//	 deathtouch counter, and a shield counter on it. (If it would be dealt
//	 damage or destroyed, remove a shield counter from it instead.)
//	 Whenever a creature you control enters, you may move a counter from
//	 this artifact onto that creature.
//	 {2}, Sacrifice this artifact: Draw a card."
//
// The four counters are one entry replacement (CR 614.1c), in place before
// anything sees the artifact. The trigger is the creature's own entry,
// including a creature that enters while the Toolkit is a Clue with fewer
// counters: the controller names which counter, or declines, as it resolves;
// a Toolkit that has left or has no counter moves nothing. The counter is
// removed from the Toolkit and put on the creature, so a shield counter keeps
// protecting. The Clue's draw is its own activated ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a87abe7a-fee9-45fd-b3d8-896e43ed6b3c",
		Name:         "Agent's Toolkit",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				for _, kind := range agentsToolkitKinds {
					ev.AddCounterAtETB(kind, 1)
				}
				return nil
			},
			Label: "Agent's Toolkit: enters with a +1/+1 counter, a flying counter, a deathtouch counter, and a shield counter",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			}, "Agent's Toolkit — you may move a counter from this artifact onto that creature", agentsToolkitMove),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}

var agentsToolkitKinds = []string{game.CounterPlusOne, game.CounterFlying, game.CounterDeathtouch, game.CounterShield}

// agentsToolkitMove offers the counters the Toolkit has now.
func agentsToolkitMove(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	creature := item.Trigger.Event.CardID
	self, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, creature) {
		return nil
	}
	kinds := sncB07Kinds([]game.Card{self})
	if len(kinds) == 0 {
		return nil
	}
	options := []game.ChoiceOption{{Label: "Don't move a counter"}}
	for _, kind := range kinds {
		options = append(options, game.ChoiceOption{Label: "A " + kind + " counter"})
	}
	return PickOption{
		Question: "Agent's Toolkit — move a counter onto that creature?",
		Options:  options,
		ThenKey:  agentsToolkitMoved,
		Carry:    []uuid.UUID{item.SourceCardID, creature},
		Words:    kinds,
	}.Apply(NewContext(g, item))
}

// agentsToolkitMoved moves the chosen counter, keyed. An on-disk key.
var agentsToolkitMoved = OptionPickThen("option-pick/agents-toolkit-move", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) < 2 || r.Index < 1 || r.Index > len(r.Words) {
		return nil
	}
	toolkit, creature, kind := r.Carry[0], r.Carry[1], r.Words[r.Index-1]
	c, ok := ctx.Game.LookupCardForEffect(toolkit)
	if !ok || !onBattlefield(ctx.Game, toolkit) || !onBattlefield(ctx.Game, creature) || c.Counters[kind] <= 0 {
		return nil
	}
	if err := (AddCounter{Target: toolkit, Kind: kind, N: -1}).Apply(ctx); err != nil {
		return err
	}
	return ctx.Game.AddCounterByForEffect(ctx.Controller(), creature, kind, 1)
})
