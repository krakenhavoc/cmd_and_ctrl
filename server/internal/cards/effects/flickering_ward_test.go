package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flickering_ward_test.go — #4843: Flickering Ward and CR 702.16n's
// "This effect doesn't remove this Aura". A white Ward that chose white
// stays on its creature; a second instance of protection from white
// from anything else knocks it off (CR 704.5m).

const flickeringWardOracle = "3e2ae59a-40a1-46ca-8f24-3231ebd4e4eb"

// fwCast casts Flickering Ward from the active seat's hand onto `host`
// in the main phase and answers its colour prompt with `color`.
func fwCast(t *testing.T, g *game.Game, host uuid.UUID, color string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Flickering Ward", TypeLine: "Enchantment — Aura", OracleID: flickeringWardOracle,
		ManaCost: "{W}", Colors: []string{"W"}, Owner: me.ID, Controller: me.ID,
	})
	fillPoolColored(me, "W", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}}}); err != nil {
		t.Fatalf("CastSpell Flickering Ward: %v", err)
	}
	adResolveSpell(t, g, id)
	answerColor(t, g, me.ID, color)
	g.RunStateChecksForTest()
	return id
}

// fwSetup is a main phase with a bear for the active seat.
func fwSetup(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedBear(g, me.ID)
	nmAdvanceTo(t, g, game.StepPrecombatMain)
	return g, me, opp, bear
}

func fwAttachedTo(t *testing.T, g *game.Game, ward, host uuid.UUID) {
	t.Helper()
	if z := adZone(g, ward); z != game.ZoneBattlefield {
		t.Fatalf("Flickering Ward is in %v, want the battlefield", z)
	}
	c, ok := battlefieldCard(g, ward)
	if !ok || !c.IsAttachedTo(host) {
		t.Fatalf("Flickering Ward is attached to %+v, want %s", c.AttachedTo, host)
	}
}

func TestFlickeringWardIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(flickeringWardOracle)
	if !ok {
		t.Fatal("Flickering Ward is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness = %v, caveats = %q", spec.Completeness, spec.Caveats)
	}
}

// CR 702.16n: the white Ward's own protection from white does not put
// it into the graveyard, and the protection still does everything else
// protection does (CR 702.16b, e, f).
func TestFlickeringWardChoosingWhiteStaysAttached(t *testing.T) {
	g, me, opp, bear := fwSetup(t)
	ward := fwCast(t, g, bear, "W")
	fwAttachedTo(t, g, ward, bear)
	if got := protectionsOn(t, g, bear); !hasString(got, "white") {
		t.Fatalf("the bear's protections = %q, want white", got)
	}

	white := game.NewCard("Swords to Plowshares", opp.ID)
	white.TypeLine, white.Colors = "Instant", []string{"W"}
	red := game.NewCard("Chaos Warp", opp.ID)
	red.TypeLine, red.Colors = "Instant", []string{"R"}
	g.WithWriteLock(func() {
		g.Stack.PushTop(white)
		g.Stack.PushTop(red)
	})
	blocker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "White Knight", TypeLine: testCreatureTypeLine, Colors: []string{"W"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	g.ReadSnapshot(func() {
		b := battlefieldCardFor(g, bear)
		if game.CanBeTargetedBy(b, game.ZoneBattlefield, game.SourceObject(opp.ID, findStackCard(g, white.InstanceID))) {
			t.Error("a white spell may target the warded bear")
		}
		if !game.CanBeTargetedBy(b, game.ZoneBattlefield, game.SourceObject(opp.ID, findStackCard(g, red.InstanceID))) {
			t.Error("a red spell may not target the bear, which has protection from white only")
		}
		if r := g.BlockPairRefusalLocked(b, battlefieldCardFor(g, blocker)); r.Reason != game.BlockReasonProtection {
			t.Errorf("a white creature blocking the bear: refusal %q, want protection", r.Reason)
		}
	})
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(white.InstanceID, bear, 2); err != nil {
			t.Fatalf("white damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if c, _ := battlefieldCard(g, bear); c.DamageMarked != 0 {
		t.Errorf("the bear has %d damage marked from a white source, want 0", c.DamageMarked)
	}
	fwAttachedTo(t, g, ward, bear)
	_ = me
}

// CR 702.16n's last sentence: another instance of protection from white
// removes the white Ward as normal.
func TestFlickeringWardFallsOffToAnotherProtectionFromWhite(t *testing.T) {
	g, me, _, bear := fwSetup(t)
	ward := fwCast(t, g, bear, "W")
	fwAttachedTo(t, g, ward, bear)
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: bear})
		if err := (GrantKeywordUntilEOT{
			Target: bear, Keywords: []string{game.ProtectionFromColor("W")}, Label: "test",
		}).Apply(ctx); err != nil {
			t.Fatalf("grant: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if !me.Graveyard.Contains(ward) {
		t.Errorf("Flickering Ward is in %v, want the graveyard (CR 704.5m)", adZone(g, ward))
	}
}

// Two white Wards that both chose white are two instances: each removes
// the other, and both go (CR 702.16n, CR 704.5m).
func TestTwoWhiteFlickeringWardsRemoveEachOther(t *testing.T) {
	g, me, _, bear := fwSetup(t)
	first := fwCast(t, g, bear, "W")
	second := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flickering Ward", TypeLine: "Enchantment — Aura", OracleID: flickeringWardOracle,
		ManaCost: "{W}", Colors: []string{"W"}, Owner: me.ID, Controller: me.ID,
		ChosenColor: "W", AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear},
	})
	g.RunStateChecksForTest()
	for _, id := range []uuid.UUID{first, second} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("a Flickering Ward is in %v, want the graveyard", adZone(g, id))
		}
	}
}

// Choosing a colour other than the Ward's own: protection from it, and
// the Ward was never at risk.
func TestFlickeringWardChoosingAnotherColor(t *testing.T) {
	g, _, _, bear := fwSetup(t)
	ward := fwCast(t, g, bear, "B")
	fwAttachedTo(t, g, ward, bear)
	got := protectionsOn(t, g, bear)
	if !hasString(got, "black") || hasString(got, "white") {
		t.Errorf("the bear's protections = %q, want black only", got)
	}
}

// {W}: Return this Aura to its owner's hand.
func TestFlickeringWardReturnsToHand(t *testing.T) {
	g, me, _, bear := fwSetup(t)
	ward := fwCast(t, g, bear, "W")
	fillPoolColored(me, "W", 1)
	if err := activateOn(t, g, me.ID, ward, 0, uuid.Nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ward) {
		t.Fatalf("Flickering Ward is in %v, want its owner's hand", adZone(g, ward))
	}
	if got := protectionsOn(t, g, bear); len(got) != 0 {
		t.Errorf("the bear kept %q after the Ward left", got)
	}
}

// The exemption is minted from the Ward's id and chosen colour, both on
// the snapshot: a restore keeps the Ward attached and the bear protected.
func TestFlickeringWardSurvivesARestore(t *testing.T) {
	g, _, _, bear := fwSetup(t)
	ward := fwCast(t, g, bear, "W")
	restored := restoreRoundTrip(t, g, true)
	restored.RunStateChecksForTest()
	fwAttachedTo(t, restored, ward, bear)
	if got := protectionsOn(t, restored, bear); !hasString(got, "white") {
		t.Errorf("restored: the bear's protections = %q, want white", got)
	}
}
