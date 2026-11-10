package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Osseous Sticktwister — Artifact Creature — Scarecrow {1}{B}, 2/2
// (Duskmourn, #4466):
//
//	"Lifelink
//	 Delirium — At the beginning of your end step, if there are four or
//	 more card types among cards in your graveyard, each opponent may
//	 sacrifice a nonland permanent of their choice or discard a card. Then
//	 this creature deals damage equal to its power to each opponent who
//	 didn't sacrifice a permanent or discard a card this way."
//
// The end step trigger is an intervening "if" (CR 603.4), read as it
// triggers (youHaveDeliriumWhen) and again as it resolves.
//
// The choice follows the rulings. The opponents decide in turn order
// starting after the controller, each knowing what the earlier ones chose
// (CR 101.4): sacrifice a nonland permanent, discard a card, or neither,
// the first two offered only to an opponent who can do them (CR 118.3). Only
// the KIND of answer is taken here; then the sacrificing opponents choose
// and sacrifice together (PlayersSacrificeThen, Fleshbag's collector) and
// the discarding opponents choose and discard together, with every answer
// made before any card leaves a hand or the battlefield except the cards
// and permanents themselves, which nobody sees chosen early.
//
// The damage is then dealt to each opponent who really neither sacrificed
// nor discarded this way (so an opponent whose sacrifice a replacement
// effect stopped is hit too), all at once, for the Sticktwister's power as
// it last existed on the battlefield if it has left (CR 608.2h, the second
// ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "33c3a81a-6cea-48c9-b966-67084d96e74f",
		Name:            "Osseous Sticktwister",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AllOf(ByYou, youHaveDeliriumWhen),
				"Osseous Sticktwister — each opponent may sacrifice a nonland permanent or discard a card; this creature deals damage equal to its power to each opponent who didn't",
				osseousSticktwisterEndStep),
		},
	})
}

const (
	osseousSacrificeOption = "Sacrifice a nonland permanent"
	osseousDiscardOption   = "Discard a card"
	osseousNeitherOption   = "Neither (take the damage)"
)

// osseousSticktwisterEndStep is the trigger's body: the delirium re-check,
// then the first opponent is asked.
func osseousSticktwisterEndStep(g *game.Game, item *game.StackItem) error {
	if b16CardTypesInGraveyard(g, item.Controller) < 4 {
		return nil
	}
	ctx := NewContext(g, item)
	everyone := opponentsInTurnOrderAfter(ctx, item.Controller)
	return osseousAsk(ctx, everyone, nil, nil, everyone)
}

// osseousNonlandSpec is "a nonland permanent" as a sacrifice clause.
func osseousNonlandSpec() *game.TargetSpec {
	return sacrificeSpec("a nonland permanent", Nonland())
}

// osseousOptions are the answers `opponent` may give now: neither is
// always there, and the other two only when they can be done.
func osseousOptions(g *game.Game, opponent uuid.UUID) []game.ChoiceOption {
	opts := []game.ChoiceOption{{Label: osseousNeitherOption}}
	if len(g.EffectSacrificeCandidatesForEffect(opponent, osseousNonlandSpec())) > 0 {
		opts = append(opts, game.ChoiceOption{Label: osseousSacrificeOption})
	}
	if b14HandSize(g, opponent) > 0 {
		opts = append(opts, game.ChoiceOption{Label: osseousDiscardOption})
	}
	return opts
}

// osseousCarry packs the state a step hands on: the opponents still to be
// asked, those who chose to sacrifice, those who chose to discard, and every
// opponent the ability affects (the ones the damage is judged for).
func osseousCarry(remaining, sacrificers, discarders, everyone []uuid.UUID) []uuid.UUID {
	return packIDLists(remaining, sacrificers, discarders, everyone)
}

func osseousUncarry(carry []uuid.UUID) (remaining, sacrificers, discarders, everyone []uuid.UUID) {
	lists := unpackIDLists(carry)
	for len(lists) < 4 {
		lists = append(lists, nil)
	}
	return lists[0], lists[1], lists[2], lists[3]
}

// osseousAsk asks the first of `remaining` that has a real choice, in turn
// order, and once nobody is left acts on the answers.
func osseousAsk(ctx *Context, remaining, sacrificers, discarders, everyone []uuid.UUID) error {
	for len(remaining) > 0 {
		who := remaining[0]
		rest := remaining[1:]
		if p := ctx.Game.PlayerByIDForEffect(who); p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		opts := osseousOptions(ctx.Game, who)
		if len(opts) == 1 {
			// Nothing they could sacrifice and nothing to discard.
			remaining = rest
			continue
		}
		return PickOption{
			Player:   who,
			Question: "Osseous Sticktwister — sacrifice a nonland permanent or discard a card, or take damage equal to its power?",
			Options:  opts,
			ThenKey:  osseousAnswered,
			Carry:    osseousCarry(rest, sacrificers, discarders, everyone),
		}.Apply(ctx)
	}
	return osseousAct(ctx, sacrificers, discarders, everyone)
}

