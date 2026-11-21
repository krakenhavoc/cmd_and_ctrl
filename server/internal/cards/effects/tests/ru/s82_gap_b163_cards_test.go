package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b163_cards_test.go — Jumpstart 2022 creatures, GAP batch 163 (#5948).
// Every test plays the real catalog entry through the engine.

const (
	gap163Clique    = "244d4807-0802-41bc-9460-55ac38a28a72"
	gap163Thrasher  = "f2d1e991-e13a-43c7-85aa-d53dd14822e6"
	gap163Weaver    = "ef0ed1b1-d092-46f0-beb2-444e3645d273"
	gap163Wingsmith = "a05bcc3e-25d5-4111-820e-a097189afb7e"
	gap163Patrol    = "dea0f359-19a5-4aef-8c96-574eaa748a31"
	gap163Mentor    = "a1fb1b77-a49f-4447-bbe0-156407f192e2"
	gap163Artificer = "d2575513-72fa-49c7-be6b-9f6f85950f88"
	gap163Alseid    = "d371031e-78be-4964-b521-b1eaa22a5480"
	gap163Protector = "cff434e1-e567-4a80-8dc5-593d70b0f348"
	gap163Anointer  = "e4f880f5-0639-4acb-a099-e48f5707f0db"
	gap163Healer    = "9442a18b-9093-4a10-ab61-6962f5af3bb6"
)

func TestGapB163EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{gap163Clique, gap163Thrasher, gap163Weaver, gap163Wingsmith, gap163Patrol, gap163Mentor,
		gap163Artificer, gap163Alseid, gap163Protector, gap163Anointer, gap163Healer} {
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

func TestGapB163PrintedKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		kws                    []string
	}{
		{"Vendilion Clique", "Legendary Creature — Faerie Wizard", gap163Clique, []string{"flash", "flying"}},
		{"Windrider Patrol", "Creature — Merfolk Wizard", gap163Patrol, []string{"flying"}},
		{"Alseid of Life's Bounty", "Enchantment Creature — Nymph", gap163Alseid, []string{"lifelink"}},
		{"Angelic Protector", "Creature — Angel", gap163Protector, []string{"flying"}},
		{"Anointer of Valor", "Creature — Angel", gap163Anointer, []string{"flying"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
		for _, kw := range tc.kws {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

func TestGapB163CliqueTucksTheChosenCardAndThePlayerDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc18Library(opp, 4)
	land := uuid.New()
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: land, Name: "Their Mountain", TypeLine: "Basic Land — Mountain", Owner: opp.ID, Controller: opp.ID})
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Their Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastHoldCreature(t, g, "Vendilion Clique", "Legendary Creature — Faerie Wizard", gap163Clique, 3, 1)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if pick.Chooser != me.ID {
		t.Fatalf("pick chooser %v, want me", pick.Chooser)
	}
	if effectstest.HasID(pick.DiscardOptions, land) {
		t.Error("a land card was offered")
	}
	if !effectstest.HasID(pick.DiscardOptions, bolt) {
		t.Fatalf("the nonland card was not offered: %v", pick.DiscardOptions)
	}
	if !pick.PickOptional {
		t.Error("the choice is not optional")
	}
	libBefore, handBefore := opp.Library.Size(), opp.Hand.Size()
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{bolt}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(bolt) {
		t.Error("the chosen card is still in their hand")
	}
	if !opp.Hand.Contains(land) {
		t.Error("the land left their hand")
	}
	if got := opp.Library.Size(); got != libBefore {
		t.Errorf("their library is %d, want %d (one tucked, one drawn)", got, libBefore)
	}
	// Zone.Cards[0] is the bottom; the card drawn afterwards came off the top.
	if len(opp.Library.Cards) == 0 || opp.Library.Cards[0].InstanceID != bolt {
		t.Error("the chosen card is not on the bottom of their library")
	}
	if opp.Hand.Size() != handBefore {
		t.Errorf("their hand is %d, want %d (one put away, one drawn)", opp.Hand.Size(), handBefore)
	}
}

func TestGapB163CliqueDecliningChoosesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc18Library(opp, 4)
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Their Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastHoldCreature(t, g, "Vendilion Clique", "Legendary Creature — Faerie Wizard", gap163Clique, 3, 1)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	hand, lib := opp.Hand.Size(), opp.Library.Size()
	if err := g.ResolvePendingChoice(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand || opp.Library.Size() != lib || !opp.Hand.Contains(bolt) {
		t.Error("declining the choice changed their hand or library")
	}
}

func TestGapB163WakeThrasherGrowsForEachPermanentThatUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thrasher := effectstest.B12Push(g, me.ID, "Wake Thrasher", "Creature — Merfolk Soldier", gap163Thrasher, 1, 1)
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	untap := func(id uuid.UUID) {
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventUntapCard, CardID: id}) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	untap(theirs)
	if got := effectstest.EffectivePower(t, g, thrasher); got != 1 {
		t.Fatalf("power %d after an opponent's permanent untapped, want 1", got)
	}
	untap(mine)
	if got := effectstest.EffectivePower(t, g, thrasher); got != 2 {
		t.Fatalf("power %d after one of mine untapped, want 2", got)
	}
	untap(mine)
	if got := effectstest.EffectivePower(t, g, thrasher); got != 3 {
		t.Errorf("power %d after a second untap, want 3", got)
	}
}

func TestGapB163WatertrapWeaverTapsAndFreezesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Watertrap Weaver", "Creature — Merfolk Wizard", gap163Weaver, 2, 2)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err == nil {
		t.Fatal("my own creature was accepted as a target")
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, theirs) {
		t.Fatal("the target was not tapped")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if !effectstest.HmlTapped(t, g, theirs) {
		t.Error("the target untapped in its controller's next untap step")
	}
}

