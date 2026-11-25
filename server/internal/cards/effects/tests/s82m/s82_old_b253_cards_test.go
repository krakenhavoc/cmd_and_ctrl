package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b253_cards_test.go — Aether Revolt, batch 253 (S82 older sets,
// #5536). Every test plays the real catalog entry through the engine.

const (
	oldB253MidnightEntourageOracle = "02d6b7f6-1fe0-4a93-8385-69ab160ab732"
	oldB253StratagemOracle         = "1225bee2-c829-4a93-9f75-3495277c23bc"
	oldB253QuicksmithSpyOracle     = "54250029-1bd7-4caf-a839-9497607ce725"
	oldB253BetrayalOracle          = "a660e7a0-0750-475d-a4bd-de5b37d53e97"
	oldB253CruelFinalityOracle     = "1285a519-b3a8-4014-8f33-89f3beb0c00b"
	oldB253YahennisOracle          = "692babfa-9b74-415e-a560-abae4ccb006b"
	oldB253WeldfastOracle          = "74adb283-9bef-4a6f-9489-bcfcba53e8d7"
	oldB253ConstructOracle         = "bccbd5ae-4e3f-43ee-a51b-d0ec2c645238"
	oldB253DeftDismissalOracle     = "7cbe863a-6654-4f54-b4ec-e1040e830174"
	oldB253QuicksmithRebelOracle   = "d69594fc-9250-41e7-a170-d4705615312c"
	oldB253PiasRevolutionOracle    = "edce515d-828e-49ae-8651-6d14f7470a62"
	oldB253ChandrasOracle          = "f72d2074-2f32-4b56-a172-947bf6104fae"
)

func TestOldB253EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB253MidnightEntourageOracle, oldB253StratagemOracle, oldB253QuicksmithSpyOracle, oldB253BetrayalOracle,
		oldB253CruelFinalityOracle, oldB253YahennisOracle, oldB253WeldfastOracle, oldB253ConstructOracle,
		oldB253DeftDismissalOracle, oldB253QuicksmithRebelOracle, oldB253PiasRevolutionOracle, oldB253ChandrasOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestOldB253ChandrasRevolutionDamagesTheCreatureAndFreezesTheLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	land := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	_ = me
	effectstest.CastCatalogSpell(t, g, "Chandra's Revolution", "Sorcery", oldB253ChandrasOracle, effectstest.CardTargets(bear, land))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1g2WantDamage(t, g, bear, 4, "the creature")
	effectstest.AssertTapped(t, g, true, land)
}

func TestOldB253ChandrasRevolutionRefusesANonLandSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Chandra's Revolution", "Sorcery", oldB253ChandrasOracle, effectstest.CardTargets(bear, other)); err == nil {
		t.Error("a creature was accepted as the land target")
	}
}

func TestOldB253IllusionistsStratagemBlinksTwoAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	library := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Illusionist's Stratagem", "Instant", oldB253StratagemOracle, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			t.Errorf("%s is still the same object on the battlefield", id)
		}
	}
	for _, name := range []string{"Bear A", "Bear B"} {
		if len(oldB253OnBattlefield(g, me.ID, name)) != 1 {
			t.Errorf("%s did not come back under its owner's control", name)
		}
	}
	if got := me.Library.Size(); got != library-1 {
		t.Errorf("library %d -> %d, want one card drawn", library, got)
	}
}

func TestOldB253IllusionistsStratagemDrawsWithNoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	before := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Illusionist's Stratagem", "Instant", oldB253StratagemOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != before-1 {
		t.Errorf("library %d -> %d, want one card drawn", before, got)
	}
}

func TestOldB253CruelFinalityShrinksAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Cruel Finality", "Instant", oldB253CruelFinalityOracle, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 1 || tt != 1 {
		t.Errorf("the creature is %d/%d, want 1/1", p, tt)
	}
	if effectstest.ScryChoiceFor(g, g.Seats[0].ID) == nil {
		t.Error("no scry prompt")
	}
}

func TestOldB253YahennisExpertiseShrinksEveryCreatureThenOffersAFreeCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	cheap := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cheap, Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	big := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: big, Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{3}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Yahenni's Expertise", "Sorcery", oldB253YahennisOracle, nil)
	effectstest.PassPriorityUntilChoice(t, g)
	if p, tt := effectstest.PtOf(t, g, theirs); p != 1 || tt != 1 {
		t.Errorf("their Ogre is %d/%d, want 1/1", p, tt)
	}
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheap)
	for _, id := range offer.DiscardOptions {
		if id == big {
			t.Error("a spell with mana value 4 was offered")
		}
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, cheap, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, mine); ok {
		t.Error("my 2/2 Bear survived -3/-3")
	}
	if got := len(oldB253OnBattlefield(g, me.ID, "Cheap Bear")); got != 1 {
		t.Errorf("the free-cast Bear is on the battlefield %d times, want once", got)
	}
}

func TestOldB253TezzeretsBetrayalDestroysAndFetchesTezzeretFromLibraryOrGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tez := effectstest.PushLibraryCard(me, game.Card{Name: "Tezzeret, Master of Metal", TypeLine: "Legendary Planeswalker — Tezzeret", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Tezzeret's Betrayal", "Sorcery", oldB253BetrayalOracle, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("the creature was not destroyed")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerSearchByID(t, g, me.ID, tez)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(tez) {
		t.Error("Tezzeret, Master of Metal is not in hand")
	}
}

