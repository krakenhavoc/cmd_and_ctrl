package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // the real catalog's declarations
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// purpose_view_test.go — ADR 0126 §6: a card's declared purpose on the
// wire, in every slot, and who may see it.

const (
	pvNightsWhisper = "7ffae8f8-3006-4969-a339-6d30678f87ea"
	pvFarewell      = "4eb813fd-2d5a-4b02-8193-662681ef4e7d"
	pvCyclonicRift  = "d75b9c82-1b49-4c3e-a1b5-aeef57d6644b"
	pvMulldrifter   = "24d0f5e7-0d9e-4b76-900e-a7274e80312d"
	pvBloodArtist   = "310f141c-7f37-4729-aed6-dd9c09db448d"
	pvMaryRead      = "5182de2d-aceb-450e-bd20-8bc7db124334"
	pvWrathOfGod    = "34515b16-c9a4-4f98-8c77-416a7a523407"
	pvMako          = "e349be42-5f14-44a9-9608-281985c10e2d"
)

func seatZones(v GameView, seat string) (hand ZoneView) {
	for _, s := range v.Seats {
		if s.ID == seat {
			return s.Hand
		}
	}
	return ZoneView{}
}

func TestPurposeRidesTheViewInEverySlot(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	whisper := rowsFixtureCard("Night's Whisper", "Sorcery", pvNightsWhisper, me.ID)
	farewell := rowsFixtureCard("Farewell", "Sorcery", pvFarewell, me.ID)
	rift := rowsFixtureCard("Cyclonic Rift", "Instant", pvCyclonicRift, me.ID)
	drifter := rowsFixtureCard("Mulldrifter", "Creature — Elemental", pvMulldrifter, me.ID)
	artist := rowsFixtureCard("Blood Artist", "Creature — Vampire", pvBloodArtist, me.ID)
	mary := rowsFixtureCard("Mary Read and Anne Bonny", "Legendary Creature — Human Assassin Pirate", pvMaryRead, me.ID)
	hidden := rowsFixtureCard("Mulldrifter", "Creature — Elemental", pvMulldrifter, me.ID)
	wrath := rowsFixtureCard("Wrath of God", "Sorcery", pvWrathOfGod, opp.ID)
	mako := rowsFixtureCard("Marauding Mako", "Creature — Shark Pirate", pvMako, me.ID)
	g.WithWriteLock(func() {
		for _, c := range []game.Card{whisper, farewell, rift, mako} {
			c.KnownBy = map[uuid.UUID]bool{me.ID: true}
			me.Hand.PushTop(c)
		}
		for _, c := range []game.Card{drifter, artist, mary} {
			c.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
			g.Battlefield.PushTop(c)
		}
		hidden.FaceDown = true
		hidden.FaceDownKind = game.FaceDownMorphed
		hidden.KnownBy = map[uuid.UUID]bool{me.ID: true}
		g.Battlefield.PushTop(hidden)
		wrath.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
		g.Stack.PushTop(wrath)
	})
	g.BumpLayerVersionForTest()
	v := ViewOfGame(g)

	t.Run("the owner's hand", func(t *testing.T) {
		hand := seatZones(FilterViewFor(v, me.ID.String()), me.ID.String())
		c := findCardView(t, hand, whisper.InstanceID)
		if c == nil || c.Purpose == nil || *c.Purpose != (PurposeView{Draws: 2}) {
			t.Fatalf("Night's Whisper purpose = %+v, want draws 2", c)
		}
		c = findCardView(t, hand, farewell.InstanceID)
		if c == nil || c.Purpose != nil {
			t.Fatalf("Farewell carries a card-level purpose %+v; a modal card declares per bullet", c)
		}
		if c.Modes == nil || len(c.Modes.Options) != 4 {
			t.Fatalf("Farewell modes = %+v", c.Modes)
		}
		if p := c.Modes.Options[1].Purpose; p == nil || p.Sweep == nil || *p.Sweep != (SweepView{Matches: "creatures", How: "exile"}) {
			t.Errorf("Farewell's creature bullet purpose = %+v", p)
		}
		if p := c.Modes.Options[3].Purpose; p != nil {
			t.Errorf("Farewell's graveyard bullet declares %+v; a graveyard is not a sweep class", p)
		}
		c = findCardView(t, hand, rift.InstanceID)
		if c == nil {
			t.Fatal("Cyclonic Rift is missing")
		}
		var overload *AlternativeCostView
		for i := range c.AlternativeCosts {
			if c.AlternativeCosts[i].Key == "overload" {
				overload = &c.AlternativeCosts[i]
			}
		}
		if overload == nil {
			t.Fatalf("Cyclonic Rift offers %+v, want its overload", c.AlternativeCosts)
		}
		want := SweepView{Matches: "nonland_permanents", How: "bounce", OpponentsOnly: true}
		if overload.Purpose == nil || overload.Purpose.Sweep == nil || *overload.Purpose.Sweep != want {
			t.Errorf("the overload's purpose = %+v, want %+v", overload.Purpose, want)
		}
	})

	t.Run("an opponent never sees a hand card's purpose", func(t *testing.T) {
		hand := seatZones(FilterViewFor(v, opp.ID.String()), me.ID.String())
		for _, c := range hand.Cards {
			if c.Purpose != nil {
				t.Errorf("a hidden hand card carries purpose %+v to an opponent", c.Purpose)
			}
			for _, r := range c.AbilityRows {
				if r.Purpose != nil {
					t.Errorf("a hidden hand card's row carries purpose %+v to an opponent", r.Purpose)
				}
			}
		}
	})

	t.Run("the owner sees a discard payoff on a hand card's row", func(t *testing.T) {
		hand := seatZones(FilterViewFor(v, me.ID.String()), me.ID.String())
		c := findCardView(t, hand, mako.InstanceID)
		if c == nil || len(c.AbilityRows) == 0 || c.AbilityRows[0].Purpose == nil {
			t.Fatalf("Marauding Mako in hand = %+v, want its discard payoff row", c)
		}
		d := c.AbilityRows[0].Purpose.DiscardPayoff
		if d == nil || !d.Any || len(d.Types) != 0 || d.Counters != 1 || d.Tokens != 0 {
			t.Errorf("Marauding Mako's discard payoff = %+v, want any card, one counter", d)
		}
	})

	for _, viewer := range []struct{ name, id string }{
		{"controller", me.ID.String()}, {"opponent", opp.ID.String()}, {"spectator", SpectatorViewerID},
	} {
		t.Run("the battlefield and the stack, "+viewer.name, func(t *testing.T) {
			fv := FilterViewFor(v, viewer.id)
			if c := findCardView(t, fv.Battlefield, drifter.InstanceID); c == nil || c.Purpose == nil || *c.Purpose != (PurposeView{Draws: 2}) {
				t.Errorf("Mulldrifter's enters purpose = %+v, want draws 2", c)
			}
			c := findCardView(t, fv.Battlefield, artist.InstanceID)
			if c == nil || len(c.AbilityRows) != 1 || c.AbilityRows[0].Purpose == nil || !c.AbilityRows[0].Purpose.DeathPayoff {
				t.Errorf("Blood Artist's row = %+v, want a death payoff", c)
			}
			c = findCardView(t, fv.Battlefield, mary.InstanceID)
			if c == nil || len(c.ActivatedAbilities) == 0 || c.ActivatedAbilities[0].Purpose == nil ||
				*c.ActivatedAbilities[0].Purpose != (PurposeView{Draws: 1, Discards: 1}) {
				t.Errorf("Mary Read's loot row = %+v, want draws 1 and discards 1", c)
			}
			var payoff *DiscardPayoffView
			for _, r := range c.AbilityRows {
				if r.Kind == "triggered" && r.Purpose != nil {
					payoff = r.Purpose.DiscardPayoff
				}
			}
			if payoff == nil || payoff.Any || strings.Join(payoff.Types, ",") != "island,pirate,vehicle" || payoff.Tokens != 1 {
				t.Errorf("Mary Read's discard trigger payoff = %+v, want an Island, Pirate or Vehicle card for one token", payoff)
			}
			if c := findCardView(t, fv.Battlefield, hidden.InstanceID); c == nil || c.Purpose != nil {
				t.Errorf("a face-down permanent carries purpose %+v (CR 708.2a: it has no text)", c)
			}
			c = findCardView(t, fv.Stack, wrath.InstanceID)
			if c == nil || c.Purpose == nil || c.Purpose.Sweep == nil || *c.Purpose.Sweep != (SweepView{Matches: "creatures", How: "destroy"}) {
				t.Errorf("Wrath of God on the stack = %+v, want a creature sweep", c)
			}
		})
	}

	t.Run("the wire shape", func(t *testing.T) {
		raw, err := json.Marshal(FilterViewFor(v, me.ID.String()))
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{
			`"purpose":{"draws":2}`,
			`"purpose":{"sweep":{"matches":"creatures","how":"destroy"}}`,
			`"purpose":{"death_payoff":true}`,
			`"purpose":{"draws":1,"discards":1}`,
			`"purpose":{"discard_payoff":{"types":["island","pirate","vehicle"],"tokens":1}}`,
			`"purpose":{"discard_payoff":{"any":true,"counters":1}}`,
		} {
			if !strings.Contains(string(raw), want) {
				t.Errorf("the frame lacks %s", want)
			}
		}
	})
}

// The redaction itself: a non-knower loses the purpose with the rows.
func TestPurposeIsRedactedForANonKnower(t *testing.T) {
	stamped := CardView{InstanceID: "x", Purpose: &PurposeView{Draws: 2}}
	if got := redactCardForViewer(stamped, false); got.Purpose != nil {
		t.Errorf("a non-knower keeps purpose %+v", got.Purpose)
	}
	if got := redactCardForViewer(stamped, true); got.Purpose == nil {
		t.Error("a knower lost the purpose")
	}
}
