package effects

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// awaken_wire_test.go — ADR 0135 §3 (#2411): the rest of the fifteen
// awaken cards end to end, the offer on the wire and in the bot's
// enumerator, and Register's guard. awaken_cards_test.go holds the rules.

// Each cast for its awaken cost: the printed instruction happens and the
// land awakens.
func TestEveryAwakenCardDoesItsOwnThingThenAwakens(t *testing.T) {
	type setup struct {
		g        *game.Game
		me, opp  *game.Player
		land     uuid.UUID
		creature uuid.UUID
	}
	for _, c := range []struct {
		name, typeLine, cost, oracle string
		n                            int
		targets                      func(s *setup) []game.TargetRef
		check                        func(t *testing.T, s *setup)
	}{
		{"Boiling Earth", "Sorcery", "{1}{R}", boilingEarthOracle, 4,
			func(s *setup) []game.TargetRef { return nil },
			func(t *testing.T, s *setup) {
				if got := pr6Marked(s.g, s.creature); got != 1 {
					t.Errorf("the opponent's creature has %d damage, want 1", got)
				}
			}},
		{"Mire's Malice", "Sorcery", "{3}{B}", miresMaliceOracle, 3,
			func(s *setup) []game.TargetRef {
				return []game.TargetRef{{Kind: game.TargetPlayer, ID: s.opp.ID}}
			},
			func(t *testing.T, s *setup) {
				found := false
				for _, pc := range s.g.PendingChoices {
					if pc != nil && pc.Chooser == s.opp.ID {
						found = true
					}
				}
				if !found {
					t.Error("the target opponent was not asked to discard")
				}
			}},
		{"Part the Waterveil", "Sorcery", "{4}{U}{U}", partTheWaterveilOracle, 6,
			func(s *setup) []game.TargetRef { return nil },
			func(t *testing.T, s *setup) {
				if len(s.g.ExtraTurns) != 1 {
					t.Errorf("%d extra turns queued, want 1", len(s.g.ExtraTurns))
				}
			}},
		{"Roil Spout", "Sorcery", "{1}{W}{U}", roilSpoutOracle, 4,
			func(s *setup) []game.TargetRef { return []game.TargetRef{awakenRef(s.creature)} },
			func(t *testing.T, s *setup) {
				if findBattlefieldCardForTest(s.g, s.creature) != nil || !s.opp.Library.Contains(s.creature) {
					t.Error("the creature is not in its owner's library")
				}
			}},
		{"Rush of Ice", "Sorcery", "{U}", rushOfIceOracle, 3,
			func(s *setup) []game.TargetRef { return []game.TargetRef{awakenRef(s.creature)} },
			func(t *testing.T, s *setup) {
				if !tappedForTest(t, s.g, s.creature) {
					t.Error("the creature was not tapped")
				}
				if tappedForTest(t, s.g, s.land) {
					t.Error("the awaken land was tapped too")
				}
			}},
		{"Sheer Drop", "Sorcery", "{2}{W}", sheerDropOracle, 3,
			func(s *setup) []game.TargetRef {
				s.g.WithWriteLock(func() { findBattlefieldCardForTest(s.g, s.creature).Tapped = true })
				return []game.TargetRef{awakenRef(s.creature)}
			},
			func(t *testing.T, s *setup) {
				if findBattlefieldCardForTest(s.g, s.creature) != nil {
					t.Error("the tapped creature was not destroyed")
				}
			}},
	} {
		t.Run(c.name, func(t *testing.T) {
			g, me, opp, land := awakenTable(t)
			s := &setup{g: g, me: me, opp: opp, land: land, creature: pr7Creature(g, opp.ID, "Theirs", 2, "G")}
			card := handCardOf(g, me, c.name, c.typeLine, c.cost, c.oracle)
			targets := append(c.targets(s), awakenRef(land))
			if err := castAwaken(g, me, card, targets...); err != nil {
				t.Fatalf("cast: %v", err)
			}
			passPriorityAroundTable(t, g)
			c.check(t, s)
			if counters, p, tough, elemental, hasty := awakenedLand(t, g, land); counters != c.n || p != c.n || tough != c.n || !elemental || !hasty {
				t.Errorf("land: %d counters, %d/%d, elemental %v, haste %v; want a hasty %d/%d Elemental",
					counters, p, tough, elemental, hasty, c.n, c.n)
			}
		})
	}
}

// Scatter to the Winds counters the spell and awakens.
func TestScatterToTheWindsCountersAndAwakens(t *testing.T) {
	g, me, opp, land := awakenTable(t)
	spell := handCardOf(g, opp, "Spell", "Instant", "{R}", "")
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent's spell: %v", err)
	}
	scatter := handCardOf(g, me, "Scatter to the Winds", "Instant", "{1}{U}{U}", scatterToTheWindsOracle)
	if err := castAwaken(g, me, scatter, awakenRef(spell), awakenRef(land)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Error("the spell was not countered into its owner's graveyard")
	}
	if counters, _, _, elemental, _ := awakenedLand(t, g, land); counters != 3 || !elemental {
		t.Errorf("land: %d counters, elemental %v", counters, elemental)
	}
}

