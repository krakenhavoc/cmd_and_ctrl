package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// protection_view_test.go — #662, the projection half.
//
// Two things reach the client, and they are the two things the badge
// row and the picker need: the PARSED quality (so a tooltip reads
// "Protection from Demons" rather than a raw token, and so the bot —
// which may not import internal/game — has the engine's answer), and
// a legal-target stamp computed against the SOURCE (so the picker
// offers exactly what the announce gate will accept).
//
// The stamp is the one that used to be impossible: before #662 the
// view passed a bare caster ID, so it could not tell a red spell from
// a white one.

func TestProtectionIsProjectedParsed(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]

	angel := game.NewCard("Baneslayer Angel", me.ID)
	angel.TypeLine = "Creature — Angel"
	angel.Power, angel.Toughness = 5, 5
	angel.Keywords = []string{"flying", "protection from Demons", "protection from Dragons"}
	angel.KnownBy = map[uuid.UUID]bool{me.ID: true}
	g.Battlefield.PushTop(angel)

	view := ViewOfGameFor(g, me.ID.String())
	var got *CardView
	for i := range view.Battlefield.Cards {
		if view.Battlefield.Cards[i].InstanceID == angel.InstanceID.String() {
			got = &view.Battlefield.Cards[i]
		}
	}
	if got == nil {
		t.Fatal("the angel is not in the view")
	}
	if len(got.Protection) != 2 {
		t.Fatalf("protection = %+v, want two qualities (CR 702.16m)", got.Protection)
	}
	if got.Protection[0].Printed != "Demons" || got.Protection[0].Kind != "subtype" || got.Protection[0].Value != "Demon" {
		t.Errorf("first quality = %+v", got.Protection[0])
	}
	// The raw tokens stay in abilities: the badge row filters them
	// out itself, and nothing else on the wire loses information.
	var raw int
	for _, a := range got.Abilities {
		if _, ok := game.ParseProtectionQuality(a); ok {
			raw++
		}
	}
	if raw != 2 {
		t.Errorf("abilities carries %d protection tokens, want 2", raw)
	}
}

// A quality the closed grammar refuses is projected as nothing, so
// the badge never promises a rule the engine does not run (ADR 0037).
func TestAnUnparseableProtectionIsNotProjected(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]

	sphinx := game.NewCard("Sphinx of the Guildpact", me.ID)
	sphinx.TypeLine = "Artifact Creature — Sphinx"
	sphinx.Keywords = []string{"flying", "protection from monocolored"}
	sphinx.KnownBy = map[uuid.UUID]bool{me.ID: true}
	g.Battlefield.PushTop(sphinx)

	view := ViewOfGameFor(g, me.ID.String())
	for _, c := range view.Battlefield.Cards {
		if c.InstanceID == sphinx.InstanceID.String() && len(c.Protection) != 0 {
			t.Errorf("protection = %+v, want none", c.Protection)
		}
	}
}

// TestLegalTargetStampReadsTheSpellsColour is the picker's half of
// CR 702.16b: the same hand card, two colours, two different legal
// sets. Before #662 the stamp saw only the caster and could not
// produce two different answers here at all.
func TestLegalTargetStampReadsTheSpellsColour(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-view-protection-spec"
	prev := game.CatalogTargetSpec
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		if id != oracle {
			return nil
		}
		return &game.TargetSpec{
			Mode:  "creature",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.IsCreature()
			},
			Min: 1, Max: 1,
		}
	}
	t.Cleanup(func() { game.CatalogTargetSpec = prev })

	firewalker := game.NewCard("Kor Firewalker", opp.ID)
	firewalker.TypeLine = "Creature — Human Soldier"
	firewalker.Keywords = []string{"protection from red"}
	firewalker.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	g.Battlefield.PushTop(firewalker)
	bear := game.NewCard("Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	bear.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	g.Battlefield.PushTop(bear)

	stamp := func(colors []string) []string {
		spell := game.NewCard("Test Spell", me.ID)
		spell.TypeLine = "Instant"
		spell.OracleID = oracle
		spell.Colors = colors
		spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
		me.Hand.PushTop(spell)
		defer func() { _, _ = me.Hand.Remove(spell.InstanceID) }()

		view := ViewOfGameFor(g, me.ID.String())
		for _, c := range view.Seats[0].Hand.Cards {
			if c.InstanceID == spell.InstanceID.String() {
				if c.LegalTargets == nil {
					t.Fatal("no legal_targets stamped")
				}
				return c.LegalTargets.Cards
			}
		}
		t.Fatal("the spell is not in the owner's hand view")
		return nil
	}

	has := func(ids []string, id uuid.UUID) bool {
		for _, s := range ids {
			if s == id.String() {
				return true
			}
		}
		return false
	}

	red := stamp([]string{"R"})
	if has(red, firewalker.InstanceID) {
		t.Error("the picker offered a pro-red creature for a RED spell")
	}
	if !has(red, bear.InstanceID) {
		t.Error("the picker dropped an ordinary creature")
	}
	white := stamp([]string{"W"})
	if !has(white, firewalker.InstanceID) {
		t.Error("the picker must offer a pro-red creature for a WHITE spell")
	}
}