// osseousAnswered records one opponent's answer and asks the next. Registered
// in init: the body asks the next opponent, which names the key again.
var osseousAnswered game.OptionPickThen

func init() {
	osseousAnswered = OptionPickThen("option-pick/osseous-sticktwister-answer", func(ctx *Context, r game.OptionPicked) error {
		remaining, sacrificers, discarders, everyone := osseousUncarry(r.Carry)
		if r.Option != nil {
			switch r.Option.Label {
			case osseousSacrificeOption:
				sacrificers = append(append([]uuid.UUID(nil), sacrificers...), r.Chooser)
			case osseousDiscardOption:
				discarders = append(append([]uuid.UUID(nil), discarders...), r.Chooser)
			}
		}
		return osseousAsk(ctx, remaining, sacrificers, discarders, everyone)
	})
}

// osseousAct runs what the opponents chose: the discarders discard, then the
// sacrificers sacrifice, then the damage. Each step with nobody to ask moves
// straight on.
func osseousAct(ctx *Context, sacrificers, discarders, everyone []uuid.UUID) error {
	if len(discarders) > 0 {
		return ctx.Game.PlayersDiscardThenKeyedForEffect(discarders, game.DiscardPrompt{
			Source:   ctx.Source(),
			N:        1,
			Question: "Osseous Sticktwister — discard a card",
		}, osseousDiscarded, packIDLists([]uuid.UUID{ctx.Controller()}, sacrificers, everyone))
	}
	return osseousSacrifice(ctx, sacrificers, nil, everyone)
}

// osseousDiscarded is the discards done: the carry holds the controller, the
// sacrificers and every opponent. An on-disk key.
var osseousDiscarded = game.RegisterPromptRunThen("prompt-run/osseous-sticktwister-discarded", func(g *game.Game, r game.RunDone) error {
	lists := unpackIDLists(r.Carry)
	for len(lists) < 3 {
		lists = append(lists, nil)
	}
	controller, sacrificers, everyone := lists[0], lists[1], lists[2]
	if len(controller) == 0 {
		return nil
	}
	var helped []uuid.UUID
	for _, id := range everyone {
		if r.Discards().Discarded(id) {
			helped = append(helped, id)
		}
	}
	return osseousSacrifice(keyedContext(g, controller[0], r.Source), sacrificers, helped, everyone)
})

// osseousSacrifice has the sacrificing opponents choose and sacrifice
// together, `helped` being those who have already discarded.
func osseousSacrifice(ctx *Context, sacrificers, helped, everyone []uuid.UUID) error {
	if len(sacrificers) == 0 {
		return osseousDamage(ctx, helped, everyone)
	}
	return ctx.PlayersSacrificeThen(sacrificers, osseousNonlandSpec(), "Osseous Sticktwister — sacrifice a nonland permanent",
		osseousSacrificed, ctx.RunData(packIDLists([]uuid.UUID{ctx.Controller()}, helped, everyone)))
}

// osseousSacrificed is the sacrifices done: the carry holds the controller,
// those who discarded and every opponent. An on-disk key.
var osseousSacrificed = game.RegisterPromptRunThen("prompt-run/osseous-sticktwister-sacrificed", func(g *game.Game, r game.RunDone) error {
	lists := unpackIDLists(r.Carry)
	for len(lists) < 3 {
		lists = append(lists, nil)
	}
	controller, helped, everyone := lists[0], lists[1], lists[2]
	if len(controller) == 0 {
		return nil
	}
	helped = append([]uuid.UUID(nil), helped...)
	for _, id := range everyone {
		if r.Sacrifices().Sacrificed(id) {
			helped = append(helped, id)
		}
	}
	return osseousDamage(keyedContext(g, controller[0], r.Source), helped, everyone)
})

// osseousDamage deals the Sticktwister's power to each opponent who neither
// sacrificed nor discarded, all at once.
func osseousDamage(ctx *Context, helped, everyone []uuid.UUID) error {
	power := b43PowerNowOrLastKnown(ctx.Game, ctx.Source())
	if power <= 0 {
		return nil
	}
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, id := range everyone {
			if b23HasID(helped, id) {
				continue
			}
			if p := ctx.Game.PlayerByIDForEffect(id); p == nil || p.Eliminated {
				continue
			}
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: power}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
