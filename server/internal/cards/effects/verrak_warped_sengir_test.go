package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// verrak_warped_sengir_test.go — #5662: the life an activation paid is
// on its event, and Verrak asks to pay it again for a copy.

const verrakOracle = "04853f06-3fd2-445d-9b0d-d68d0f116635"

func verrakTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newBigLibraryCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	verrak := pushCatalogPermanent(g, me.ID, "Verrak, Warped Sengir", "Legendary Creature — Vampire", verrakOracle, false)
	return g, me, opp, verrak
}

// verrakSettle passes priority until the stack is empty, answering every
// "pay that much life again" prompt with `yes` and every new-targets
// prompt for a copy with `retarget`. It returns the life each prompt
// asked for, in order.
func verrakSettle(t *testing.T, g *game.Game, me uuid.UUID, yes bool, retarget game.TargetRef) []int {
	t.Helper()
	var asked []int
	for i := 0; i < 64; i++ {
		if p := triggerOrderPromptFor(g, me); p != nil {
			order := append([]uuid.UUID(nil), p.TriggerOrderIDs...)
			if err := g.ResolveTriggerOrder(p.ID, me, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if c := latestConfirmFor(g, me); c != nil {
			asked = append(asked, c.LifeCost)
			if err := g.ResolveConfirm(c.ID, me, yes); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			continue
		}
		if p := latestPickTarget(g, me); p != nil {
			if err := g.ResolvePickTarget(p.ID, me, retarget); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return asked
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
	return nil
}

func verrakGriselbrand(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Griselbrand", "Legendary Creature — Demon", griselbrandOracle, false)
}

func TestVerrakHasFlyingDeathtouchAndLifelink(t *testing.T) {
	g, _, _, verrak := verrakTable(t)
	for _, kw := range []string{"flying", "deathtouch", "lifelink"} {
		if !hasEffectiveKeyword(t, g, verrak, kw) {
			t.Errorf("Verrak lacks %s", kw)
		}
	}
}

func TestVerrakPayingThatMuchLifeAgainCopiesTheAbility(t *testing.T) {
	g, me, _, _ := verrakTable(t)
	gris := verrakGriselbrand(g, me.ID)
	life, hand := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gris, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Griselbrand: %v", err)
	}
	asked := verrakSettle(t, g, me.ID, true, game.TargetRef{})
	// One prompt, for the 7 the activation paid: the second payment is
	// not an activation and the copy was not activated (CR 707.10), so
	// Verrak does not trigger again.
	if len(asked) != 1 || asked[0] != 7 {
		t.Fatalf("asked to pay %v, want one prompt for 7 life", asked)
	}
	if got := life - me.Life; got != 14 {
		t.Errorf("life paid %d, want 14: 7 to activate and 7 again", got)
	}
	if got := me.Hand.Size() - hand; got != 14 {
		t.Errorf("drew %d, want 14: the ability and its copy", got)
	}
}

func TestVerrakDecliningPaysNothingAndCopiesNothing(t *testing.T) {
	g, me, _, _ := verrakTable(t)
	gris := verrakGriselbrand(g, me.ID)
	life, hand := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gris, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Griselbrand: %v", err)
	}
	if asked := verrakSettle(t, g, me.ID, false, game.TargetRef{}); len(asked) != 1 {
		t.Fatalf("asked %v, want one prompt", asked)
	}
	if got := life - me.Life; got != 7 {
		t.Errorf("life paid %d, want only the activation's 7", got)
	}
	if got := me.Hand.Size() - hand; got != 7 {
		t.Errorf("drew %d, want 7: no copy without the payment", got)
	}
}

// CR 119.4: a player who can't pay that much life again is not asked.
func TestVerrakDoesNotAskAPlayerWhoCannotPayAgain(t *testing.T) {
	g, me, _, _ := verrakTable(t)
	gris := verrakGriselbrand(g, me.ID)
	g.WithWriteLock(func() { me.Life = 13 })
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gris, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Griselbrand: %v", err)
	}
	if asked := verrakSettle(t, g, me.ID, true, game.TargetRef{}); len(asked) != 0 {
		t.Fatalf("asked %v at 6 life, want no prompt for 7", asked)
	}
	if me.Life != 6 {
		t.Errorf("life %d, want 6", me.Life)
	}
	if got := me.Hand.Size() - hand; got != 7 {
		t.Errorf("drew %d, want 7", got)
	}
}

// The payment is re-checked when the answer arrives: life that went
// between the question and the answer pays nothing and copies nothing.
func TestVerrakRechecksThePaymentWhenTheAnswerArrives(t *testing.T) {
	g, me, _, _ := verrakTable(t)
	gris := verrakGriselbrand(g, me.ID)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gris, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Griselbrand: %v", err)
	}
	for i := 0; i < 16 && latestConfirmFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := latestConfirmFor(g, me.ID)
	if c == nil {
		t.Fatal("Verrak never asked")
	}
	g.WithWriteLock(func() { me.Life = 5 })
	// The prompt itself may refuse a yes that can't be paid; either way
	// nothing is paid and nothing is copied.
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		if err := g.ResolveConfirm(c.ID, me.ID, false); err != nil {
			t.Fatalf("ResolveConfirm(no): %v", err)
		}
	}
	verrakSettle(t, g, me.ID, false, game.TargetRef{})
	if me.Life != 5 {
		t.Errorf("life %d, want 5: a payment that can't be made is not made", me.Life)
	}
	if got := me.Hand.Size() - hand; got != 7 {
		t.Errorf("drew %d, want 7: no copy without the payment", got)
	}
}

