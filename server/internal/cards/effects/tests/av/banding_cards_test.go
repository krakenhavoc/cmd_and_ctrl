package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// banding_cards_test.go — #5723 (ADR 0224). The four Limited Edition
// Alpha banding cards: Benalish Hero, Mesa Pegasus and Timber Wolves
// print banding, and Helm of Chatzuk grants it until end of turn. The
// rule itself is pinned in game/banding_test.go; these pin the cards.

const (
	bandingHeroOracle    = "4c81cfb7-8765-4e28-ae33-4287fa9a86cc"
	bandingPegasusOracle = "8161f5b8-6aab-4133-ba2c-2e7b5774153e"
	bandingWolvesOracle  = "35d07ac9-b184-4b5f-8192-34b1db042f69"
	bandingHelmOracle    = "948e3bb7-8265-4e95-acd1-a0c4f22441df"
)

func TestBandingCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bandingHeroOracle: "Benalish Hero", bandingPegasusOracle: "Mesa Pegasus",
		bandingWolvesOracle: "Timber Wolves", bandingHelmOracle: "Helm of Chatzuk",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	if got, ok := game.CanonicalKeyword("Banding"); !ok || got != game.KeywordBanding {
		t.Fatalf(`CanonicalKeyword("Banding") = %q, %v; want %q, true`, got, ok, game.KeywordBanding)
	}
	for _, oracle := range []string{bandingHeroOracle, bandingPegasusOracle, bandingWolvesOracle} {
		c := game.Card{OracleID: oracle, Name: oracle, TypeLine: "Creature — Test"}
		if !game.HasKeyword(&c, game.KeywordBanding) {
			t.Errorf("%s: the catalog card does not have banding", oracle)
		}
	}
	pegasus := game.Card{OracleID: bandingPegasusOracle, Name: "Mesa Pegasus", TypeLine: "Creature — Pegasus"}
	if !game.HasKeyword(&pegasus, "flying") {
		t.Error("Mesa Pegasus does not fly")
	}
}

// bandingDamagePrompt is the one damage-assignment prompt queued, or nil.
func bandingDamagePrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDamageAssignment {
			return c
		}
	}
	return nil
}

// bandingAttack declares `ids` attacking `defender` as one band and
// walks into the declare blockers step.
func bandingAttack(t *testing.T, g *game.Game, defender uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	decls := make([]game.AttackDeclaration, 0, len(ids))
	for _, id := range ids {
		decls = append(decls, game.AttackDeclaration{Attacker: id, Target: defender, Band: 1})
	}
	if _, err := g.DeclareAttackersWith(decls, game.DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare the band: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
}

// The Alpha play: Benalish Hero bands with a bigger creature; the
// defender blocks the big one with a creature that would kill it, and
// the Hero's controller puts the blocker's damage on the Hero instead.
func TestBenalishHeroBandsAndTakesTheBlockersDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	hero := effectstest.B12Push(g, me.ID, "Benalish Hero", "Creature — Human Soldier", bandingHeroOracle, 1, 1)
	giant := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	bandingAttack(t, g, opp.ID, hero, giant)
	if err := g.DeclareBlocker(ogre, giant); err != nil {
		t.Fatal(err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	if got := effectstest.ApaLive(g, ogre).BlockedAttackers(); len(got) != 2 {
		t.Fatalf("the ogre blocks %v, want the whole band", got)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	choice := bandingDamagePrompt(g)
	if choice == nil || choice.Chooser != me.ID || !choice.DamageAssignment.Banding {
		t.Fatalf("prompt = %+v, want the Hero's controller dividing the ogre's damage", choice)
	}
	if err := g.ResolveDamageAssignment(choice.ID, me.ID, []game.DamageAssignmentEntry{
		{BlockerID: giant, Amount: 0}, {BlockerID: hero, Amount: 3},
	}, 0); err != nil {
		t.Fatal(err)
	}
	if effectstest.ApaLive(g, giant) == nil || effectstest.ApaLive(g, giant).DamageMarked != 0 {
		t.Error("the Hill Giant took damage it was spared")
	}
	if effectstest.ApaLive(g, hero) != nil {
		t.Error("Benalish Hero took 3 and lived")
	}
	if effectstest.ApaLive(g, ogre) != nil {
		t.Error("the ogre took 4 from the band and lived")
	}
}

// Timber Wolves blocking: the defending player divides the attacker's
// damage among the Wolves and the creature blocking beside them.
func TestTimberWolvesBlockingDividesTheAttackersDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	giant := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	wolves := effectstest.B12Push(g, opp.ID, "Timber Wolves", "Creature — Wolf", bandingWolvesOracle, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(giant, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range []uuid.UUID{wolves, bear} {
		if err := g.DeclareBlocker(b, giant); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	choice := bandingDamagePrompt(g)
	if choice == nil || choice.Chooser != opp.ID || !choice.DamageAssignment.Banding {
		t.Fatalf("prompt = %+v, want the defending player dividing", choice)
	}
	if err := g.ResolveDamageAssignment(choice.ID, opp.ID, []game.DamageAssignmentEntry{
		{BlockerID: wolves, Amount: 3}, {BlockerID: bear, Amount: 0},
	}, 0); err != nil {
		t.Fatal(err)
	}
	if effectstest.ApaLive(g, bear) == nil {
		t.Error("the bear died though the defending player spared it")
	}
	if effectstest.ApaLive(g, giant) != nil {
		t.Error("the Hill Giant took 3 from its blockers and lived")
	}
}

// Mesa Pegasus bands and flies: a band of the Pegasus and a ground
// creature is blocked whole by a flier that blocks the Pegasus.
func TestMesaPegasusBandIsBlockedByAFlierBlockingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	pegasus := effectstest.B12Push(g, me.ID, "Mesa Pegasus", "Creature — Pegasus", bandingPegasusOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 4, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"flying"},
	})
	bandingAttack(t, g, opp.ID, pegasus, bear)
	if err := g.DeclareBlocker(bird, pegasus); err != nil {
		t.Fatalf("a flier blocking Mesa Pegasus: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	if got := effectstest.ApaLive(g, bird).BlockedAttackers(); len(got) != 2 {
		t.Fatalf("the bird blocks %v, want the Pegasus and the bear", got)
	}
}

// Helm of Chatzuk: a creature that gained banding forms a band with a
// creature without it, and the banding is gone at the next turn.
func TestHelmOfChatzukGrantsBandingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	helm := effectstest.B12Push(g, me.ID, "Helm of Chatzuk", "Artifact", bandingHelmOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	giant := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}")
	effectstest.P7Activate(t, g, me, helm, 0, game.ActivateAbilityParams{Targets: effectstest.RfCardTarget(bear)})
	if !effectstest.HasEffectiveKeyword(t, g, bear, game.KeywordBanding) {
		t.Fatal("the bear did not gain banding")
	}
	if !effectstest.TappedNow(g, helm) {
		t.Error("the Helm did not tap")
	}
	bandingAttack(t, g, opp.ID, bear, giant)
	if err := g.DeclareBlocker(wall, giant); err != nil {
		t.Fatal(err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	if got := effectstest.ApaLive(g, wall).BlockedAttackers(); len(got) != 2 {
		t.Fatalf("the wall blocks %v, want the band of two", got)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, bear, game.KeywordBanding) {
		t.Error("the bear still has banding on the next turn")
	}
}
