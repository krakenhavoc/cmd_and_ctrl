package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b600_cards_test.go — Betrayers of Kamigawa, card draw, library,
// Equipment and bushido, batch 600 (#5602, S82 OLD). Every test plays the real
// catalog entry through the engine.

const (
	oldB600Lifespinner = "0fb26acd-891d-471d-98d8-73c6f940052b"
	oldB600Tallowisp   = "25187c05-ed6a-4690-b92b-89c532fdb780"
	oldB600Heed        = "49330aaf-c005-4e07-ae3a-2579bea3c326"
	oldB600Tomorrow    = "51d517c9-2812-44ce-ab4d-e5422b5ecf6c"
	oldB600Warclub     = "11b969f2-ff68-4c9e-b3f3-839736084303"
	oldB600Powder      = "75e13f7f-aded-405c-97ce-124fa0cb1929"
	oldB600NekoTe      = "a1d2944e-6668-4157-aac5-54572759c05c"
	oldB600Shuriken    = "c6e92e21-d3e7-4982-b9f3-af56d4014c07"
	oldB600Isao        = "15ec1bfc-b0a0-4058-87bb-8dbaf17b7f3b"
	oldB600Kentaro     = "bfaf20a7-a11c-4f57-8eb6-b03b5223704d"
	oldB600Cavalry     = "c5577514-d2af-4904-9f94-b472228c6e3d"
	oldB600Cliffrider  = "ce8fc1bb-5dd0-41b1-909a-bbfe6b165158"
)

func TestOldB600CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB600Lifespinner: "Lifespinner", oldB600Tallowisp: "Tallowisp", oldB600Heed: "Heed the Mists",
		oldB600Tomorrow: "Tomorrow, Azami's Familiar", oldB600Warclub: "Ronin Warclub", oldB600Powder: "Blinding Powder",
		oldB600NekoTe: "Neko-Te", oldB600Shuriken: "Shuriken", oldB600Isao: "Isao, Enlightened Bushi",
		oldB600Kentaro: "Kentaro, the Smiling Cat", oldB600Cavalry: "Takeno's Cavalry", oldB600Cliffrider: "Ronin Cliffrider",
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

func oldB600Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB600Attach(t *testing.T, g *game.Game, attachment, host uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		err = g.AttachForEffect(attachment, game.TargetRef{Kind: game.TargetCard, ID: host})
	})
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
}

func TestOldB600BushidoBodiesGrowWhenBlocked(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough, n        int
	}{
		{"Isao, Enlightened Bushi", "Legendary Creature — Human Samurai", oldB600Isao, 2, 1, 2},
		{"Kentaro, the Smiling Cat", "Legendary Creature — Human Samurai", oldB600Kentaro, 2, 1, 1},
		{"Takeno's Cavalry", "Creature — Human Samurai Archer", oldB600Cavalry, 1, 1, 1},
		{"Ronin Cliffrider", "Creature — Human Samurai", oldB600Cliffrider, 2, 2, 1},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: id, CardID: id, Target: id})
		})
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != tc.power+tc.n || tough != tc.tough+tc.n {
			t.Errorf("%s is %d/%d after becoming blocked, want %d/%d", tc.name, p, tough, tc.power+tc.n, tc.tough+tc.n)
		}
	}
}

func TestOldB600IsaoCantBeCounteredAndRegeneratesSamurai(t *testing.T) {
	spec, _ := effects.Lookup(oldB600Isao)
	if !spec.CantBeCountered {
		t.Error("Isao's spell can be countered")
	}
	g, me, opp := oldB600Table(t)
	isao := effectstest.PushCatalogPermanent(g, me.ID, "Isao, Enlightened Bushi", "Legendary Creature — Human Samurai", oldB600Isao, false)
	samurai := effectstest.B12Creature(g, opp.ID, "Their Samurai", "Creature — Human Samurai", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.RavB2Activate(t, g, me.ID, isao, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}, "C", "C"); err == nil {
		t.Fatal("a non-Samurai was a legal target")
	}
	effectstest.B06AddMana(me, "C", "C")
	effectstest.RavB2MustActivate(t, g, me.ID, isao, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(samurai)})
	effectstest.DealDamageTo(t, g, bear, samurai, 5)
	if !g.Battlefield.Contains(samurai) {
		t.Error("the regenerated Samurai died")
	}
	if !effectstest.IsTapped(g, samurai) {
		t.Error("the regeneration did not tap the Samurai")
	}
}

