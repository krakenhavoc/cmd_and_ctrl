package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_timing_cards_test.go — #1195's proof cards. Four cards across
// the two homes the amendment gives a per-player timing statement:
//
//	Vedalken Orrery         derived, no filter, the grant
//	Leyline of Anticipation derived, no filter, and a second source
//	Emergence Zone          stored, "this turn", source sacrificed
//	Teferi, Time Raveler    derived restriction AND stored grant
//
// Every assertion drives CastSpell rather than the predicate, because
// the announce path is what a player and a bot both hit.

const (
	vedalkenOrreryOracle   = "60d640a4-032b-4279-ac75-8f200ca776fb"
	leylineAnticipation    = "9dc65ffe-17fc-4280-b4bd-78073ac7e12b"
	emergenceZoneOracle    = "7536eb66-959d-4dca-9b75-895572ef733c"
	teferiTimeRavelerCTest = "ae7604bb-4818-45a3-960c-cf3d83f15964"
)

// ctHandCard seeds one card into a seat's hand.
func ctHandCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// ctStoredTimings counts the cast-timing statements stored on a
// player. Since #1195 folded onto #1197's registry a granted timing
// statement is a game.PlayerStatic carrying a rule and no Keyword, so
// it shares one slice, one sweep and one duration with "you gain
// protection from everything until your next turn".
func ctStoredTimings(p *game.Player) int {
	n := 0
	for _, st := range p.Statics {
		if st.Timing.Timing != game.TimingNormal {
			n++
		}
	}
	return n
}

// Vedalken Orrery: a CREATURE spell in the active seat's own end
// step, which is not a sorcery window by any reading — refused
// without the artifact, accepted with it.
func TestVedalkenOrreryOpensTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepEnd)

	first := ctHandCard(me, "Proof Bear", "Creature — Bear")
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("a creature spell in an end step with no Orrery = %v, want ErrSorcerySpeedRequired", err)
	}

	pushPermanentForTest(g, me.ID, "Vedalken Orrery", vedalkenOrreryOracle, "Artifact")
	second := ctHandCard(me, "Proof Bear Two", "Creature — Bear")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("a creature spell in an end step under a Vedalken Orrery: %v", err)
	}
	if me.Hand.Contains(second) {
		t.Error("the creature spell never left the hand")
	}
}

// Leyline of Anticipation is the same sentence on an enchantment, and
// the test is that a card the ACTIVE seat does not control opens that
// seat's own turn for its controller: a sorcery, on somebody else's
// turn, which the sorcery-speed gate refuses twice over.
func TestLeylineOfAnticipationOpensAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)

	first := ctHandCard(me, "Proof Ritual", "Sorcery")
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("a sorcery on somebody else's turn = %v, want ErrSorcerySpeedRequired", err)
	}

	pushPermanentForTest(g, me.ID, "Leyline of Anticipation", leylineAnticipation, "Enchantment")
	second := ctHandCard(me, "Proof Ritual Two", "Sorcery")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("a sorcery on an opponent's turn under a Leyline: %v", err)
	}
}

// Emergence Zone is the STORED half: the land sacrifices itself to
// pay for the ability, so nothing is left on the battlefield for a
// derivation to read and the permission has to be written onto the
// player with a duration.
func TestEmergenceZoneGrantsFlashForTheTurnAndDiesDoingIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	zone := pushPermanentForTest(g, me.ID, "Emergence Zone", emergenceZoneOracle, "Land")
	b06AddMana(me, "C")

	// Ability index 0 is the activated one — the mana ability lives on
	// its own list.
	if err := g.ActivateCatalogAbility(me.ID, zone, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility(Emergence Zone): %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(zone) {
		t.Error("the land was not sacrificed to its own cost")
	}
	if n := ctStoredTimings(me); n != 1 {
		t.Fatalf("the grant did not land on the player: %d statements", n)
	}
	if k := me.Statics[0].Duration.Kind; k != game.UntilEndOfTurn {
		t.Errorf("duration kind = %v, want UntilEndOfTurn — the clause is \"this turn\"", k)
	}

	// The window is open in this turn's end step, with the land gone.
	advanceTo(t, g, game.StepEnd)
	bear := ctHandCard(me, "Proof Bear", "Creature — Bear")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("a creature spell in the end step under Emergence Zone: %v", err)
	}
	passPriorityAroundTable(t, g)

	// And it ends with the turn (CR 514.2).
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if n := ctStoredTimings(me); n != 0 {
		t.Errorf("the grant outlived its turn: %d statements left", n)
	}
	later := ctHandCard(me, "Proof Bear Two", "Creature — Bear")
	if err := g.CastSpell(me.ID, later, game.CastSpellParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("a creature spell after the grant expired = %v, want ErrSorcerySpeedRequired", err)
	}
}

// Teferi, Time Raveler's STATIC, and the caveat it used to carry:
// "opponents can still cast spells at instant speed" is no longer
// true.
func TestTeferiTimeRavelerStopsOpponentsInstants(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)

	bolt := ctHandCard(them, "Proof Bolt", "Instant")
	if err := g.CastSpell(them.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("setup: an opponent's instant on my main phase: %v", err)
	}
	passPriorityAroundTable(t, g)

	pushPermanentForTest(g, me.ID, "Teferi, Time Raveler", teferiTimeRavelerCTest, "Legendary Planeswalker — Teferi")
	second := ctHandCard(them, "Proof Bolt Two", "Instant")
	if err := g.CastSpell(them.ID, second, game.CastSpellParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("an opponent's instant under Teferi = %v, want ErrSorcerySpeedRequired", err)
	}
	// "Each OPPONENT": Teferi's own controller keeps instant speed.
	mine := ctHandCard(me, "My Proof Bolt", "Instant")
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{}); err != nil {
		t.Errorf("Teferi restricted his own controller: %v", err)
	}
}

// Teferi's +1, and the other caveat: "the +1 only adds loyalty" is no
// longer true either. A SORCERY at instant speed, until your next
// turn — and a creature spell beside it to prove the filter narrows.
func TestTeferiTimeRavelerPlusOneOpensSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	teferi := pushPermanentForTest(g, me.ID, "Teferi, Time Raveler", teferiTimeRavelerCTest, "Legendary Planeswalker — Teferi")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(teferi, "loyalty", 4) })

	if err := g.ActivateCatalogAbility(me.ID, teferi, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility(Teferi +1): %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := ctStoredTimings(me); n != 1 {
		t.Fatalf("the +1 granted nothing: %d statements", n)
	}
	if k := me.Statics[0].Duration.Kind; k != game.UntilYourNextTurn {
		t.Errorf("duration kind = %v, want UntilYourNextTurn", k)
	}

	advanceTo(t, g, game.StepEnd)
	ritual := ctHandCard(me, "Proof Ritual", "Sorcery")
	if err := g.CastSpell(me.ID, ritual, game.CastSpellParams{}); err != nil {
		t.Fatalf("a sorcery in the end step under Teferi's +1: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The filter is SORCERY spells: a creature spell is not opened.
	bear := ctHandCard(me, "Proof Bear", "Creature — Bear")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("a creature spell under a sorcery-only grant = %v, want ErrSorcerySpeedRequired", err)
	}
}
