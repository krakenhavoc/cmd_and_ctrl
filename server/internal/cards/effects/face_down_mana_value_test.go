package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// face_down_mana_value_test.go — #4682. CR 708.2a: a face-down permanent
// has no mana cost, so its mana value is 0 (CR 202.3a), whatever card it
// hides. These are the cards that compare mana values, played against a
// face-down permanent whose hidden card has mana value 6: the sweep, the
// targeted bound, the offers and the numbers the view ships.
//
// The engine half (every read, every way to be face down) is
// game/face_down_mana_value_test.go.

// faceDownSixDrop puts a face-down permanent onto the battlefield under
// owner. The card it hides is a {4}{B}{B} creature.
func faceDownSixDrop(g *game.Game, owner *game.Player, kind game.FaceDownKind) uuid.UUID {
	c := game.Card{
		Name: "Griselbrand", TypeLine: "Legendary Creature — Demon",
		ManaCost: "{4}{B}{B}", Power: 7, Toughness: 7, Colors: []string{"B"},
		KnownBy: map[uuid.UUID]bool{owner.ID: true},
	}
	c.SetFaceDown(kind)
	return apaPush(g, owner.ID, owner.ID, c)
}

// TestManaValueCeilingTreatsAFaceDownPermanentAsZero: Ritual of Soot
// destroys each creature with mana value 3 or less. The face-down
// permanent is a 2/2 with no mana cost, so it goes, and the face-up
// mana value 4 creature beside it does not.
func TestManaValueCeilingTreatsAFaceDownPermanentAsZero(t *testing.T) {
	for _, kind := range []game.FaceDownKind{
		game.FaceDownManifested, game.FaceDownMorphed, game.FaceDownDisguised,
		game.FaceDownCloaked, game.FaceDownTurned,
	} {
		t.Run(string(kind), func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			hidden := faceDownSixDrop(g, opp, kind)
			four := apaPush(g, opp.ID, opp.ID, game.Card{
				Name: "Four", TypeLine: "Creature — Wurm", ManaCost: "{2}{G}{G}", Power: 4, Toughness: 4,
			})

			castCatalogSpell(t, g, "Ritual of Soot", "Sorcery", ritualOfSootOracle, nil)
			passPriorityAroundTable(t, g)

			if g.Battlefield.Contains(hidden) {
				t.Error("a face-down permanent (mana value 0) survived \"destroy each creature with mana value 3 or less\": " +
					"the sweep read the card it hides")
			}
			if !g.Battlefield.Contains(four) {
				t.Error("a face-up mana value 4 creature was destroyed by Ritual of Soot")
			}
		})
	}
}

// TestManaValueFloorDoesNotCountAFaceDownPermanent is the same read
// from the other side: Ritual of Soot is the ceiling, and a floor ("mana
// value 4 or greater") must not catch the face-down 2/2 either.
func TestManaValueFloorDoesNotCountAFaceDownPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	hidden := faceDownSixDrop(g, opp, game.FaceDownManifested)
	c, ok := g.LookupCardForEffect(hidden)
	if !ok {
		t.Fatal("the face-down permanent is in no zone")
	}
	if ManaValueGE(4)(g, uuid.Nil, c) {
		t.Error("ManaValueGE(4) passes a face-down permanent whose card has mana value 6")
	}
	if !ManaValueLE(0)(g, uuid.Nil, c) {
		t.Error("ManaValueLE(0) refuses a face-down permanent, which has mana value 0")
	}
}

// heliosWithEnergy is HELIOS One on `me`'s side, with the energy to pay any X up
// to 6 and the mana for its {3}.
func heliosWithEnergy(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	helios := edCard(t, g, me, "HELIOS One", edHeliosOne, "Land", 0, 0)
	setEnergy(t, g, me, 6)
	floatMana(t, g, me, "{C}{C}{C}")
	return helios
}

