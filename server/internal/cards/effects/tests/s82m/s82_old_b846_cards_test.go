package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b846_cards_test.go — Ice Age, batch 846 (#5636): each test plays the
// real catalog entry through the real engine paths.

const (
	b846RunedArch    = "32717625-46c6-45c5-ad0a-fb6a91f9878c"
	b846KjeldoranDed = "61e00946-7df7-4298-91b2-c656169e6601"
	b846CurseMarit   = "6fde4790-4186-4786-aa1a-b6a53bfadc1c"
	b846CallToArms   = "72eaaa73-3263-46e9-aa86-5862d3aed20a"
	b846LostOrder    = "75f047dc-4a51-494d-8765-fa7502d1fd06"
	b846Seraph       = "a1c50bd1-120d-44ac-af20-13a90fc991e6"
	b846JestersMask  = "bcc797b9-d2e5-4fde-8d80-c890c871f344"
	b846Earthlink    = "c6e4a5c6-3c9b-49d5-8288-1be46be657dc"
)

func TestS82OldB846CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b846RunedArch: "Runed Arch", b846KjeldoranDed: "Kjeldoran Dead", b846CurseMarit: "Curse of Marit Lage",
		b846CallToArms: "Call to Arms", b846LostOrder: "Lost Order of Jarkeld", b846Seraph: "Seraph",
		b846JestersMask: "Jester's Mask", b846Earthlink: "Earthlink",
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
}

// b846Enter puts the card onto the battlefield from the hand, so its entry
// replacements and enters triggers apply.
func b846Enter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.OracleID, c.Power, c.Toughness = typeLine, oracle, power, toughness
	owner.Hand.PushTop(c)
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		id, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, game.HandEntryOptions{Controller: owner.ID})
		if err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, id).SummonedThisTurn = false })
	return id
}

func b846Tapped(g *game.Game, id uuid.UUID) bool {
	c := effectstest.FindBattlefieldCardByID(g, id)
	return c != nil && c.Tapped
}

func TestS82OldB846RunedArchEntersTappedAndMakesXSmallCreaturesUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arch := b846Enter(t, g, me, "Runed Arch", "Artifact", b846RunedArch, 0, 0)
	if !b846Tapped(g, arch) {
		t.Fatal("the Arch did not enter tapped")
	}
	small := effectstest.B16Creature(g, me.ID, "Small", "Creature — Bear", 2, 2, "G")
	other := effectstest.B16Creature(g, me.ID, "Other", "Creature — Bear", 1, 1, "G")
	big := effectstest.B16Creature(g, me.ID, "Big", "Creature — Bear", 3, 3, "G")
	blocker := effectstest.B16Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardByID(g, arch).Tapped = false
		for _, id := range []uuid.UUID{small, other, big} {
			effectstest.FindBattlefieldCardByID(g, id).SummonedThisTurn = false
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	params := game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(small, big)}
	if err := g.ActivateCatalogAbility(me.ID, arch, 0, params); err == nil {
		t.Fatal("a power 3 creature was a legal target")
	}
	params.Targets = effectstest.CardRefs(small, other)
	if err := g.ActivateCatalogAbility(me.ID, arch, 0, params); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(arch) {
		t.Error("the Arch was not sacrificed")
	}
	effectstest.DeclareAttack(t, g, opp.ID, small, other, big)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, small); err == nil {
		t.Error("an unblockable creature was blocked")
	}
	if err := g.DeclareBlocker(blocker, other); err == nil {
		t.Error("the second unblockable creature was blocked")
	}
	if err := g.DeclareBlocker(blocker, big); err != nil {
		t.Errorf("the creature that was not targeted could not be blocked: %v", err)
	}
}

// b846CastDead casts a real 3/1 Kjeldoran Dead from the hand.
func b846CastDead(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	c := game.NewCard("Kjeldoran Dead", me.ID)
	c.TypeLine, c.OracleID, c.ManaCost, c.Power, c.Toughness, c.Colors = "Creature — Skeleton", b846KjeldoranDed, "{B}", 3, 1, []string{"B"}
	me.Hand.PushTop(c)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B")
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	return c.InstanceID
}

func TestS82OldB846KjeldoranDeadSacrificesACreatureAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.B16Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2, "G")
	dead := b846CastDead(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the chosen creature was not sacrificed")
	}
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the Dead was sacrificed too")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.New(), dead, 3); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(dead) || !b846Tapped(g, dead) {
		t.Errorf("the shielded Dead did not regenerate (alive %v, tapped %v)", g.Battlefield.Contains(dead), b846Tapped(g, dead))
	}
}

func TestS82OldB846KjeldoranDeadMayHaveToSacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := b846CastDead(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		effectstest.AnswerSacrifice(t, g, me.ID, dead)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(dead) {
		t.Error("the lone Dead was not sacrificed to its own trigger")
	}
}

func TestS82OldB846CurseOfMaritLageTapsAndHoldsEveryIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.BlbB1Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Island", "Basic Land — Island", "", 0, 0)
	mountain := effectstest.BlbB1Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
	b846Enter(t, g, me, "Curse of Marit Lage", "Enchantment", b846CurseMarit, 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !b846Tapped(g, mine) || !b846Tapped(g, theirs) {
		t.Fatal("not every Island was tapped")
	}
	if b846Tapped(g, mountain) {
		t.Error("a Mountain was tapped")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, mountain).Tapped = true })
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if !b846Tapped(g, theirs) {
		t.Error("their Island untapped")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if !b846Tapped(g, mine) {
		t.Error("my Island untapped")
	}
	if b846Tapped(g, mountain) {
		t.Error("my Mountain stayed tapped")
	}
}

