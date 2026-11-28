package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b774_test.go — S82 OLD batch 774 (Weatherlight: Auras, #5629), each
// Aura cast and attached through the real spell path.

const (
	oldb774Abduction  = "d0e1904e-1a37-41f6-8582-b9ea794bb886"
	oldb774FireWhip   = "9d4372f0-eedf-4897-9f03-f96fb14ac4a6"
	oldb774Coils      = "02a8a4e6-2081-4903-81ad-96e5a949dff8"
	oldb774Empyrial   = "ec9e8611-e785-4d5d-b6e1-bf63daf408df"
	oldb774Betrothed  = "77849657-c872-4d54-9149-a08bbe90f787"
	oldb774NaturesKis = "e3e87fee-5717-4ca1-adfd-4df8732020f8"
	oldb774BriarShld  = "9d437391-d476-4976-8714-6a1d3185ab52"
	oldb774Apathy     = "09ac5adf-f722-417e-b0fc-7ed0d8c9abd6"
)

func TestOldB774EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldb774Abduction: "Abduction", oldb774FireWhip: "Fire Whip", oldb774Coils: "Coils of the Medusa",
		oldb774Empyrial: "Empyrial Armor", oldb774Betrothed: "Betrothed of Fire", oldb774NaturesKis: "Nature's Kiss",
		oldb774BriarShld: "Briar Shield", oldb774Apathy: "Apathy",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

func oldb774Size(t *testing.T, g *game.Game, id uuid.UUID, p, tough int) {
	t.Helper()
	if gp, gt := effectstest.PtOf(t, g, id); gp != p || gt != tough {
		t.Errorf("creature is %d/%d, want %d/%d", gp, gt, p, tough)
	}
}

func oldb774Controller(g *game.Game, id uuid.UUID) uuid.UUID {
	var c uuid.UUID
	g.ReadSnapshot(func() {
		if card := effectstest.FindBattlefieldCardForTest(g, id); card != nil {
			c = card.Controller
		}
	})
	return c
}

func oldb774Graveyard(p *game.Player, typeLines ...string) []uuid.UUID {
	var ids []uuid.UUID
	for _, tl := range typeLines {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Card " + tl, TypeLine: tl, Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

func TestOldB774AbductionStealsUntapsAndReturnsTheCreatureToItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.Enchant(t, g, "Abduction", oldb774Abduction, bear)
	if got := oldb774Controller(g, bear); got != me.ID {
		t.Fatalf("controller is %v, want the Aura's controller", got)
	}
	if effectstest.IsTapped(g, bear) {
		t.Error("the enchanted creature stayed tapped")
	}
	effectstest.KillCreature(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature did not return to the battlefield")
	}
	if got := oldb774Controller(g, bear); got != opp.ID {
		t.Errorf("the creature returned under %v, want its owner", got)
	}
}

func TestOldB774AbductionOfYourOwnCreatureReturnsItUnderYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.Enchant(t, g, "Abduction", oldb774Abduction, bear)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || oldb774Controller(g, bear) != me.ID {
		t.Error("the creature did not come back under its owner")
	}
}

func TestOldB774FireWhipGivesThePingAndSacrificesForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	aura := effectstest.Enchant(t, g, "Fire Whip", oldb774FireWhip, host)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, host, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("the enchanted creature's granted ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 || !effectstest.IsTapped(g, host) {
		t.Fatalf("life %d (want %d) tapped %v", opp.Life, life-1, effectstest.IsTapped(g, host))
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("sacrifice the Aura: %v", err)
	}
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
}

func TestOldB774CoilsDestroysNonWallBlockersOfTheEnchantedAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 3)
	aura := effectstest.Enchant(t, g, "Coils of the Medusa", oldb774Coils, host)
	oldb774Size(t, g, host, 3, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, host)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range []uuid.UUID{bear, wall} {
		if err := g.DeclareBlocker(b, host); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("sacrifice the Aura: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a non-Wall blocker survived")
	}
	if !g.Battlefield.Contains(wall) || !g.Battlefield.Contains(bystander) {
		t.Error("a Wall or a non-blocker was destroyed")
	}
	oldb774Size(t, g, host, 2, 3)
}

