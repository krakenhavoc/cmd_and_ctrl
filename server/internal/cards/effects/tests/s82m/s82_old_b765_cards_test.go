package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b765_cards_test.go — S82 OLD batch 765 (#5628, Tempest:
// enchantments). Each test plays the real catalog entry.

const (
	oldB765NaturesRevolt = "0df0503b-25e7-499b-a703-a867f2da790c"
	oldB765Aluren        = "44e02e70-07ab-4795-8037-55185bd11d54"
	oldB765HannasCustody = "57b0205d-ad9d-45b7-8556-0733aa7a4987"
	oldB765Precognition  = "6d1a9d06-028f-487a-b490-27100c128ea3"
	oldB765Chill         = "70a93d0f-9a02-4e60-a2b8-827fd2410604"
	oldB765HandToHand    = "bf45f579-539c-4ffe-89df-4896a394158f"
	oldB765LightOfDay    = "cd854e00-c9da-4593-8402-4bae637f5185"
	oldB765NoQuarter     = "d238a5c1-f992-4f89-851a-de75b59f6793"
	oldB765Humility      = "ed7bdb3e-5c51-4547-9266-76a791e0b2b0"
	oldB765Sorcerer      = "5e961d15-5972-4e4b-9385-1cd7cd7c6bbe"
)

func TestOldB765EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB765NaturesRevolt: "Nature's Revolt",
		oldB765Aluren:        "Aluren",
		oldB765HannasCustody: "Hanna's Custody",
		oldB765Precognition:  "Precognition",
		oldB765Chill:         "Chill",
		oldB765HandToHand:    "Hand to Hand",
		oldB765LightOfDay:    "Light of Day",
		oldB765NoQuarter:     "No Quarter",
		oldB765Humility:      "Humility",
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

func TestOldB765HumilityMakesEveryCreatureAOneOneWithNoAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Big Flyer", TypeLine: "Creature — Dragon", Power: 5, Toughness: 5, Keywords: []string{"flying", "trample"}})
	theirs := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 7, Keywords: []string{"defender"}})
	if !effectstest.HasEffectiveKeyword(t, g, mine, "flying") {
		t.Fatal("setup: the Dragon should fly before Humility")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Humility", "Enchantment", oldB765Humility, false)
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, th := effectstest.PtOf(t, g, id); p != 1 || th != 1 {
			t.Errorf("a creature is %d/%d under Humility, want 1/1", p, th)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, mine, "flying") || effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("the Dragon kept a keyword under Humility")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "defender") {
		t.Error("an opponent's creature kept defender under Humility")
	}
}

func TestOldB765HumilityLeavesCountersOnTopOfTheBaseSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PushCatalogPermanent(g, me.ID, "Humility", "Enchantment", oldB765Humility, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, bear).Counters = map[string]int{game.CounterPlusOne: 2}
	})
	if p, th := effectstest.PtOf(t, g, bear); p != 3 || th != 3 {
		t.Errorf("a Bear with two +1/+1 counters is %d/%d under Humility, want 3/3", p, th)
	}
}

func TestOldB765NaturesRevoltMakesEveryLandATwoTwoCreatureStillALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	island := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	ring := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter", Power: 0, Toughness: 2})
	effectstest.PushCatalogPermanent(g, me.ID, "Nature's Revolt", "Enchantment", oldB765NaturesRevolt, false)
	for _, id := range []uuid.UUID{forest, island} {
		types := strings.Join(effectstest.EffectiveTypes(t, g, id), " ")
		if !strings.Contains(types, "Land") || !strings.Contains(types, "Creature") {
			t.Errorf("a land's types are %q, want Land and Creature", types)
		}
		if p, th := effectstest.PtOf(t, g, id); p != 2 || th != 2 {
			t.Errorf("a land is %d/%d, want 2/2", p, th)
		}
	}
	if p, th := effectstest.PtOf(t, g, ring); p != 0 || th != 2 {
		t.Errorf("a non-land creature changed to %d/%d", p, th)
	}
}

func TestOldB765HannasCustodyGivesEveryArtifactShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Sol Ring", TypeLine: "Artifact"})
	theirs := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 3, Toughness: 3})
	bear := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PushCatalogPermanent(g, me.ID, "Hanna's Custody", "Enchantment", oldB765HannasCustody, false)
	for _, id := range []uuid.UUID{mine, theirs} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "shroud") {
			t.Error("an artifact has no shroud under Hanna's Custody")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "shroud") {
		t.Error("a non-artifact creature got shroud")
	}
	// Shroud is a real bar: an ability can't target the opponent's artifact
	// creature, though it can target the creature without shroud.
	sorcerer := effectstest.PushCatalogPermanent(g, me.ID, "Prodigal Sorcerer", "Creature — Human Wizard", oldB765Sorcerer, false)
	if err := g.ActivateCatalogAbility(me.ID, sorcerer, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Error("an ability targeted an artifact with shroud")
	}
	if err := g.ActivateCatalogAbility(me.ID, sorcerer, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Errorf("an ability could not target a creature without shroud: %v", err)
	}
}