func TestOldB253WeldfastEngineerPumpsAnArtifactCreatureAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Weldfast Engineer", "Creature — Human Artificer", oldB253WeldfastOracle, false)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityUntilChoice(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTarget(t, g, golem)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, golem); p != 4 {
		t.Errorf("the artifact creature has power %d, want 4", p)
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("the non-artifact creature has power %d, want 2", p)
	}
}

func TestOldB253CrackdownConstructGrowsOnANonManaActivationOfAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	construct := effectstest.B12Push(g, me.ID, "Crackdown Construct", "Artifact Creature — Construct", oldB253ConstructOracle, 2, 2)
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	// A Quicksmith Spy entering gives the Rock "{T}: Draw a card".
	effectstest.CastCatalogSpell(t, g, "Quicksmith Spy", "Creature — Human Artificer", oldB253QuicksmithSpyOracle, nil)
	effectstest.PassPriorityUntilChoice(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTarget(t, g, rock)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, construct); p != 2 || tt != 2 {
		t.Fatalf("Construct is %d/%d before any activation, want 2/2", p, tt)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, rock, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
	if p, tt := effectstest.PtOf(t, g, construct); p != 3 || tt != 3 {
		t.Errorf("Construct is %d/%d after the activation, want 3/3", p, tt)
	}
}

func TestOldB253QuicksmithSpyGrantEndsWhenTheSpyLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	effectstest.CastCatalogSpell(t, g, "Quicksmith Spy", "Creature — Human Artificer", oldB253QuicksmithSpyOracle, nil)
	effectstest.PassPriorityUntilChoice(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTarget(t, g, rock)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, rock, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the Rock did not gain the ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, spyOn(g, me.ID, "Quicksmith Spy"))
	effectstest.P1g3Untap(g, rock)
	if err := g.ActivateCatalogAbility(me.ID, rock, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the Rock kept the ability after the Spy left")
	}
}

func spyOn(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestOldB253QuicksmithRebelGrantsAPingTheArtifactDeals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	effectstest.CastCatalogSpell(t, g, "Quicksmith Rebel", "Creature — Human Artificer", oldB253QuicksmithRebelOracle, nil)
	effectstest.PassPriorityUntilChoice(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTarget(t, g, rock)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, rock, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("the Rock did not gain the ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want 2 damage", life, opp.Life)
	}
}

func TestOldB253DeftDismissalDividesAmongAttackingOrBlockingCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	a := effectstest.B12Creature(g, me.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Attacker B", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	spell := effectstest.HandCardOf(g, me, "Deft Dismissal", "Instant", "{3}{W}", oldB253DeftDismissalOracle)
	effectstest.FillPoolColored(me, "W", 4)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Strict: true, Targets: effectstest.CardTargets(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1},
	}); err != nil {
		t.Fatalf("Deft Dismissal: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, a); ok {
		t.Error("the attacker dealt 2 survived")
	}
	effectstest.P1g2WantDamage(t, g, b, 1, "the attacker dealt 1")
	if err := effectstest.CastCatalogSpellErr(t, g, "Deft Dismissal", "Instant", oldB253DeftDismissalOracle, effectstest.CardTargets(idle)); err == nil {
		t.Error("a creature that is neither attacking nor blocking was accepted")
	}
}

func TestOldB253MidnightEntourageDrawsWhenAnAetherbornDiesAndPumpsOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Midnight Entourage", "Creature — Aetherborn Rogue", oldB253MidnightEntourageOracle, false)
	other := effectstest.B12Creature(g, me.ID, "Other Aetherborn", "Creature — Aetherborn", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	if p, tt := effectstest.PtOf(t, g, other); p != 2 || tt != 2 {
		t.Errorf("the other Aetherborn is %d/%d, want 2/2", p, tt)
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 1 {
		t.Errorf("a Bear got the anthem: power %d", p)
	}
	hand, life := me.Hand.Size(), me.Life
	effectstest.P1Destroy(t, g, other)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d -> %d, life %d -> %d: want a card and 1 life lost", hand, me.Hand.Size(), life, me.Life)
	}
	effectstest.P1Destroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Error("a Bear dying drew a card")
	}
}

func TestOldB253PiasRevolutionOpponentChoosesDamageOrTheReturn(t *testing.T) {
	for _, takeDamage := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		effectstest.AdvanceToMain(t, g)
		effectstest.PushCatalogPermanent(g, me.ID, "Pia's Revolution", "Enchantment", oldB253PiasRevolutionOracle, false)
		rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
		life := opp.Life
		effectstest.P1Destroy(t, g, rock)
		effectstest.PassPriorityUntilChoice(t, g)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
		}
		effectstest.PassPriorityUntilChoice(t, g)
		effectstest.AnswerMayChoice(t, g, opp.ID, takeDamage)
		effectstest.PassPriorityAroundTable(t, g)
		if takeDamage {
			if opp.Life != life-3 || me.Hand.Contains(rock) {
				t.Errorf("take damage: life %d -> %d, returned %v", life, opp.Life, me.Hand.Contains(rock))
			}
		} else if opp.Life != life || !me.Hand.Contains(rock) {
			t.Errorf("decline: life %d -> %d, returned %v", life, opp.Life, me.Hand.Contains(rock))
		}
	}
}

func oldB253OnBattlefield(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}