func TestOldB774EmpyrialArmorCountsTheAurasControllersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "Empyrial Armor", oldb774Empyrial, theirs)
	n := len(me.Hand.Cards)
	oldb774Size(t, g, theirs, 2+n, 2+n)
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	oldb774Size(t, g, theirs, 3+n, 3+n)
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatal(err)
	}
	oldb774Size(t, g, theirs, 3+n, 3+n)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		me.Hand.Cards[0].InstanceID); err != nil {
		t.Fatal(err)
	}
	oldb774Size(t, g, theirs, 2+n, 2+n)
}

func TestOldB774BetrothedOfFireBothAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	aura := effectstest.Enchant(t, g, "Betrothed of Fire", oldb774Betrothed, host)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, tapped).Tapped = true })
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{tapped}}); err == nil {
		t.Fatal("sacrificed a tapped creature to the first ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oldb774Size(t, g, host, 4, 2)
	if g.Battlefield.Contains(fodder) {
		t.Error("the untapped creature was not sacrificed")
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{other}}); err == nil {
		t.Fatal("sacrificed a creature that is not the enchanted one")
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{host}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Error("the enchanted creature was not sacrificed")
	}
	oldb774Size(t, g, other, 3, 1)
	oldb774Size(t, g, tapped, 3, 1)
	oldb774Size(t, g, theirs, 1, 1)
}

func TestOldB774BriarShieldSacrificesForPlusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	aura := effectstest.Enchant(t, g, "Briar Shield", oldb774BriarShld, host)
	oldb774Size(t, g, host, 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed")
	}
	oldb774Size(t, g, host, 5, 5)
	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
	oldb774Size(t, g, host, 2, 2)
}

func TestOldB774NaturesKissExilesTheTopCardForPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	aura := effectstest.Enchant(t, g, "Nature's Kiss", oldb774NaturesKis, host)
	effectstest.AdvanceToMain(t, g)
	ids := oldb774Graveyard(me, "Creature — Bear", "Sorcery")
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: ids[:1]}); err == nil {
		t.Fatal("a card below the top paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: ids[1:]}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[1]) || !me.Graveyard.Contains(ids[0]) {
		t.Error("the top card was not the one exiled")
	}
	oldb774Size(t, g, host, 3, 3)
}

func TestOldB774ApathyKeepsItTappedUnlessItsControllerDiscardsAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	aura := effectstest.Enchant(t, g, "Apathy", oldb774Apathy, bear)
	for i := 0; i < 2; i++ {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TriggerOnStack(g, aura) != nil {
		t.Error("Apathy triggered in another player's upkeep")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.IsTapped(g, bear) {
		t.Fatal("the enchanted creature untapped in its controller's untap step")
	}
	if effectstest.TriggerOnStack(g, aura) == nil {
		t.Fatal("Apathy's upkeep trigger is not on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.UntapStaticPendingPay(t, g, me.ID)
	hand := len(me.Hand.Cards)
	if err := g.ResolvePayUnless(choice.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if !effectstest.IsTapped(g, bear) || len(me.Hand.Cards) != hand {
		t.Error("declining untapped the creature or discarded a card")
	}
}

func TestOldB774ApathyDiscardUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.Enchant(t, g, "Apathy", oldb774Apathy, bear)
	for i := 0; i < 2; i++ {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.UntapStaticPendingPay(t, g, me.ID)
	hand := len(me.Hand.Cards)
	if err := g.ResolvePayUnless(choice.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if effectstest.IsTapped(g, bear) {
		t.Error("discarding a card did not untap the creature")
	}
	if len(me.Hand.Cards) != hand-1 {
		t.Errorf("hand went from %d to %d, want one card discarded", hand, len(me.Hand.Cards))
	}
}
