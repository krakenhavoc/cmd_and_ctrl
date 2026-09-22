package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// player_protection_cards_test.go — #1197's four cards, each proved
// through the engine rather than by reading its own Spec back.
//
// The assertions are deliberately about what a PLAYER at the table
// would notice: an opponent's Lightning Bolt is refused at announce,
// the life total does not move, the shield is gone on your next turn.
// A test that only checked `spec.PlayerKeywords` would pass with
// every choke point backed out.

const (
	leylineOfSanctityOracle = "492e0e6c-8c27-4376-938b-f8a8b6205810"
	aegisOfTheGodsOracle    = "c5bfc1b9-a55d-4608-a6f7-bb62cb8dc3c6"
	teferisProtectionOracle = "0d4ecdb1-ec90-497f-a7a4-1c68092b8757"
)

// playerAbilities reads a seat's abilities the way every consumer
// does, under the read lock.
func playerAbilities(g *game.Game, p *game.Player) []string {
	var out []string
	g.ReadSnapshot(func() { out = g.PlayerAbilitiesForEffect(p) })
	return out
}

func hasPlayerAbility(list []string, want string) bool {
	for _, a := range list {
		if a == want {
			return true
		}
	}
	return false
}

// boltAtPlayerErr is b08CastBolt's announce half: the cast that is
// SUPPOSED to be refused hands its error back.
func boltAtPlayerErr(t *testing.T, g *game.Game, caster *game.Player, target uuid.UUID) error {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: lightningBoltOracle, Colors: []string{"R"}, ManaCost: "{R}",
		Owner: caster.ID, Controller: caster.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}},
	})
}

// --- Leyline of Sanctity -------------------------------------------

// TestLeylineOfSanctityGivesItsControllerHexproof is the card, end to
// end: the active seat plays it, and the seat to its left can no
// longer aim a Bolt at them. Cast by the ACTIVE seat because the
// announce gate runs for whoever has priority — the Leyline's
// controller is seat 0 and the Bolt comes from seat 0's own hand at
// the opposite seat, which is the direction that has to keep working.
func TestLeylineOfSanctityGivesItsControllerHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, opp.ID, "Leyline of Sanctity", leylineOfSanctityOracle, "Enchantment")

	if got := playerAbilities(g, opp); !hasPlayerAbility(got, KeywordHexproof) {
		t.Fatalf("the Leyline's controller has %v, want hexproof", got)
	}
	if got := playerAbilities(g, me); len(got) != 0 {
		t.Errorf("a seat with no Leyline has %v, want nothing", got)
	}

	if err := boltAtPlayerErr(t, g, me, opp.ID); err != game.ErrIllegalTarget {
		t.Errorf("Bolt at a hexproof opponent: got %v, want ErrIllegalTarget", err)
	}
	// And the Leyline's own controller can still target themselves,
	// which is the half that makes the card playable.
	if err := boltAtPlayerErr(t, g, opp, opp.ID); err != nil {
		t.Errorf("the Leyline's controller targeting themselves: %v", err)
	}
}

// TestLeylineLeavingTakesTheHexproofWithIt — the grant is derived, so
// there is nothing to unwind and nothing to strand.
func TestLeylineLeavingTakesTheHexproofWithIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	leyline := pushPermanentForTest(g, me.ID, "Leyline of Sanctity", leylineOfSanctityOracle, "Enchantment")
	if !hasPlayerAbility(playerAbilities(g, me), KeywordHexproof) {
		t.Fatal("no hexproof with the Leyline out")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(leyline); err != nil {
			t.Fatalf("destroy the Leyline: %v", err)
		}
	})
	if hasPlayerAbility(playerAbilities(g, me), KeywordHexproof) {
		t.Error("the hexproof outlived the Leyline")
	}
}

// --- Aegis of the Gods ---------------------------------------------

// TestAegisOfTheGodsAndLeylineCompose is the composition case, and
// the reason the grant is derived rather than written: two sources of
// one ability, one of them dies, the player is still hexproof. A
// "set on enter, restore on leave" design gets exactly this wrong.
func TestAegisOfTheGodsAndLeylineCompose(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aegis := pushPermanentForTest(g, me.ID, "Aegis of the Gods", aegisOfTheGodsOracle,
		"Enchantment Creature — Human Soldier")
	pushPermanentForTest(g, me.ID, "Leyline of Sanctity", leylineOfSanctityOracle, "Enchantment")

	if !hasPlayerAbility(playerAbilities(g, me), KeywordHexproof) {
		t.Fatal("two sources and no hexproof")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aegis); err != nil {
			t.Fatalf("destroy the Aegis: %v", err)
		}
	})
	if !hasPlayerAbility(playerAbilities(g, me), KeywordHexproof) {
		t.Error("killing the Aegis revoked the Leyline's grant")
	}
}

// --- Teferi's Protection -------------------------------------------

