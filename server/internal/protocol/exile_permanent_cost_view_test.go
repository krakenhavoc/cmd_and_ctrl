package protocol

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exile_permanent_cost_view_test.go — #1600: the wire half of the
// exile-a-permanent cost. Both ability views carry
// `exile_permanent_label` / `exile_permanent_options`, and the options
// are the engine's own walk, so the client's picker, the bot's moves and
// the validator agree (#544).

func creatureExileClause() *game.ExilePermanentsCost {
	return &game.ExilePermanentsCost{Count: 1, CardType: "creature", Label: "a creature you control"}
}

func seatExileCostSource(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Exile Cost Source",
		TypeLine:   "Artifact",
		OracleID:   "00000000-0000-0000-0000-0000000016a1",
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "Exile a creature you control: mark",
			Cost:   game.AbilityCost{ExilePermanents: creatureExileClause()},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	return id
}

func seatExilePermanentManaSource(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Food Chain Stand-in",
		TypeLine:   "Enchantment",
		OracleID:   "00000000-0000-0000-0000-0000000016a2",
		Owner:      owner,
		Controller: owner,
		ManaAbilities: []game.ManaAbilityShape{{
			ExilePermanents: creatureExileClause(),
			Produced:        "{G}{G}",
			Label:           "Exile a creature you control: Add {G}{G}",
		}},
	})
	return id
}

func seatExileCreatureFor(g *game.Game, controller uuid.UUID, name string, keywords ...string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: controller, Controller: controller, Keywords: keywords,
	})
	return id
}

// Both views carry the clause and the creatures that could pay: mine,
// a hexproof one included (paying a cost does not target, CR 601.2h),
// and never an opponent's or a non-creature.
func TestAbilityViewsCarryExilePermanentCostAndOptions(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatExileCostSource(g, owner)
	mana := seatExilePermanentManaSource(g, owner)
	plain := seatExileCreatureFor(g, owner, "Plain Bear")
	hexproof := seatExileCreatureFor(g, owner, "Hexproof Bear", "hexproof")
	seatExileCreatureFor(g, g.Seats[1].ID, "Their Bear")
	seatArtifactFor(g, owner, "Not A Creature")
	g.BumpLayerVersionForTest()

	want := []string{plain.String(), hexproof.String()}
	sort.Strings(want)

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	ma := vehicleView(t, g, mana).ManaAbilities[0]
	for name, got := range map[string]struct {
		label string
		opts  *LegalTargetsView
	}{
		"activated": {ab.ExilePermanentLabel, ab.ExilePermanentOptions},
		"mana":      {ma.ExilePermanentLabel, ma.ExilePermanentOptions},
	} {
		if got.label != "a creature you control" {
			t.Errorf("%s: exile_permanent_label = %q, want the clause as printed", name, got.label)
		}
		if got.opts == nil {
			t.Fatalf("%s: exile_permanent_options absent on an ability that prints the clause", name)
		}
		if got.opts.Min != 1 || got.opts.Max != 1 {
			t.Errorf("%s: bounds %d/%d, want 1/1", name, got.opts.Min, got.opts.Max)
		}
		cards := append([]string(nil), got.opts.Cards...)
		sort.Strings(cards)
		if !sameStrings(cards, want) {
			t.Errorf("%s: exile_permanent_options = %v, want my two creatures (hexproof included)", name, got.opts.Cards)
		}
	}
}

