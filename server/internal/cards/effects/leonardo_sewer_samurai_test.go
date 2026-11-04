package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// leonardo_sewer_samurai_test.go — #4964: "During your turn, you may cast
// creature spells with power or toughness 1 or less from your graveyard.
// If you cast a spell this way, that creature enters with a finality
// counter on it", plus sneak and double strike.

const leonardoSewerSamuraiOracle = "74e3243b-6707-4769-8767-7cf2ef71009a"

// leonardoTable is seat 0's precombat main with Leonardo on the
// battlefield under seat 0.
func leonardoTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	leo := pushCatalogPermanent(g, me.ID, "Leonardo, Sewer Samurai",
		"Legendary Creature — Mutant Ninja Turtle Samurai", leonardoSewerSamuraiOracle, false)
	return g, me, leo
}

// leonardoGraveCreature puts a {0} creature card with these printed
// numbers into p's graveyard.
func leonardoGraveCreature(p *game.Player, name string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear", ManaCost: "{0}",
		Power: power, Toughness: toughness,
		Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

// leonardoOpens is the one read: does Leonardo's permission let `me` cast
// this card out of their graveyard?
func leonardoOpens(g *game.Game, me *game.Player, card uuid.UUID) bool {
	return grantedPermissionOn(g, me.ID, card, game.ZoneGraveyard).Granted()
}

func TestLeonardoSewerSamuraiIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(leonardoSewerSamuraiOracle)
	if !ok {
		t.Fatal("Leonardo, Sewer Samurai is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	alt := game.AlternativeCostByKey(leonardoSewerSamuraiOracle, "sneak")
	if alt == nil || !alt.Sneak || alt.ManaCost != "{2}{W}{W}" || alt.ReturnToHand == nil {
		t.Errorf("sneak offer %+v, want Sneak {2}{W}{W} returning an unblocked attacker", alt)
	}
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "double strike" {
		t.Errorf("printed keywords %v, want double strike", spec.PrintedKeywords)
	}
	if len(spec.CastPermissions) != 1 {
		t.Fatalf("%d cast permissions, want 1", len(spec.CastPermissions))
	}
	perm := spec.CastPermissions[0]
	if perm.Zone != game.ZoneGraveyard || !perm.CastOnly || perm.Timing != game.TimingYourTurnOnly {
		t.Errorf("permission zone %v castOnly %v timing %q, want a cast-only graveyard permission on your turn",
			perm.Zone, perm.CastOnly, perm.Timing)
	}
	if !perm.Filter.CreatureOnly || !perm.Filter.PowerOrToughnessCapped || perm.Filter.PowerOrToughnessAtMost != 1 {
		t.Errorf("filter %+v, want creature spells with power or toughness 1 or less", perm.Filter)
	}
	if perm.EntersWithCounterKind != game.CounterFinality || perm.EntersWithCounters != 1 {
		t.Errorf("rider %q x%d, want one finality counter", perm.EntersWithCounterKind, perm.EntersWithCounters)
	}
	if len(perm.PerType) != 0 || perm.CastsLeft != 0 {
		t.Errorf("the permission is limited (perType %v castsLeft %d); the card prints no limit", perm.PerType, perm.CastsLeft)
	}
}

// The headline clause: power OR toughness. A 2/1 and a 1/2 both qualify,
// a 2/2 does not, and a 0/7 does on its power alone.
func TestLeonardoOpensCreaturesWithPowerOrToughnessOneOrLess(t *testing.T) {
	g, me, _ := leonardoTable(t)
	for _, tc := range []struct {
		name             string
		power, toughness int
		want             bool
	}{
		{"1/1", 1, 1, true},
		{"2/1", 2, 1, true},
		{"1/2", 1, 2, true},
		{"2/2", 2, 2, false},
		{"0/7", 0, 7, true},
		{"7/0", 7, 0, true},
		{"5/5", 5, 5, false},
	} {
		id := leonardoGraveCreature(me, tc.name, tc.power, tc.toughness)
		if got := leonardoOpens(g, me, id); got != tc.want {
			t.Errorf("a %s in the graveyard: opened = %v, want %v", tc.name, got, tc.want)
		}
		if err := muldrothaCast(g, me, id, ""); (err == nil) != tc.want {
			t.Errorf("casting the %s from the graveyard: err = %v, want allowed = %v", tc.name, err, tc.want)
		}
		if tc.want {
			resolveAllPT(t, g)
		}
	}
}

// "Creature spells": no other card type is opened, whatever its numbers.
func TestLeonardoOpensNoOtherCardType(t *testing.T) {
	g, me, _ := leonardoTable(t)
	for _, typeLine := range []string{"Instant", "Sorcery", "Artifact", "Enchantment", "Basic Land — Forest"} {
		id := ptGraveCard(me, "Grave "+typeLine, typeLine)
		if leonardoOpens(g, me, id) {
			t.Errorf("a %s card in the graveyard was opened", typeLine)
		}
		if err := muldrothaCast(g, me, id, ""); err == nil {
			t.Errorf("a %s card was cast or played from the graveyard", typeLine)
		}
	}
}

// A `*/*` card's characteristic-defining ability functions in the
// graveyard (CR 604.3, CR 113.6a, CR 208.2a): Tarmogoyf with only itself
// in the graveyards is a 1/2 and qualifies; add an instant card to a
// graveyard and it is a 2/3 and does not. Its PRINTED numbers (0/1) are
// never the answer.
func TestLeonardoJudgesAStarCardByItsOwnAbilityInTheGraveyard(t *testing.T) {
	g, me, _ := leonardoTable(t)
	goyf := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{
			InstanceID: goyf, Name: "Tarmogoyf", TypeLine: "Creature — Lhurgoyf",
			OracleID: tarmogoyfOracle, ManaCost: "{0}", Power: 0, Toughness: 1,
			VariableToughness: true,
			Owner:             me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	// One card type (creature) among all graveyards: a 1/2.
	if !leonardoOpens(g, me, goyf) {
		t.Error("a 1/2 Tarmogoyf in the graveyard was not opened")
	}
	pushGraveyardCardWithTypeLine(g, g.Seats[1].ID, "Lightning Bolt", "Instant")
	// Two card types: a 2/3, and neither number is 1 or less.
	if leonardoOpens(g, me, goyf) {
		t.Error("a 2/3 Tarmogoyf in the graveyard was opened")
	}
	if err := muldrothaCast(g, me, goyf, ""); err == nil {
		t.Error("a 2/3 Tarmogoyf was cast from the graveyard")
	}
}

// A `*` the engine cannot compute is not "1 or less": an unknown number
// makes the permission narrower than printed, never wider.
func TestLeonardoDoesNotOpenAStarCardTheEngineCannotCompute(t *testing.T) {
	g, me, _ := leonardoTable(t)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: "Unknown Lhurgoyf", TypeLine: "Creature — Lhurgoyf",
			ManaCost: "{0}", Power: 0, Toughness: 0, VariableToughness: true,
			Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	if leonardoOpens(g, me, id) {
		t.Error("a */* card with no computable power or toughness was opened")
	}
	if err := muldrothaCast(g, me, id, ""); err == nil {
		t.Error("a */* card with no computable power or toughness was cast from the graveyard")
	}
}

// "During your turn" (TimingYourTurnOnly), and "your graveyard": an
// opponent's turn closes the permission, and an opponent's pile is never
// opened.
func TestLeonardoOpensOnlyYourTurnAndYourOwnGraveyard(t *testing.T) {
	g, me, _ := leonardoTable(t)
	theirs := leonardoGraveCreature(g.Seats[1], "Their Bear", 1, 1)
	if leonardoOpens(g, me, theirs) {
		t.Error("a 1/1 in an opponent's graveyard was opened")
	}
	mine := leonardoGraveCreature(me, "My Bear", 1, 1)
	if !leonardoOpens(g, me, mine) {
		t.Fatal("a 1/1 in my own graveyard was not opened on my turn (sanity)")
	}
	if !enumeratedFor(g, me.ID, mine) {
		t.Fatal("the cast is not a legal move on my own turn (sanity)")
	}
	advanceToUpkeepOf(t, g, 1)
	advanceToMain(t, g)
	// The permission is still there (a static ability's permission always
	// is); what "during your turn" closes is the cast.
	if enumeratedFor(g, me.ID, mine) {
		t.Error("the cast is a legal move on an opponent's turn")
	}
	if err := muldrothaCast(g, me, mine, ""); err == nil {
		t.Error("a creature was cast from the graveyard on an opponent's turn")
	}
}

// The rider: a creature cast this way enters with a finality counter, and
// the same creature cast from hand enters with none.
func TestLeonardoGraveyardCastEntersWithAFinalityCounter(t *testing.T) {
	g, me, _ := leonardoTable(t)
	bear := leonardoGraveCreature(me, "Grave Bear", 1, 1)
	if err := muldrothaCast(g, me, bear, ""); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	resolveAllPT(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature cast from the graveyard did not resolve")
	}
	if got := counterCount(g, bear, game.CounterFinality); got != 1 {
		t.Errorf("it entered with %d finality counters, want 1", got)
	}
	hand := castFromHandForTest(t, g, me, "Hand Bear", "Creature — Bear", "{0}", "", game.CastSpellParams{})
	resolveAllPT(t, g)
	if got := counterCount(g, hand, game.CounterFinality); got != 0 {
		t.Errorf("the same creature cast from hand entered with %d finality counters, want 0", got)
	}
}

// And the counter does what CR 122.1h says: the creature is exiled
// rather than put into a graveyard.
func TestLeonardoFinalityCounterExilesTheCreatureInstead(t *testing.T) {
	g, me, _ := leonardoTable(t)
	bear := leonardoGraveCreature(me, "Grave Bear", 1, 1)
	if err := muldrothaCast(g, me, bear, ""); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	resolveAllPT(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !g.Exile.Contains(bear) || me.Graveyard.Contains(bear) {
		t.Error("the creature with a finality counter was not exiled instead of dying")
	}
}

// Double strike is Leonardo's own printed keyword on the battlefield.
func TestLeonardoHasDoubleStrike(t *testing.T) {
	g, _, leo := leonardoTable(t)
	if !hasEffectiveKeyword(t, g, leo, "double strike") {
		t.Error("Leonardo does not have double strike")
	}
}

// A restore with the permission open: the restored Leonardo still opens
// the graveyard and still gives the counter.
func TestLeonardoPermissionSurvivesARestore(t *testing.T) {
	g, me, _ := leonardoTable(t)
	bear := leonardoGraveCreature(me, "Grave Bear", 2, 1)
	restored := restoreThroughJSON(t, g)
	rme := restored.Seats[0]
	if rme.ID != me.ID {
		t.Fatal("the restored seat 0 is another player")
	}
	if err := muldrothaCast(restored, rme, bear, ""); err != nil {
		t.Fatalf("cast from the restored graveyard: %v", err)
	}
	resolveAllPT(t, restored)
	if got := counterCount(restored, bear, game.CounterFinality); got != 1 {
		t.Errorf("after a restore: %d finality counters, want 1", got)
	}
}