func TestOldB765ChillTaxesEveryRedSpellTwo(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, mana string
		opponent, refused          bool
	}{
		{"Shock", "Instant", "{R}", "R", false, true},
		{"Shock", "Instant", "{R}", "CCR", false, false},
		{"Shock", "Instant", "{R}", "CR", false, true},
		// Whoever casts it: an opponent's red spell is taxed too.
		{"Shock", "Instant", "{R}", "R", true, true},
		{"Shock", "Instant", "{R}", "CCR", true, false},
		// A spell that is not red is not.
		{"Opt", "Instant", "{U}", "U", false, false},
		{"Giant Growth", "Instant", "{G}", "G", true, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.PushCatalogPermanent(g, me.ID, "Chill", "Enchantment", oldB765Chill, false)
		effectstest.AdvanceToMain(t, g)
		caster := me
		if tc.opponent {
			caster = opp
		}
		effectstest.FloatForTest(g, caster, tc.mana)
		id := uuid.New()
		g.WithWriteLock(func() {
			caster.Hand.PushTop(game.Card{InstanceID: id, Name: tc.name, TypeLine: tc.typeLine, ManaCost: tc.cost, Owner: caster.ID, Controller: caster.ID})
		})
		err := g.CastSpell(caster.ID, id, game.CastSpellParams{Strict: true})
		if (err != nil) != tc.refused {
			t.Errorf("%s with %s (opponent %v): error %v, want refused %v", tc.name, tc.mana, tc.opponent, err, tc.refused)
		}
	}
}

func TestOldB765LightOfDayStopsBlackCreaturesAttackingAndBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	black := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Zombie", TypeLine: "Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}})
	green := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	theirBlack := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Ghoul", TypeLine: "Creature — Zombie", Power: 1, Toughness: 1, Colors: []string{"B"}})
	theirGreen := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Colors: []string{"G"}})
	effectstest.PushCatalogPermanent(g, me.ID, "Light of Day", "Enchantment", oldB765LightOfDay, false)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(black, opp.ID); err == nil {
		t.Error("a black creature attacked under Light of Day")
	}
	if err := g.DeclareAttacker(green, opp.ID); err != nil {
		t.Errorf("a green creature could not attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(theirBlack, green); err == nil {
		t.Error("a black creature blocked under Light of Day")
	}
	if err := g.DeclareBlocker(theirGreen, green); err != nil {
		t.Errorf("a green creature could not block: %v", err)
	}
}

func TestOldB765HandToHandBarsInstantsAndNonManaAbilitiesDuringCombatOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Hand to Hand", "Enchantment", oldB765HandToHand, false)
	sorcerer := effectstest.PushCatalogPermanent(g, me.ID, "Prodigal Sorcerer", "Creature — Human Wizard", oldB765Sorcerer, false)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}

	// Outside combat, both are open.
	effectstest.FloatForTest(g, me, "RR")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Shock", "Instant", "", target); err != nil {
		t.Fatalf("an instant was refused in main phase: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sorcerer, 0, game.ActivateAbilityParams{Targets: target}); err != nil {
		t.Fatalf("an ability was refused in main phase: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// During combat both are shut, whoever tries.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatForTest(g, me, "RR")
	effectstest.FloatForTest(g, opp, "RR")
	for _, p := range []*game.Player{me, opp} {
		id := uuid.New()
		g.WithWriteLock(func() {
			p.Hand.PushTop(game.Card{InstanceID: id, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: p.ID, Controller: p.ID})
		})
		err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: target})
		if err == nil || !strings.Contains(err.Error(), "During combat") {
			t.Errorf("an instant was not barred during combat for %s: %v", p.Name, err)
		}
	}
	effectstest.UntapForTest(g, sorcerer)
	if err := g.ActivateCatalogAbility(me.ID, sorcerer, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Error("a non-mana ability was activated during combat")
	}
}

