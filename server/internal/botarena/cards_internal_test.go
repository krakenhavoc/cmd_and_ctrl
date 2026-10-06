package botarena

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// cardsWindow builds one decision window: the seat's hand and
// battlefield, and a move list over them.
type cardsWindow struct {
	seat, opp                                 uuid.UUID
	ring, petal, study, seer, land, token, op uuid.UUID
}

func newCardsWindow() cardsWindow {
	return cardsWindow{
		seat: uuid.New(), opp: uuid.New(),
		ring: uuid.New(), petal: uuid.New(), study: uuid.New(), seer: uuid.New(), land: uuid.New(), token: uuid.New(), op: uuid.New(),
	}
}

func (w cardsWindow) view() protocol.GameView {
	me, them := w.seat.String(), w.opp.String()
	return protocol.GameView{
		Battlefield: protocol.ZoneView{Cards: []protocol.CardView{
			{InstanceID: w.seer.String(), Name: "Viscera Seer", Owner: me, TypeLine: "Creature — Vampire Wizard"},
			{InstanceID: w.land.String(), Name: "High Market", Owner: me, TypeLine: "Land"},
			{InstanceID: w.token.String(), Name: "Treasure", Owner: me, TypeLine: "Token Artifact — Treasure", IsToken: true},
			{InstanceID: w.op.String(), Name: "Opponent's Rock", Owner: them, TypeLine: "Artifact"},
		}},
		Seats: []protocol.PlayerView{{ID: me, Hand: protocol.ZoneView{Cards: []protocol.CardView{
			{InstanceID: w.ring.String(), Name: "Sol Ring", Owner: me, TypeLine: "Artifact",
				ManaAbilities: []protocol.ManaAbilityView{{TapCost: true}}},
			{InstanceID: w.petal.String(), Name: "Lotus Petal", Owner: me, TypeLine: "Artifact",
				ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, SacrificeCost: true}}},
			{InstanceID: w.study.String(), Name: "Rhystic Study", Owner: me, TypeLine: "Enchantment"},
		}}}},
	}
}

func (w cardsWindow) moves() []legal.Move {
	return []legal.Move{
		{Kind: legal.KindPass, Label: "pass"},
		{Kind: legal.KindCast, Source: w.ring, Label: "Cast Sol Ring"},
		{Kind: legal.KindCast, Source: w.study, Label: "Cast Rhystic Study"},
		// A one-shot mana source is not a rock (ADR 0126 §2).
		{Kind: legal.KindCast, Source: w.petal, Label: "Cast Lotus Petal"},
		// Two activations of one card are one offer of that card.
		{Kind: legal.KindActivate, Source: w.seer, Label: "Seer: sac A"},
		{Kind: legal.KindActivate, Source: w.seer, Label: "Seer: sac B"},
		// None of these is a deck card the section counts.
		{Kind: legal.KindActivate, Source: w.land, Label: "High Market"},
		{Kind: legal.KindActivate, Source: w.token, Label: "Treasure"},
		{Kind: legal.KindActivate, Source: w.op, Label: "Opponent's Rock"},
		{Kind: legal.KindMana, Source: w.ring, Label: "Tap Sol Ring for mana"},
	}
}

func (w cardsWindow) event(index int, applied bool) aiseat.DecisionEvent {
	return aiseat.DecisionEvent{
		Seat:    w.seat,
		Input:   aiseat.Input{View: w.view(), Seat: w.seat, Moves: w.moves()},
		Index:   index,
		Applied: applied,
	}
}

func TestCardTallyCountsOffersAndUses(t *testing.T) {
	w := newCardsWindow()
	tally := newCardTally()
	tally.Observe(w.event(1, true))              // cast Sol Ring
	tally.Observe(w.event(5, true))              // the second Seer activation
	tally.Observe(w.event(2, false))             // Rhystic Study, rejected: not a use
	tally.Observe(w.event(aiseat.Decline, true)) // nothing dispatched
	tally.Observe(w.event(9, true))              // a mana ability is not a cast

	got := map[string]CardUse{}
	for _, u := range tally.list() {
		got[u.Name+"/"+u.Action] = u
	}
	want := map[string]CardUse{
		"Sol Ring/cast":         {Name: "Sol Ring", Action: ActionCast, ManaSource: true, Offered: 5, Taken: 1},
		"Rhystic Study/cast":    {Name: "Rhystic Study", Action: ActionCast, Offered: 5, Taken: 0},
		"Viscera Seer/activate": {Name: "Viscera Seer", Action: ActionActivate, Offered: 5, Taken: 1},
		"Lotus Petal/cast":      {Name: "Lotus Petal", Action: ActionCast, Offered: 5, Taken: 0},
	}
	if len(got) != len(want) {
		t.Fatalf("tallied %v, want exactly %v", got, want)
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s = %+v, want %+v", k, got[k], w)
		}
	}
}