// "If life was paid to activate it" (CR 603.4): an ability whose cost
// paid no life does not trigger Verrak at all.
func TestVerrakIgnoresAnAbilityThatPaidNoLife(t *testing.T) {
	g, me, _, _ := verrakTable(t)
	rock := pushCatalogPermanent(g, me.ID, "Basalt Monolith", "Artifact", acBasaltMonolithOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, rock, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Basalt Monolith's untap: %v", err)
	}
	if n := len(g.StackMeta); n != 1 {
		t.Fatalf("%d items on the stack, want only the ability: Verrak must not trigger", n)
	}
	if asked := verrakSettle(t, g, me.ID, true, game.TargetRef{}); len(asked) != 0 {
		t.Errorf("asked %v for an ability that paid no life", asked)
	}
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

// "Whenever YOU activate": an opponent's life payment is not Verrak's
// controller's.
func TestVerrakIgnoresAnOpponentsActivation(t *testing.T) {
	g, me, opp, _ := verrakTable(t)
	gris := verrakGriselbrand(g, opp.ID)
	life := me.Life
	if err := g.ActivateCatalogAbility(opp.ID, gris, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the opponent activates Griselbrand: %v", err)
	}
	if n := len(g.StackMeta); n != 1 {
		t.Fatalf("%d items on the stack, want only the opponent's ability", n)
	}
	if asked := verrakSettle(t, g, me.ID, true, game.TargetRef{}); len(asked) != 0 {
		t.Errorf("asked %v for an opponent's activation", asked)
	}
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

// "That isn't a mana ability" (CR 605.3b): Mana Confluence's life is a
// mana ability's cost.
func TestVerrakIgnoresAManaAbilityThatPaidLife(t *testing.T) {
	g, me, _, _ := verrakTable(t)
	land := seedPermanentWithOracle(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != life-1 {
		t.Fatalf("life %d, want %d: Mana Confluence costs 1 life", me.Life, life-1)
	}
	if latestConfirmFor(g, me.ID) != nil || len(g.StackMeta) != 0 || len(g.PendingTriggers) != 0 {
		t.Error("Verrak triggered on a mana ability")
	}
}

// CR 107.4f: life paid for Phyrexian mana symbols is life paid to
// activate the ability, and "that much" is all of it.
func TestVerrakCountsLifePaidForPhyrexianMana(t *testing.T) {
	g, me, _, _ := verrakTable(t)
	drivnod := pushCatalogPermanent(g, me.ID, "Drivnod, Carnage Dominus", "Legendary Creature — Phyrexian Horror", drivnodOracle, false)
	a := pushGraveyardCardTyped(me, "Dead One", "Creature — Zombie")
	b := pushGraveyardCardTyped(me, "Dead Two", "Creature — Zombie")
	c := pushGraveyardCardTyped(me, "Dead Three", "Creature — Zombie")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, drivnod, 0, game.ActivateAbilityParams{
		Strict: true, ExileIDs: []uuid.UUID{a, b, c}, PhyrexianLife: 2,
	}); err != nil {
		t.Fatalf("activate Drivnod: %v", err)
	}
	want := 2 * game.PhyrexianLifePerSymbol
	asked := verrakSettle(t, g, me.ID, true, game.TargetRef{})
	if len(asked) != 1 || asked[0] != want {
		t.Fatalf("asked to pay %v, want one prompt for %d", asked, want)
	}
	if got := life - me.Life; got != 2*want {
		t.Errorf("life paid %d, want %d", got, 2*want)
	}
	// Nothing is exiled again for the copy: a copy pays no costs.
	if n := counterCount(g, drivnod, "indestructible"); n != 2 {
		t.Errorf("%d indestructible counters, want 2: the ability and its copy", n)
	}
}

// The same symbols paid with mana pay no life, and Verrak does not
// trigger.
func TestVerrakIgnoresPhyrexianManaPaidWithMana(t *testing.T) {
	g, me, _, _ := verrakTable(t)
	drivnod := pushCatalogPermanent(g, me.ID, "Drivnod, Carnage Dominus", "Legendary Creature — Phyrexian Horror", drivnodOracle, false)
	a := pushGraveyardCardTyped(me, "Dead One", "Creature — Zombie")
	b := pushGraveyardCardTyped(me, "Dead Two", "Creature — Zombie")
	c := pushGraveyardCardTyped(me, "Dead Three", "Creature — Zombie")
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "B"})
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, drivnod, 0, game.ActivateAbilityParams{
		Strict: true, ExileIDs: []uuid.UUID{a, b, c},
	}); err != nil {
		t.Fatalf("activate Drivnod: %v", err)
	}
	if asked := verrakSettle(t, g, me.ID, true, game.TargetRef{}); len(asked) != 0 {
		t.Errorf("asked %v though no life was paid", asked)
	}
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
	if n := counterCount(g, drivnod, "indestructible"); n != 1 {
		t.Errorf("%d indestructible counters, want 1", n)
	}
}