// The wire and the bot: the awaken offer ships two clauses, the spell's
// own then "target land you control" (lands the caster controls only),
// and its purpose; the enumerator offers an awaken cast naming both, and
// the engine accepts it.
func TestRuinousPathAwakenOfferOnTheWireAndInTheEnumerator(t *testing.T) {
	g, me, opp, _ := awakenTable(t)
	for i := 0; i < 6; i++ {
		pushEarthbendLand(g, me.ID, "Swamp", "Basic Land — Swamp")
	}
	pr7Creature(g, opp.ID, "Victim", 2, "G")
	pushEarthbendLand(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	path := handCardOf(g, me, "Ruinous Path", "Sorcery", "{1}{B}{B}", ruinousPathOracle)

	var offer *protocol.AlternativeCostView
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			if s.Hand.Cards[i].InstanceID != path.String() {
				continue
			}
			for j := range s.Hand.Cards[i].AlternativeCosts {
				if s.Hand.Cards[i].AlternativeCosts[j].Key == "awaken" {
					offer = &s.Hand.Cards[i].AlternativeCosts[j]
				}
			}
		}
	}
	if offer == nil {
		t.Fatal("the awaken offer is not on the wire")
	}
	if offer.Label != "Awaken 4—{5}{B}{B}" || offer.Purpose == nil || offer.Purpose.AwakenLand != 4 {
		t.Errorf("offer label %q, purpose %+v", offer.Label, offer.Purpose)
	}
	if len(offer.Clauses) != 2 || offer.Clauses[1].Label != "target land you control" || len(offer.Clauses[1].Cards) == 0 {
		t.Fatalf("offer clauses %+v, want the spell's and the land", offer.Clauses)
	}
	for _, id := range offer.Clauses[1].Cards {
		c := findBattlefieldCardForTest(g, uuid.MustParse(id))
		if c == nil || c.Controller != me.ID || !c.IsLand() {
			t.Errorf("land clause lists %s, which is not a land the caster controls", id)
		}
	}

	type move struct {
		AlternativeCost string `json:"alternative_cost"`
		Targets         []struct {
			ID string `json:"id"`
		} `json:"targets"`
	}
	found := false
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != path {
			continue
		}
		var p move
		if err := json.Unmarshal(m.Params, &p); err == nil && p.AlternativeCost == "awaken" && len(p.Targets) == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("the enumerator offers no awaken cast naming a creature and a land")
	}
}

// effects.Register holds an awaken offer to the shape AwakenIfPaid reads.
func TestRegisterRefusesMalformedAwakenOffers(t *testing.T) {
	printed := TargetCreature("target creature")
	other := TargetCreature("target creature you control", YouControl())
	notAwaken := Overload("{4}")
	notAwaken.Purpose = game.Purpose{AwakenLand: 2}
	noPurpose := Awaken(2, "{4}", printed)
	noPurpose.Purpose = game.Purpose{}
	for _, c := range []struct {
		name string
		spec Spec
		want string
	}{
		{"drifted clause", Spec{Targets: printed, AlternativeCosts: []game.AlternativeCost{Awaken(2, "{4}", other)}}, "not the spell's own"},
		{"missing printed", Spec{Targets: printed, AlternativeCosts: []game.AlternativeCost{Awaken(2, "{4}", nil)}}, "clauses"},
		{"modal", Spec{Modes: ChooseOne(Mode("A."), Mode("B.")), AlternativeCosts: []game.AlternativeCost{Awaken(2, "{4}", nil)}}, "modal"},
		{"not awaken", Spec{Targets: printed, AlternativeCosts: []game.AlternativeCost{notAwaken}}, "not awaken"},
		{"no purpose", Spec{Targets: printed, AlternativeCosts: []game.AlternativeCost{noPurpose}}, "no AwakenLand"},
		{"purpose on the card", Spec{Purpose: game.Purpose{AwakenLand: 2}}, "AwakenLand off an alternative cost"},
	} {
		c.spec.OracleID, c.spec.Name = "test-0135-awaken-"+c.name, "Test "+c.name
		msg := registerPanics(c.spec)
		if !strings.Contains(msg, c.want) {
			t.Errorf("%s: panic %q, want one mentioning %q", c.name, msg, c.want)
		}
	}
	if printed.ClauseCount() != 1 {
		t.Errorf("Awaken grew the printed statement to %d clauses", printed.ClauseCount())
	}
}
