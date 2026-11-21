package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const interceptorOracle = "8571bd94-66c0-4cca-b10f-2c34062d008e"

// putInterceptorInGraveyard seeds Interceptor, Shadow's Hound into a
// player's graveyard with its catalog identity.
func putInterceptorInGraveyard(p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       "Interceptor, Shadow's Hound",
		TypeLine:   "Legendary Creature — Dog",
		OracleID:   interceptorOracle,
		ManaCost:   "{2}{B}{B}",
		Power:      4,
		Toughness:  3,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

// The printed keyword and the Assassin lord: Interceptor is a Dog, so
// it has menace on its own, and an Assassin of yours has it too while
// an opponent's Assassin does not.
func TestInterceptorMenaceAndTheAssassinGrant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hound := effectstest.PushCatalogPermanent(g, me.ID, "Interceptor, Shadow's Hound", "Legendary Creature — Dog", interceptorOracle, false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Assassin", "Creature — Human Assassin", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Assassin", "Creature — Human Assassin", "", false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)

	for id, want := range map[uuid.UUID]bool{hound: true, mine: true, theirs: false, bear: false} {
		if got := effectstest.EffectiveAbilitiesContain(t, g, id, "menace"); got != want {
			t.Errorf("menace on %v = %v, want %v", id, got, want)
		}
	}
}

// Paying {2}{B} when a legendary creature attacks returns the same
// card from the graveyard tapped and attacking the player the
// controller named, and it never triggers anything that watches attacks.
func TestInterceptorReturnsTappedAndAttackingTheChosenDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	legend := effectstest.PushCatalogPermanent(g, me.ID, "Legend", "Legendary Creature — Human", "", false)
	hound := putInterceptorInGraveyard(me)

	effectstest.DeclareAttack(t, g, opp.ID, legend)
	if n := effectstest.TriggersOnStackFrom(g, hound); n != 1 {
		t.Fatalf("Interceptor triggers on the stack = %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	// Three opponents: the controller names who it attacks.
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no \"what does it attack?\" question with three possible defenders")
	}
	idx := -1
	var refs []game.AttackTargetRef
	g.ReadSnapshot(func() { refs = g.AttackTargetsForEffect(me.ID) })
	for i, r := range refs {
		if r.ID == third.ID {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("the third seat is not an attack target: %v", refs)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, idx)

	c, ok := effectstest.InterceptorOnBattlefield(g, hound)
	if !ok {
		t.Fatal("Interceptor did not return to the battlefield")
	}
	if me.Graveyard.Contains(hound) {
		t.Error("Interceptor is still in the graveyard")
	}
	if !c.Tapped || c.AttackingTarget != third.ID || c.Controller != me.ID {
		t.Errorf("tapped=%v attacking=%v controller=%v, want tapped, attacking seat 2, mine", c.Tapped, c.AttackingTarget, c.Controller)
	}
	// Entering attacking is not declaring an attacker (CR 508.4): no
	// attack event names it, so nothing that watches attacks fires.
	for _, ev := range g.Events {
		if ev.Kind == game.EventAttack && ev.CardID == hound {
			t.Error("the returned Interceptor was announced as a declared attacker")
		}
	}
}

// "One or more": three legendary attackers make one offer, not three.
func TestInterceptorTriggersOncePerAttackDeclaration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushCatalogPermanent(g, me.ID, "Legend A", "Legendary Creature — Human", "", false)
	b := effectstest.PushCatalogPermanent(g, me.ID, "Legend B", "Legendary Creature — Elf", "", false)
	c := effectstest.PushCatalogPermanent(g, me.ID, "Legend C", "Legendary Creature — Orc", "", false)
	hound := putInterceptorInGraveyard(me)

	effectstest.DeclareAttack(t, g, opp.ID, a, b, c)
	if n := effectstest.TriggersOnStackFrom(g, hound); n != 1 {
		t.Fatalf("Interceptor triggers = %d, want 1 for a three-legend attack", n)
	}
}

// A nonlegendary attacker starts nothing.
func TestInterceptorIgnoresANonlegendaryAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	hound := putInterceptorInGraveyard(me)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if n := effectstest.TriggersOnStackFrom(g, hound); n != 0 {
		t.Fatalf("a nonlegendary attack triggered Interceptor %d times", n)
	}
}

// Declining leaves the card where it was.
func TestInterceptorDeclineLeavesItInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legend := effectstest.PushCatalogPermanent(g, me.ID, "Legend", "Legendary Creature — Human", "", false)
	hound := putInterceptorInGraveyard(me)

	effectstest.DeclareAttack(t, g, opp.ID, legend)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if !me.Graveyard.Contains(hound) {
		t.Error("declining {2}{B} moved Interceptor")
	}
}

// The ability works from the graveyard only: on the battlefield the
// card has nothing to return.
func TestInterceptorDoesNotTriggerFromTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legend := effectstest.PushCatalogPermanent(g, me.ID, "Legend", "Legendary Creature — Human", "", false)
	hound := effectstest.PushCatalogPermanent(g, me.ID, "Interceptor, Shadow's Hound", "Legendary Creature — Dog", interceptorOracle, false)

	effectstest.DeclareAttack(t, g, opp.ID, legend)
	if n := effectstest.TriggersOnStackFrom(g, hound); n != 0 {
		t.Fatalf("Interceptor on the battlefield triggered %d times", n)
	}
}

// CR 400.7: a card that left the graveyard after the trigger went on
// the stack is a new object, and the offer never appears.
func TestInterceptorTriggerDoesNotFollowANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legend := effectstest.PushCatalogPermanent(g, me.ID, "Legend", "Legendary Creature — Human", "", false)
	hound := putInterceptorInGraveyard(me)

	effectstest.DeclareAttack(t, g, opp.ID, legend)
	if n := effectstest.TriggersOnStackFrom(g, hound); n != 1 {
		t.Fatalf("triggers = %d, want 1", n)
	}
	// In response: exile it from the graveyard and bring it back.
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneExile}, hound,
	); err != nil {
		t.Fatalf("exile Interceptor: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the stale trigger offered a payment for a card that left the graveyard")
	}
	if _, ok := effectstest.InterceptorOnBattlefield(g, hound); ok {
		t.Error("an exiled Interceptor came back")
	}
}

// The same rule at the second pause: the card leaves while the
// controller is deciding what it attacks, and nothing is put onto the
// battlefield (CR 400.7) — the move checks the zone it is about to take
// the card from, not only the zone it was in when the trigger resolved.
func TestInterceptorLeavingWhileNamingTheDefenderPutsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legend := effectstest.PushCatalogPermanent(g, me.ID, "Legend", "Legendary Creature — Human", "", false)
	hound := putInterceptorInGraveyard(me)

	effectstest.DeclareAttack(t, g, opp.ID, legend)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatal("no defender question")
	}
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneExile}, hound,
	); err != nil {
		t.Fatalf("exile Interceptor: %v", err)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if _, ok := effectstest.InterceptorOnBattlefield(g, hound); ok {
		t.Error("a card that left the graveyard was put onto the battlefield")
	}
	// A quiet no-op, not an effect error: the instruction named an
	// object that is gone, so there is nothing to fail at.
	for _, ev := range g.Events {
		if ev.Kind == game.EventEffectError {
			t.Errorf("effect error raised: %s", ev.ErrorMsg)
		}
	}
}
