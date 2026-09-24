package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// fblthp_lost_on_the_range_test.go — #1391, the first special action
// granted to ANOTHER card: Fblthp, Lost on the Range lets its
// controller plot the nonland card on top of their library, for its
// mana cost or for its own plot cost (ADR 0062 amendment 2026-09-24).

const fblthpOracle = "764f6412-c4ff-4eec-9a9d-870661f97f8b"

// pushFblthp puts a Fblthp onto the battlefield under p's control.
func pushFblthp(g *game.Game, p *game.Player) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fblthp, Lost on the Range",
		TypeLine: "Legendary Creature — Homunculus", ManaCost: "{1}{U}{U}",
		OracleID: fblthpOracle, Owner: p.ID, Controller: p.ID,
		Power: 1, Toughness: 1,
	})
}

// pushOnTop puts c on top of p's library WITHOUT emptying the rest of
// it, so walking to a later turn's draw step does not deck the seat.
func pushOnTop(p *game.Player, c game.Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}

func grizzlyBears() game.Card {
	return game.Card{
		Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		OracleID: "test-fblthp-bears", Power: 2, Toughness: 2,
	}
}

// libraryContains reports whether id is still anywhere in p's library.
func libraryContains(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Library.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// offersOn is the engine's special-action offers on p's library top.
func offersOn(g *game.Game, p *game.Player, card uuid.UUID) []game.SpecialAction {
	var out []game.SpecialAction
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		out = g.SpecialActionsOfferedLocked(p.ID, c, game.ZoneLibrary)
	})
	return out
}

// The whole card end to end. In your main phase, pay the top card's
// mana cost and it goes to exile, plotted. It can't be cast this turn;
// on your next turn it is cast for nothing.
func TestFblthpPlotsTheTopCardForItsManaCostAndItCastsFreeLater(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	p := g.Seats[seat]
	pushFblthp(g, p)
	advanceToMainOf(t, g, seat)
	bears := pushOnTop(p, grizzlyBears())
	p.ManaPool = nil
	p.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "G"})

	if err := g.PerformSpecialAction(p.ID, bears, game.SpecialActionPlot, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("plot the top card: %v", err)
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("mana pool: %d tokens left, want 0 — the plot cost was not the Bears' {1}{G}", len(p.ManaPool))
	}
	if libraryContains(p, bears) {
		t.Fatal("the plotted card is still in the library")
	}
	if !g.Exile.Contains(bears) {
		t.Fatal("the plotted card is not in exile")
	}
	if err := g.CastSpell(p.ID, bears, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Fatal("cast a plotted card on the turn it was plotted (CR 702.170d)")
	}

	toNextTurnMainOf(t, g, seat)
	p.ManaPool = nil
	if err := g.CastSpell(p.ID, bears, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("free cast of the plotted card on a later turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bears) {
		t.Fatal("the plotted Bears never reached the battlefield")
	}
}