// TestTeferisProtectionShieldsYouAndExilesItself is the whole of what
// the card ships: protection from everything until your next turn,
// and the card in exile rather than the graveyard.
func TestTeferisProtectionShieldsYouAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	spell := castCatalogSpell(t, g, "Teferi's Protection", "Instant", teferisProtectionOracle, nil)
	passPriorityAroundTable(t, g)

	if got := playerAbilities(g, me); !hasPlayerAbility(got, ProtectionFromEverything) {
		t.Fatalf("the caster has %v, want protection from everything", got)
	}
	// "Exile Teferi's Protection", not the graveyard: #489's
	// self-move check is what keeps the resolution frame from
	// putting it back.
	var inExile, inGraveyard bool
	g.ReadSnapshot(func() {
		inExile = g.Exile.Contains(spell)
		inGraveyard = me.Graveyard.Contains(spell)
	})
	if !inExile || inGraveyard {
		t.Errorf("Teferi's Protection exiles itself: exile=%v graveyard=%v", inExile, inGraveyard)
	}

	// Targeting: refused for everybody, including the caster —
	// protection asks WHAT is casting, not who.
	if err := boltAtPlayerErr(t, g, opp, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("an opponent's Bolt at a protected player: got %v, want ErrIllegalTarget", err)
	}
	if err := boltAtPlayerErr(t, g, me, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("your OWN Bolt at yourself under protection from everything: got %v, want ErrIllegalTarget", err)
	}
}

// TestTeferisProtectionPreventsDamageUntilYourNextTurn is CR 702.16e
// on a player, through the card, plus the boundary: the shield sits
// through every opponent's turn and is gone when yours begins.
func TestTeferisProtectionPreventsDamageUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Teferi's Protection", "Instant", teferisProtectionOracle, nil)
	passPriorityAroundTable(t, g)

	life := me.Life
	source := pushPermanentForTest(g, g.Seats[1].ID, "Pinger", "test-player-prot-pinger", "Artifact")
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, me.ID, 5); err != nil {
			t.Fatalf("damage at the protected player: %v", err)
		}
	})
	if me.Life != life {
		t.Errorf("protected player lost %d life; CR 702.16e prevents damage from every source", life-me.Life)
	}

	// Three opponents' turns, then back to seat 0 — a four-seat
	// table, so this is the full rotation the card is famous for.
	for i := 0; i < 3; i++ {
		advanceOneTurnForTest(t, g)
		if !hasPlayerAbility(playerAbilities(g, me), ProtectionFromEverything) {
			t.Fatalf("the shield ended after %d opponent turns; it lasts until YOUR next turn", i+1)
		}
	}
	advanceOneTurnForTest(t, g)
	if hasPlayerAbility(playerAbilities(g, me), ProtectionFromEverything) {
		t.Error("the shield survived the beginning of its own player's next turn")
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, me.ID, 5); err != nil {
			t.Fatalf("damage after the shield expired: %v", err)
		}
	})
	if me.Life != life-5 {
		t.Errorf("life %d after the shield expired, want %d", me.Life, life-5)
	}
}

// --- The One Ring ---------------------------------------------------

// TestTheOneRingShieldsYouWhenYouCastIt — the clause the card carried
// a caveat for from S40 until #1197.
func TestTheOneRingShieldsYouWhenYouCastIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	castCatalogSpell(t, g, "The One Ring", "Legendary Artifact", theOneRingOracle, nil)
	passPriorityAroundTable(t, g) // the Ring resolves; the ETB trigger lands
	passPriorityAroundTable(t, g) // the trigger resolves

	if got := playerAbilities(g, me); !hasPlayerAbility(got, ProtectionFromEverything) {
		t.Fatalf("casting The One Ring gave %v, want protection from everything", got)
	}
	if err := boltAtPlayerErr(t, g, opp, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("an opponent's Bolt at the Ring-bearer: got %v, want ErrIllegalTarget", err)
	}
}

// TestTheOneRingPutOntoTheBattlefieldGivesNoShield is the
// intervening-if (CR 603.4): "if you cast it". A Ring reanimated or
// cheated into play is the draw engine and nothing else — as printed,
// and the assertion that fails if b16EnteredFromStack is dropped for
// a bare ETB.
func TestTheOneRingPutOntoTheBattlefieldGivesNoShield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "The One Ring", theOneRingOracle, "Legendary Artifact")
	passPriorityAroundTable(t, g)

	if got := playerAbilities(g, me); hasPlayerAbility(got, ProtectionFromEverything) {
		t.Errorf("a Ring that was not CAST still shielded its controller: %v", got)
	}
}

// advanceOneTurnForTest walks the cursor until the active seat
// changes. The effects package has no turn helper of its own; this is
// duration_test.go's, locally.
func advanceOneTurnForTest(t *testing.T, g *game.Game) {
	t.Helper()
	start := g.Turn.ActiveSeat
	for i := 0; i < 64; i++ {
		if g.Turn.ActiveSeat != start {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("cursor never left seat %d's turn", start)
}
