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
