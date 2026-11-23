package testsp

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Leyline of Sanctity -------------------------------------------

// TestLeylineOfSanctityGivesItsControllerHexproof is the card, end to
// end: the active seat plays it, and the seat to its left can no
// longer aim a Bolt at them. Cast by the ACTIVE seat because the
// announce gate runs for whoever has priority — the Leyline's
// controller is seat 0 and the Bolt comes from seat 0's own hand at
// the opposite seat, which is the direction that has to keep working.
func TestLeylineOfSanctityGivesItsControllerHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, opp.ID, "Leyline of Sanctity", effectstest.LeylineOfSanctityOracle, "Enchantment")

	if got := effectstest.PlayerAbilities(g, opp); !effectstest.HasPlayerAbility(got, effects.KeywordHexproof) {
		t.Fatalf("the Leyline's controller has %v, want hexproof", got)
	}
	if got := effectstest.PlayerAbilities(g, me); len(got) != 0 {
		t.Errorf("a seat with no Leyline has %v, want nothing", got)
	}

	if err := effectstest.BoltAtPlayerErr(t, g, me, opp.ID); err != game.ErrIllegalTarget {
		t.Errorf("Bolt at a hexproof opponent: got %v, want ErrIllegalTarget", err)
	}
	// And the Leyline's own controller can still target themselves,
	// which is the half that makes the card playable.
	if err := effectstest.BoltAtPlayerErr(t, g, opp, opp.ID); err != nil {
		t.Errorf("the Leyline's controller targeting themselves: %v", err)
	}
}

// TestLeylineLeavingTakesTheHexproofWithIt — the grant is derived, so
// there is nothing to unwind and nothing to strand.
func TestLeylineLeavingTakesTheHexproofWithIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	leyline := effectstest.PushPermanentForTest(g, me.ID, "Leyline of Sanctity", effectstest.LeylineOfSanctityOracle, "Enchantment")
	if !effectstest.HasPlayerAbility(effectstest.PlayerAbilities(g, me), effects.KeywordHexproof) {
		t.Fatal("no hexproof with the Leyline out")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(leyline); err != nil {
			t.Fatalf("destroy the Leyline: %v", err)
		}
	})
	if effectstest.HasPlayerAbility(effectstest.PlayerAbilities(g, me), effects.KeywordHexproof) {
		t.Error("the hexproof outlived the Leyline")
	}
}

// --- Aegis of the Gods ---------------------------------------------

// TestAegisOfTheGodsAndLeylineCompose is the composition case, and
// the reason the grant is derived rather than written: two sources of
// one ability, one of them dies, the player is still hexproof. A
// "set on enter, restore on leave" design gets exactly this wrong.
func TestAegisOfTheGodsAndLeylineCompose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aegis := effectstest.PushPermanentForTest(g, me.ID, "Aegis of the Gods", effectstest.AegisOfTheGodsOracle,
		"Enchantment Creature — Human Soldier")
	effectstest.PushPermanentForTest(g, me.ID, "Leyline of Sanctity", effectstest.LeylineOfSanctityOracle, "Enchantment")

	if !effectstest.HasPlayerAbility(effectstest.PlayerAbilities(g, me), effects.KeywordHexproof) {
		t.Fatal("two sources and no hexproof")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aegis); err != nil {
			t.Fatalf("destroy the Aegis: %v", err)
		}
	})
	if !effectstest.HasPlayerAbility(effectstest.PlayerAbilities(g, me), effects.KeywordHexproof) {
		t.Error("killing the Aegis revoked the Leyline's grant")
	}
}

// --- Absolute Virtue -----------------------------------------------

