package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 181 (Ravnica Allegiance: token makers, enters and dies
// triggers, instants and sorceries, #5521). Every test plays the real catalog
// entry through the engine.

const (
	oldB181Rampage    = "5dc34e93-97e3-434b-9f12-f45c6d556f13"
	oldB181March      = "945eaa8b-0ba5-45a6-a02a-763e5bfe3af8"
	oldB181Absolution = "e6f95e3a-61a6-40a9-8070-ffd2767764ad"
	oldB181Benthid    = "f36d3d87-5596-4c26-b8d0-f2979a166b57"
	oldB181Chill      = "604d2a0e-224c-4a7b-aabf-033e7dcee3f2"
	oldB181Mangler    = "6416dae0-0d40-4d88-b30b-e1a0c8380e78"
	oldB181Battlement = "dcf7d099-a726-4388-9744-70d37c5ea860"
	oldB181Forbidding = "efd8dc6c-8233-45b2-99bd-e38ca2e0b35f"
	oldB181Gathering  = "01f06642-015f-40b4-a940-af576ad9e82c"
	oldB181Mass       = "68e87559-7cae-4dcf-b7cd-a61f19bcd840"
	oldB181Shimmer    = "85e2b073-e953-46f9-a1d4-0dfe030b3751"
)

func TestOldB181EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB181Rampage: "Rampage of the Clans", oldB181March: "Mirror March",
		oldB181Absolution: "Ethereal Absolution", oldB181Benthid: "Mesmerizing Benthid",
		oldB181Chill: "Chillbringer", oldB181Mangler: "Spire Mangler",
		oldB181Battlement: "Lumbering Battlement", oldB181Forbidding: "Forbidding Spirit",
		oldB181Gathering: "Goblin Gathering", oldB181Mass: "Mass Manipulation",
		oldB181Shimmer: "Shimmer of Possibility",
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

func TestOldB181RampageOfTheClansGivesEachControllerACentaurPerPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment — Aura")
	effectstest.B12Permanent(g, opp.ID, "Their Idol", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Rampage of the Clans", "Instant", oldB181Rampage, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TokensNamedControlledBy(g, "Centaur", me.ID); got != 1 {
		t.Errorf("I have %d Centaurs, want 1", got)
	}
	if got := effectstest.TokensNamedControlledBy(g, "Centaur", opp.ID); got != 3 {
		t.Errorf("the opponent has %d Centaurs, want 3", got)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was destroyed")
	}
}

