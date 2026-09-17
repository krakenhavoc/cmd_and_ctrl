package protocol

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_n_view_test.go — #747 (ADR 0020 addendum §16): no new wire
// field. sacrifice_options.min / max carry the count on all three
// views (activated_abilities, mana_abilities, additional_cost), and
// the cards come in the enumerator's payment order, which is what the
// client's "Choose for me" button takes the first N of.

const (
	kuldothaForgemasterOracle = "b0f99367-4313-4bb1-a9fd-7711bc4ce40e"
	sacNViewRitesOracle       = "protocol-test-sacrifice-n-rites"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:       sacNViewRitesOracle,
		Name:           "Two-Creature Rites",
		AdditionalCost: effects.SacrificeNCost(2, "two creatures", effects.Creature()),
		OnResolve:      func(*game.StackItem, *effects.Context) error { return nil },
	})
}

func pushSacNCard(g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func idStringsOf(ids ...uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

func TestActivatedSacrificeOptionsCarryNAndPaymentOrder(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	// The source is an artifact with mana value 0, so only the
	// source-last rule puts it behind the other zero-cost artifact.
	forge := pushSacNCard(g, me, game.Card{Name: "Kuldotha Forgemaster", TypeLine: "Artifact Creature — Construct", OracleID: kuldothaForgemasterOracle, Power: 3, Toughness: 5})
	wurm := pushSacNCard(g, me, game.Card{Name: "Wurmcoil", TypeLine: "Artifact Creature — Phyrexian Wurm", ManaCost: "{6}", Power: 6, Toughness: 6})
	ring := pushSacNCard(g, me, game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"})
	orni := pushSacNCard(g, me, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter", ManaCost: "{0}"})
	treasure := pushSacNCard(g, me, game.Card{Name: "Treasure", TypeLine: "Token Artifact — Treasure"})
	pushSacNCard(g, g.Seats[1].ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})

	c := battlefieldCardView(t, ViewOfGame(g), forge)
	if len(c.ActivatedAbilities) != 1 || c.ActivatedAbilities[0].SacrificeOptions == nil {
		t.Fatalf("Kuldotha Forgemaster's ability view: %+v", c.ActivatedAbilities)
	}
	opts := c.ActivatedAbilities[0].SacrificeOptions
	if opts.Min != 3 || opts.Max != 3 {
		t.Errorf("sacrifice_options min/max = %d/%d, want 3/3", opts.Min, opts.Max)
	}
	want := idStringsOf(treasure, orni, forge, ring, wurm)
	if fmt.Sprint(opts.Cards) != fmt.Sprint(want) {
		t.Errorf("sacrifice_options.cards = %v, want payment order %v (token, mana value 0, the source, mana value 1, mana value 6)", opts.Cards, want)
	}
}

func TestManaSacrificeOptionsCarryNAndDropTheSourceWhenItPaysToo(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	a := seatCreature(g, me, "Bear A", 2, false, false)
	b := seatCreature(g, me, "Bear B", 2, false, false)
	spec := creatureSacrificeSpec().WithCount(2, 2)

	altar := game.NewCard("Test Altar", me)
	altar.TypeLine = "Artifact Creature — Golem"
	altar.OracleID = "00000000-0000-0000-0000-0000000000df"
	altar.ManaAbilities = []game.ManaAbilityShape{{
		SacrificeCost:  true,
		SacrificeOther: spec,
		Produced:       "{C}{C}{C}{C}",
		Label:          "Sacrifice this and two creatures: Add {C}{C}{C}{C}",
	}}
	g.Battlefield.PushTop(altar)

	c := battlefieldCardView(t, ViewOfGame(g), altar.InstanceID)
	if len(c.ManaAbilities) != 1 || c.ManaAbilities[0].SacrificeOptions == nil {
		t.Fatalf("mana ability view: %+v", c.ManaAbilities)
	}
	opts := c.ManaAbilities[0].SacrificeOptions
	if opts.Min != 2 || opts.Max != 2 {
		t.Errorf("min/max = %d/%d, want 2/2", opts.Min, opts.Max)
	}
	if containsID(opts, altar.InstanceID) {
		t.Error("the source is sacrificed by the cost already and cannot also be one of the two")
	}
	if !containsID(opts, a) || !containsID(opts, b) {
		t.Errorf("both creatures are options: %v", opts.Cards)
	}
}

func TestAdditionalCostSacrificeOptionsCarryN(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	big := pushSacNCard(g, me.ID, game.Card{Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{5}", Power: 5, Toughness: 5})
	small := pushSacNCard(g, me.ID, game.Card{Name: "Small", TypeLine: "Creature — Elf", ManaCost: "{G}", Power: 1, Toughness: 1})
	tok := pushSacNCard(g, me.ID, game.Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1})

	rites := game.NewCard("Two-Creature Rites", me.ID)
	rites.OracleID = sacNViewRitesOracle
	rites.TypeLine = "Sorcery"
	rites.ManaCost = "{B}"
	me.Hand.PushTop(rites)

	v := ViewOfGame(g)
	var hand *CardView
	for i := range v.Seats[0].Hand.Cards {
		if v.Seats[0].Hand.Cards[i].InstanceID == rites.InstanceID.String() {
			hand = &v.Seats[0].Hand.Cards[i]
		}
	}
	if hand == nil || hand.AdditionalCost == nil || hand.AdditionalCost.SacrificeOptions == nil {
		t.Fatalf("additional_cost.sacrifice_options absent: %+v", hand)
	}
	opts := hand.AdditionalCost.SacrificeOptions
	if opts.Min != 2 || opts.Max != 2 {
		t.Errorf("min/max = %d/%d, want 2/2", opts.Min, opts.Max)
	}
	if want := idStringsOf(tok, small, big); fmt.Sprint(opts.Cards) != fmt.Sprint(want) {
		t.Errorf("cards = %v, want %v", opts.Cards, want)
	}
	if hand.AdditionalCost.Label != "Sacrifice two creatures" {
		t.Errorf("label = %q", hand.AdditionalCost.Label)
	}
}
