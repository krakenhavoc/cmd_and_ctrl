package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b9_helpers.go — Secrets of Strixhaven Commander, batch 9 (#4834):
// the pieces two cards, or a card and its test, share.

// socB9TeamKeywordAndPump is "creatures you control gain <keyword> and
// get +X/+X until end of turn, where X is the number of creatures you
// control" (Craterhoof Behemoth's trample, Moonshaker Cavalry's flying).
// X is counted as the ability resolves, the source included if it is
// still there, and the affected set is the creatures you control right
// then (CR 611.2c); neither changes later in the turn. A count of zero
// does nothing.
func socB9TeamKeywordAndPump(name, keyword string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		x := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.IsCreature() && c.Controller == item.Controller {
				x++
			}
		}
		if x == 0 {
			return nil
		}
		yours := And(Creature(), YouControl())
		if err := (BoostUntilEOT{
			Match:     yours,
			Power:     x,
			Toughness: x,
			Label:     name + " — +X/+X",
		}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{
			Match:    yours,
			Keywords: []string{keyword},
			Label:    name + " — " + keyword,
		}.Apply(ctx)
	}
}

// socB9HasXInManaCost reports whether the spell an EventCast names has an
// {X} in its own mana cost: the cost printed on the card, parsed, so an
// {X} in an additional or alternative cost does not count (Brass
// Infiniscope's reading).
func socB9HasXInManaCost(g *game.Game, cardID uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return false
	}
	cost, err := game.ParseCost(c.ManaCost)
	return err == nil && cost.XSlots > 0
}

// socB9FirstXSpellThisTurn is "whenever you cast your first spell with
// {X} in its mana cost each turn" (Nev, the Practical Dean; Owlin
// Spiralmancer): the source's controller cast a spell with {X} in its
// mana cost, and no earlier spell they cast this turn had one. The scan
// covers the whole turn, before the source was on the battlefield too,
// so a second X spell after the permanent arrives does not trigger it
// (the rulings).
func socB9FirstXSpellThisTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == uuid.Nil || ev.Actor != source.Controller || !socB9HasXInManaCost(g, ev.CardID) {
		return false
	}
	for _, prev := range g.EventsThisTurn() {
		if prev.Seq >= ev.Seq {
			break
		}
		if prev.Kind == game.EventCast && prev.Actor == ev.Actor && prev.CardID != ev.CardID && socB9HasXInManaCost(g, prev.CardID) {
			return false
		}
	}
	return true
}

// socB9CountersOn is the number of counters of every kind on c — "the
// number of counters on that creature" (Nils, Discipline Enforcer's
// ruling: all counters, not only +1/+1).
func socB9CountersOn(c game.Card) int {
	n := 0
	for _, k := range c.Counters {
		if k > 0 {
			n += k
		}
	}
	return n
}

// socB9CreatureYouControlWithACounter is "creatures you control with
// counters on them": any counter of any kind, read off the permanent.
func socB9CreatureYouControlWithACounter(target *game.Card, _ *game.Game, source *game.Card) bool {
	return source != nil && target.IsCreature() && target.Controller == source.Controller && socB9CountersOn(*target) > 0
}

// socB9CounterKinds is c's counter kinds with at least one counter, in
// sorted order, so the events a move or a removal emits are
// deterministic.
func socB9CounterKinds(c game.Card) []string {
	kinds := make([]string, 0, len(c.Counters))
	for kind, n := range c.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	return kinds
}

// KeywordGravestorm is the printed keyword Gravestorm's trigger row
// carries.
const KeywordGravestorm = "gravestorm"

// gravestormKey is the Key of every gravestorm row (ADR 0041 P9).
const gravestormKey = "gravestorm"

// Gravestorm is the keyword on the card that has it (CR 702.69a):
//
//	"When you cast this spell, copy it for each permanent that was put
//	 into a graveyard from the battlefield this turn. If the spell has
//	 any targets, you may choose new targets for any of the copies."
//
// Storm's shape (storm.go): a FromStack cast trigger that names the
// spell and does not target it, so a spell countered in response is
// still copied from last-known information (CR 608.2h). The count is
// read as the trigger resolves, off the turn's event log: every
// EventLTB this turn whose destination was a graveyard, whoever
// controlled or owned the permanent and token or not (the rulings). The
// copies are created, not cast, so they trigger nothing that watches a
// cast. Two instances trigger separately (CR 702.69b).
func Gravestorm() game.TriggeredAbility {
	return game.TriggeredAbility{
		Keyword:   KeywordGravestorm,
		FromStack: true,
		Key:       gravestormKey,
		Watches:   []game.EventKind{game.EventCast},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return &game.StackItem{
				Kind:         game.StackItemTriggered,
				Controller:   ev.Actor,
				Owner:        ev.Actor,
				SourceCardID: source.InstanceID,
				Label:        source.Name + " — gravestorm",
			}
		},
		Effect: gravestormEffect,
	}
}

// socB9PermanentsPutIntoGraveyardsThisTurn is gravestorm's count: the
// permanents put into a graveyard from the battlefield this turn.
func socB9PermanentsPutIntoGraveyardsThisTurn(g *game.Game) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventLTB && ev.CardID != uuid.Nil && ev.NewZone == game.ZoneGraveyard {
			n++
		}
	}
	return n
}

func gravestormEffect(g *game.Game, item *game.StackItem) error {
	n := socB9PermanentsPutIntoGraveyardsThisTurn(g)
	if n <= 0 {
		return nil
	}
	return CopySpell{
		StackID:          item.SourceCardID,
		Controller:       item.Controller,
		Count:            n,
		ChooseNewTargets: true,
		FromLastKnown:    true,
	}.Apply(NewContext(g, item))
}

// socB9PlusOneCountersOnSource is "for each +1/+1 counter on this
// creature": the source as it is now, or as it last existed on the
// battlefield (CR 608.2h).
func socB9PlusOneCountersOnSource(ctx *Context) int {
	self, ok := ctx.SourcePermanent()
	if !ok {
		return 0
	}
	return self.Counters[game.CounterPlusOne]
}

// socB9APlayerHasAtMostOneCardInHand is "if a player has one or fewer
// cards in hand" (Naktamun Lorespinner): any player still in the game.
func socB9APlayerHasAtMostOneCardInHand(g *game.Game) bool {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.Hand == nil {
			continue
		}
		if p.Hand.Size() <= 1 {
			return true
		}
	}
	return false
}

// socB9RemoveCounters removes n counters of kind from the permanent and
// reports how many really came off (a counter that can't be removed
// stays, #1824), read back off the board.
func socB9RemoveCounters(ctx *Context, id uuid.UUID, kind string, n int) (int, error) {
	// The count is read into a local first: the card's Counters map is
	// shared with the board, so a later read of `before` sees the removal.
	before, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || n <= 0 {
		return 0, nil
	}
	was := before.Counters[kind]
	if was <= 0 {
		return 0, nil
	}
	if err := (AddCounter{Target: id, Kind: kind, N: -n}).Apply(ctx); err != nil {
		return 0, err
	}
	now := 0
	if after, ok := ctx.Game.LookupCardForEffect(id); ok {
		now = after.Counters[kind]
	}
	return max(was-now, 0), nil
}

// socB9CounterCounts is a private copy of c's counters, so a read after a
// removal still sees what was there before it.
func socB9CounterCounts(c game.Card) map[string]int {
	out := make(map[string]int, len(c.Counters))
	for kind, n := range c.Counters {
		out[kind] = n
	}
	return out
}