// "You may choose new targets for the copy" (CR 707.10c).
func TestVerrakCopyMayTakeANewTarget(t *testing.T) {
	g, me, opp, _ := verrakTable(t)
	vilis := pushCatalogPermanent(g, me.ID, "Vilis, Broker of Blood", "Legendary Creature — Demon", vilisOracle, false)
	first := b12Creature(g, opp.ID, "First Ogre", "Creature — Ogre", 3, 3)
	second := b12Creature(g, opp.ID, "Second Ogre", "Creature — Ogre", 3, 3)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, vilis, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: first}},
	}); err != nil {
		t.Fatalf("activate Vilis: %v", err)
	}
	asked := verrakSettle(t, g, me.ID, true, game.TargetRef{Kind: game.TargetCard, ID: second})
	if len(asked) != 1 || asked[0] != 2 {
		t.Fatalf("asked to pay %v, want one prompt for 2", asked)
	}
	if got := life - me.Life; got != 4 {
		t.Errorf("life paid %d, want 4", got)
	}
	for _, id := range []uuid.UUID{first, second} {
		if p, tg := socB14PT(t, g, id); p != 2 || tg != 2 {
			t.Errorf("an Ogre is %d/%d, want 2/2: the ability hit one and its copy the other", p, tg)
		}
	}
}

// The amount rides the triggering event, so a table restored while
// Verrak is asking still knows what was paid and which ability to copy.
func TestVerrakPromptSurvivesARestore(t *testing.T) {
	g, me, _, _ := verrakTable(t)
	gris := verrakGriselbrand(g, me.ID)
	life, hand := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gris, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Griselbrand: %v", err)
	}
	// Restored with the trigger on the stack: its event keeps the 7.
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	for i := 0; i < 16 && latestConfirmFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if c := latestConfirmFor(g, me.ID); c == nil || c.LifeCost != 7 {
		t.Fatalf("after the first restore Verrak asks %+v, want a prompt for 7 life", c)
	}
	// Restored again while the question is open.
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	asked := verrakSettle(t, g, me.ID, true, game.TargetRef{})
	if len(asked) != 1 || asked[0] != 7 {
		t.Fatalf("after the second restore asked %v, want one prompt for 7", asked)
	}
	if got := life - me.Life; got != 14 {
		t.Errorf("life paid %d, want 14", got)
	}
	if got := me.Hand.Size() - hand; got != 14 {
		t.Errorf("drew %d, want 14", got)
	}
}
