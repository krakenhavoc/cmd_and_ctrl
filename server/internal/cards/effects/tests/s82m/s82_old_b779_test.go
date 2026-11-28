package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b779_test.go — Weatherlight, batch 779 (#5629): other creatures,
// instants and sorceries, removal. Each test plays the real catalog entry.

const (
	oldB779Ghoul     = "1610875f-48da-490c-a371-c439dc36f533"
	oldB779Entrancer = "50c8c7e2-1b45-44b2-a50b-3b514d4c4c42"
	oldB779Maraxus   = "af4b67c5-53dc-4e4d-9ec4-fa51cc3cd49a"
	oldB779Peace     = "e82ce5c6-1e42-4e87-b696-67aa3822ec9c"
	oldB779Memories  = "2ac9d397-10ca-43ce-bff9-45d631553486"
	oldB779Debt      = "7a5cfc97-89cf-43bc-8e48-f8bf7a0c5fcd"
	oldB779Justice   = "d8e2dea9-0e96-4ce4-a070-9893f44cf0cc"
	oldB779Tariff    = "e96b0560-1d91-48df-b42f-3b7d9c6d94bd"
	oldB779Grove     = "2f2a3588-b9a6-4c1a-89f8-b11339851b8f"
	oldB779Paradigm  = "7179c2c5-c59c-4793-a3b5-28d26d755157"
	oldB779Heat      = "a8127c04-f67f-4780-86bf-bb24480ec51a"
	oldB779Grenadier = "be83a518-cfab-46b1-9fe9-7e3d54311c37"

	oldB779Bear = "Creature — Bear"
)

func TestOldB779EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB779Ghoul: "Barrow Ghoul", oldB779Entrancer: "Tolarian Entrancer", oldB779Maraxus: "Maraxus of Keld",
		oldB779Peace: "Peacekeeper", oldB779Memories: "Agonizing Memories", oldB779Debt: "Debt of Loyalty",
		oldB779Justice: "Urborg Justice", oldB779Tariff: "Tariff", oldB779Grove: "Tranquil Grove",
		oldB779Paradigm: "Paradigm Shift", oldB779Heat: "Heat Stroke", oldB779Grenadier: "Goblin Grenadiers",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v; want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB779InZone(z *game.Zone, id uuid.UUID) bool {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestOldB779BarrowGhoulExilesTheTopCreatureCardOrIsSacrificedAtUpkeep(t *testing.T) {
	t.Run("pays with the top creature card", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		ghoul := effectstest.B12Push(g, me.ID, "Barrow Ghoul", "Creature — Zombie", oldB779Ghoul, 4, 4)
		older := effectstest.PushGraveyardCardTyped(me, "Older Bear", oldB779Bear)
		top := effectstest.PushGraveyardCardTyped(me, "Top Bear", oldB779Bear)
		spell := effectstest.PushGraveyardCardTyped(me, "Spell on top", "Sorcery")
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.UpkeepPromptFor(t, g, me.ID)
		if prompt == nil {
			t.Fatal("no upkeep prompt")
		}
		if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{older}); err == nil {
			t.Error("a creature card below the top one paid")
		}
		if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{top}); err != nil {
			t.Fatalf("ResolvePayUnlessWithCards: %v", err)
		}
		if !g.Battlefield.Contains(ghoul) {
			t.Error("the Ghoul was sacrificed although its payment was made")
		}
		if !effectstest.InExile(g, top) || effectstest.InExile(g, older) || effectstest.InExile(g, spell) {
			t.Errorf("the wrong card paid: top exiled %v, older %v, spell %v", effectstest.InExile(g, top), effectstest.InExile(g, older), effectstest.InExile(g, spell))
		}
	})
	t.Run("declines", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		ghoul := effectstest.B12Push(g, me.ID, "Barrow Ghoul", "Creature — Zombie", oldB779Ghoul, 4, 4)
		top := effectstest.PushGraveyardCardTyped(me, "Top Bear", oldB779Bear)
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, false)
		if g.Battlefield.Contains(ghoul) {
			t.Error("the Ghoul stayed although nothing was exiled")
		}
		if effectstest.InExile(g, top) {
			t.Error("a card was exiled after declining")
		}
	})
	t.Run("no creature card in the graveyard", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		ghoul := effectstest.B12Push(g, me.ID, "Barrow Ghoul", "Creature — Zombie", oldB779Ghoul, 4, 4)
		effectstest.PushGraveyardCardTyped(me, "Spell", "Sorcery")
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.HasPayUnlessFor(g, me.ID) {
			effectstest.AnswerPayUnless(t, g, me.ID, true)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(ghoul) {
			t.Error("the Ghoul stayed with no creature card to exile")
		}
	})
}