// Every refusal, and a refusal pays nothing and moves nothing: no
// Fblthp, the wrong timing, a land, a card that is not on top, and a
// Fblthp an opponent controls ("your library" is its controller's).
func TestFblthpPlotIsRefusedWithoutTheGrantOrOutsideItsWindow(t *testing.T) {
	for _, tc := range []struct {
		name    string
		fblthp  string // "", "mine", "theirs"
		main    bool
		topCard game.Card
		buried  bool
		want    error
	}{
		{"no Fblthp", "", true, grizzlyBears(), false, game.ErrCardNotFound},
		{"an opponent's Fblthp", "theirs", true, grizzlyBears(), false, game.ErrCardNotFound},
		{"not your main phase", "mine", false, grizzlyBears(), false, game.ErrSpecialActionTiming},
		// A printed land has no mana cost, so the mana-cost offer would
		// refuse it anyway. The fixture gives it one so that the
		// "nonland" clause is the only thing refusing it.
		{"a land on top", "mine", true, game.Card{Name: "Costed Land", TypeLine: "Land", ManaCost: "{G}"}, false, game.ErrSpecialActionNotOffered},
		{"a card with no mana cost", "mine", true, game.Card{Name: "Ancestral Vision", TypeLine: "Sorcery"}, false, game.ErrSpecialActionNotOffered},
		{"a card second from the top", "mine", true, grizzlyBears(), true, game.ErrCardNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			seat := g.Turn.ActiveSeat
			p := g.Seats[seat]
			switch tc.fblthp {
			case "mine":
				pushFblthp(g, p)
			case "theirs":
				pushFblthp(g, g.Seats[(seat+1)%len(g.Seats)])
			}
			if tc.main {
				advanceToMainOf(t, g, seat)
			} else {
				// The upkeep of the seat's NEXT turn: its own turn, and
				// not a main phase.
				advanceToMainOf(t, g, seat)
				for g.Turn.Step != game.StepUpkeep || g.Turn.ActiveSeat != seat {
					if _, err := g.AdvanceStep(); err != nil {
						t.Fatalf("AdvanceStep: %v", err)
					}
				}
			}
			id := pushOnTop(p, tc.topCard)
			if tc.buried {
				pushOnTop(p, game.Card{Name: "On Top", TypeLine: "Sorcery", ManaCost: "{U}"})
			}
			p.ManaPool = nil
			p.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "G"}, game.ManaToken{Color: "G"})
			libBefore := p.Library.Size()

			err := g.PerformSpecialAction(p.ID, id, game.SpecialActionPlot, game.SpecialActionParams{Strict: true})
			if !errors.Is(err, tc.want) {
				t.Fatalf("plot: got %v, want %v", err, tc.want)
			}
			if len(p.ManaPool) != 3 {
				t.Errorf("a refused plot spent mana: %d tokens left, want 3", len(p.ManaPool))
			}
			if p.Library.Size() != libBefore || !libraryContains(p, id) {
				t.Error("a refused plot moved the card")
			}
		})
	}
}

// A card that prints plot has two plots on top of the library: its
// own, and the one at its mana cost. The player chooses by price.
func TestFblthpOffersAPlotCardBothItsPlotCostAndItsManaCost(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	p := g.Seats[seat]
	pushFblthp(g, p)
	advanceToMainOf(t, g, seat)
	djinn := pushOnTop(p, game.Card{
		Name: "Djinn of Fool's Fall", TypeLine: "Creature — Djinn", ManaCost: "{4}{U}",
		OracleID: djinnOfFoolsFallOracle, Power: 4, Toughness: 3,
	})

	offers := offersOn(g, p, djinn)
	var costs []string
	for _, sa := range offers {
		if sa.Kind != game.SpecialActionPlot || sa.Zone != game.ZoneLibrary {
			t.Errorf("offer %+v: want a plot from the library", sa)
		}
		costs = append(costs, sa.Cost)
	}
	if len(costs) != 2 || costs[0] != "{3}{U}" || costs[1] != "{4}{U}" {
		t.Fatalf("plot offers: got %v, want [{3}{U} {4}{U}]", costs)
	}

	// Pick the dearer one by name: five mana, all spent.
	p.ManaPool = nil
	p.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"},
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "U"})
	if err := g.PerformSpecialAction(p.ID, djinn, game.SpecialActionPlot, game.SpecialActionParams{Strict: true, Cost: "{4}{U}"}); err != nil {
		t.Fatalf("plot for {4}{U}: %v", err)
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("mana pool: %d tokens left, want 0 — the {4}{U} offer was not the one paid", len(p.ManaPool))
	}
	if !g.Exile.Contains(djinn) {
		t.Fatal("the Djinn is not in exile")
	}
}

