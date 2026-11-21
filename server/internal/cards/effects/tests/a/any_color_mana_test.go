package testsa

import (
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.RiderGiveCommander(g, me, "{2}{G}")
			src := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, src, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			pick := effectstest.RiderLatestManaPick(g, me.ID)
			if pick == nil || !reflect.DeepEqual(pick.ColorOptions, effectstest.MonoGreenFirst) {
				t.Fatalf("pick = %+v, want %v", pick, effectstest.MonoGreenFirst)
			}
			// An off-identity colour is a legal answer: the card
			// prints "any color".
			if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
				t.Fatalf("ResolveManaChoice R: %v", err)
			}
			if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "R" {
				t.Errorf("pool = %+v, want one {R}", me.ManaPool)
			}
		})
	}

	t.Run("Treasure", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.RiderGiveCommander(g, me, "{2}{G}")
		tok := effects.TreasureToken()
		tok.InstanceID = uuid.New()
		tok.Owner, tok.Controller = me.ID, me.ID
		g.Battlefield.PushTop(tok)
		if err := g.ActivateManaAbility(me.ID, tok.InstanceID, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		pick := effectstest.RiderLatestManaPick(g, me.ID)
		if pick == nil || !reflect.DeepEqual(pick.ColorOptions, effectstest.MonoGreenFirst) {
			t.Fatalf("pick = %+v, want %v", pick, effectstest.MonoGreenFirst)
		}
	})

	t.Run("Phyrexian Altar", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.RiderGiveCommander(g, me, "{2}{G}")
		altar := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Altar", "Artifact", "8d02b297-97c4-4379-9862-0a462400f66f", false)
		fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Bear", "", false)
		if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		pick := effectstest.RiderLatestManaPick(g, me.ID)
		if pick == nil || !reflect.DeepEqual(pick.ColorOptions, effectstest.MonoGreenFirst) {
			t.Fatalf("pick = %+v, want %v", pick, effectstest.MonoGreenFirst)
		}
	})
}

// A two-colour land names its colours; both stay on offer, the
// commander's colour first.
func TestDualLandOffersBothColoursIdentityFirst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.RiderGiveCommander(g, me, "{B}")
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Temple of Silence", "Land", effectstest.TempleOfSilenceOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
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
		{"Commander's Sphere", "Artifact", effectstest.CommandersSphereOracle},
		{"Path of Ancestry", "Land", effectstest.PathOfAncestryOracle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.RiderGiveCommander(g, me, "{2}{G}")
			src := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, src, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			pick := effectstest.RiderLatestManaPick(g, me.ID)
			if pick == nil || !reflect.DeepEqual(pick.ColorOptions, []string{"G"}) {
				t.Fatalf("pick = %+v, want narrowed to [G]", pick)
			}
		})
	}
}

// CR 903.4f (#844), and the official rulings on all four cards: with
// no commander, or with a commander whose colour identity is
// colourless, "any color in your commander's color identity" adds no
// mana. No token, and no picker to answer.
//
// A catalog game's seats start with no commander at all, which is the
// first case; the second stamps an imported colourless one, the shape
// deck.ToGameCard gives a real Kozilek.
func TestCommanderIdentityTextAddsNothingWithNoIdentity(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
	}{
		{"Command Tower", "Land", "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"},
		{"Arcane Signet", "Artifact", "0bc7f093-bef0-4f1a-852c-4b75ebf54838"},
		{"Commander's Sphere", "Artifact", effectstest.CommandersSphereOracle},
		{"Path of Ancestry", "Land", effectstest.PathOfAncestryOracle},
	}
	for _, seat := range []struct {
		name string
		give func(g *game.Game, p *game.Player)
	}{
		{"no commander", func(*game.Game, *game.Player) {}},
		{"colourless commander", func(g *game.Game, p *game.Player) {
			effectstest.RiderGiveImportedCommander(g, p, "Kozilek, Butcher of Truth", "{10}", nil)
		}},
	} {
		t.Run(seat.name, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					g := effectstest.NewCatalogGame(t)
					me := g.Seats[0]
					seat.give(g, me)
					src := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
					if err := g.ActivateManaAbility(me.ID, src, 0, game.ManaAbilityParams{}); err != nil {
						t.Fatalf("ActivateManaAbility: %v", err)
					}
					if pick := effectstest.RiderLatestManaPick(g, me.ID); pick != nil {
						t.Errorf("queued a pick of %v; the ability adds no mana", pick.ColorOptions)
					}
					if me.ManaPool.Total() != 0 {
						t.Errorf("pool = %+v, want empty", me.ManaPool)
					}
				})
			}
		})
	}
}

// identityNarrowingCards is the whole list of catalog cards allowed to
// narrow. Adding a card here needs the printed clause, which
// TestNarrowToCommanderIdentityMatchesOracleText checks against the
// real dump.
var identityNarrowingCards = []string{"Arcane Signet", "Command Tower", "Commander's Sphere", "Hidden Hideout", "Opal Palace", "Path of Ancestry"}

func TestOnlyCommanderIdentityCardsNarrow(t *testing.T) {
	var got []string
	for _, s := range effects.All() {
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
	for _, s := range effects.All() {
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
