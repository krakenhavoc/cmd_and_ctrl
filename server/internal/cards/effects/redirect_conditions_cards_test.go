package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// redirect_conditions_cards_test.go — ADR 0108 §9, amendment 2026-10-16
// (#2056): Blood of the Martyr (an optional redirection from every
// creature), Kjeldoran Royal Guard (from unblocked creatures) and Ascent
// of the Worthy (from your creatures, never you), each against the damage
// it must move and the damage it must leave alone.

const (
	rcBloodOfTheMartyr    = "0e69e6c2-fdc5-4c1a-bd96-c4283b82743a"
	rcKjeldoranRoyalGuard = "8137e765-4df9-469f-a527-dee91d58fb7e"
	rcAscentOfTheWorthy   = "d748bad4-dd4c-4553-9fcc-e462260f6ff3"
)

func TestRedirectConditionCardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		rcBloodOfTheMartyr:    "Blood of the Martyr",
		rcKjeldoranRoyalGuard: "Kjeldoran Royal Guard",
		rcAscentOfTheWorthy:   "Ascent of the Worthy",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s: not registered (got %q)", name, spec.Name)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

// rcMayPrompt is the open optional_replacement prompt for `chooser`.
func rcMayPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionalReplacement {
			if c.Chooser != chooser {
				t.Fatalf("the may is asked of %v, want %v", c.Chooser, chooser)
			}
			return c
		}
	}
	t.Fatal("no optional_replacement prompt open")
	return nil
}

// Blood of the Martyr: damage to any creature, an opponent's included,
// is offered to the caster; yes moves it to them, no leaves it; damage to
// a player is not asked about.
func TestBloodOfTheMartyrOffersEachCreaturesDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pr7Creature(g, opp.ID, "Theirs", 2, "G")
	mine := pr7Creature(g, me.ID, "Mine", 2, "W")
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	castCatalogSpell(t, g, "Blood of the Martyr", "Instant", rcBloodOfTheMartyr, nil)
	passPriorityAroundTable(t, g)
	life := me.Life

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(src, theirs, 3); err != nil {
			t.Fatal(err)
		}
	})
	p := rcMayPrompt(t, g, me.ID)
	if !strings.Contains(p.Reason, "Blood of the Martyr") || !strings.Contains(p.Reason, "3 damage") {
		t.Errorf("reason %q", p.Reason)
	}
	if err := g.ResolveOptionalReplacement(p.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if me.Life != life-3 || pr6Marked(g, theirs) != 0 {
		t.Fatalf("yes: life %d (want %d), their creature %d (want 0)", me.Life, life-3, pr6Marked(g, theirs))
	}

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(src, mine, 2); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.ResolveOptionalReplacement(rcMayPrompt(t, g, me.ID).ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if me.Life != life-3 || pr6Marked(g, mine) != 2 {
		t.Fatalf("no: life %d (want %d), my creature %d (want 2)", me.Life, life-3, pr6Marked(g, mine))
	}

	pr7Hit(t, g, src, me.ID, 1)
	if me.Life != life-4 {
		t.Fatalf("life %d, want %d", me.Life, life-4)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionalReplacement {
			t.Fatal("damage to a player was offered")
		}
	}
}

// Blood of the Martyr in combat: each creature's combat damage is its own
// event and its own question, asked in the middle of the combat damage
// step; the yes and the no each land.
func TestBloodOfTheMartyrInCombat(t *testing.T) {
	g := newCatalogGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	raider := apaPush(g, atk.ID, atk.ID, game.Card{Name: "Raider", TypeLine: "Creature — Test", Power: 3, Toughness: 5})
	wall := apaPush(g, def.ID, def.ID, game.Card{Name: "Wall", TypeLine: "Creature — Test", Power: 2, Toughness: 5})
	g.WithWriteLock(func() {
		if err := g.RedirectDamageThisTurnForEffect(game.DamageRedirection{
			Controller: def.ID, ProtectSet: game.DamageRecipientFilter{Permanents: true, Match: []game.PermanentQuery{QueryTypes("creature")}},
			To: def.ID, Optional: true, Label: "Blood of the Martyr",
		}); !err {
			t.Fatal("no redirection")
		}
	})
	life := def.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, def.ID); err != nil {
		t.Fatal(err)
	}
	b35Block(t, g, wall, raider)
	for i := 0; i < 20 && g.Turn.Step != game.StepCombatDamage; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	answered := 0
	for i := 0; i < 4; i++ {
		var p *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionalReplacement {
				p = c
			}
		}
		if p == nil {
			break
		}
		amount, from, _ := func() (int, uuid.UUID, uuid.UUID) {
			var a int
			var f, to uuid.UUID
			g.ReadSnapshot(func() { a, f, to = g.DamageRedirectOfferLocked(p) })
			return a, f, to
		}()
		// Yes for the defender's own Wall, no for the Raider.
		if err := g.ResolveOptionalReplacement(p.ID, def.ID, from == wall); err != nil {
			t.Fatal(err)
		}
		if (from == wall && amount != 3) || (from == raider && amount != 2) {
			t.Errorf("offer of %d to %v", amount, from)
		}
		answered++
	}
	if answered != 2 {
		t.Fatalf("answered %d questions, want 2: one per creature dealt damage", answered)
	}
	if def.Life != life-3 || pr6Marked(g, wall) != 0 || pr6Marked(g, raider) != 2 {
		t.Fatalf("life %d (want %d), Wall %d (want 0), Raider %d (want 2)", def.Life, life-3, pr6Marked(g, wall), pr6Marked(g, raider))
	}
}

