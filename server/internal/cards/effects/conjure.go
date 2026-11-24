package effects

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conjure.go — the catalog side of the Alchemy actions ADR 0099's
// 2026-11-24 amendment allows (#5948): conjure, perpetual changes and
// boons. The engine halves are game/conjure.go, Card.PerpetualKeywords
// and game.EmblemDef.Boon; a boon is declared as EmblemSpec{Boon: true}
// and made by the ordinary CreateEmblem primitive.

func init() {
	var recs []cards.Card
	if err := json.Unmarshal([]byte(conjurableRecordsJSON), &recs); err != nil {
		panic("effects: conjurable_cards.go does not parse: " + err.Error())
	}
	for _, r := range recs {
		game.RegisterConjurable(deck.ToGameCard(r, false))
	}
}

// ConjurableRecords are the printed records conjure builds its cards
// from, for the audit that holds them to the dump.
func ConjurableRecords() []cards.Card {
	var recs []cards.Card
	if err := json.Unmarshal([]byte(conjurableRecordsJSON), &recs); err != nil {
		panic(err)
	}
	return recs
}

// Conjure is "conjure a card named <name> into your hand / into your
// library / onto the battlefield" for the item's controller.
//
//   - Names are the cards conjured, one card per entry: "conjure four
//     cards named Lightning Bolt" lists it four times, "the Power Nine"
//     lists nine names.
//   - AtRandom conjures ONE card, a name drawn at random from Names
//     ("a card named A, B, or C at random"), from the keyed pick stream.
//   - Reveal reveals a card conjured into the hand ("and reveal it").
//   - Shuffle shuffles the library afterwards ("…, then shuffle").
//
// A name with no record in conjurable_cards.go fails at boot
// (CheckConjure), not at the table.
type Conjure struct {
	Names    []string
	Zone     game.ZoneKind
	AtRandom bool
	Reveal   bool
	Shuffle  bool
	// ThenTail runs once the cards have landed, with their IDs, as a
	// keyed tail carrying this effect's controller and source
	// (Sigardian Evangel discards the card it conjured). Zero for none.
	ThenTail game.TailRef
}

func (c Conjure) Apply(ctx *Context) error {
	player := ctx.Controller()
	var then game.KeyedTail
	if c.ThenTail != (game.TailRef{}) {
		then = ctx.Tail(c.ThenTail)
	}
	names := c.Names
	if c.AtRandom && len(names) > 0 {
		names = []string{ctx.Game.ConjureRandomNameForEffect(player, ctx.Source(), c.Names)}
	}
	_, err := ctx.Game.ConjureForEffect(game.ConjureSpec{
		Player: player,
		Source: ctx.Source(),
		Names:  names,
		Zone:   c.Zone,
		Reveal: c.Reveal,
		Then:   then,
	})
	if err != nil {
		return err
	}
	if c.Shuffle {
		return ctx.Game.ShuffleLibraryForEffect(player)
	}
	return nil
}

// CheckConjure is the boot guard for a Conjure: every name must have a
// printed record, and the zone must be one conjure knows.
func CheckConjure(card string, c Conjure) Conjure {
	if len(c.Names) == 0 {
		panic(fmt.Sprintf("effects: %s conjures nothing", card))
	}
	for _, n := range c.Names {
		if _, ok := game.ConjurableTemplate(n); !ok {
			panic(fmt.Sprintf("effects: %s conjures %q, which has no record in conjurable_cards.go", card, n))
		}
	}
	switch c.Zone {
	case game.ZoneHand, game.ZoneLibrary, game.ZoneBattlefield:
	default:
		panic(fmt.Sprintf("effects: %s conjures into %q", card, c.Zone))
	}
	return c
}

// PowerNine is "the Power Nine": Black Lotus, Ancestral Recall, Time
// Walk, Timetwister and the five Moxen (Oracle of the Alpha's reminder
// of what it means is the cards' own names).
var PowerNine = []string{
	"Black Lotus", "Ancestral Recall", "Time Walk", "Timetwister",
	"Mox Pearl", "Mox Sapphire", "Mox Jet", "Mox Ruby", "Mox Emerald",
}

// PerpetuallyGainKeyword is "<target> perpetually gains <keyword>": the
// keyword becomes part of the card itself (Card.PerpetualKeywords) and
// stays with it in every zone (Pearl Collector). Target is the item's
// target slot; a target that is gone gains nothing.
type PerpetuallyGainKeyword struct {
	Keyword string
	Target  int
}

func (p PerpetuallyGainKeyword) Apply(ctx *Context) error {
	if ctx.Item == nil || p.Target >= len(ctx.Item.Targets) {
		return nil
	}
	id := ctx.Item.Targets[p.Target].ID
	if id == uuid.Nil {
		return nil
	}
	return ctx.Game.GainKeywordPerpetuallyForEffect(id, p.Keyword)
}

// PermanentsYouControlNamed lists the permanents `you` control whose
// name is `name` ("each permanent you control named Chitterspitter").
// The name is read through the layered view, so a permanent that has
// become a copy of something else is named what it is now.
func PermanentsYouControlNamed(g *game.Game, you uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller != you {
			continue
		}
		if c.Effective().Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// CounterOnEachPermanentYouControlNamed is "put a <kind> counter on each
// permanent you control named <name>" (Euru, Acorn Scrounger's acorn;
// Rusko, Clockmaker's hour).
type CounterOnEachPermanentYouControlNamed struct {
	Name string
	Kind string
}

func (s CounterOnEachPermanentYouControlNamed) Apply(ctx *Context) error {
	for _, id := range PermanentsYouControlNamed(ctx.Game, ctx.Controller(), s.Name) {
		if err := (AddCounter{Target: id, Kind: s.Kind, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