// TestManaValueTargetBoundSeesAFaceDownPermanentAsZero: HELIOS One pays
// X energy to destroy a nonland permanent with mana value X. Paying 6
// does not reach a face-down permanent whose card costs six, which would
// be a way to learn it; paying 0 does.
func TestManaValueTargetBoundSeesAFaceDownPermanentAsZero(t *testing.T) {
	g, me, opp := p7Table(t)
	helios := heliosWithEnergy(t, g, me)
	hidden := faceDownSixDrop(g, opp, game.FaceDownManifested)

	if err := g.ActivateCatalogAbility(me.ID, helios, 1, game.ActivateAbilityParams{
		XValue: 6, Targets: edCardTarget(hidden),
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("X=6 at a face-down permanent whose card costs six: err = %v, want ErrIllegalTarget "+
			"(its mana value is 0, and the refusal must not say otherwise)", err)
	}
	if energyOf(me) != 6 {
		t.Errorf("a refused activation changed energy 6 → %d", energyOf(me))
	}

	p7Activate(t, g, me, helios, 1, game.ActivateAbilityParams{XValue: 0, Targets: edCardTarget(hidden)})
	if g.Battlefield.Contains(hidden) {
		t.Error("X=0 did not destroy the face-down permanent (mana value 0)")
	}
}

// TestManaValueOffersDoNotNameAFaceDownPermanentsCost: neither the moves
// the enumerator lists nor the numbers the view ships tell what a
// face-down permanent hides. HELIOS One's offers pair an X with the
// permanents of that mana value; the face-down permanent may only be
// paired with 0, never with 6, and the view's per-card mana value for it
// is 0.
func TestManaValueOffersDoNotNameAFaceDownPermanentsCost(t *testing.T) {
	g, me, opp := p7Table(t)
	helios := heliosWithEnergy(t, g, me)
	hidden := faceDownSixDrop(g, opp, game.FaceDownManifested)
	three := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Three Drop", TypeLine: "Artifact", ManaCost: "{3}"})

	// The legal moves.
	offered := map[int]map[string]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Source != helios || m.Kind != legal.KindActivate {
			continue
		}
		var p struct {
			AbilityIndex int `json:"ability_index"`
			XValue       int `json:"x_value"`
			Targets      []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.AbilityIndex != 1 {
			continue
		}
		if offered[p.XValue] == nil {
			offered[p.XValue] = map[string]bool{}
		}
		for _, tg := range p.Targets {
			offered[p.XValue][tg.ID] = true
		}
	}
	if offered[6][hidden.String()] {
		t.Error("the enumerator offers X=6 at the face-down permanent: the move names the hidden card's mana value")
	}
	if !offered[3][three.String()] {
		t.Errorf("the enumerator does not offer X=3 at the face-up mana value 3 artifact (offered = %v)", offered)
	}
	for x, targets := range offered {
		if x != 0 && targets[hidden.String()] {
			t.Errorf("the enumerator offers X=%d at the face-down permanent, which has mana value 0", x)
		}
	}

	// The view the client narrows the bound with. A permanent pushed
	// straight onto the battlefield has no knower stamp and the viewer
	// filter redacts it, so HELIOS One alone is marked as public; the
	// face-down permanent keeps the knowers the rules give it.
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, helios)
		c.KnownBy = map[uuid.UUID]bool{}
		for _, p := range g.Seats {
			c.KnownBy[p.ID] = true
		}
	})
	v := protocol.ViewOfGameFor(g, me.ID.String())
	var row *protocol.ActivatedAbilityView
	for _, cv := range v.Battlefield.Cards {
		if cv.InstanceID != helios.String() {
			continue
		}
		for i := range cv.ActivatedAbilities {
			if cv.ActivatedAbilities[i].Index == 1 {
				row = &cv.ActivatedAbilities[i]
			}
		}
	}
	if row == nil || row.LegalTargets == nil {
		t.Fatalf("no legal-targets view on HELIOS One's destroy ability: %+v", row)
	}
	lt := row.LegalTargets
	if !lt.ManaValueEqualsX {
		t.Fatalf("the clause is not shipped as bound by X: %+v", lt)
	}
	if got, ok := lt.ManaValues[hidden.String()]; !ok || got != 0 {
		t.Errorf("the view ships the face-down permanent's mana value as %d (present = %v), want 0: "+
			"the wire must not tell what it hides", got, ok)
	}
	if lt.ManaValues[three.String()] != 3 {
		t.Errorf("the view ships the face-up artifact's mana value as %d, want 3", lt.ManaValues[three.String()])
	}
}

// TestFaceDownSpellIsNotAFiveDropToThryx is the cast side of the same
// rule (CR 708.4: the spell has only the face-down characteristics).
// Thryx, the Sudden Storm makes spells with mana value 5 or greater cost
// {1} less and uncounterable. A mana value 5 card cast face down for
// morph's {3} is a spell with no mana cost, so it gets neither: it costs
// the full {3}, and it can be countered. Before #4682 Thryx read the card
// the spell hides, so the discount and the shield told every player that
// it was a five-drop or better.
func TestFaceDownSpellIsNotAFiveDropToThryx(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Thryx, the Sudden Storm", "Legendary Creature — Elemental Giant", csThryxOracle, false)

	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Willbender", TypeLine: "Creature — Human Wizard",
			OracleID: willbenderOracle, ManaCost: "{4}{U}", Owner: me.ID, Controller: me.ID,
		})
	})
	floatForTest(g, me, "CC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "morph", Strict: true}); err == nil {
		t.Fatal("a morph cast for {3} was paid with {2}: Thryx discounted a face-down spell as a mana value 5 spell")
	}

	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "morph", Strict: true}); err != nil {
		t.Fatalf("cast face down for {3}: %v", err)
	}
	if g.SpellCantBeCounteredForEffect(id) {
		t.Error("a face-down spell can't be countered: Thryx's shield read the mana value of the card it hides")
	}
}