func TestOldB779TolarianEntrancerStealsEachBlockerAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ent := effectstest.B12Push(g, me.ID, "Tolarian Entrancer", "Creature — Human Wizard", oldB779Entrancer, 1, 1)
	wall := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4})
	other := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Other", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.DeclareAttack(t, g, opp.ID, ent)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, ent); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := controllerOfCard(t, g, wall); got != opp.ID {
		t.Fatalf("the blocker changed hands before end of combat: %v", got)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := controllerOfCard(t, g, wall); got != me.ID {
		t.Errorf("the blocker is controlled by %v, want the Entrancer's controller", got)
	}
	if got := controllerOfCard(t, g, other); got != opp.ID {
		t.Error("a creature that did not block changed hands")
	}
	effectstest.NextTurnOfMine(t, g)
	if got := controllerOfCard(t, g, wall); got != me.ID {
		t.Error("the control change did not last past the turn")
	}
}

func controllerOfCard(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	var who uuid.UUID
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				who = c.Controller
			}
		}
	})
	if who == uuid.Nil {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return who
}

func TestOldB779MaraxusCountsUntappedArtifactsCreaturesAndLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mx := effectstest.B12Push(g, me.ID, "Maraxus of Keld", "Legendary Creature — Human Warrior", oldB779Maraxus, 0, 0)
	if p, tough := effectstest.PtOf(t, g, mx); p != 1 || tough != 1 {
		t.Fatalf("alone, Maraxus is %d/%d, want 1/1 (itself)", p, tough)
	}
	land := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	rock := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Rock", "Artifact", "")
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB779Bear, 2, 2)
	effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Golem", "Artifact Creature — Golem", "")
	effectstest.PushLandFor(g, opp.ID, "Their Forest", "Basic Land — Forest")
	effectstest.B12Push(g, me.ID, "Enchant", "Enchantment", "", 0, 0)
	// Maraxus, land, rock, bear and a golem (one permanent, both types).
	if p, tough := effectstest.PtOf(t, g, mx); p != 5 || tough != 5 {
		t.Errorf("Maraxus is %d/%d, want 5/5", p, tough)
	}
	for _, id := range []uuid.UUID{land, rock, bear} {
		if err := g.TapTargetForEffect(id); err != nil {
			t.Fatalf("tap: %v", err)
		}
	}
	if p, tough := effectstest.PtOf(t, g, mx); p != 2 || tough != 2 {
		t.Errorf("with three tapped Maraxus is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB779PeacekeeperStopsEveryAttackAndTaxesItsUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Peacekeeper", "Creature — Human", oldB779Peace, 1, 1)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mine, opp.ID); err == nil {
		t.Error("a creature attacked under Peacekeeper")
	}
	if err := g.DeclareAttacker(theirs, me.ID); err == nil {
		t.Error("an opponent's creature attacked under Peacekeeper")
	}
}

func TestOldB779PeacekeeperIsSacrificedUnlessItsUpkeepIsPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		pk := effectstest.B12Push(g, me.ID, "Peacekeeper", "Creature — Human", oldB779Peace, 1, 1)
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.FloatMana(t, g, me, "{W}{W}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(pk); got != pay {
			t.Errorf("paid %v: Peacekeeper on the battlefield = %v", pay, got)
		}
	}
}

func TestOldB779TranquilGroveDestroysEveryOtherEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grove := effectstest.B12Push(g, me.ID, "Tranquil Grove", "Enchantment", oldB779Grove, 0, 0)
	mine := effectstest.B12Push(g, me.ID, "My Ward", "Enchantment", "", 0, 0)
	theirs := effectstest.B12Push(g, opp.ID, "Their Ward", "Enchantment", "", 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB779Bear, 2, 2)
	effectstest.FillPoolColored(me, "G", 3)
	effectstest.Pr7Activate(t, g, me.ID, grove, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("another enchantment survived")
	}
	if !g.Battlefield.Contains(grove) || !g.Battlefield.Contains(bear) {
		t.Error("Tranquil Grove or a nonenchantment was destroyed")
	}
}

