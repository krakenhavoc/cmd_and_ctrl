package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// storm_windrider_test.go — #4601: "Creatures with flying can't attack you
// or block creatures you control. Whenever you cast a spell that targets
// one or more creatures, those creatures gain flying until end of turn."

const stormWindriderOracle = "06e9287b-56d4-443e-97ba-b5dcdc4fa989"

// stormTable is seat 1 controlling Storm, with a planeswalker; seat 0 (the
// active seat) has a flier and a ground creature ready to attack.
type stormTable struct {
	g           *game.Game
	s0, s1      *game.Player
	storm       uuid.UUID
	walker      uuid.UUID
	flier, bear uuid.UUID
}

func newStormTable(t *testing.T) *stormTable {
	t.Helper()
	g := newCatalogGame(t)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: seat %d is active, want seat 0", g.Turn.ActiveSeat)
	}
	s0, s1 := g.Seats[0], g.Seats[1]
	tb := &stormTable{g: g, s0: s0, s1: s1}
	tb.storm = pushCatalogPermanent(g, s1.ID, "Storm, Windrider", "Legendary Creature — Mutant Hero", stormWindriderOracle, false)
	tb.walker = pushWalkerForTest(g, s1.ID, "Test Walker", "", 3)
	tb.flier = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hawk", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: s0.ID, Controller: s0.ID, Keywords: []string{"flying"},
	})
	tb.bear = b12Creature(g, s0.ID, "Bear", "Creature — Bear", 2, 2)
	return tb
}

// A creature with flying can't attack Storm's controller, and the refusal
// says so; the other seats and a non-flier are unaffected.
func TestStormWindriderFlierCantAttackItsControllerButGroundCreatureCan(t *testing.T) {
	tb := newStormTable(t)
	advanceToDeclareAttackersOf(t, tb.g, 0)

	got := arbTargets(t, tb.g, tb.s0.ID, tb.flier)
	if got[tb.s1.ID.String()] {
		t.Fatalf("a flier is offered Storm's controller: %v", got)
	}
	if !got[tb.s0.ID.String()] && (!got[tb.g.Seats[2].ID.String()] || !got[tb.g.Seats[3].ID.String()]) {
		t.Fatalf("the flier should still reach the other opponents: %v", got)
	}
	err := tb.g.Clone().DeclareAttacker(tb.flier, tb.s1.ID)
	var refused *game.AttackTargetRestrictionError
	if !errors.As(err, &refused) || !errors.Is(err, game.ErrIllegalAttackTarget) ||
		!strings.Contains(refused.Sentence(), "Hawk can't attack "+tb.s1.Name) || strings.Contains(refused.Sentence(), "planeswalkers") {
		t.Fatalf("flier attacking Storm's controller: err = %v, want a players-only refusal", err)
	}

	ground := arbTargets(t, tb.g, tb.s0.ID, tb.bear)
	if !ground[tb.s1.ID.String()] {
		t.Fatalf("a non-flier is not offered Storm's controller: %v", ground)
	}
	if err := tb.g.Clone().DeclareAttacker(tb.bear, tb.s1.ID); err != nil {
		t.Fatalf("a creature without flying attacking Storm's controller: %v", err)
	}
}

// "You", not "you or planeswalkers you control": a flier may attack the
// controller's planeswalker.
func TestStormWindriderFlierCanAttackThePlayersPlaneswalker(t *testing.T) {
	tb := newStormTable(t)
	advanceToDeclareAttackersOf(t, tb.g, 0)

	if got := arbTargets(t, tb.g, tb.s0.ID, tb.flier); !got[tb.walker.String()] {
		t.Fatalf("the flier is not offered Storm's controller's planeswalker: %v", got)
	}
	if err := tb.g.Clone().DeclareAttacker(tb.flier, tb.walker); err != nil {
		t.Fatalf("flier attacking the planeswalker: %v", err)
	}
}