// The mana-cost offer reads the rules for the awkward costs: {X} is 0,
// a split card costs both halves (CR 709.4b), and a Fblthp that has
// lost its abilities grants nothing.
func TestFblthpManaCostOfferEdgeCases(t *testing.T) {
	t.Run("X is zero", func(t *testing.T) {
		g := newCatalogGame(t)
		p := g.Seats[g.Turn.ActiveSeat]
		pushFblthp(g, p)
		id := pushOnTop(p, game.Card{Name: "Blaze", TypeLine: "Sorcery", ManaCost: "{X}{R}"})
		offers := offersOn(g, p, id)
		if len(offers) != 1 || offers[0].Cost != "{R}" || offers[0].Label != "Plot {R}" {
			t.Fatalf("offers on an {X}{R} card: got %+v, want one Plot {R}", offers)
		}
	})
	t.Run("a zero-cost card is a real {0} offer", func(t *testing.T) {
		g := newCatalogGame(t)
		seat := g.Turn.ActiveSeat
		p := g.Seats[seat]
		pushFblthp(g, p)
		advanceToMainOf(t, g, seat)
		id := pushOnTop(p, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter", ManaCost: "{0}", Power: 0, Toughness: 2})
		offers := offersOn(g, p, id)
		if len(offers) != 1 || offers[0].Cost != "{0}" {
			t.Fatalf("offers on a {0} card: got %+v, want one Plot {0}", offers)
		}
		p.ManaPool = nil
		if err := g.PerformSpecialAction(p.ID, id, game.SpecialActionPlot, game.SpecialActionParams{Strict: true, Cost: "{0}"}); err != nil {
			t.Fatalf("plot a {0} card with an empty pool: %v", err)
		}
		if !g.Exile.Contains(id) {
			t.Fatal("the {0} card is not in exile")
		}
	})
	t.Run("a split card costs both halves", func(t *testing.T) {
		g := newCatalogGame(t)
		p := g.Seats[g.Turn.ActiveSeat]
		pushFblthp(g, p)
		split := game.Card{
			Name: "Fire // Ice", Layout: game.LayoutSplit,
			Faces: []game.Face{
				{Name: "Fire", TypeLine: "Instant", ManaCost: "{1}{R}"},
				{Name: "Ice", TypeLine: "Instant", ManaCost: "{1}{U}"},
			},
		}
		split.SetFace(0)
		id := pushOnTop(p, split)
		offers := offersOn(g, p, id)
		if len(offers) != 1 || offers[0].Cost != "{2}{R}{U}" {
			t.Fatalf("offers on Fire // Ice: got %+v, want one Plot {2}{R}{U}", offers)
		}
	})
	t.Run("a Fblthp with no abilities grants nothing", func(t *testing.T) {
		g := newCatalogGame(t)
		p := g.Seats[g.Turn.ActiveSeat]
		fb := pushFblthp(g, p)
		id := pushOnTop(p, grizzlyBears())
		enchant(t, g, "Darksteel Mutation", darksteelMutationOracle, fb)
		settle(t, g)
		if c := layeredCard(t, g, fb); !c.HasLostAllAbilities() {
			t.Fatal("fixture is wrong: Fblthp should have lost its abilities")
		}
		if offers := offersOn(g, p, id); len(offers) != 0 {
			t.Fatalf("offers under a Fblthp that lost its abilities: got %+v, want none", offers)
		}
	})
}

// The bot sees the move: the enumerator offers the plot from the top,
// priced, and the move it offers is one the dispatcher accepts.
func TestFblthpPlotIsEnumeratedForTheBot(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	p := g.Seats[seat]
	pushFblthp(g, p)
	advanceToMainOf(t, g, seat)
	bears := pushOnTop(p, grizzlyBears())

	find := func() *legal.Move {
		for _, m := range legal.EnumerateFor(g, p.ID) {
			if m.Kind == legal.KindSpecialAction && m.Source == bears {
				mm := m
				return &mm
			}
		}
		return nil
	}
	p.ManaPool = nil
	if m := find(); m != nil {
		t.Fatalf("offered an unaffordable plot: %q", m.Label)
	}

	p.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "G"})
	m := find()
	if m == nil {
		t.Fatal("the enumerator did not offer plotting the top card")
	}
	var params struct {
		Kind string `json:"kind"`
		Cost string `json:"cost"`
	}
	if err := json.Unmarshal(m.Params, &params); err != nil {
		t.Fatalf("move params: %v", err)
	}
	if params.Kind != "plot" || params.Cost != "{1}{G}" {
		t.Errorf("move params: kind %q cost %q, want plot {1}{G}", params.Kind, params.Cost)
	}
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(m.Type), Player: m.Player, Caller: m.Player, Params: m.Params,
	}); err != nil {
		t.Fatalf("dispatch the enumerated move: %v", err)
	}
	if !g.Exile.Contains(bears) {
		t.Fatal("the enumerated plot did not exile the card")
	}
}