func TestOldB181MirrorMarchCopiesOncePerWonFlipThenExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Mirror March", "Enchantment", oldB181March, 0, 0)
	effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	flips := 0
	for ; flips < 60; flips++ {
		call := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if call == nil {
			break
		}
		if err := g.ResolveCoinCall(call.ID, me.ID, "heads"); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if flips == 0 || flips >= 60 {
		t.Fatalf("%d flips: the chain did not start or never ended", flips)
	}
	wins := flips - 1
	copies := effectstest.TokensNamed(g, "Bear")
	if len(copies) != wins {
		t.Fatalf("%d copies after %d flips, want %d (one per won flip)", len(copies), flips, wins)
	}
	for _, c := range copies {
		if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "haste") {
			t.Error("a copy has no haste")
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if left := effectstest.TokensNamed(g, "Bear"); len(left) != 0 {
		t.Errorf("%d copies were not exiled at the end step", len(left))
	}
}

func TestOldB181EtherealAbsolutionAnthemsAndExilesForASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	abs := effectstest.B12Push(g, me.ID, "Ethereal Absolution", "Enchantment", oldB181Absolution, 0, 0)
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 3 {
		t.Errorf("my creature has power %d, want 3", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("their creature has power %d, want 2", p)
	}
	creature := effectstest.B17GraveyardCard(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	spell := effectstest.B17GraveyardCard(opp, "Dead Bolt", "Instant", "{R}")
	mineDead := effectstest.B17GraveyardCard(me, "My Dead Bear", "Creature — Bear", "{1}{G}")
	effectstest.FloatForTest(g, me, "CCWB")
	if err := g.ActivateCatalogAbility(me.ID, abs, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(mineDead)}); err == nil {
		t.Fatal("targeted a card in my own graveyard")
	}
	for _, tc := range []struct {
		card   uuid.UUID
		spirit int
	}{{creature, 1}, {spell, 1}} {
		effectstest.FloatForTest(g, me, "CCWB")
		if err := g.ActivateCatalogAbility(me.ID, abs, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(tc.card)}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Graveyard.Contains(tc.card) {
			t.Errorf("the card stayed in the graveyard")
		}
		if got := effectstest.TokensNamedControlledBy(g, "Spirit", me.ID); got != tc.spirit {
			t.Errorf("%d Spirits, want %d (only a creature card makes one)", got, tc.spirit)
		}
	}
}

func TestOldB181MesmerizingBenthidMakesIllusionsThatFreezeWhatTheyBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	benthid := effectstest.CastHoldCreature(t, g, "Mesmerizing Benthid", "Creature — Octopus", oldB181Benthid, 4, 5)
	illusions := effectstest.TokensNamed(g, "Illusion")
	if len(illusions) != 2 {
		t.Fatalf("%d Illusions, want 2", len(illusions))
	}
	if p, to := effectstest.EffectivePTOf(t, g, illusions[0].InstanceID); p != 0 || to != 2 {
		t.Errorf("an Illusion is %d/%d, want 0/2", p, to)
	}
	if !effectstest.HasEffectiveKeyword(t, g, benthid, "hexproof") {
		t.Error("no hexproof while I control an Illusion")
	}
	// An Illusion that blocks freezes the attacker.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	attacker2 := effectstest.B12Creature(g2, me2.ID, "Attacker", "Creature — Bear", 1, 5)
	effectstest.AdvanceToMain(t, g2)
	// The opponent's Benthid makes the blockers.
	g2.WithWriteLock(func() {
		g2.Turn.ActiveSeat = 1
	})
	effectstest.CastHoldCreature(t, g2, "Mesmerizing Benthid", "Creature — Octopus", oldB181Benthid, 4, 5)
	g2.WithWriteLock(func() {
		g2.Turn.ActiveSeat = 0
	})
	ill := effectstest.TokensNamed(g2, "Illusion")
	if len(ill) != 2 {
		t.Fatalf("%d Illusions, want 2", len(ill))
	}
	effectstest.DeclareAttack(t, g2, opp2.ID, attacker2)
	effectstest.AdvanceTo(t, g2, game.StepDeclareBlockers)
	if err := g2.DeclareBlocker(ill[0].InstanceID, attacker2); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g2)
	effectstest.PassPriorityAroundTable(t, g2)
	if skips := effectstest.HmlUntapSkips(t, g2, attacker2); len(skips) == 0 {
		t.Error("the blocked creature will untap normally")
	}
}

func TestOldB181ChillbringerTapsAndFreezes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Chillbringer", "Creature — Elemental", oldB181Chill, 3, 3)
	if err := g.ResolvePickTarget(effectstest.PickTargetPrompt(t, g).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("targeted my own creature")
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertTapped(t, g, true, theirs)
	if skips := effectstest.HmlUntapSkips(t, g, theirs); len(skips) == 0 {
		t.Error("the creature will untap normally")
	}
}

func TestOldB181SpireManglerPumpsAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird", Keywords: []string{"flying"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	ground := effectstest.B12Creature(g, me.ID, "Ground", "Creature — Bear", 2, 2)
	mangler := effectstest.CastHoldCreature(t, g, "Spire Mangler", "Creature — Insect", oldB181Mangler, 2, 1)
	if err := g.ResolvePickTarget(effectstest.PickTargetPrompt(t, g).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: ground}); err == nil {
		t.Fatal("targeted a creature without flying")
	}
	effectstest.AnswerPickTarget(t, g, flyer)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, flyer); p != 3 {
		t.Errorf("the flyer has power %d, want 3", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, mangler); p != 2 {
		t.Errorf("the Mangler has power %d, want 2", p)
	}
}

func TestOldB181LumberingBattlementGrowsPerExiledCreatureAndReturnsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID, TokenKey: "test-token",
	})
	_ = token
	wall := effectstest.CastHoldCreature(t, g, "Lumbering Battlement", "Creature — Beast", oldB181Battlement, 4, 5)
	prompt := effectstest.ChooseCardsPrompt(t, g)
	for _, c := range prompt.ChooseCards {
		if c == token || c == wall {
			t.Fatalf("offered %s, which is a token or the Battlement itself", c)
		}
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Fatal("the chosen creatures were not exiled")
	}
	if p, to := effectstest.EffectivePTOf(t, g, wall); p != 8 || to != 9 {
		t.Errorf("the Battlement is %d/%d, want 8/9", p, to)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wall) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"Bear A", "Bear B"} {
		if _, ok := effectstest.E2BattlefieldNamed(g, name); !ok {
			t.Errorf("%s did not come back when the Battlement left", name)
		}
	}
}

