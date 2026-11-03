package effects

import (
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b19_helpers.go — shared pieces for the nineteenth Secrets of
// Strixhaven batch, the first of the main set (#4456).

// KeywordIncrement is the printed keyword the Increment trigger row
// carries.
const KeywordIncrement = "increment"

// incrementLabel is the label (and effect key) of every increment row.
const incrementLabel = "Increment — put a +1/+1 counter on this creature"

// Increment is the keyword on the card that has it (CR 702.191a):
//
//	"Whenever you cast a spell, if this permanent is a creature and the
//	 amount of mana spent to cast that spell is greater than this
//	 creature's power or this creature's toughness, put a +1/+1 counter
//	 on this creature."
//
// The mana spent is read off the spell's payment record as it is cast
// (Aetherflux Conduit's Build) and carried on the item, so a spell
// countered in response still counts. The "if" is an intervening if
// (CR 603.4): asked as the spell is cast, with the creature's power and
// toughness as they are then (counters and pumps included), and again as
// the ability resolves, against the same amount and the creature as it is
// by then. The counter goes on this permanent, not on a creature that
// left and came back (CR 400.7). Two instances trigger separately
// (CR 702.191b), each its own row.
func Increment() game.TriggeredAbility {
	t := On(game.EventCast, incrementTriggers, incrementLabel, incrementResolve)
	t.Keyword = KeywordIncrement
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, incrementLabel)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	return t
}

// incrementExceeds is the increment condition on a permanent as it is
// now: a creature whose power or toughness is less than the mana spent.
func incrementExceeds(c game.Card, spent int) bool {
	return c.IsCreature() && (spent > c.PowerForComparison() || spent > c.CurrentToughness())
}

// incrementTriggers is "whenever you cast a spell, if this permanent is a
// creature and the mana spent is greater than its power or toughness".
func incrementTriggers(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	self, ok := g.LookupCardForEffect(source.InstanceID)
	if !ok {
		return false
	}
	return incrementExceeds(self, g.StackItemPaidForEffect(ev.CardID).Spent().Total())
}

// incrementResolve re-checks the condition (CR 603.4) and puts the
// counter on this creature.
func incrementResolve(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	self, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !incrementExceeds(self, item.Params.Amount) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
}

// socB19FractalWithTheDeadsCounters is "create a 0/0 green and blue
// Fractal creature token, then put this creature's counters on that
// token" (Ambitious Augmenter): every counter the creature had as it died,
// of every kind (its last-known information, CR 603.10a, 608.2h), put on
// the new Fractal. The token is created first and the counters are then
// PUT on it (Deekah's shape, socB8FractalWithCounters), as two steps so a
// creation that waits on a question still gets them. `key` names the step
// record that carries the event cursor across a pause: an on-disk
// identity.
//
// Both steps declare a footprint (footprint.go), because the printed line
// is one sentence: the creation, and the counters as a counter-placement
// event, the conservative reading. The counters come from the dead
// creature's frozen record and land on the new token, so they read and
// change nothing another trigger in the batch can reach.
func socB19FractalWithTheDeadsCounters(key string) Effect {
	return Do(
		socB19CreateFractal{CreateToken: CreateToken{Template: TokenCard("0/0 green and blue Fractal"), N: 1}, key: key},
		socB19PutDeadCounters{CounterOnThis: CounterOnThis{Kind: game.CounterPlusOne, N: 1}, key: key},
	)
}

// socB19CreateFractal is the first step: record where the event log is,
// then create the Fractal. It embeds CreateToken for the footprint the
// printed one-sentence line needs: the Apply below makes the same token
// with the same events.
type socB19CreateFractal struct {
	CreateToken
	key string
}

func (s socB19CreateFractal) Apply(ctx *Context) error {
	ctx.RecordStep(s.key, nil, int(b25LastEventSeq(ctx.Game)))
	_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("0/0 green and blue Fractal")), 1, game.TokenEntryOptions{})
	return err
}

// socB19PutDeadCounters is the second step: every counter the triggering
// permanent had, put on the Fractal the first step made. It embeds
// CounterOnThis for the footprint: a counter-placement event, the
// conservative reading of counters put on the new token. The counters come
// from the dead creature's frozen record and land on that token, so they
// read and change nothing another trigger in the batch can reach.
type socB19PutDeadCounters struct {
	CounterOnThis
	key string
}

func (s socB19PutDeadCounters) Apply(ctx *Context) error {
	rec := ctx.StepRecord(s.key)
	if len(rec.Ns) < 1 {
		return nil
	}
	dead, ok := ctx.TriggeringPermanent()
	if !ok || len(dead.Counters) == 0 {
		return nil
	}
	kinds := make([]string, 0, len(dead.Counters))
	for kind, n := range dead.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	for _, id := range b27TokensCreatedByAfter(ctx.Game, ctx.Controller(), uint64(rec.Ns[0])) {
		if !onBattlefield(ctx.Game, id) || !socB8IsFractal(ctx.Game, id) {
			continue
		}
		for _, kind := range kinds {
			if err := (AddCounter{Target: id, Kind: kind, N: dead.Counters[kind]}).Apply(ctx.asGroupMember()); err != nil {
				return err
			}
		}
	}
	return nil
}

// socB19WithCreatureAndSubtypeInAddition is a type line gaining the
// creature card type and `subtype`, both "in addition to its other types"
// (CR 707.9b, Applied Geometry): a copy of an artifact is an Artifact
// Creature with the new subtype added to its own.
func socB19WithCreatureAndSubtypeInAddition(printed, subtype string) string {
	super, types, subs := game.ParseTypeLine(printed)
	hasCreature := false
	for _, t := range types {
		if t == "Creature" {
			hasCreature = true
		}
	}
	head := make([]string, 0, len(super)+len(types)+1)
	head = append(head, super...)
	head = append(head, types...)
	if !hasCreature {
		head = append(head, "Creature")
	}
	hasSub := false
	for _, s := range subs {
		if strings.EqualFold(s, subtype) {
			hasSub = true
		}
	}
	if !hasSub {
		subs = append(append([]string(nil), subs...), subtype)
	}
	line := strings.Join(head, " ")
	if len(subs) == 0 {
		return line
	}
	return line + " — " + strings.Join(subs, " ")
}

// socB19PutCountersOnNewTokens is the second half of a "create ... put N
// +1/+1 counters on it" pair: the counters go on the tokens `controller`
// created after the event cursor filed under key, the ones still on the
// battlefield.
func socB19PutCountersOnNewTokens(ctx *Context, key string, n int) error {
	rec := ctx.StepRecord(key)
	if len(rec.Ns) < 1 || n <= 0 {
		return nil
	}
	var ids []uuid.UUID
	for _, id := range b27TokensCreatedByAfter(ctx.Game, ctx.Controller(), uint64(rec.Ns[0])) {
		if onBattlefield(ctx.Game, id) {
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
