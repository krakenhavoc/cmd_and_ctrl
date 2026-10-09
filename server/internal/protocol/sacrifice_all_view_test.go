package protocol

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_all_view_test.go — #2097: a "sacrifice all creatures you
// control" additional cost (Soulblast) ships sacrifice_options with
// `all`, min = max = the number listed, and the cards the cast will
// take: the viewer's creatures, not an opponent's and not a phased-out
// one.

const soulblastViewOracle = "18d4c57b-e2bf-47a0-8823-c4a79498a7ff"

func soulblastHandView(t *testing.T, g *game.Game, me *game.Player) *CardView {
	t.Helper()
	sb := game.NewCard("Soulblast", me.ID)
	sb.OracleID = soulblastViewOracle
	sb.TypeLine = "Instant"
	sb.ManaCost = "{3}{R}{R}{R}"
	me.Hand.PushTop(sb)
	v := ViewOfGame(g)
	for i := range v.Seats[0].Hand.Cards {
		if v.Seats[0].Hand.Cards[i].InstanceID == sb.InstanceID.String() {
			return &v.Seats[0].Hand.Cards[i]
		}
	}
	t.Fatal("Soulblast is not in the hand view")
	return nil
}

func TestSacrificeAllCostShipsEveryCreatureAndTheFlag(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	big := pushSacNCard(g, me.ID, game.Card{Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{5}", Power: 5, Toughness: 5})
	tok := pushSacNCard(g, me.ID, game.Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1})
	away := pushSacNCard(g, me.ID, game.Card{Name: "Away", TypeLine: "Creature — Elf", Power: 1, Toughness: 1})
	pushSacNCard(g, g.Seats[1].ID, game.Card{Name: "Theirs", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3})
	pushSacNCard(g, me.ID, game.Card{Name: "Relic", TypeLine: "Artifact"})
	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, away); err != nil {
			t.Fatal(err)
		}
	})

	hand := soulblastHandView(t, g, me)
	if hand.AdditionalCost == nil || hand.AdditionalCost.SacrificeOptions == nil {
		t.Fatalf("additional_cost.sacrifice_options absent: %+v", hand.AdditionalCost)
	}
	opts := hand.AdditionalCost.SacrificeOptions
	if !opts.All {
		t.Error("all is not set")
	}
	if opts.Min != 2 || opts.Max != 2 {
		t.Errorf("min/max = %d/%d, want 2/2", opts.Min, opts.Max)
	}
	if want := idStringsOf(tok, big); fmt.Sprint(opts.Cards) != fmt.Sprint(want) {
		t.Errorf("cards = %v, want %v", opts.Cards, want)
	}
	if hand.AdditionalCost.Label != "Sacrifice all creatures you control" {
		t.Errorf("label = %q", hand.AdditionalCost.Label)
	}
}

func TestSacrificeAllCostWithNoCreaturesShipsAnEmptyDemand(t *testing.T) {
	g := buildActiveGame(t)
	hand := soulblastHandView(t, g, g.Seats[0])
	opts := hand.AdditionalCost.SacrificeOptions
	if opts == nil || !opts.All || opts.Min != 0 || opts.Max != 0 || len(opts.Cards) != 0 {
		t.Fatalf("sacrifice_options = %+v, want all with nothing listed", opts)
	}
}