// A plot card on top gives the bot two moves, one per price, and each
// move's params name its price, so dispatching the dearer move pays the
// dearer price. That is the `cost` param end to end: enumerator, wire,
// dispatcher, engine.
func TestFblthpBotMovesNameTheirPrice(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	p := g.Seats[seat]
	pushFblthp(g, p)
	advanceToMainOf(t, g, seat)
	djinn := pushOnTop(p, game.Card{
		Name: "Djinn of Fool's Fall", TypeLine: "Creature — Djinn", ManaCost: "{4}{U}",
		OracleID: djinnOfFoolsFallOracle, Power: 4, Toughness: 3,
	})
	p.ManaPool = nil
	p.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"},
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "U"})

	byCost := map[string]legal.Move{}
	for _, m := range legal.EnumerateFor(g, p.ID) {
		if m.Kind != legal.KindSpecialAction || m.Source != djinn {
			continue
		}
		var params struct {
			Cost string `json:"cost"`
		}
		if err := json.Unmarshal(m.Params, &params); err != nil {
			t.Fatalf("move params: %v", err)
		}
		byCost[params.Cost] = m
	}
	if len(byCost) != 2 {
		t.Fatalf("plot moves by price: got %d (%v), want {3}{U} and {4}{U}", len(byCost), byCost)
	}
	dear, ok := byCost["{4}{U}"]
	if !ok {
		t.Fatal("no {4}{U} move")
	}
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(dear.Type), Player: dear.Player, Caller: dear.Player, Params: dear.Params,
	}); err != nil {
		t.Fatalf("dispatch the {4}{U} move: %v", err)
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("mana pool: %d tokens left, want 0 — the {4}{U} move paid another price", len(p.ManaPool))
	}
}

// libraryTopView is the card `viewer` sees on top of `owner`'s
// library, or nil when the view hides it.
func libraryTopView(g *game.Game, viewer, owner uuid.UUID) *protocol.CardView {
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, s := range v.Seats {
		if s.ID != owner.String() || len(s.Library.Cards) == 0 {
			continue
		}
		c := s.Library.Cards[len(s.Library.Cards)-1]
		if !c.KnownByYou {
			return nil
		}
		return &c
	}
	return nil
}

// The wire: the owner sees the top card with its plot row, available
// in their main phase. An opponent does not see the card at all. Once
// something reveals the top card (Courser of Kruphix), the opponent
// sees the card, but not the owner's row.
func TestFblthpPlotRowIsTheOwnersAndLeaksNothing(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	p := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pushFblthp(g, p)
	advanceToMainOf(t, g, seat)
	bears := pushOnTop(p, grizzlyBears())

	mine := libraryTopView(g, p.ID, p.ID)
	if mine == nil || mine.InstanceID != bears.String() {
		t.Fatalf("the owner does not see their top card under Fblthp: %+v", mine)
	}
	if len(mine.SpecialActions) != 1 {
		t.Fatalf("owner's special-action rows on the top card: got %+v, want one", mine.SpecialActions)
	}
	row := mine.SpecialActions[0]
	if row.Kind != "plot" || row.Cost != "{1}{G}" || row.Label != "Plot {1}{G}" || !row.Available {
		t.Errorf("owner's plot row: got %+v, want an available Plot {1}{G}", row)
	}

	if theirs := libraryTopView(g, opp.ID, p.ID); theirs != nil {
		t.Fatalf("an opponent sees the top card of a library only Fblthp's controller may look at: %+v", theirs)
	}
	v := protocol.ViewOfGameFor(g, opp.ID.String())
	for _, s := range v.Seats {
		if s.ID != p.ID.String() {
			continue
		}
		for _, c := range s.Library.Cards {
			if c.Name != "" || len(c.SpecialActions) != 0 {
				t.Fatalf("an opponent's view of the library carries %q with rows %+v", c.Name, c.SpecialActions)
			}
		}
	}

	// Revealed to the table: the card is public, the row is not.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Courser of Kruphix",
		TypeLine: "Enchantment Creature — Centaur", OracleID: courserOracle,
		Owner: p.ID, Controller: p.ID, Power: 2, Toughness: 4,
	})
	theirs := libraryTopView(g, opp.ID, p.ID)
	if theirs == nil || theirs.Name != "Grizzly Bears" {
		t.Fatalf("an opponent does not see a revealed top card: %+v", theirs)
	}
	if len(theirs.SpecialActions) != 0 {
		t.Errorf("an opponent sees the owner's plot row on a revealed card: %+v", theirs.SpecialActions)
	}
}