func TestOldB600KentaroLetsYouPayItsManaValueInGenericForSamuraiOnly(t *testing.T) {
	g, me, _ := oldB600Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Kentaro, the Smiling Cat", "Legendary Creature — Human Samurai", oldB600Kentaro, false)
	samurai := effectstest.HandCardFull(me, "Test Samurai", "Creature — Human Samurai", "{3}{R}", "", []string{"R"})
	bear := effectstest.HandCardFull(me, "Test Bear", "Creature — Bear", "{3}{R}", "", []string{"R"})
	offer := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, samurai, game.ZoneHand), effects.GrantedAltCostXForSamuraiSpells)
	if offer == nil {
		t.Fatal("a Samurai spell is not offered Kentaro's cost")
	}
	if offer.ManaCost != "{4}" {
		t.Errorf("the offered cost is %q, want {4} (its mana value, generic)", offer.ManaCost)
	}
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, bear, game.ZoneHand), effects.GrantedAltCostXForSamuraiSpells) != nil {
		t.Error("a non-Samurai spell is offered Kentaro's cost")
	}
	// Four colourless mana pay for the {3}{R} Samurai with the alternative cost
	// but not with its printed cost.
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, samurai, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the printed {3}{R} was paid with colourless mana")
	}
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, samurai, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostXForSamuraiSpells}); err != nil {
		t.Fatalf("cast with Kentaro's cost: %v", err)
	}
}

func TestOldB600TakenosCavalryPingsOnlyAnAttackingOrBlockingSpirit(t *testing.T) {
	g, me, opp := oldB600Table(t)
	cav := effectstest.PushCatalogPermanent(g, me.ID, "Takeno's Cavalry", "Creature — Human Samurai Archer", oldB600Cavalry, false)
	idle := effectstest.B12Creature(g, opp.ID, "Idle Spirit", "Creature — Spirit", 2, 3)
	if err := g.ActivateCatalogAbility(me.ID, cav, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("a Spirit that is neither attacking nor blocking was a legal target")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, idle).BlockingTarget = me.ID })
	if err := g.ActivateCatalogAbility(me.ID, cav, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(idle)}); err != nil {
		t.Fatalf("activate against a blocking Spirit: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DamageMarkedOn(g, idle); n != 1 {
		t.Errorf("the Spirit has %d damage, want 1", n)
	}
}

func TestOldB600RoninCliffriderMayDamageEachDefendingCreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rider := effectstest.B12Push(g, me.ID, "Ronin Cliffrider", "Creature — Human Samurai", oldB600Cliffrider, 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
		effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
		if err := g.DeclareAttacker(rider, opp.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		want := 0
		if yes {
			want = 1
		}
		if n := effectstest.DamageMarkedOn(g, theirs); n != want {
			t.Errorf("yes=%v: their creature has %d damage, want %d", yes, n, want)
		}
		if n := effectstest.DamageMarkedOn(g, mine); n != 0 {
			t.Errorf("yes=%v: my own creature was damaged", yes)
		}
	}
}

func TestOldB600LifespinnerSacrificesThreeSpiritsToFetchALegendarySpirit(t *testing.T) {
	g, me, _ := oldB600Table(t)
	spinner := effectstest.PushCatalogPermanent(g, me.ID, "Lifespinner", "Creature — Spirit", oldB600Lifespinner, false)
	a := effectstest.B12Creature(g, me.ID, "Spirit A", "Creature — Spirit", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Spirit B", "Creature — Spirit", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "A Bear", "Creature — Bear", 1, 1)
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Plain Spirit", "Creature — Spirit"),
		effectstest.SearchTestLand("Legendary Bear", "Legendary Creature — Bear"),
		effectstest.SearchTestLand("Legendary Spirit", "Legendary Creature — Spirit"),
	)
	if err := g.ActivateCatalogAbility(me.ID, spinner, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{a, b, bear}}); err == nil {
		t.Fatal("a non-Spirit was sacrificed for a Spirit")
	}
	if err := g.ActivateCatalogAbility(me.ID, spinner, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{a, b, spinner}}); err != nil {
		t.Fatalf("activate sacrificing the Lifespinner among the three: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search offered")
	}
	if len(c.SearchCards) != 1 {
		t.Errorf("the search offers %d cards, want only the legendary Spirit", len(c.SearchCards))
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[3])
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[3]) {
		t.Error("the legendary Spirit did not enter the battlefield")
	}
	for _, id := range []uuid.UUID{a, b, spinner} {
		if g.Battlefield.Contains(id) {
			t.Error("a Spirit paid as a cost is still on the battlefield")
		}
	}
}