// Flying is read as attackers are declared: a creature that has lost
// flying (Gravity Sphere) may attack Storm's controller, and one that has
// gained it (a spell Storm's controller... or anyone's) may not. The chip
// the view reads follows the same fact.
func TestStormWindriderReadsFlyingLiveAtDeclaration(t *testing.T) {
	tb := newStormTable(t)
	g := tb.g
	// A creature that has gained flying (a granted keyword, not a printed
	// one) is judged by it at declaration.
	g.WithWriteLock(func() {
		if err := (ScopedEffectFor{Target: tb.bear, Mods: []game.Mod{game.AddKeywordsMod("flying")}, Duration: game.IndefiniteDuration()}).Apply(NewContext(g, nil)); err != nil {
			t.Fatal(err)
		}
	})
	advanceToDeclareAttackersOf(t, g, 0)
	if got := arbTargets(t, g, tb.s0.ID, tb.bear); got[tb.s1.ID.String()] {
		t.Fatalf("a creature that has gained flying is offered Storm's controller: %v", got)
	}

	// Gravity Sphere: every creature loses flying, so the flier attacks.
	b12Push(g, tb.s0.ID, "Gravity Sphere", "World Enchantment", gravitySphereOracle, 0, 0)
	if got := arbTargets(t, g, tb.s0.ID, tb.flier); !got[tb.s1.ID.String()] {
		t.Fatalf("a creature that lost flying is not offered Storm's controller: %v", got)
	}
	if err := g.Clone().DeclareAttacker(tb.flier, tb.s1.ID); err != nil {
		t.Fatalf("a creature that lost flying attacking Storm's controller: %v", err)
	}
}

// Storm's controller's own creatures are untouched, and the static
// follows Storm: with Storm gone, a flier attacks again.
func TestStormWindriderRestrictionLeavesWithStorm(t *testing.T) {
	tb := newStormTable(t)
	advanceToDeclareAttackersOf(t, tb.g, 0)
	if got := arbTargets(t, tb.g, tb.s0.ID, tb.flier); got[tb.s1.ID.String()] {
		t.Fatalf("setup: a flier is offered Storm's controller: %v", got)
	}
	tb.g.WithWriteLock(func() {
		_, _ = tb.g.Battlefield.Remove(tb.storm)
		// The zone move is what tells the layer pass the board changed.
		tb.g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: tb.storm, OldZone: game.ZoneBattlefield, NewZone: game.ZoneGraveyard})
	})
	if got := arbTargets(t, tb.g, tb.s0.ID, tb.flier); !got[tb.s1.ID.String()] {
		t.Fatalf("with Storm gone the flier is still refused: %v", got)
	}
}

// "…or block creatures you control": a flier can't block Storm's
// controller's creature, a ground creature can, and a flier may still
// block anyone else's.
func TestStormWindriderFliersCantBlockItsControllersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	s0, s1 := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, s0.ID, "Storm, Windrider", "Legendary Creature — Mutant Hero", stormWindriderOracle, false)
	mine := b12Creature(g, s0.ID, "My Bear", "Creature — Bear", 2, 2)
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hawk", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: s1.ID, Controller: s1.ID, Keywords: []string{"flying"},
	})
	ground := b12Creature(g, s1.ID, "Ox", "Creature — Ox", 2, 2)
	theirs := b12Creature(g, g.Seats[2].ID, "Their Bear", "Creature — Bear", 2, 2)
	refused := func(attacker, blocker uuid.UUID) bool {
		var r game.BlockRefusal
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			var a, b *game.Card
			for i := range g.Battlefield.Cards {
				switch g.Battlefield.Cards[i].InstanceID {
				case attacker:
					a = &g.Battlefield.Cards[i]
				case blocker:
					b = &g.Battlefield.Cards[i]
				}
			}
			r = g.BlockPairRefusalLocked(a, b)
		})
		return r.Reason != ""
	}
	if !refused(mine, flier) {
		t.Error("a flier may block a creature Storm's controller controls")
	}
	if refused(mine, ground) {
		t.Error("a ground creature is refused a block on Storm's controller's creature")
	}
	if refused(theirs, flier) {
		t.Error("a flier is refused a block on a creature Storm's controller does not control")
	}
}

// "Whenever you cast a spell that targets one or more creatures, those
// creatures gain flying until end of turn": any controller's creature, once
// per spell, and only as the objects they were; a spell that targets no
// creature does nothing.
func TestStormWindriderCastGivesTargetedCreaturesFlying(t *testing.T) {
	g := newCatalogGame(t)
	s0, s1 := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, s0.ID, "Storm, Windrider", "Legendary Creature — Mutant Hero", stormWindriderOracle, false)
	mine := b12Creature(g, s0.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, s1.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, s1.ID, "Bystander", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Split Pump", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"My Bear": mine, "Their Bear": theirs} {
		if !wpHas(g, wpCard(t, g, id), "flying") {
			t.Errorf("%s did not gain flying", name)
		}
	}
	if wpHas(g, wpCard(t, g, other), "flying") {
		t.Error("an untargeted creature gained flying")
	}

	castCatalogSpell(t, g, "Shock", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: s1.ID}})
	passPriorityAroundTable(t, g)
	if wpHas(g, wpCard(t, g, other), "flying") {
		t.Error("a spell that targets no creature gave flying")
	}
}