func b846Chosen(g *game.Game, id uuid.UUID, color string, player uuid.UUID) {
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardByID(g, id)
		c.ChosenColor, c.ChosenPlayer = color, player
	})
	g.BumpLayerVersionForTest()
}

func TestS82OldB846CallToArmsPumpsWhiteWhileTheColorLeadsAndSacrificesWhenItDoesNot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	call := effectstest.B12Push(g, me.ID, "Call to Arms", "Enchantment", b846CallToArms, 0, 0)
	white := effectstest.B16Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2, "W")
	green := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B16Creature(g, opp.ID, "Bear 2", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	// Unchosen: nothing happens and nothing is sacrificed.
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(call) {
		t.Fatal("Call to Arms was sacrificed before its choices were made")
	}
	b846Chosen(g, call, "G", opp.ID)
	g.RunStateChecksForTest()
	if p, tough := effectstest.PtOf(t, g, white); p != 3 || tough != 3 {
		t.Errorf("white creature is %d/%d, want 3/3 while green leads", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, green); p != 2 {
		t.Errorf("the green creature got the bonus: power %d", p)
	}
	if !g.Battlefield.Contains(call) {
		t.Fatal("Call to Arms left while green led")
	}
	// Tie it: a second colour with as many nontoken permanents.
	effectstest.B16Creature(g, opp.ID, "Red 1", "Creature — Goblin", 1, 1, "R")
	effectstest.B16Creature(g, opp.ID, "Red 2", "Creature — Goblin", 1, 1, "R")
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(call) {
		t.Error("Call to Arms stayed with the chosen color tied for most common")
	}
}

func TestS82OldB846LostOrderOfJarkeldCountsTheChosenPlayersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	order := effectstest.B12Push(g, me.ID, "Lost Order of Jarkeld", "Creature — Human Knight", b846LostOrder, 0, 0)
	effectstest.AdvanceToMain(t, g)
	b846Chosen(g, order, "", opp.ID)
	if p, tough := effectstest.PtOf(t, g, order); p != 1 || tough != 1 {
		t.Fatalf("against an empty board it is %d/%d, want 1/1", p, tough)
	}
	effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B16Creature(g, opp.ID, "Bear 2", "Creature — Bear", 2, 2, "G")
	effectstest.B16Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2, "G")
	if p, tough := effectstest.PtOf(t, g, order); p != 3 || tough != 3 {
		t.Errorf("against two creatures it is %d/%d, want 3/3", p, tough)
	}
}

func TestS82OldB846SeraphBringsBackWhatItKilledAndLosesItWithControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seraph := effectstest.B12Push(g, me.ID, "Seraph", "Creature — Angel", b846Seraph, 4, 4)
	victim := effectstest.B16Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2, "G")
	bystander := effectstest.B16Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(seraph, victim, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.KillCreature(t, g, opp.ID, bystander)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the victim came back before the end step")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardByID(g, victim)
	if c == nil || c.Controller != me.ID {
		t.Fatal("the creature Seraph damaged did not return under my control")
	}
	if g.Battlefield.Contains(bystander) {
		t.Error("a creature Seraph never damaged came back")
	}
	// Losing Seraph sacrifices what it brought back.
	effectstest.KillCreature(t, g, me.ID, seraph)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(victim) {
		t.Error("the creature stayed after Seraph left")
	}
}

func TestS82OldB846JestersMaskPutsTheHandOnTopAndSearchesThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mask := b846Enter(t, g, me, "Jester's Mask", "Artifact", b846JestersMask, 0, 0)
	if !b846Tapped(g, mask) {
		t.Fatal("the Mask did not enter tapped")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, mask).Tapped = false })
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	var hand []uuid.UUID
	for _, n := range []string{"A", "B", "C"} {
		c := game.NewCard("Hand "+n, opp.ID)
		c.TypeLine = "Sorcery"
		opp.Hand.PushTop(c)
		hand = append(hand, c.InstanceID)
	}
	var lib []uuid.UUID
	for _, n := range []string{"X", "Y", "Z", "W"} {
		c := game.NewCard("Lib "+n, opp.ID)
		c.TypeLine = "Sorcery"
		opp.Library.PushTop(c)
		lib = append(lib, c.InstanceID)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	myHand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, mask, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PutInLibraryChoiceFor(g, opp.ID)
	if ch == nil {
		t.Fatal("the opponent was not asked in what order to put the hand on top")
	}
	if err := g.ResolvePutInLibrary(ch.ID, opp.ID, nil, hand); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 0 {
		t.Fatalf("the opponent still holds %d cards after putting the hand on top", opp.Hand.Size())
	}
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("the Mask's controller was not asked to search")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, hand[0], lib[0], lib[1])
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 3 {
		t.Errorf("the opponent's hand is %d cards, want 3", opp.Hand.Size())
	}
	if me.Hand.Size() != myHand {
		t.Errorf("my hand went from %d to %d cards: the found cards went to the wrong hand", myHand, me.Hand.Size())
	}
}

func TestS82OldB846EarthlinkTaxesUpkeepAndCostsEachDeathALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	link := effectstest.B12Push(g, me.ID, "Earthlink", "Enchantment", b846Earthlink, 0, 0)
	land1 := effectstest.BlbB1Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	land2 := effectstest.BlbB1Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	mine := effectstest.BlbB1Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
	victim := effectstest.B16Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, land1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land1) || !g.Battlefield.Contains(land2) {
		t.Error("the dead creature's controller did not sacrifice exactly the land they chose")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("Earthlink's controller lost a land to the opponent's creature")
	}
	// Upkeep: decline to pay and it is sacrificed.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(link) {
		t.Error("Earthlink stayed after its tax was declined")
	}
}