func TestOldB765NoQuarterDestroysTheLesserPowerCreatureOfEachBlock(t *testing.T) {
	for _, tc := range []struct {
		name                           string
		attackPower, blockPower        int
		attackerSurvives, blockerLives bool
	}{
		{"the blocker has lesser power and is destroyed", 3, 1, true, false},
		{"the blocker has greater power and the attacker is destroyed", 1, 3, false, true},
		{"equal power destroys neither", 2, 2, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			effectstest.PushCatalogPermanent(g, me.ID, "No Quarter", "Enchantment", oldB765NoQuarter, false)
			attacker := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Attacker", TypeLine: "Creature — Giant", Power: tc.attackPower, Toughness: 20})
			blocker := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Blocker", TypeLine: "Creature — Wall", Power: tc.blockPower, Toughness: 20})
			effectstest.DeclareAttack(t, g, opp.ID, attacker)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			if err := g.DeclareBlocker(blocker, attacker); err != nil {
				t.Fatalf("DeclareBlocker: %v", err)
			}
			effectstest.LockInBlocks(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if got := g.Battlefield.Contains(attacker); got != tc.attackerSurvives {
				t.Errorf("attacker on the battlefield = %v, want %v", got, tc.attackerSurvives)
			}
			if got := g.Battlefield.Contains(blocker); got != tc.blockerLives {
				t.Errorf("blocker on the battlefield = %v, want %v", got, tc.blockerLives)
			}
		})
	}
}

func TestOldB765AlurenOffersAnyPlayerAFreeFlashCastOfASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Aluren", "Enchantment", oldB765Aluren, false)
	three := effectstest.EacHandCard(g, opp, "Three Drop", "Creature — Test", "{1}{G}{G}")
	four := effectstest.EacHandCard(g, opp, "Four Drop", "Creature — Test", "{2}{G}{G}")
	spell := effectstest.EacHandCard(g, opp, "Some Instant", "Instant", "{G}")
	mine := effectstest.EacHandCard(g, me, "My Drop", "Creature — Test", "{G}")
	key := effects.GrantedAltCostAnyPlayerSmallCreaturesFreeFlash

	if o := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, opp.ID, three, game.ZoneHand), key); o == nil || !o.AsThoughFlash {
		t.Fatalf("an opponent of the controller is not offered the free flash cast: %+v", o)
	}
	if o := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, mine, game.ZoneHand), key); o == nil {
		t.Error("the controller is not offered it either")
	}
	if o := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, opp.ID, four, game.ZoneHand), key); o != nil {
		t.Error("a mana value 4 creature was offered")
	}
	if o := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, opp.ID, spell, game.ZoneHand), key); o != nil {
		t.Error("a noncreature spell was offered")
	}

	// It is my main phase and the opponent holds no priority window a sorcery-speed
	// creature could use: the free cast carries flash, the printed cost does not.
	if err := g.CastSpell(opp.ID, three, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the opponent cast a creature at sorcery timing on my turn for its printed cost")
	}
	if err := g.CastSpell(opp.ID, three, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("the opponent could not cast it free, as though it had flash: %v", err)
	}
	if err := g.CastSpell(opp.ID, four, game.CastSpellParams{Strict: true, AlternativeCost: key}); err == nil {
		t.Error("a mana value 4 creature was cast for free")
	}
}

func TestOldB765PrecognitionLooksAtAnOpponentsTopCardAndMayBuryIt(t *testing.T) {
	for _, bury := range []bool{true, false} {
		name := "left on top"
		if bury {
			name = "put on the bottom"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			opp := g.Seats[2]
			effectstest.PushCatalogPermanent(g, me.ID, "Precognition", "Enchantment", oldB765Precognition, false)
			for _, seat := range []int{1, 2, 3} {
				effectstest.AdvanceToUpkeepOf(t, g, seat)
			}
			lib := effectstest.SeedLibrary(opp, "Their Top", "Their Next")
			effectstest.AdvanceToUpkeepOf(t, g, 0)
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickPlayer(t, g, me.ID, opp.ID)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
			ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
			if ask == nil || ask.LibraryPlacement != game.LibraryPlaceTopOrBottom || len(ask.ScryCards) != 1 || ask.ScryCards[0] != lib[0] {
				t.Fatalf("prompt %+v, want top-or-bottom over the target's top card", ask)
			}
			var top, bottom []uuid.UUID
			if bury {
				bottom = []uuid.UUID{lib[0]}
			} else {
				top = []uuid.UUID{lib[0]}
			}
			if err := g.ResolvePutInLibrary(ask.ID, me.ID, bottom, top); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			if got := effectstest.LibraryTopIDs(opp, 1)[0]; bury == (got == lib[0]) {
				t.Errorf("top card is %v after bury=%v", got, bury)
			}
		})
	}
}

func TestOldB765PrecognitionDeclinedLooksAtNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Precognition", "Enchantment", oldB765Precognition, false)
	for _, seat := range []int{1, 2, 3} {
		effectstest.AdvanceToUpkeepOf(t, g, seat)
	}
	lib := effectstest.SeedLibrary(opp, "Their Top", "Their Next")
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if ask := effectstest.PutInLibraryChoiceFor(g, me.ID); ask != nil {
		t.Errorf("declining the look still asked about the card: %+v", ask)
	}
	if effectstest.KnownTo(g, lib[0], me.ID) {
		t.Error("the controller learned the card without looking")
	}
}