// The view's list and the enumerator's payments are one list: every
// creature the view offers is a move, and every move pays with one the
// view offered.
func TestExilePermanentCostViewAndEnumeratorAgree(t *testing.T) {
	g := busyTable(t, 0)
	me := g.Seats[g.Turn.ActiveSeat]
	src := seatExileCostSource(g, me.ID)
	seatExileCreatureFor(g, me.ID, "Bear A")
	seatExileCreatureFor(g, me.ID, "Bear B", "hexproof")
	for _, s := range g.Seats {
		if s.ID != me.ID {
			seatExileCreatureFor(g, s.ID, "Their Bear")
		}
	}
	g.BumpLayerVersionForTest()

	opts := vehicleView(t, g, src).ActivatedAbilities[0].ExilePermanentOptions
	if opts == nil {
		t.Fatal("exile_permanent_options absent on an ability that prints the clause")
	}
	view := map[string]bool{}
	for _, id := range opts.Cards {
		view[id] = true
	}
	paid := map[string]bool{}
	for _, m := range legal.EnumerateLocked(g, me.ID, legal.Options{}) {
		if m.Type != legal.TypeActivateAbility || m.Source != src {
			continue
		}
		var p struct {
			ExilePermanentIDs []string `json:"exile_permanent_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		for _, id := range p.ExilePermanentIDs {
			paid[id] = true
		}
	}
	if len(view) == 0 || !sameStrings(keysOfSet(view), keysOfSet(paid)) {
		t.Errorf("view offers %v, enumerator pays with %v — want one list", keysOfSet(view), keysOfSet(paid))
	}
}

// ADR 0137 (#2124): craft's materials span two zones (CR 702.167b). The
// view offers my creatures AND the creature cards in my own graveyard —
// never a card in another player's graveyard or a noncreature card in
// mine — with min and max the clause's count, and the enumerator pays
// out of exactly that list.
func TestCraftMaterialOptionsSpanTheBattlefieldAndYourGraveyard(t *testing.T) {
	g := busyTable(t, 0)
	me := g.Seats[g.Turn.ActiveSeat]
	src := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: src, Name: "Craft Stand-in", TypeLine: "Artifact",
		OracleID: "00000000-0000-0000-0000-000000002124",
		Owner:    me.ID, Controller: me.ID,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Craft with two creatures",
			Cost: game.AbilityCost{ExileSelf: true, ExilePermanents: &game.ExilePermanentsCost{
				Count: 2, CardType: "creature", ExcludeSource: true, FromGraveyard: true,
				Label: "two from among creatures you control and/or creature cards in your graveyard",
			}},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	live := seatExileCreatureFor(g, me.ID, "Live Bear")
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	rock := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: rock, Name: "Dead Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	refused := map[string]bool{src.String(): true, rock.String(): true}
	for _, s := range g.Seats {
		if s.ID != me.ID {
			theirs := uuid.New()
			s.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Dead Bear", TypeLine: "Creature — Bear", Owner: s.ID, Controller: s.ID})
			refused[theirs.String()] = true
		}
	}
	g.BumpLayerVersionForTest()

	opts := vehicleView(t, g, src).ActivatedAbilities[0].ExilePermanentOptions
	if opts == nil || opts.Min != 2 || opts.Max != 2 {
		t.Fatalf("exile_permanent_options = %+v, want two to pick", opts)
	}
	offered := map[string]bool{}
	for _, id := range opts.Cards {
		offered[id] = true
		if refused[id] {
			t.Errorf("craft options offer %s, which the clause refuses (the source, a noncreature card, another player's graveyard)", id)
		}
	}
	if !offered[live.String()] || !offered[dead.String()] {
		t.Fatalf("craft options = %v, want my creature and my graveyard creature card among them", opts.Cards)
	}

	paid := map[string]bool{}
	for _, m := range legal.EnumerateLocked(g, me.ID, legal.Options{}) {
		if m.Type != legal.TypeActivateAbility || m.Source != src {
			continue
		}
		var p struct {
			ExilePermanentIDs []string `json:"exile_permanent_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		for _, id := range p.ExilePermanentIDs {
			paid[id] = true
		}
	}
	if len(paid) != 2 {
		t.Fatalf("enumerator pays with %v, want one payment of two", keysOfSet(paid))
	}
	for id := range paid {
		if !offered[id] {
			t.Errorf("enumerator pays with %s, which the view does not offer", id)
		}
	}
}