func TestOldB181ForbiddingSpiritTaxesAttacksUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastHoldCreature(t, g, "Forbidding Spirit", "Creature — Spirit Cleric", oldB181Forbidding, 3, 3)
	var onPlayer, onWalker game.AttackTaxPrice
	attacker := effectstest.B12Creature(g, opp.ID, "Raider", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		onPlayer = g.PriceAttackDeclarationForEffect([]game.AttackDeclaration{{Attacker: attacker, Target: me.ID}})
	})
	if onPlayer.Cost != "{2}" {
		t.Errorf("attacking me costs %q, want {2}", onPlayer.Cost)
	}
	walker := effectstest.PushWalkerForTest(g, me.ID, "Walker", "", 3)
	g.WithWriteLock(func() {
		onWalker = g.PriceAttackDeclarationForEffect([]game.AttackDeclaration{{Attacker: attacker, Target: walker}})
	})
	if onWalker.Cost != "{2}" {
		t.Errorf("attacking my planeswalker costs %q, want {2}", onWalker.Cost)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	g.WithWriteLock(func() {
		onPlayer = g.PriceAttackDeclarationForEffect([]game.AttackDeclaration{{Attacker: attacker, Target: me.ID}})
	})
	if onPlayer.Cost != "{2}" {
		t.Errorf("the tax ended early, in another player's turn: %q", onPlayer.Cost)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	g.WithWriteLock(func() {
		onPlayer = g.PriceAttackDeclarationForEffect([]game.AttackDeclaration{{Attacker: attacker, Target: me.ID}})
	})
	if !onPlayer.IsFree() {
		t.Errorf("the tax outlived its controller's next turn: %q", onPlayer.Cost)
	}
}

func TestOldB181GoblinGatheringCountsCopiesInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Goblin Gathering", "Sorcery", oldB181Gathering, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TokensNamedControlledBy(g, "Goblin", me.ID); got != 2 {
		t.Fatalf("%d Goblins, want 2", got)
	}
	effectstest.B17GraveyardCard(me, "Goblin Gathering", "Sorcery", "{2}{R}")
	effectstest.B17GraveyardCard(me, "Goblin Gathering", "Sorcery", "{2}{R}")
	effectstest.CastCatalogSpell(t, g, "Goblin Gathering", "Sorcery", oldB181Gathering, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The first cast went to the graveyard, so the second cast counts it and the
	// two seeded copies: two plus three.
	if got := effectstest.TokensNamedControlledBy(g, "Goblin", me.ID); got != 2+5 {
		t.Errorf("%d Goblins, want 7", got)
	}
}

func TestOldB181MassManipulationStealsXCreaturesAndPlaneswalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	walker := effectstest.PushWalkerForTest(g, opp.ID, "Walker", "", 4)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	id := effectstest.HandCardFull(me, "Mass Manipulation", "Sorcery", "{X}{X}{U}{U}{U}{U}", oldB181Mass, []string{"U"})
	effectstest.FloatForTest(g, me, "CCCCUUUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(bear, rock)}); err == nil {
		t.Fatal("targeted an artifact")
	}
	effectstest.FloatForTest(g, me, "CCCCUUUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(bear, walker)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, p := range map[string]uuid.UUID{"creature": bear, "planeswalker": walker} {
		c, ok := effectstest.BattlefieldCard(g, p)
		if !ok || c.Controller != me.ID {
			t.Errorf("the %s was not taken (on the battlefield %v)", name, ok)
		}
	}
}

func TestOldB181ShimmerOfPossibilityTakesOneOfFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four", "Five")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Shimmer of Possibility", "Sorcery", oldB181Shimmer, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsPrompt(t, g)
	if len(prompt.ChooseCards) != 4 {
		t.Fatalf("%d cards offered, want the top four", len(prompt.ChooseCards))
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, nil); err == nil {
		t.Fatal("took nothing from a mandatory pick")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatal(err)
	}
	if got := me.Hand.Size(); got != hand+1 || !me.Hand.Contains(ids[2]) {
		t.Errorf("hand %d (want %d), has the pick %v", got, hand+1, me.Hand.Contains(ids[2]))
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != ids[4] {
		t.Errorf("the top card is %v, want the fifth card (the rest of the look went to the bottom)", top.Name)
	}
	if bottom := me.Library.Cards[:3]; len(bottom) != 3 {
		t.Error("the library lost cards")
	}
}