func TestOldB600TallowispFetchesAnAuraForASpiritOrArcaneSpell(t *testing.T) {
	g, me, _ := oldB600Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Tallowisp", "Creature — Spirit", oldB600Tallowisp, false)
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Test Aura", "Enchantment — Aura"),
	)
	effectstest.CastCatalogSpell(t, g, "Bear Spell", "Creature — Bear", "", nil)
	if effectstest.OwedMays(g, me.ID) != 0 {
		t.Fatal("a Bear spell triggered Tallowisp")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "An Arcane Spell", "Instant — Arcane", "", nil)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the Aura card did not reach the hand")
	}
}

func TestOldB600HeedTheMistsDrawsTheMilledCardsManaValue(t *testing.T) {
	g, me, _ := oldB600Table(t)
	deep := effectstest.SeedLibraryTop(me, "Deep One", "Creature — Bear", "{1}")
	deep2 := effectstest.SeedLibraryTop(me, "Deep Two", "Creature — Bear", "{1}")
	deep3 := effectstest.SeedLibraryTop(me, "Deep Three", "Creature — Bear", "{1}")
	milled := effectstest.SeedLibraryTop(me, "Milled Card", "Creature — Bear", "{2}{U}")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Heed the Mists", "Sorcery — Arcane", oldB600Heed, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !me.Graveyard.Contains(milled) {
		t.Error("the top card was not milled")
	}
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d, want %d (three cards for a mana value 3)", me.Hand.Size(), hand+3)
	}
	for _, id := range []uuid.UUID{deep, deep2, deep3} {
		if !me.Hand.Contains(id) {
			t.Error("a drawn card is not in hand")
		}
	}
}

func TestOldB600TomorrowReplacesEachDrawWithALookAtThree(t *testing.T) {
	g, me, _ := oldB600Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Tomorrow, Azami's Familiar", "Legendary Creature — Spirit", oldB600Tomorrow, false)
	bottom := effectstest.SeedLibraryTop(me, "Bottom Marker", "Creature — Bear", "{1}")
	c3 := effectstest.SeedLibraryTop(me, "Third", "Creature — Bear", "{1}")
	c2 := effectstest.SeedLibraryTop(me, "Second", "Creature — Bear", "{1}")
	c1 := effectstest.SeedLibraryTop(me, "First", "Creature — Bear", "{1}")
	hand, drawn := me.Hand.Size(), len(g.DrawnThisTurn[me.ID])
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Tomorrow did not ask which card to take")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("prompt over %d cards, %d..%d; want exactly one of three", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if me.Hand.Size() != hand {
		t.Fatal("nothing moves before the answer")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, c2)
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 2 {
		t.Fatalf("want a bottom placement over the two declined cards, got %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{c3, c1}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if !me.Hand.Contains(c2) || me.Hand.Size() != hand+1 {
		t.Error("the chosen card is not the one card added to the hand")
	}
	if got := effectstest.LibraryBottomIDs(me, 2); !effectstest.SameIDs(got, []uuid.UUID{c3, c1}) {
		t.Errorf("the bottom two are %v, want %v in the order chosen", got, []uuid.UUID{c3, c1})
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != bottom {
		t.Error("the card under the look is not the new top of the library")
	}
	if len(g.DrawnThisTurn[me.ID]) != drawn {
		t.Error("the replaced draw drew a card")
	}
}

func TestOldB600RoninWarclubAttachesToEachCreatureThatEnters(t *testing.T) {
	g, me, _ := oldB600Table(t)
	club := effectstest.SeedEquipment(g, me.ID, "Ronin Warclub", oldB600Warclub)
	first := effectstest.B12Creature(g, me.ID, "First Bear", "Creature — Bear", 2, 2)
	oldB600Attach(t, g, club, first)
	if p, tough := effectstest.EffectivePTOf(t, g, first); p != 4 || tough != 3 {
		t.Errorf("the equipped creature is %d/%d, want 4/3", p, tough)
	}
	effectstest.CastCatalogSpell(t, g, "Newcomer", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	var newcomer uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Newcomer" {
			newcomer = c.InstanceID
		}
	}
	if newcomer == uuid.Nil {
		t.Fatal("the creature did not enter")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, club, newcomer) || effectstest.AttachedTo(t, g, club, first) {
		t.Error("the Warclub did not move to the creature that entered")
	}
	// An opponent's creature entering under their control does nothing.
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	if effectstest.AttachedTo(t, g, club, theirs) {
		t.Error("the Warclub attached to an opponent's creature")
	}
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, club, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(first)}); err == nil {
		t.Fatal("equipped for {4}")
	}
}

