package testsm

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func upfrontTapped(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	return false
}

func upfrontPool(p *game.Player) []string {
	out := make([]string, 0, p.ManaPool.Total())
	for _, tok := range p.ManaPool.Units() {
		out = append(out, tok.Color)
	}
	return out
}

// A painland's coloured half, the colour named up front: one {R} in
// the pool, the rider's point of damage, and no second question.
func TestUpfrontColorPainlandIsOneStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Shivan Reef", "Land", effectstest.ShivanReefOracle)

	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("ActivateManaAbility with colour R: %v", err)
	}
	if got := upfrontPool(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want [R]", got)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d: the rider still deals its 1 damage", before, me.Life)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick != nil {
		t.Errorf("a colour named up front must queue no mana_pick, got %+v", pick)
	}
}

// Birds of Paradise: any of its five colours, off-identity included,
// straight into the pool.
func TestUpfrontColorBirdsTakesAnyOfItsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.RiderGiveCommander(g, me, "{2}{G}")
	birds := effectstest.PushCatalogPermanent(g, me.ID, "Birds of Paradise", "Creature — Bird", effectstest.UpfrontBirdsOracle, false)

	if err := g.ActivateManaAbility(me.ID, birds, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("ActivateManaAbility with colour U: %v", err)
	}
	if got := upfrontPool(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool = %v, want [U]", got)
	}
	if !upfrontTapped(g, birds) {
		t.Error("Birds did not tap")
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick != nil {
		t.Errorf("queued a mana_pick anyway: %+v", pick)
	}
}

// Mystic Gate's two independent pipe slots take one colour each, in
// output order.
func TestUpfrontColorsFillEachPickingSlot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gate := effectstest.SeedPermanentWithOracle(g, me.ID, "Mystic Gate", "Land", effectstest.UpfrontMysticGateOracle)
	me.ManaPool = append(me.ManaPool, game.ManaToken{Color: "W", Source: uuid.New()})

	row := effectstest.UpfrontManaRow(t, g, me.ID, gate, 1)
	if want := [][]string{{"W", "U"}, {"W", "U"}}; !reflect.DeepEqual(row.ColorOptions, want) {
		t.Fatalf("Mystic Gate color_options = %v, want %v", row.ColorOptions, want)
	}
	if err := g.ActivateManaAbility(me.ID, gate, 1, game.ManaAbilityParams{Colors: []string{"W", "U"}}); err != nil {
		t.Fatalf("ActivateManaAbility with colours W, U: %v", err)
	}
	if got := upfrontPool(me); !reflect.DeepEqual(got, []string{"W", "U"}) {
		t.Errorf("pool = %v, want [W U]", got)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick != nil {
		t.Errorf("queued a mana_pick anyway: %+v", pick)
	}
}

// Every illegal naming is refused before anything is paid: no tap, no
// damage, no mana, no prompt.
func TestUpfrontIllegalColorIsRefusedWithNothingTapped(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		idx                    int
		colors                 []string
	}{
		{"Shivan Reef", "Land", effectstest.ShivanReefOracle, 1, []string{"G"}},      // not printed
		{"Shivan Reef", "Land", effectstest.ShivanReefOracle, 1, []string{"U", "R"}}, // one slot, two answers
		{"Shivan Reef", "Land", effectstest.ShivanReefOracle, 0, []string{"C"}},      // {C} picks nothing
		{"Birds of Paradise", "Creature — Bird", effectstest.UpfrontBirdsOracle, 0, []string{"C"}},
		{"Birds of Paradise", "Creature — Bird", effectstest.UpfrontBirdsOracle, 0, []string{"g"}}, // exact letters only
		{"Command Tower", "Land", effectstest.UpfrontCommandTowerOracle, 0, []string{"G"}},         // outside a mono-W identity
		{"Mystic Gate", "Land", effectstest.UpfrontMysticGateOracle, 1, []string{"W"}},             // two slots, one answer
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.RiderGiveCommander(g, me, "{W}")
			src := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			if tc.name == "Mystic Gate" {
				me.ManaPool = append(me.ManaPool, game.ManaToken{Color: "W", Source: uuid.New()})
			}
			life, pool := me.Life, me.ManaPool.Total()
			err := g.ActivateManaAbility(me.ID, src, tc.idx, game.ManaAbilityParams{Colors: tc.colors})
			if !errors.Is(err, game.ErrIllegalManaColor) {
				t.Fatalf("colours %v: err = %v, want ErrIllegalManaColor", tc.colors, err)
			}
			if upfrontTapped(g, src) {
				t.Error("the refused activation tapped the source")
			}
			if me.Life != life || me.ManaPool.Total() != pool {
				t.Errorf("the refused activation paid or produced: life %d → %d, pool %d → %d",
					life, me.Life, pool, me.ManaPool.Total())
			}
			if pick := effectstest.RiderLatestManaPick(g, me.ID); pick != nil {
				t.Errorf("the refused activation queued a mana_pick: %+v", pick)
			}
		})
	}
}