func TestGapB163WeldfastWingsmithFliesWhenAnArtifactEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	smith := effectstest.B12Push(g, me.ID, "Weldfast Wingsmith", "Creature — Human Artificer", gap163Wingsmith, 3, 3)
	if effectstest.HasKeywordOnBattlefield(t, g, smith, "flying") {
		t.Fatal("flying before any artifact entered")
	}
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, smith, "flying") {
		t.Fatal("flying after a nonartifact entered")
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, smith, "flying") {
		t.Error("no flying after an artifact entered")
	}
}

func TestGapB163WindriderPatrolScriesTwoOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 6)
	patrol := effectstest.B12Push(g, me.ID, "Windrider Patrol", "Creature — Merfolk Wizard", gap163Patrol, 4, 3)
	effectstest.DealCombatDamage(g, me.ID, patrol, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil || len(scry.ScryCards) != 2 {
		t.Fatalf("scry prompt = %+v, want scry 2", scry)
	}
}

func TestGapB163WizardMentorReturnsItselfAndTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mentor := effectstest.B12Push(g, me.ID, "Wizard Mentor", "Creature — Human Wizard", gap163Mentor, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaActivate(t, g, me, mentor, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mentor) || !me.Hand.Contains(mentor) {
		t.Error("the Mentor did not return to hand")
	}
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the Bear did not return to hand")
	}
}

func TestGapB163WizardMentorTargetingItselfReturnsOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mentor := effectstest.B12Push(g, me.ID, "Wizard Mentor", "Creature — Human Wizard", gap163Mentor, 2, 2)
	effectstest.ApaActivate(t, g, me, mentor, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mentor}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mentor) || g.Battlefield.Contains(mentor) {
		t.Error("the Mentor did not return to hand")
	}
}

func TestGapB163AethershieldArtificerPumpsAnArtifactCreatureAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Aethershield Artificer", "Creature — Dwarf Artificer", gap163Artificer, 3, 3)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	effectstest.AnswerPickTarget(t, g, golem)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, golem); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, golem, "indestructible") {
		t.Error("the Golem is not indestructible")
	}
}

func TestGapB163AlseidGivesProtectionFromAChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	alseid := effectstest.B12Push(g, me.ID, "Alseid of Life's Bounty", "Enchantment Creature — Nymph", gap163Alseid, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "W")
	effectstest.ApaActivate(t, g, me, alseid, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if g.Battlefield.Contains(alseid) {
		t.Error("the Alseid was not sacrificed as a cost")
	}
	for i := 0; i < 8 && effectstest.PendingOfKind(g, game.PendingChoiceColor) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerColor(t, g, me.ID, "R")
	c, _ := effectstest.BattlefieldCard(g, bear)
	if !game.HasKeyword(&c, game.ProtectionFromColor("R")) {
		t.Error("the Bear did not gain protection from red")
	}
}

func TestGapB163AngelicProtectorGainsToughnessWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	angel := effectstest.B12Push(g, me.ID, "Angelic Protector", "Creature — Angel", gap163Protector, 2, 2)
	if got := effectstest.ApaLive(g, angel).CurrentToughness(); got != 2 {
		t.Fatalf("toughness %d before being targeted, want 2", got)
	}
	effectstest.CastCatalogSpell(t, g, "Pacifism", "Enchantment — Aura", effectstest.PacifismAuraOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: angel}})
	effectstest.PassPriorityAroundTable(t, g)
	live := effectstest.ApaLive(g, angel)
	if live == nil {
		t.Fatal("the Angelic Protector left the battlefield")
	}
	if live.CurrentToughness() != 5 {
		t.Errorf("toughness %d, want 5 after becoming a target", live.CurrentToughness())
	}
}

func TestGapB163AnointerOfValorCountersAnAttackerForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Anointer of Valor", "Creature — Angel", gap163Anointer, 3, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B09AddColorless(me, 3)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, bear); got != 1 {
		t.Errorf("+1/+1 counters on the attacker = %d, want 1", got)
	}
}

func TestGapB163AnointerOfValorDeclinedPutsNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Anointer of Valor", "Creature — Angel", gap163Anointer, 3, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B09AddColorless(me, 3)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, bear); got != 0 {
		t.Errorf("+1/+1 counters on the attacker = %d, want 0", got)
	}
}

func TestGapB163AttendedHealerMakesACatOnTheFirstLifeGainEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Attended Healer", "Creature — Kor Cleric", gap163Healer, 2, 3)
	gain := func() {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	gain()
	if got := effectstest.BattlefieldNamed(g, me.ID, "Cat"); got != 1 {
		t.Fatalf("%d Cat tokens after the first life gain, want 1", got)
	}
	gain()
	if got := effectstest.BattlefieldNamed(g, me.ID, "Cat"); got != 1 {
		t.Errorf("%d Cat tokens after the second life gain, want still 1", got)
	}
}

func TestGapB163AttendedHealerGivesAnotherClericLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	healer := effectstest.B12Push(g, me.ID, "Attended Healer", "Creature — Kor Cleric", gap163Healer, 2, 3)
	cleric := effectstest.B12Creature(g, me.ID, "Other Cleric", "Creature — Human Cleric", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, healer, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: healer}}}); err == nil {
		t.Error("the Healer was accepted as its own target (\"another\")")
	}
	if err := g.ActivateCatalogAbility(me.ID, healer, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Error("a non-Cleric was accepted as a target")
	}
	effectstest.ApaActivate(t, g, me, healer, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: cleric}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, cleric, "lifelink") {
		t.Error("the Cleric did not gain lifelink")
	}
}