// TestAbsoluteVirtueProtectsYouFromYourOpponentsOnly is #2745's card:
// its controller has protection from each of their opponents, so an
// opponent's Bolt is refused and their own is not, and the grant ends
// when Absolute Virtue leaves.
func TestAbsoluteVirtueProtectsYouFromYourOpponentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	virtue := effectstest.PushPermanentForTest(g, opp.ID, "Absolute Virtue", effectstest.AbsoluteVirtueOracle,
		"Legendary Creature — Avatar Warrior")

	if got := effectstest.PlayerAbilities(g, opp); !effectstest.HasPlayerAbility(got, effects.ProtectionFromEachOfYourOpponents) {
		t.Fatalf("Absolute Virtue's controller has %v, want %q", got, effects.ProtectionFromEachOfYourOpponents)
	}
	if err := effectstest.BoltAtPlayerErr(t, g, me, opp.ID); err != game.ErrIllegalTarget {
		t.Errorf("an opponent's Bolt at Absolute Virtue's controller: got %v, want ErrIllegalTarget", err)
	}
	if err := effectstest.BoltAtPlayerErr(t, g, opp, opp.ID); err != nil {
		t.Errorf("Absolute Virtue's controller targeting themselves: %v", err)
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(virtue); err != nil {
			t.Fatalf("destroy Absolute Virtue: %v", err)
		}
	})
	if err := effectstest.BoltAtPlayerErr(t, g, me, opp.ID); err != nil {
		t.Errorf("the protection outlived Absolute Virtue: %v", err)
	}
}

// --- Teferi's Protection -------------------------------------------

// TestTeferisProtectionShieldsYouAndExilesItself is the whole of what
// the card ships: protection from everything until your next turn,
// and the card in exile rather than the graveyard.
func TestTeferisProtectionShieldsYouAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	spell := effectstest.CastCatalogSpell(t, g, "Teferi's Protection", "Instant", effectstest.TeferisProtectionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.PlayerAbilities(g, me); !effectstest.HasPlayerAbility(got, effects.ProtectionFromEverything) {
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
	if err := effectstest.BoltAtPlayerErr(t, g, opp, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("an opponent's Bolt at a protected player: got %v, want ErrIllegalTarget", err)
	}
	if err := effectstest.BoltAtPlayerErr(t, g, me, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("your OWN Bolt at yourself under protection from everything: got %v, want ErrIllegalTarget", err)
	}
}

// TestTeferisProtectionPreventsDamageUntilYourNextTurn is CR 702.16e
// on a player, through the card, plus the boundary: the shield sits
// through every opponent's turn and is gone when yours begins.
func TestTeferisProtectionPreventsDamageUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Teferi's Protection", "Instant", effectstest.TeferisProtectionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	life := me.Life
	source := effectstest.PushPermanentForTest(g, g.Seats[1].ID, "Pinger", "test-player-prot-pinger", "Artifact")
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
		effectstest.AdvanceOneTurnForTest(t, g)
		if !effectstest.HasPlayerAbility(effectstest.PlayerAbilities(g, me), effects.ProtectionFromEverything) {
			t.Fatalf("the shield ended after %d opponent turns; it lasts until YOUR next turn", i+1)
		}
	}
	effectstest.AdvanceOneTurnForTest(t, g)
	if effectstest.HasPlayerAbility(effectstest.PlayerAbilities(g, me), effects.ProtectionFromEverything) {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	effectstest.CastCatalogSpell(t, g, "The One Ring", "Legendary Artifact", effectstest.TheOneRingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g) // the Ring resolves; the ETB trigger lands
	effectstest.PassPriorityAroundTable(t, g) // the trigger resolves

	if got := effectstest.PlayerAbilities(g, me); !effectstest.HasPlayerAbility(got, effects.ProtectionFromEverything) {
		t.Fatalf("casting The One Ring gave %v, want protection from everything", got)
	}
	if err := effectstest.BoltAtPlayerErr(t, g, opp, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("an opponent's Bolt at the Ring-bearer: got %v, want ErrIllegalTarget", err)
	}
}

// TestTheOneRingPutOntoTheBattlefieldGivesNoShield is the
// intervening-if (CR 603.4): "if you cast it". A Ring reanimated or
// cheated into play is the draw engine and nothing else — as printed,
// and the assertion that fails if b16EnteredFromStack is dropped for
// a bare ETB.
func TestTheOneRingPutOntoTheBattlefieldGivesNoShield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "The One Ring", effectstest.TheOneRingOracle, "Legendary Artifact")
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.PlayerAbilities(g, me); effectstest.HasPlayerAbility(got, effects.ProtectionFromEverything) {
		t.Errorf("a Ring that was not CAST still shielded its controller: %v", got)
	}
}