// Kjeldoran Royal Guard: the unblocked attacker's combat damage to its
// controller is dealt to the Guard; the blocked attacker's damage to its
// blocker is left alone, and so is the unblocked attacker's noncombat
// damage.
func TestKjeldoranRoyalGuardTakesTheUnblockedDamage(t *testing.T) {
	g := newCatalogGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	guard := apaPush(g, def.ID, def.ID, apaCreature("Kjeldoran Royal Guard", rcKjeldoranRoyalGuard, 2, 5))
	free := apaPush(g, atk.ID, atk.ID, game.Card{Name: "Free", TypeLine: "Creature — Test", Power: 3, Toughness: 3})
	stopped := apaPush(g, atk.ID, atk.ID, game.Card{Name: "Stopped", TypeLine: "Creature — Test", Power: 2, Toughness: 3})
	chump := apaPush(g, def.ID, def.ID, game.Card{Name: "Chump", TypeLine: "Creature — Test", Power: 0, Toughness: 4})
	pr7Activate(t, g, def.ID, guard, 0, game.ActivateAbilityParams{})
	if c := findBattlefieldCardForTest(g, guard); c == nil || !c.Tapped {
		t.Fatal("the Guard is not tapped")
	}
	life := def.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{free, stopped} {
		if err := g.DeclareAttacker(a, def.ID); err != nil {
			t.Fatal(err)
		}
	}
	b35Block(t, g, chump, stopped)
	advanceTo(t, g, game.StepCombatDamage)
	if def.Life != life || pr6Marked(g, guard) != 3 || pr6Marked(g, chump) != 2 {
		t.Fatalf("life %d (want %d), Guard %d (want 3), Chump %d (want 2)", def.Life, life, pr6Marked(g, guard), pr6Marked(g, chump))
	}
	// Still unblocked, but noncombat damage is not combat damage.
	pr7Hit(t, g, free, def.ID, 1)
	if def.Life != life-1 || pr6Marked(g, guard) != 3 {
		t.Fatalf("noncombat: life %d (want %d), Guard %d (want 3)", def.Life, life-1, pr6Marked(g, guard))
	}
}

// rcChooseOwn answers the open "choose a creature you control" prompt.
func rcChooseOwn(t *testing.T, g *game.Game, chooser, pick uuid.UUID) {
	t.Helper()
	for i := 0; i < 4; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOwnPermanents {
				if err := g.ResolveOwnPermanents(c.ID, chooser, []uuid.UUID{pick}); err != nil {
					t.Fatal(err)
				}
				return
			}
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("no creature choice open")
}

// Ascent of the Worthy: I redirects your creatures' damage to the chosen
// one, never yours and never an opponent's creature's; III returns a
// creature card with a flying counter as an Angel Warrior as well, and the
// Saga goes.
func TestAscentOfTheWorthy(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	chosen := pr7Creature(g, me.ID, "Chosen", 1, "W")
	other := pr7Creature(g, me.ID, "Other", 1, "W")
	theirs := pr7Creature(g, opp.ID, "Theirs", 1, "G")
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})

	saga := castCatalogSpell(t, g, "Ascent of the Worthy", "Enchantment — Saga", rcAscentOfTheWorthy, nil)
	passPriorityAroundTable(t, g)
	rcChooseOwn(t, g, me.ID, chosen)
	passPriorityAroundTable(t, g)

	life := me.Life
	pr6Damage(t, g, src, other, 2)
	pr6Damage(t, g, src, theirs, 1)
	pr7Hit(t, g, src, me.ID, 1)
	if pr6Marked(g, other) != 0 || pr6Marked(g, chosen) != 2 || pr6Marked(g, theirs) != 1 || me.Life != life-1 {
		t.Fatalf("other %d (want 0), chosen %d (want 2), theirs %d (want 1), life %d (want %d)",
			pr6Marked(g, other), pr6Marked(g, chosen), pr6Marked(g, theirs), me.Life, life-1)
	}

	// II at your next precombat main, then III at the one after.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	rcChooseOwn(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters %d, want 2", got)
	}
	advanceToPrecombatMainOf(t, g, seat)
	for i := 0; i < 4 && !g.Battlefield.Contains(bear); i++ {
		if p := latestPickTarget(g, me.ID); p != nil {
			pickCard(t, g, me.ID, bear)
		}
		passPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("chapter III did not return the creature card")
	}
	if got := counterCount(g, bear, game.CounterFlying); got != 1 {
		t.Errorf("flying counters %d, want 1", got)
	}
	ch := lilianaDMEffective(t, g, bear)
	for _, want := range []string{"Bear", "Angel", "Warrior"} {
		if !typeListHasForTest(ch.Subtypes, want) {
			t.Errorf("subtypes %v, want %s", ch.Subtypes, want)
		}
	}
	if loreCountersOn(g, saga) != -1 {
		t.Error("the Saga was not sacrificed after III")
	}
}
