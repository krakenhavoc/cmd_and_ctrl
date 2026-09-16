package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ward_life_sacrifice_test.go — the two non-mana ward costs.
//
//	ward—pay N life        Refraction Elemental, Sedgemoor Witch
//	ward—sacrifice a creature   Vein Ripper
//
// ward_test.go pins what every ward shares (a legal target, the caster
// pays, own spells don't trigger it). These pin what the cost changes:
// who is asked, what the prompt charges, what "can't pay" means for
// that cost, and that paying actually takes the life or the creature.

const (
	refractionElementalOracle = invasionOfKarsusOracleID + "#1"
	sedgemoorWitchOracle      = "25dce517-ac0a-4577-89ed-04296c7c4069"
	veinRipperOracle          = "f83a768c-162d-46f0-8ebd-d9c6b2c69322"
)

// wardTable puts a catalog ward creature on seat 0's battlefield, moves
// to a main phase and hands priority to seat 1, the caster.
func wardTable(t *testing.T, name, typeLine, oracle string) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warded := pushCatalogPermanent(g, me.ID, name, typeLine, oracle, false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	return g, me, opp, warded
}

// passUntilConfirmFor passes priority until the ward trigger resolves
// and its prompt reaches chooser, or the stack settles without one.
func passUntilConfirmFor(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 8; i++ {
		if c := confirmChoiceFor(g, chooser); c != nil {
			return c
		}
		if stackFullyEmpty(g) {
			return nil
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return confirmChoiceFor(g, chooser)
}

// settleDrainingAt passes priority to an empty stack, answering every
// Vein Ripper drain target prompt addressed to controller with victim.
func settleDrainingAt(t *testing.T, g *game.Game, controller, victim uuid.UUID) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if p := latestPickTarget(g, controller); p != nil {
			if err := g.ResolvePickTarget(p.ID, controller, game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("stack did not settle")
}

// --- ward—pay N life -------------------------------------------------

// Declining counters the spell and costs nothing. The prompt goes to
// the caster and declares the life it would charge, so the bot's move
// list can price the accept branch.
func TestLifeWardDeclinedCountersTheSpell(t *testing.T) {
	g, me, opp, elemental := wardTable(t, "Refraction Elemental", "Creature — Elemental", refractionElementalOracle)
	blade := castAtWardedCreature(t, g, opp, elemental)
	life := opp.Life

	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if confirmChoiceFor(g, me.ID) != nil {
		t.Error("the ward permanent's controller must not be asked to pay")
	}
	if prompt.LifeCost != 2 {
		t.Errorf("prompt LifeCost = %d, want 2", prompt.LifeCost)
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(elemental) {
		t.Error("a declined ward must counter the spell — the Elemental should live")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell should be in its owner's graveyard")
	}
	if opp.Life != life {
		t.Errorf("declining cost %d life, want 0", life-opp.Life)
	}
}

// Paying takes the life and the spell resolves.
func TestLifeWardPaidTakesTheLifeAndTheSpellResolves(t *testing.T) {
	g, _, opp, elemental := wardTable(t, "Refraction Elemental", "Creature — Elemental", refractionElementalOracle)
	castAtWardedCreature(t, g, opp, elemental)
	life := opp.Life

	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)

	if opp.Life != life-2 {
		t.Errorf("paying the ward: life %d -> %d, want -2", life, opp.Life)
	}
	if g.Battlefield.Contains(elemental) {
		t.Error("the ward was paid — the removal spell should have resolved")
	}
}

// CR 119.4: a caster below the payment cannot pay, so the spell is
// countered and nobody is asked.
func TestLifeWardCountersWithoutAPromptWhenTheCasterCannotPay(t *testing.T) {
	g, _, opp, elemental := wardTable(t, "Refraction Elemental", "Creature — Elemental", refractionElementalOracle)
	opp.Life = 1
	blade := castAtWardedCreature(t, g, opp, elemental)

	if prompt := passUntilConfirmFor(t, g, opp.ID); prompt != nil {
		t.Fatalf("a caster at 1 life was offered a 2-life payment: %q", prompt.Reason)
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(elemental) || !opp.Graveyard.Contains(blade) {
		t.Error("a ward the caster cannot pay must counter the spell")
	}
	if opp.Life != 1 {
		t.Errorf("life %d, want 1 — nothing was paid", opp.Life)
	}
}

// Sedgemoor Witch carries the same cost at 3.
func TestSedgemoorWitchWardAsksForThreeLife(t *testing.T) {
	g, _, opp, witch := wardTable(t, "Sedgemoor Witch", "Creature — Human Warlock", sedgemoorWitchOracle)
	castAtWardedCreature(t, g, opp, witch)

	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if prompt.LifeCost != 3 {
		t.Errorf("prompt LifeCost = %d, want 3", prompt.LifeCost)
	}
}

// Magecraft: an instant its controller casts makes a Pest; an
// opponent's does not.
func TestSedgemoorWitchMakesAPestWhenYouCastAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Sedgemoor Witch", "Creature — Human Warlock", sedgemoorWitchOracle, false)

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if n := countBattlefieldNamed(g, me.ID, "Pest"); n != 1 {
		t.Errorf("%d Pests after casting an instant, want 1", n)
	}
}

// --- ward—sacrifice a creature ------------------------------------

// Paying is a two-link chain: the confirm, then a pick over the
// CASTER's creatures only — never the Ripper's controller's. The pick
// is sacrificed, and the removal resolves. Both deaths feed the
// Ripper's drain, which is how the test sees them.
func TestSacrificeWardPaidSacrificesTheCastersCreature(t *testing.T) {
	g, me, opp, ripper := wardTable(t, "Vein Ripper", "Creature — Vampire Assassin", veinRipperOracle)
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	castAtWardedCreature(t, g, opp, ripper)
	myLife, oppLife := me.Life, opp.Life

	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if prompt.LifeCost != 0 {
		t.Errorf("a sacrifice ward declared LifeCost %d", prompt.LifeCost)
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("accepting the ward queued no sacrifice pick")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != bear {
		t.Fatalf("candidates = %v, want only the caster's creature %s", pick.ChooseCards, bear)
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("pick bounds %d..%d, want exactly one", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Fatal("the chosen creature was not sacrificed")
	}
	settleDrainingAt(t, g, me.ID, opp.ID)

	if !opp.Graveyard.Contains(bear) {
		t.Error("the sacrificed creature should be in its owner's graveyard")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the Ripper's controller's own creature must be untouched")
	}
	if g.Battlefield.Contains(ripper) {
		t.Error("the ward was paid — the removal spell should have resolved")
	}
	// Two creatures died (the sacrifice, then the Ripper): two drains.
	if opp.Life != oppLife-4 || me.Life != myLife+4 {
		t.Errorf("two drains of 2: me %+d opp %+d, want +4 / -4", me.Life-myLife, opp.Life-oppLife)
	}
}

// Declining counters the spell and nothing is sacrificed.
func TestSacrificeWardDeclinedCountersTheSpell(t *testing.T) {
	g, _, opp, ripper := wardTable(t, "Vein Ripper", "Creature — Vampire Assassin", veinRipperOracle)
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	blade := castAtWardedCreature(t, g, opp, ripper)

	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if chooseCardsChoiceFor(g, opp.ID) != nil {
		t.Error("declining must not ask for a sacrifice")
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(ripper) || !opp.Graveyard.Contains(blade) {
		t.Error("a declined ward must counter the spell")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("declining sacrificed a creature")
	}
}

// A caster with no creature cannot pay: countered, no prompt. The
// Ripper's controller having creatures does not count.
func TestSacrificeWardCountersWithoutAPromptWhenTheCasterHasNoCreature(t *testing.T) {
	g, me, opp, ripper := wardTable(t, "Vein Ripper", "Creature — Vampire Assassin", veinRipperOracle)
	b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	blade := castAtWardedCreature(t, g, opp, ripper)

	if prompt := passUntilConfirmFor(t, g, opp.ID); prompt != nil {
		t.Fatalf("a caster with no creature was offered the sacrifice: %q", prompt.Reason)
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(ripper) || !opp.Graveyard.Contains(blade) {
		t.Error("a ward the caster cannot pay must counter the spell")
	}
}

// --- construction ----------------------------------------------------

// A ward cost is exactly one component. Anything else fails at catalog
// init rather than charging whichever component the resolver checks
// first.
func TestWardCostMustHaveExactlyOneComponent(t *testing.T) {
	for name, cost := range map[string]WardCost{
		"empty":         {},
		"mana and life": {Mana: "{1}", Life: 1},
		"life and sac":  {Life: 2, Sacrifice: &WardSacrificeCost{Label: "a creature", OK: Creature()}},
		"nil predicate": {Sacrifice: &WardSacrificeCost{Label: "a creature"}},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: Ward accepted %+v", name, cost)
				}
			}()
			Ward(cost, "")
		}()
	}
	for _, ok := range []WardCost{WardMana("{2}"), WardLife(3), WardSacrifice("a creature", Creature())} {
		Ward(ok, "")
	}
}