func TestOldB779ParadigmShiftExilesTheLibraryThenShufflesTheGraveyardIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := effectstest.SeedLibrary(me, "L1", "L2", "L3", "L4")
	g1 := effectstest.PushGraveyardCardTyped(me, "Gone 1", "Sorcery")
	g2 := effectstest.PushGraveyardCardTyped(me, "Gone 2", oldB779Bear)
	effectstest.CastCatalogSpell(t, g, "Paradigm Shift", "Sorcery", oldB779Paradigm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range lib {
		if !effectstest.InExile(g, id) {
			t.Errorf("library card %s was not exiled", id)
		}
	}
	if !oldB779InZone(me.Library, g1) || !oldB779InZone(me.Library, g2) {
		t.Error("the graveyard cards are not in the library")
	}
	if len(me.Library.Cards) != 2+0 {
		t.Errorf("library has %d cards, want only the 2 shuffled in (Paradigm Shift itself stays in the graveyard)", len(me.Library.Cards))
	}
}

func TestOldB779HeatStrokeDestroysEveryCreatureThatBlockedOrWasBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Heat Stroke", "Enchantment", oldB779Heat, 0, 0)
	atk := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 1, 9)
	free := effectstest.PushVanillaCreature(g, me.ID, "Unblocked", 1, 9)
	idle := effectstest.PushVanillaCreature(g, me.ID, "Idle", 1, 9)
	wall := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 9})
	bystander := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Bystander", TypeLine: "Creature — Bear", Power: 0, Toughness: 9})
	effectstest.DeclareAttack(t, g, opp.ID, atk, free)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, atk); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{atk, wall} {
		if g.Battlefield.Contains(id) {
			t.Errorf("creature %s that blocked or was blocked survived", id)
		}
	}
	for _, id := range []uuid.UUID{free, idle, bystander} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("creature %s that neither blocked nor was blocked was destroyed", id)
		}
	}
}

func TestOldB779GoblinGrenadiersSacrificesItselfToDestroyACreatureAndALand(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		gren := effectstest.B12Push(g, me.ID, "Goblin Grenadiers", "Creature — Goblin", oldB779Grenadier, 2, 2)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB779Bear, 2, 2)
		land := effectstest.PushLandFor(g, opp.ID, "Their Forest", "Basic Land — Forest")
		effectstest.DeclareAttack(t, g, opp.ID, gren)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		effectstest.LockInBlocks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		oldB779PickTargets(t, g, me.ID, victim, land)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(gren); got == yes {
			t.Errorf("answered %v: the Grenadiers on the battlefield = %v", yes, got)
		}
		if got := g.Battlefield.Contains(victim); got == yes {
			t.Errorf("answered %v: the creature on the battlefield = %v", yes, got)
		}
		if got := g.Battlefield.Contains(land); got == yes {
			t.Errorf("answered %v: the land on the battlefield = %v", yes, got)
		}
	}
}

// oldB779PickTargets answers the open trigger-target prompts with the given
// cards, whichever of them each prompt offers.
func oldB779PickTargets(t *testing.T, g *game.Game, chooser uuid.UUID, cards ...uuid.UUID) {
	t.Helper()
	for i := 0; i < 4; i++ {
		p := effectstest.LatestPickTarget(g, chooser)
		if p == nil {
			return
		}
		picked := false
		for _, id := range cards {
			if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
				picked = true
				break
			}
		}
		if !picked {
			t.Fatalf("no card fits the open target prompt %s", p.ID)
		}
	}
}

func TestOldB779DebtOfLoyaltyRegeneratesAndStealsOnlyWhenTheShieldIsUsed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB779Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Debt of Loyalty", "Instant", oldB779Debt, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := controllerOfCard(t, g, bear); got != opp.ID {
		t.Fatalf("control changed before the creature was destroyed: %v", got)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the shield did not regenerate the creature")
	}
	if got := controllerOfCard(t, g, bear); got != me.ID {
		t.Errorf("the regenerated creature is controlled by %v, want the caster", got)
	}
}

