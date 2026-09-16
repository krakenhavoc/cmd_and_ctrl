package protocol

import (
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // March of the Multitudes' convoke clause
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cost_option_gate_view_test.go — the option lists the view ships for
// COSTS must not apply the targeting gate. Tapping a creature for
// convoke and sacrificing one to a cost are not targeting (CR 601.2f
// and 601.2h), so a shrouded creature you control pays either one.
// The engine already accepts it (tap_cost.go and activated.go pass
// targeting=false, ADR 0038 §4). These tests pin the wire projection
// to the same rule, so the client offers what the server accepts.
//
// Shroud is the keyword under test because it is the one that bites
// its own controller (CR 702.18a). Hexproof only refuses opponents
// (CR 702.11b), and every one of these lists is built from the
// controller's point of view, so hexproof never narrowed them.

const marchOfTheMultitudesOracle = "0c26ab0d-80f6-4e5b-9d0e-af17c1519583"

// seatShroudedCreature puts an untapped creature with shroud (as
// under Lightning Greaves) on the battlefield under owner.
func seatShroudedCreature(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := game.NewCard("Shrouded Bear", owner)
	c.TypeLine = "Creature — Bear"
	c.Power, c.Toughness = 2, 2
	c.Keywords = []string{"shroud"}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// creatureSacrificeSpec is "Sacrifice a creature": the clause
// Ashnod's Altar puts on a mana ability and Carrion Feeder puts on
// an activated one.
func creatureSacrificeSpec() *game.TargetSpec {
	return &game.TargetSpec{
		Mode:  "permanent",
		Label: "a creature",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1,
		Max: 1,
	}
}

func battlefieldCardView(t *testing.T, v GameView, id uuid.UUID) CardView {
	t.Helper()
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("card %s missing from the battlefield view", id)
	return CardView{}
}

func containsID(opts *LegalTargetsView, id uuid.UUID) bool {
	if opts == nil {
		return false
	}
	for _, s := range opts.Cards {
		if s == id.String() {
			return true
		}
	}
	return false
}

func TestConvokeOptionsOfferAShroudedCreatureYouControl(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	shrouded := seatShroudedCreature(g, me)
	plain := seatCreature(g, me, "Plain Bear", 2, false, false)

	march := game.NewCard("March of the Multitudes", me)
	march.OracleID = marchOfTheMultitudesOracle
	march.ManaCost = "{X}{G}{W}{W}"
	march.TypeLine = "Instant"
	g.Seats[0].Hand.PushTop(march)

	v := ViewOfGame(g)
	var hand *CardView
	for i := range v.Seats[0].Hand.Cards {
		if v.Seats[0].Hand.Cards[i].InstanceID == march.InstanceID.String() {
			hand = &v.Seats[0].Hand.Cards[i]
		}
	}
	if hand == nil {
		t.Fatal("March of the Multitudes missing from the hand view")
	}
	if hand.TapCost == nil {
		t.Fatal("tap_cost is absent on a convoke card")
	}
	if !containsID(hand.TapCost.Options, plain) {
		t.Fatal("control: an ordinary untapped creature is missing from the convoke options")
	}
	if !containsID(hand.TapCost.Options, shrouded) {
		t.Error("a shrouded creature you control is missing from the convoke options; convoke does not target (CR 601.2h)")
	}
}

func TestManaAbilitySacrificeOptionsOfferAShroudedCreatureYouControl(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	shrouded := seatShroudedCreature(g, me)
	plain := seatCreature(g, me, "Plain Bear", 2, false, false)

	altar := game.NewCard("Test Altar", me)
	altar.TypeLine = "Artifact"
	// Non-empty so the battlefield stamp doesn't skip it; the
	// intrinsic mana ability wins over the catalog lookup.
	altar.OracleID = "00000000-0000-0000-0000-0000000000dd"
	altar.ManaAbilities = []game.ManaAbilityShape{{
		SacrificeOther: creatureSacrificeSpec(),
		Produced:       "{C}{C}",
		Label:          "Sacrifice a creature: Add {C}{C}",
	}}
	g.Battlefield.PushTop(altar)

	c := battlefieldCardView(t, ViewOfGame(g), altar.InstanceID)
	if len(c.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(c.ManaAbilities))
	}
	opts := c.ManaAbilities[0].SacrificeOptions
	if !containsID(opts, plain) {
		t.Fatal("control: an ordinary creature is missing from the mana ability's sacrifice options")
	}
	if !containsID(opts, shrouded) {
		t.Error("a shrouded creature you control is missing from the mana ability's sacrifice options; sacrificing does not target (CR 601.2h)")
	}
}

func TestActivatedAbilitySacrificeOptionsOfferAShroudedCreatureYouControl(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	shrouded := seatShroudedCreature(g, me)
	plain := seatCreature(g, me, "Plain Bear", 2, false, false)
	theirs := seatCreature(g, g.Seats[1].ID, "Their Bear", 2, false, false)

	feeder := game.NewCard("Test Feeder", me)
	feeder.TypeLine = "Artifact"
	feeder.OracleID = "00000000-0000-0000-0000-0000000000ee"
	feeder.ActivatedAbilities = []game.ActivatedAbilityShape{{
		Label: "Sacrifice a creature: do a thing",
		Cost:  game.AbilityCost{SacrificeOther: creatureSacrificeSpec()},
	}}
	g.Battlefield.PushTop(feeder)

	c := battlefieldCardView(t, ViewOfGame(g), feeder.InstanceID)
	if len(c.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(c.ActivatedAbilities))
	}
	opts := c.ActivatedAbilities[0].SacrificeOptions
	if !containsID(opts, plain) {
		t.Fatal("control: an ordinary creature is missing from the activated ability's sacrifice options")
	}
	if !containsID(opts, shrouded) {
		t.Error("a shrouded creature you control is missing from the activated ability's sacrifice options; sacrificing does not target (CR 601.2h)")
	}
	if containsID(opts, theirs) {
		t.Error("an opponent's creature is in the sacrifice options; CR 701.21a limits them to permanents you control")
	}
	if opts.Min != 1 || opts.Max != 1 {
		t.Errorf("sacrifice options min/max = %d/%d, want 1/1 from the spec", opts.Min, opts.Max)
	}
}