func TestOldB600BlindingPowderUnattachesToPreventCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)
	powder := effectstest.SeedEquipment(g, me.ID, "Blinding Powder", oldB600Powder)
	oldB600Attach(t, g, powder, bear)
	blocker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 3, Toughness: 9, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, bear); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.AttachedTo(t, g, powder, bear) {
		t.Error("the Powder was not unattached as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, bear) == nil || effectstest.DamageMarkedOn(g, bear) != 0 {
		t.Fatalf("the creature took combat damage (marked %d)", effectstest.DamageMarkedOn(g, bear))
	}
	effectstest.Pr6Damage(t, g, blocker, bear, 1)
	if n := effectstest.DamageMarkedOn(g, bear); n != 1 {
		t.Errorf("noncombat damage was prevented too: %d marked, want 1", n)
	}
}

func TestOldB600NekoTeTapsAndHoldsDamagedCreaturesAndDrainsPlayers(t *testing.T) {
	g, me, opp := oldB600Table(t)
	bear := effectstest.SeedBear(g, me.ID)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	neko := effectstest.SeedEquipment(g, me.ID, "Neko-Te", oldB600NekoTe)
	oldB600Attach(t, g, neko, bear)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 5)
	// A creature the equipped one did not damage is untouched.
	effectstest.DealDamageTo(t, g, other, victim, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, victim) {
		t.Fatal("damage from an unequipped creature tapped it")
	}
	effectstest.DealDamageTo(t, g, bear, victim, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, victim) {
		t.Fatal("the damaged creature was not tapped")
	}
	// It stays held through its controller's untap step while the Neko-Te remains.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if !effectstest.IsTapped(g, victim) {
		t.Error("the held creature untapped during its controller's untap step")
	}
	// The player drain.
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bear, opp.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2-1 {
		t.Errorf("life %d, want %d (2 damage and 1 life lost)", opp.Life, life-3)
	}
}

func TestOldB600ShurikenDamagesAndHandsItToTheTargetsControllerUnlessFromANinja(t *testing.T) {
	for _, ninja := range []bool{false, true} {
		g, me, opp := oldB600Table(t)
		typeLine := "Creature — Bear"
		if ninja {
			typeLine = "Creature — Human Ninja"
		}
		thrower := effectstest.B12Creature(g, me.ID, "Thrower", typeLine, 2, 2)
		shuriken := effectstest.SeedEquipment(g, me.ID, "Shuriken", oldB600Shuriken)
		oldB600Attach(t, g, shuriken, thrower)
		victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 5)
		if err := g.ActivateCatalogAbility(me.ID, thrower, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err != nil {
			t.Fatalf("ninja=%v: activate: %v", ninja, err)
		}
		if effectstest.AttachedTo(t, g, shuriken, thrower) || !effectstest.IsTapped(g, thrower) {
			t.Errorf("ninja=%v: the cost was not paid (unattach and tap)", ninja)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.DamageMarkedOn(g, victim); n != 2 {
			t.Errorf("ninja=%v: the target has %d damage, want 2", ninja, n)
		}
		var controller uuid.UUID
		g.WithWriteLock(func() { controller = effectstest.FindBattlefieldCardForTest(g, shuriken).Controller })
		want := opp.ID
		if ninja {
			want = me.ID
		}
		if controller != want {
			t.Errorf("ninja=%v: the Shuriken is controlled by the wrong player", ninja)
		}
	}
}