func TestOldB779DebtOfLoyaltyTakesNothingIfNothingDestroysTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB779Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Debt of Loyalty", "Instant", oldB779Debt, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NextTurnOfMine(t, g)
	if got := controllerOfCard(t, g, bear); got != opp.ID {
		t.Errorf("the creature changed hands with no regeneration: %v", got)
	}
}

func TestOldB779UrborgJusticeMakesTheOpponentSacrificeOnePerCreatureYouLost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Mine A", oldB779Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Mine B", oldB779Bear, 2, 2)
	theirDead := effectstest.B12Creature(g, opp.ID, "Their Dead", oldB779Bear, 2, 2)
	x := effectstest.B12Creature(g, opp.ID, "X", oldB779Bear, 3, 3)
	y := effectstest.B12Creature(g, opp.ID, "Y", oldB779Bear, 3, 3)
	z := effectstest.B12Creature(g, opp.ID, "Z", oldB779Bear, 3, 3)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, theirDead); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.CastCatalogSpell(t, g, "Urborg Justice", "Instant", oldB779Justice, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, x)
	effectstest.AnswerSacrifice(t, g, opp.ID, y)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(x) || g.Battlefield.Contains(y) {
		t.Error("the opponent kept a creature they chose to sacrifice")
	}
	if !g.Battlefield.Contains(z) {
		t.Error("the opponent sacrificed a third creature although only two of yours died")
	}
}

func TestOldB779TariffMakesEachPlayerPayForTheirMostExpensiveCreatureOrSacrificeIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := effectstest.B12Push(g, me.ID, "Cheap", oldB779Bear, "", 1, 1)
	_ = cheap
	big := effectstest.B12Push(g, me.ID, "Big", oldB779Bear, "", 4, 4)
	theirBig := effectstest.B12Push(g, opp.ID, "Their Big", oldB779Bear, "", 4, 4)
	setCost := func(id uuid.UUID, cost string) {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].ManaCost = cost
				}
			}
		})
	}
	setCost(cheap, "{1}")
	setCost(big, "{3}{G}")
	setCost(theirBig, "{2}{U}")
	effectstest.CastCatalogSpell(t, g, "Tariff", "Sorcery", oldB779Tariff, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{G}{G}{G}{G}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(cheap) {
		t.Error("the caster paid for their creature but lost one")
	}
	if g.Battlefield.Contains(theirBig) {
		t.Error("the opponent declined to pay but kept their creature")
	}
}

func TestOldB779AgonizingMemoriesPutsTwoChosenCardsOnTopOfTheirLibraryInYourOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	other := g.Seats[2]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	keep := effectstest.HandCardFull(opp, "Keep", "Sorcery", "{1}", "", nil)
	a := effectstest.HandCardFull(opp, "A", "Sorcery", "{1}", "", nil)
	b := effectstest.HandCardFull(opp, "B", "Instant", "{1}", "", nil)
	effectstest.SeedLibrary(opp, "L1", "L2", "L3")
	effectstest.CastCatalogSpell(t, g, "Agonizing Memories", "Sorcery", oldB779Memories, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.OpenVariant(t, g)
	if c.Chooser != me.ID || c.Count != 2 {
		t.Fatalf("the caster should choose two cards: chooser %v count %d", c.Chooser, c.Count)
	}
	for _, id := range []uuid.UUID{keep, a, b} {
		if !effectstest.HasID(c.DiscardOptions, id) {
			t.Errorf("card %s was not on offer", id)
		}
	}
	for _, card := range opp.Hand.Cards {
		if card.IsKnownTo(other.ID) {
			t.Error("a third player saw the hand")
		}
	}
	if err := g.ResolvePendingChoice(c.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("the caster was not asked to order the cards")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{b, a}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	// The top of a library is the last card of its zone.
	lib := effectstest.LibraryOrder(opp)
	if n := len(lib); n < 2 || lib[n-1] != b || lib[n-2] != a {
		t.Errorf("the top of the library is %v, want %s on top, then %s", lib, b, a)
	}
	if !opp.Hand.Contains(keep) || opp.Hand.Contains(a) || opp.Hand.Contains(b) {
		t.Error("the wrong cards left the hand")
	}
}
