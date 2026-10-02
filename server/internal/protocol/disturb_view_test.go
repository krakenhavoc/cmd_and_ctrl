package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// disturb_view_test.go — ADR 0107 §4 (#1855). The graveyard cast menu
// is read straight off this projection, so a disturb card in its
// owner's graveyard has to arrive as a cast surface whose only price is
// "Disturb {cost}", carrying the BACK face's target clause: the client
// takes an offer's target_mode / legal_targets as the spell's clause
// (castTargetOverride), and the spell is the back face (CR 712.8c).

const drogskolInfantryOracle = "389bcb9f-4e66-4704-9968-a1c1574ec2c8"

func TestDisturbOfferCarriesTheBackFacesTargetClause(t *testing.T) {
	g := buildActiveGame(t)
	castWindowOpen(t, g)
	me := g.Seats[0]
	seen := map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}

	bear := game.NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Power, bear.Toughness = 2, 2
	bear.KnownBy = seen
	g.Battlefield.PushTop(bear)

	infantry := game.Card{
		InstanceID: uuid.New(), OracleID: drogskolInfantryOracle,
		Owner: me.ID, Controller: me.ID, Layout: game.LayoutTransform, KnownBy: seen,
		Faces: []game.Face{
			{Name: "Drogskol Infantry", TypeLine: "Creature — Spirit Soldier", ManaCost: "{1}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2},
			{Name: "Drogskol Armaments", TypeLine: "Enchantment — Aura", Colors: []string{"W"}},
		},
	}
	infantry.SetFace(0)
	me.Graveyard.PushTop(infantry)

	v := ViewOfGameFor(g, me.ID.String())
	var cv *CardView
	for i := range v.Seats[0].Graveyard.Cards {
		if v.Seats[0].Graveyard.Cards[i].InstanceID == infantry.InstanceID.String() {
			cv = &v.Seats[0].Graveyard.Cards[i]
		}
	}
	if cv == nil {
		t.Fatal("graveyard view is missing the disturb card")
	}
	if !cv.CastableHere {
		t.Error("a disturb card in its owner's graveyard is not marked castable_here")
	}
	if !cv.AlternativeCostRequired {
		t.Error("the printed cost is offered from the graveyard; disturb is the only price there")
	}
	if len(cv.AlternativeCosts) != 1 {
		t.Fatalf("graveyard offers = %+v, want disturb alone", cv.AlternativeCosts)
	}
	offer := cv.AlternativeCosts[0]
	if offer.Key != "disturb" || offer.Label != "Disturb {3}{W}" || offer.ManaCost != "{3}{W}" {
		t.Errorf("offer = %+v, want Disturb {3}{W}", offer)
	}
	if offer.TargetMode == "" || offer.LegalTargets == nil {
		t.Fatalf("the disturb offer carries no target clause, so the client would cast the Aura with no creature: %+v", offer)
	}
	found := false
	for _, c := range offer.LegalTargets.Cards {
		if c == bear.InstanceID.String() {
			found = true
		}
	}
	if !found {
		t.Errorf("the bear is not among the disturbed Aura's legal targets: %+v", offer.LegalTargets)
	}
}

// The cast gate judges the face the cast puts on the stack. Under
// Aether Storm ("Creature spells can't be cast."), a disturbed Aura is
// castable out of the graveyard and a disturbed creature is not — the
// front faces are both creatures, and asking the gate of them would
// refuse both.
func TestDisturbCastGateReadsTheBackFace(t *testing.T) {
	g := buildActiveGame(t)
	castWindowOpen(t, g)
	me := g.Seats[0]
	seen := map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}

	storm := game.NewCard("Aether Storm", me.ID)
	storm.TypeLine = "Enchantment"
	storm.OracleID = "ff4297d3-3d96-4bd6-a606-1bdc20a6df2b"
	storm.KnownBy = seen
	g.Battlefield.PushTop(storm)
	bear := game.NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Power, bear.Toughness = 2, 2
	bear.KnownBy = seen
	g.Battlefield.PushTop(bear)

	transform := func(oracle string, front, back game.Face) game.Card {
		c := game.Card{InstanceID: uuid.New(), OracleID: oracle, Owner: me.ID, Controller: me.ID,
			Layout: game.LayoutTransform, KnownBy: seen, Faces: []game.Face{front, back}}
		c.SetFace(0)
		me.Graveyard.PushTop(c)
		return c
	}
	aura := transform(drogskolInfantryOracle,
		game.Face{Name: "Drogskol Infantry", TypeLine: "Creature — Spirit Soldier", ManaCost: "{1}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2},
		game.Face{Name: "Drogskol Armaments", TypeLine: "Enchantment — Aura", Colors: []string{"W"}})
	creature := transform("c6bb4b41-8dae-429a-b928-ae9d39c74711",
		game.Face{Name: "Baithook Angler", TypeLine: "Creature — Human Peasant", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 2, Toughness: 1},
		game.Face{Name: "Hook-Haunt Drifter", TypeLine: "Creature — Spirit", Colors: []string{"U"}, Power: 1, Toughness: 2})

	v := ViewOfGameFor(g, me.ID.String())
	stamps := map[string]CardView{}
	for _, c := range v.Seats[0].Graveyard.Cards {
		stamps[c.InstanceID] = c
	}
	if got := stamps[aura.InstanceID.String()]; !got.CastableHere || got.CantCast != "" {
		t.Errorf("a disturbed Aura under Aether Storm: castable_here %v, cant_cast %q; want castable", got.CastableHere, got.CantCast)
	}
	if got := stamps[creature.InstanceID.String()]; got.CastableHere || got.CantCast == "" {
		t.Errorf("a disturbed creature under Aether Storm: castable_here %v, cant_cast %q; want refused", got.CastableHere, got.CantCast)
	}

	// And the engine agrees.
	if err := g.CastSpell(me.ID, creature.InstanceID, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb"}); err == nil {
		t.Error("Hook-Haunt Drifter was disturbed under Aether Storm")
	}
	if err := g.CastSpell(me.ID, aura.InstanceID, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear.InstanceID}}}); err != nil {
		t.Errorf("Drogskol Armaments was refused under Aether Storm: %v", err)
	}
}
