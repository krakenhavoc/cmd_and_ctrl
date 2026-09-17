package effects

import (
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Owner decision (2026-09-17): "any color" mana offers ALL FIVE
// colours, with the commander's colour identity listed first — no
// narrowing. Only the cards whose printed text says "any color in your
// commander's color identity" still narrow.

// monoGreenFirst is what a five-colour pick looks like under a
// mono-green commander: G first, then the rest in WUBRG order.
var monoGreenFirst = []string{"G", "W", "U", "B", "R"}

func TestAnyColorManaOffersAllFiveIdentityFirst(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
	}{
		{"Birds of Paradise", "Creature — Bird", "d3a0b660-358c-41bd-9cd2-41fbf3491b1a"},
		{"Lotus Petal", "Artifact", "32e5339e-9e4f-46f8-b305-f9d6d3ba8bb5"},
		{"Coalition Relic", "Artifact", "008cb342-79f5-4df6-a6b7-0e9e22ed693f"},
		{"Ornithopter of Paradise", "Artifact Creature — Thopter", "3a940dfa-a026-4969-8981-ef90cdfbe9ef"},
		{"Cultivator's Caravan", "Artifact — Vehicle", "c1eb530c-dd36-40ae-8617-6bb6969565e1"},
		{"Dragonstorm Globe", "Artifact", "f6de5bd7-7704-4a0c-a27a-9e565e49f5e9"},
		{"Decanter of Endless Water", "Artifact", "8ae98ef8-8f52-4877-a08c-1fae5514184e"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			riderGiveCommander(g, me, "{2}{G}")
			src := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, src, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			pick := riderLatestManaPick(g, me.ID)
			if pick == nil || !reflect.DeepEqual(pick.ColorOptions, monoGreenFirst) {
				t.Fatalf("pick = %+v, want %v", pick, monoGreenFirst)
			}
			// An off-identity colour is a legal answer: the card
			// prints "any color".
			if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
				t.Fatalf("ResolveManaChoice R: %v", err)
			}
			if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "R" {
				t.Errorf("pool = %+v, want one {R}", me.ManaPool)
			}
		})
	}

	t.Run("Treasure", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		riderGiveCommander(g, me, "{2}{G}")
		tok := TreasureToken()
		tok.InstanceID = uuid.New()
		tok.Owner, tok.Controller = me.ID, me.ID
		g.Battlefield.PushTop(tok)
		if err := g.ActivateManaAbility(me.ID, tok.InstanceID, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		pick := riderLatestManaPick(g, me.ID)
		if pick == nil || !reflect.DeepEqual(pick.ColorOptions, monoGreenFirst) {
			t.Fatalf("pick = %+v, want %v", pick, monoGreenFirst)
		}
	})

	t.Run("Phyrexian Altar", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		riderGiveCommander(g, me, "{2}{G}")
		altar := pushCatalogPermanent(g, me.ID, "Phyrexian Altar", "Artifact", "8d02b297-97c4-4379-9862-0a462400f66f", false)
		fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Bear", "", false)
		if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		pick := riderLatestManaPick(g, me.ID)
		if pick == nil || !reflect.DeepEqual(pick.ColorOptions, monoGreenFirst) {
			t.Fatalf("pick = %+v, want %v", pick, monoGreenFirst)
		}
	})
}

// A two-colour land names its colours; both stay on offer, the
// commander's colour first.
func TestDualLandOffersBothColoursIdentityFirst(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	riderGiveCommander(g, me, "{B}")
	land := seedPermanentWithOracle(g, me.ID, "Temple of Silence", "Land", templeOfSilenceOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || !reflect.DeepEqual(pick.ColorOptions, []string{"B", "W"}) {
		t.Fatalf("Temple of Silence under a mono-black commander: pick %+v, want [B W]", pick)
	}
}

// The contrast that keeps the flag honest: a card whose printed text
// says "in your commander's color identity" still narrows.
func TestCommanderIdentityTextStillNarrows(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
	}{
		{"Command Tower", "Land", "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"},
		{"Arcane Signet", "Artifact", "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{"Commander's Sphere", "Artifact", commandersSphereOracle},
		{"Path of Ancestry", "Land", pathOfAncestryOracle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			riderGiveCommander(g, me, "{2}{G}")
			src := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, src, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			pick := riderLatestManaPick(g, me.ID)
			if pick == nil || !reflect.DeepEqual(pick.ColorOptions, []string{"G"}) {
				t.Fatalf("pick = %+v, want narrowed to [G]", pick)
			}
		})
	}
}

// identityNarrowingCards is the whole list of catalog cards allowed to
// narrow. Adding a card here needs the printed clause, which
// TestNarrowToCommanderIdentityMatchesOracleText checks against the
// real dump.
var identityNarrowingCards = []string{"Arcane Signet", "Command Tower", "Commander's Sphere", "Path of Ancestry"}

func TestOnlyCommanderIdentityCardsNarrow(t *testing.T) {
	var got []string
	for _, s := range All() {
		for _, a := range s.ManaAbilities {
			if a.NarrowToCommanderIdentity {
				got = append(got, s.Name)
				break
			}
		}
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, identityNarrowingCards) {
		t.Errorf("cards narrowing to commander identity = %v, want exactly %v", got, identityNarrowingCards)
	}
}

// TestNarrowToCommanderIdentityMatchesOracleText holds every catalog
// mana ability to its printed text: a card narrows if and only if its
// oracle text says "commander's color identity".
//
// Dump-gated like the other real-dump tests:
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json go test ./internal/cards/effects/ -run NarrowToCommanderIdentity
func TestNarrowToCommanderIdentityMatchesOracleText(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to check identity narrowing against the real oracle text")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("load %s: %v", path, err)
	}
	for _, s := range All() {
		if len(s.ManaAbilities) == 0 {
			continue
		}
		oid, err := uuid.Parse(s.OracleID)
		if err != nil {
			continue // a face key ("<oracle>#1"), checked through its front
		}
		card, ok := idx.FindByOracleID(oid)
		if !ok {
			continue
		}
		text := card.OracleText
		for _, f := range card.CardFaces {
			text += "\n" + f.OracleText
		}
		prints := strings.Contains(text, "commander's color identity")
		narrows := false
		for _, a := range s.ManaAbilities {
			narrows = narrows || a.NarrowToCommanderIdentity
		}
		if prints != narrows {
			t.Errorf("%s: oracle text mentions commander's color identity = %v, but NarrowToCommanderIdentity = %v:\n%s",
				s.Name, prints, narrows, text)
		}
	}
}