// With no colour named, nothing changes: the pick is still queued
// after the tap, with the same options the view published.
func TestUpfrontAbsentColorKeepsTheTwoStepActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Shivan Reef", "Land", effectstest.ShivanReefOracle)

	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if !upfrontTapped(g, land) || me.Life != before-1 {
		t.Errorf("tapped=%v life %d → %d: the cost and the rider still happen first", upfrontTapped(g, land), before, me.Life)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %v before the pick is answered, want empty", upfrontPool(me))
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || !reflect.DeepEqual(pick.ColorOptions, []string{"U", "R"}) {
		t.Fatalf("pick = %+v, want a mana_pick over [U R]", pick)
	}
}

// The published list IS the prompt's list, narrowing and order
// included, for every shape that asks: an any-colour source (identity
// first, #843), the two "commander's color identity" cards (narrowed,
// CR 903.4f) and a dual whose second colour is the identity's.
func TestUpfrontColorOptionsMatchTheManaPickPrompt(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle, commander string
		idx                               int
		want                              []string
	}{
		{"Birds of Paradise", "Creature — Bird", effectstest.UpfrontBirdsOracle, "{2}{G}", 0, []string{"G", "W", "U", "B", "R"}},
		{"Birds of Paradise", "Creature — Bird", effectstest.UpfrontBirdsOracle, "{B}{G}", 0, []string{"B", "G", "W", "U", "R"}},
		{"Command Tower", "Land", effectstest.UpfrontCommandTowerOracle, "{U}{R}", 0, []string{"U", "R"}},
		{"Command Tower", "Land", effectstest.UpfrontCommandTowerOracle, "{G}", 0, []string{"G"}},
		{"Arcane Signet", "Artifact", effectstest.UpfrontArcaneSignetOracle, "{W}{B}", 0, []string{"W", "B"}},
		{"Shivan Reef", "Land", effectstest.ShivanReefOracle, "{R}", 1, []string{"R", "U"}},
	}
	for _, tc := range cases {
		t.Run(tc.name+" "+tc.commander, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.RiderGiveCommander(g, me, tc.commander)
			src := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)

			row := effectstest.UpfrontManaRow(t, g, me.ID, src, tc.idx)
			if len(row.ColorOptions) != 1 || !reflect.DeepEqual(row.ColorOptions[0], tc.want) {
				t.Fatalf("color_options = %v, want [%v]", row.ColorOptions, tc.want)
			}
			if err := g.ActivateManaAbility(me.ID, src, tc.idx, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			pick := effectstest.RiderLatestManaPick(g, me.ID)
			if pick == nil {
				t.Fatal("no mana_pick")
			}
			if !reflect.DeepEqual(row.ColorOptions[0], pick.ColorOptions) {
				t.Errorf("view color_options %v != mana_pick color_options %v", row.ColorOptions[0], pick.ColorOptions)
			}
		})
	}
}

// Nothing to pick publishes nothing: a fixed output, and CR 903.4f's
// Command Tower with no commander (which adds no mana at all).
func TestUpfrontColorOptionsAbsentWhenNothingIsPicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	reef := effectstest.SeedPermanentWithOracle(g, me.ID, "Shivan Reef", "Land", effectstest.ShivanReefOracle)
	tower := effectstest.SeedPermanentWithOracle(g, me.ID, "Command Tower", "Land", effectstest.UpfrontCommandTowerOracle)
	me.Command.Cards = nil

	if row := effectstest.UpfrontManaRow(t, g, me.ID, reef, 0); row.ColorOptions != nil {
		t.Errorf("Shivan Reef {C}: color_options = %v, want none", row.ColorOptions)
	}
	row := effectstest.UpfrontManaRow(t, g, me.ID, tower, 0)
	if !row.AddsNoMana || row.ColorOptions != nil {
		t.Errorf("Command Tower, no commander: adds_no_mana=%v color_options=%v, want true and none",
			row.AddsNoMana, row.ColorOptions)
	}
}
