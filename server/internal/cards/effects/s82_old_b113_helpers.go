package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b113_helpers.go — shared pieces of Ikoria: Lair of Behemoths
// batch 113 (#5512): "enters with your choice of …" keyword counters, and
// the menace/vigilance mentors.

// S82OldB113EntersWithChoiceOfCounters is "This creature enters with your
// choice of <A> or <B> counter on it" (Flycatcher Giraffid, Ferocious
// Tigorilla) and, with a choice that names several kinds, "your choice of two
// different counters" (Grimdancer): each choice is the kinds one answer puts
// on the creature, in printed order. It is the permanent's AsEnters hook, so
// the controller is asked as the creature arrives and the counters land
// before anything else can act (the same shape as the as-enters colour and
// creature-type choices).
func S82OldB113EntersWithChoiceOfCounters(name string, choices ...[]string) func(*game.Card, *Context) error {
	words := make([]string, len(choices))
	options := make([]game.ChoiceOption, len(choices))
	for i, kinds := range choices {
		words[i] = strings.Join(kinds, ",")
		options[i] = game.ChoiceOption{Label: s82OldB113CounterWords(kinds)}
	}
	return func(card *game.Card, ctx *Context) error {
		ctx.Game.QueueOptionPickForEffect(game.OptionPickPrompt{
			Chooser:  card.Controller,
			Source:   card.InstanceID,
			Question: name + " — choose the counters it enters with",
			Options:  options,
			ThenKey:  s82OldB113EntersWithCountersThen,
			Carry:    []uuid.UUID{card.Controller, card.InstanceID},
			Words:    words,
		})
		return nil
	}
}

func s82OldB113CounterWords(kinds []string) string {
	if len(kinds) == 1 {
		return strings.ToUpper(kinds[0][:1]) + kinds[0][1:] + " counter"
	}
	return strings.ToUpper(kinds[0][:1]) + kinds[0][1:] + " and " + strings.Join(kinds[1:], " and ") + " counters"
}

// s82OldB113EntersWithCountersThen puts the chosen answer's counters on the
// permanent, if it is still there (CR 608.2b). An on-disk key: never renamed,
// never reused.
var s82OldB113EntersWithCountersThen = OptionPickThen("option-pick/s82-old-b113-enters-with-counters", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Index < 0 || r.Index >= len(r.Words) {
		return nil
	}
	id := r.Carry[0]
	if !OnBattlefield(ctx.Game, id) {
		return nil
	}
	for _, kind := range strings.Split(r.Words[r.Index], ",") {
		if err := (AddCounter{Target: id, Kind: kind, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// S82OldB113CounterOnEachWithKeyword is "put a +1/+1 counter on each creature
// you control with <keyword>" — Frillscare Mentor (menace) and Keensight
// Mentor (vigilance). The keyword is read through the layered view, so a
// keyword counter or a granted keyword counts.
func S82OldB113CounterOnEachWithKeyword(keyword string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, c := range MatchingBattlefield(ctx, And(YouControl(), Creature(), HasKeyword(keyword))) {
			if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// S82OldB113KeywordCounterOnTargetNonHuman is the mentors' enters trigger,
// "put a <keyword> counter on target non-Human creature you control".
func S82OldB113KeywordCounterOnTargetNonHuman(name, kind string) game.TriggeredAbility {
	return Targeting(
		WhenThisEnters(name+" — put a "+kind+" counter on target non-Human creature you control", PutACounterOnTheTarget(kind)),
		TargetCreature("target non-Human creature you control", YouControl(), Not(Subtype("Human"))),
	)
}

// S82OldB113CounterOnACreatureYouControl is "put a <kind> counter on a
// creature you control": the choice is made as the effect resolves and does
// not target. One candidate is taken without a prompt; none does nothing.
func S82OldB113CounterOnACreatureYouControl(name, kind string) Instruction {
	return func(ctx *Context) error {
		g, me := ctx.Game, ctx.Controller()
		var mine []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == me && c.IsCreature() {
				mine = append(mine, c.InstanceID)
			}
		}
		switch len(mine) {
		case 0:
			return nil
		case 1:
			return NothingIfGone(g.AddCounterForEffect(mine[0], kind, 1))
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:   me,
			Source:    ctx.Source(),
			Question:  name + " — choose a creature you control to put a " + kind + " counter on",
			Cards:     mine,
			Min:       1,
			Max:       1,
			CarryWord: kind,
			ThenKey:   s82OldB113CounterPick,
			LiveZone:  game.ZoneBattlefield,
		})
		return nil
	}
}

// s82OldB113CounterPick puts the counter (its kind rides the prompt as
// data) on the creature picked. An on-disk key: never renamed, never reused.
var s82OldB113CounterPick = game.RegisterChooseCardsThen("choose-cards/s82-old-b113-counter-on-a-creature",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) != 1 || r.Word == "" {
			return nil
		}
		return NothingIfGone(g.AddCounterForEffect(r.Picked[0], r.Word, 1))
	})

// S82OldB113BountyOnOpponentsPermanent reports whether a permanent an
// opponent of `you` controls has a bounty counter on it.
func S82OldB113BountyOnOpponentsPermanent(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you && c.Counters["bounty"] > 0 {
			return true
		}
	}
	return false
}

// S82OldB113ThisHasFirstStrike is the intervening "if it has first strike"
// (want true) or "if it doesn't have first strike" (want false) of Momentum
// Rumbler, read off the source through the layers: asked as the ability
// triggers and again as it resolves (CR 603.4). A source that has left reads
// as having neither.
type S82OldB113ThisHasFirstStrike bool

func (w S82OldB113ThisHasFirstStrike) Holds(ctx *Context) bool {
	c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	return ok && OnBattlefield(ctx.Game, ctx.Source()) && game.HasKeyword(&c, "first strike") == bool(w)
}

func (S82OldB113ThisHasFirstStrike) Condition() (game.FootprintStep, bool) {
	return CondStep(game.FootprintCondBoard, 0), true
}

// S82OldB113OpponentsBountyPermanentDied is Chevill's "a permanent an
// opponent controls with a bounty counter on it dies": the permanent went
// from the battlefield to a graveyard, an opponent of the source's controller
// controlled it as it left, and it had a bounty counter then (CR 603.10a).
func S82OldB113OpponentsBountyPermanentDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	return LeftUnderControlOf(ev, c) != source.Controller && g.LastKnownCountersForEffect(ev.CardID)["bounty"] > 0
}

// S82OldB113ThisAttackedWith is "whenever this creature attacks, if it has
// (or doesn't have) first strike": the attack, with the "if" asked as the
// attack is declared (CR 603.4); the resolution asks it again through
// S82OldB113ThisHasFirstStrike.
func S82OldB113ThisAttackedWith(want bool) When {
	return func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		if !ThisAttacked(ev, source, lki, g) {
			return false
		}
		c, ok := g.LookupCardForEffect(source.InstanceID)
		return ok && game.HasKeyword(&c, "first strike") == want
	}
}