func TestCardsSectionNeverAndCanaries(t *testing.T) {
	acc := newCardsAcc()
	spec := SeatSpec{Tier: "heuristic", Deck: "esper-control"}
	game := func(cards ...CardUse) GameResult {
		return GameResult{Seats: []SeatResult{{Spec: spec, Cards: cards}}}
	}
	acc.add(game(
		CardUse{Name: "Sol Ring", Action: ActionCast, ManaSource: true, Offered: 3, Taken: 1},
		CardUse{Name: "Rhystic Study", Action: ActionCast, Offered: 4},
	))
	acc.add(game(
		CardUse{Name: "Sol Ring", Action: ActionCast, ManaSource: true, Offered: 2},
		CardUse{Name: "Rhystic Study", Action: ActionCast, Offered: 1},
	))
	acc.add(game())
	cards := acc.totals()
	if len(cards) != 1 {
		t.Fatalf("%d contestants", len(cards))
	}
	cc := cards[0]
	if cc.SeatGames != 3 || cc.Never != 1 || cc.Contestant() != "heuristic · esper-control" {
		t.Fatalf("contestant %+v", cc)
	}
	byName := map[string]CardTotals{}
	for _, c := range cc.Cards {
		byName[c.Name] = c
	}
	if c := byName["Sol Ring"]; c.Windows != 5 || c.Taken != 1 || c.GamesOffered != 2 || c.GamesUsed != 1 || c.Never {
		t.Errorf("Sol Ring %+v", c)
	}
	// Five windows and no use is `never`, four would not be.
	if c := byName["Rhystic Study"]; c.Windows != 5 || !c.Never {
		t.Errorf("Rhystic Study %+v", c)
	}

	rows := canaries(cards)
	var ring2, ring3, study3, seer3 *CanaryResult
	for i := range rows {
		r := &rows[i]
		switch {
		case r.Bar == "A2" && r.Name == "Sol Ring":
			ring2 = r
		case r.Bar == "A3" && r.Name == "Sol Ring":
			ring3 = r
		case r.Bar == "A3" && r.Name == "Rhystic Study":
			study3 = r
		case r.Bar == "A3" && r.Name == "Viscera Seer":
			seer3 = r
		}
	}
	if ring2 == nil || ring2.Rate != 0.5 || ring2.Meets || ring2.Want != ManaSourceBar {
		t.Errorf("A2 Sol Ring: %+v", ring2)
	}
	if ring3 == nil || !ring3.Meets || ring3.Want != CanaryBar {
		t.Errorf("A3 Sol Ring at 1 of 2 games meets a 50%% bar: %+v", ring3)
	}
	if study3 == nil || study3.Meets || study3.Offered != 2 {
		t.Errorf("A3 Rhystic Study: %+v", study3)
	}
	if seer3 == nil || seer3.Contestant != "" || seer3.Meets {
		t.Errorf("an A3 canary nobody was offered is reported unmeasured: %+v", seer3)
	}

	var b strings.Builder
	writeCards(&b, Summary{Cards: cards, Canaries: rows})
	md := b.String()
	for _, want := range []string{
		"| heuristic · esper-control | 3 | 2 | 1 | Rhystic Study |",
		"| A2 | Sol Ring | mana rock or dork | heuristic · esper-control | 2 | 1 | 50% | 80% | no |",
		"| A3 | Viscera Seer | sacrifice outlet | — | 0 | 0 | — | 50% | not offered |",
		"| A3 | Mary Read and Anne Bonny (activate) |",
		"<details><summary>heuristic · esper-control: 2 cards offered, 1 never</summary>",
		"| Rhystic Study | cast |  | 5 | 0 | 2 | 0 | **never** |",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("the Cards section is missing %q:\n%s", want, md)
		}
	}
}
