package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m19_b196_cards_test.go — Core Set 2019, batch 196 (#5527). Each test plays
// the real catalog entry and asserts what its text says.

const (
	m19B196SurgeMareOracle      = "8342de86-5b11-486d-8012-f0f9d2ded8c4"
	m19B196ClancallerOracle     = "a8f5440e-5d4b-420a-8cfd-27a3c72be0d2"
	m19B196RebuildingOracle     = "df9fe645-6f7b-44a0-ab69-b68cb7e525cd"
	m19B196ArcadesOracle        = "e7fc5ad6-f2f0-4b06-a61b-05022db9d93b"
	m19B196PyromancerOracle     = "f040310d-d6ac-4d77-9f76-ca325eddf306"
	m19B196ProdigiousOracle     = "15d63e5d-dc4d-4e5f-8043-78039a49732b"
	m19B196AlterationOracle     = "2518a4f2-607f-41bb-9389-2009138d6b09"
	m19B196TalonsOracle         = "2ff309e3-90e0-40e7-a277-3df057f8f198"
	m19B196DwindleOracle        = "48a8fce6-b493-4b1d-8ac8-2a4f4b4781b5"
	m19B196CreatureType         = "Creature — Bear"
	m19B196DefenderWallTypeLine = "Creature — Wall"
)

func TestM19B196EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		m19B196SurgeMareOracle:  "Surge Mare",
		m19B196ClancallerOracle: "Elvish Clancaller",
		m19B196RebuildingOracle: "Patient Rebuilding",
		m19B196ArcadesOracle:    "Arcades, the Strategist",
		m19B196PyromancerOracle: "Dismissive Pyromancer",
		m19B196ProdigiousOracle: "Prodigious Growth",
		m19B196AlterationOracle: "Metamorphic Alteration",
		m19B196TalonsOracle:     "Talons of Wildwood",
		m19B196DwindleOracle:    "Dwindle",
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

func m19B196Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors []string, keywords ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Keywords: keywords,
		Owner: owner, Controller: owner,
	})
}

func TestM19B196ProdigiousGrowthGivesSevenAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Prodigious Growth", m19B196ProdigiousOracle, host)
	if p, tough := effectstest.EffectivePTOf(t, g, host); p != 9 || tough != 9 {
		t.Errorf("host is %d/%d, want 9/9", p, tough)
	}
	if !effectstest.HasKeywordOf(g, host, "trample") {
		t.Error("the host has no trample")
	}
}

func TestM19B196TalonsGivesOneAndTrampleAndReturnsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	host := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Talons of Wildwood", m19B196TalonsOracle, host)
	if p, tough := effectstest.EffectivePTOf(t, g, host); p != 3 || tough != 3 {
		t.Fatalf("host is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasKeywordOf(g, host, "trample") {
		t.Error("the host has no trample")
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the return ability activated while the Aura was on the battlefield")
	}
	effectstest.WoeB1Destroy(t, g, host)
	if !me.Graveyard.Contains(aura) {
		t.Fatal("the Aura is not in the graveyard after its creature died")
	}
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the return activated without {G}")
	}
	effectstest.FloatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{2}{G}: return this card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("Talons of Wildwood did not return to hand")
	}
}

func TestM19B196DismissivePyromancerLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pyro := effectstest.B12Push(g, me.ID, "Dismissive Pyromancer", "Creature — Human Wizard", m19B196PyromancerOracle, 2, 2)
	me.Hand.Cards = me.Hand.Cards[:0]
	effectstest.ApaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, pyro, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no card to discard")
	}
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, pyro, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discard cost was not paid")
	}
	if !effectstest.B16Tapped(t, g, pyro) {
		t.Error("the Pyromancer did not tap")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 1 {
		t.Errorf("hand is %d, want the one drawn card", me.Hand.Size())
	}
}

func TestM19B196DismissivePyromancerSacrificesForFourDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pyro := effectstest.B12Push(g, me.ID, "Dismissive Pyromancer", "Creature — Human Wizard", m19B196PyromancerOracle, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", m19B196CreatureType, 5, 5)
	small := effectstest.B12Creature(g, opp.ID, "Small", m19B196CreatureType, 4, 4)
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, pyro, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(small)}); err == nil {
		t.Fatal("activated {2}{R} without red mana")
	}
	effectstest.ApaMana(me, "R")
	effectstest.B16Activate(t, g, me.ID, pyro, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(small)})
	if g.Battlefield.Contains(pyro) {
		t.Error("the Pyromancer was not sacrificed")
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 4/4 survived 4 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 5/5 died")
	}
}

func TestM19B196SurgeMareCantBeBlockedByGreenCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mare := m19B196Push(g, me.ID, "Surge Mare", "Creature — Horse Fish", m19B196SurgeMareOracle, 0, 5, []string{"U"})
	green := m19B196Push(g, opp.ID, "Green Bear", m19B196CreatureType, "", 2, 2, []string{"G"})
	red := m19B196Push(g, opp.ID, "Red Bear", m19B196CreatureType, "", 2, 2, []string{"R"})
	effectstest.BrAttack(t, g, mare)
	if err := g.DeclareBlocker(green, mare); err == nil {
		t.Error("a green creature blocked Surge Mare")
	}
	if err := g.DeclareBlocker(red, mare); err != nil {
		t.Errorf("a red creature couldn't block Surge Mare: %v", err)
	}
}

func TestM19B196SurgeMareLootsOnDamageToAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B")
	mare := m19B196Push(g, me.ID, "Surge Mare", "Creature — Horse Fish", m19B196SurgeMareOracle, 0, 5, []string{"U"})
	effectstest.EmitNoncombatDamage(g, mare, me.ID, 1)
	if len(g.PendingChoices) != 0 {
		t.Fatal("Surge Mare asked about damage dealt to its own controller")
	}
	hand := me.Hand.Size()
	effectstest.EmitNoncombatDamage(g, mare, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("declining still drew: hand %d -> %d", hand, me.Hand.Size())
	}
	effectstest.EmitNoncombatDamage(g, mare, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d after the discard", me.Hand.Size(), hand)
	}
}

func TestM19B196SurgeMarePumpsForTwoMinusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mare := effectstest.B12Push(g, me.ID, "Surge Mare", "Creature — Horse Fish", m19B196SurgeMareOracle, 0, 5)
	effectstest.ApaMana(me, "C", "U")
	effectstest.B16Activate(t, g, me.ID, mare, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.EffectivePTOf(t, g, mare); p != 2 || tough != 3 {
		t.Errorf("the Mare is %d/%d, want 2/3", p, tough)
	}
}

func TestM19B196ElvishClancallerPumpsOtherElvesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	clan := effectstest.B12Push(g, me.ID, "Elvish Clancaller", "Creature — Elf Druid", m19B196ClancallerOracle, 1, 1)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m19B196CreatureType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf Warrior", 2, 2)
	for _, tc := range []struct {
		name       string
		id         uuid.UUID
		power, tou int
	}{{"the Clancaller", clan, 1, 1}, {"another Elf", elf, 3, 3}, {"a Bear", bear, 2, 2}, {"an opposing Elf", theirs, 2, 2}} {
		if p, tough := effectstest.EffectivePTOf(t, g, tc.id); p != tc.power || tough != tc.tou {
			t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.tou)
		}
	}
}

func TestM19B196ElvishClancallerSearchesForAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	clan := effectstest.PushCatalogPermanent(g, me.ID, "Elvish Clancaller", "Creature — Elf Druid", m19B196ClancallerOracle, false)
	other := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: other, Name: "Elvish Clancaller", TypeLine: "Creature — Elf Druid", OracleID: m19B196ClancallerOracle,
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "C", "C", "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, clan, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {4}{G}{G} with five mana")
	}
	effectstest.ApaMana(me, "G")
	effectstest.ApaActivate(t, g, me, clan, 0, game.ActivateAbilityParams{})
	effectstest.AnswerSearchByID(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(other) {
		t.Fatalf("the second Clancaller is not on the battlefield: %+v", g.PendingChoices)
	}
	if !effectstest.B16Tapped(t, g, clan) {
		t.Error("the first Clancaller did not tap")
	}
	if effectstest.B16Tapped(t, g, other) {
		t.Error("the fetched Clancaller entered tapped")
	}
}

func TestM19B196ArcadesDrawsForADefenderAndTurnsToughnessIntoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	m19B196Push(g, me.ID, "Arcades, the Strategist", "Legendary Creature — Elder Dragon", m19B196ArcadesOracle, 3, 5, []string{"G", "W", "U"}, "flying", "vigilance")
	wall := m19B196Push(g, me.ID, "Wall", m19B196DefenderWallTypeLine, "", 0, 4, nil, "defender")
	bear := effectstest.B12Creature(g, me.ID, "Glass Cannon", m19B196CreatureType, 4, 1)
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, wall, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if want := before - (4 + 4); opp.Life != want {
		t.Errorf("opponent at %d, want %d (the Wall deals its toughness 4, the Bear its power 4)", opp.Life, want)
	}
}

func TestM19B196ArcadesLeavesOpposingDefendersAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	m19B196Push(g, me.ID, "Arcades, the Strategist", "Legendary Creature — Elder Dragon", m19B196ArcadesOracle, 3, 5, []string{"G", "W", "U"}, "flying", "vigilance")
	theirs := m19B196Push(g, opp.ID, "Their Wall", m19B196DefenderWallTypeLine, "", 0, 4, nil, "defender")
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(theirs, me.ID); err == nil {
		t.Error("an opposing defender was allowed to attack")
	}
}

func TestM19B196ArcadesDrawsWhenACreatureWithDefenderEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	m19B196Push(g, me.ID, "Arcades, the Strategist", "Legendary Creature — Elder Dragon", m19B196ArcadesOracle, 3, 5, []string{"G", "W", "U"}, "flying", "vigilance")
	hand := me.Hand.Size()
	wall := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: wall, Name: "Wall", TypeLine: m19B196DefenderWallTypeLine, Power: 0, Toughness: 4,
			Keywords: []string{"defender"}, Owner: me.ID, Controller: me.ID})
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: wall, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d, want %d after a defender entered", me.Hand.Size(), hand+1)
	}
	plain := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: plain, Name: "Bear", TypeLine: m19B196CreatureType, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: plain, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	effectstest.PassPriorityAroundTable(t, g)
	theirs := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: theirs, Name: "Their Wall", TypeLine: m19B196DefenderWallTypeLine, Power: 0, Toughness: 4,
			Keywords: []string{"defender"}, Owner: opp.ID, Controller: opp.ID})
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirs, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: only your own creature with defender draws", me.Hand.Size(), hand+1)
	}
}

func TestM19B196PatientRebuildingMillsAndDrawsPerLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Patient Rebuilding", "Enchantment", m19B196RebuildingOracle, false)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.SeedLibraryTop(opp, "Spell", "Sorcery", "{1}")
	effectstest.SeedLibraryTop(opp, "Forest", "Basic Land — Forest", "")
	effectstest.SeedLibraryTop(opp, "Island", "Basic Land — Island", "")
	oppLib := opp.Library.Size()
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLib - opp.Library.Size(); got != 3 {
		t.Errorf("the opponent milled %d cards, want 3", got)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (two lands milled)", got)
	}
}

func TestM19B196DwindleShrinksAndDestroysABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", m19B196CreatureType, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", m19B196CreatureType, 4, 4)
	effectstest.AuraCast(t, g, "Dwindle", m19B196DwindleOracle, blocker)
	if p, tough := effectstest.EffectivePTOf(t, g, blocker); p != -2 || tough != 4 {
		t.Fatalf("the enchanted creature is %d/%d, want -2/4", p, tough)
	}
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Error("the enchanted creature survived blocking")
	}
}

func TestM19B196DwindleDoesNothingWhenTheEnchantedCreatureAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Dwindle", m19B196DwindleOracle, host)
	effectstest.BrAttack(t, g, host)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(host) {
		t.Error("the enchanted creature died without blocking")
	}
}

func TestM19B196MetamorphicAlterationMakesTheHostACopyOfTheChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.AuraBear(g, me.ID)
	original := m19B196Push(g, opp.ID, "Surge Mare", "Creature — Horse Fish", m19B196SurgeMareOracle, 0, 5, []string{"U"})
	aura := effectstest.CastCatalogSpell(t, g, "Metamorphic Alteration", effectstest.AuraTypeLine, m19B196AlterationOracle, effectstest.CardRefs(host))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no choose-a-creature prompt: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 1 {
		t.Errorf("the choice is optional (min %d)", c.ChooseMin)
	}
	if !effectstest.HasID(c.ChooseCards, original) || !effectstest.HasID(c.ChooseCards, host) {
		t.Errorf("offered %v, want every creature on the battlefield", c.ChooseCards)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{original}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, host)
	if !ok {
		t.Fatal("the host is gone")
	}
	if eff := card.Effective(); eff.Name != "Surge Mare" || eff.Power != 0 || eff.Toughness != 5 {
		t.Errorf("the host is %q %d/%d, want a 0/5 Surge Mare", eff.Name, eff.Power, eff.Toughness)
	}
	// The Aura leaving ends the copy.
	effectstest.WoeB1Destroy(t, g, aura)
	if g.Battlefield.Contains(aura) {
		t.Fatal("the Aura could not be removed")
	}
	if card, ok := effectstest.BattlefieldCard(g, host); !ok || card.Effective().Name != "Bear" {
		t.Errorf("the host did not return to being a Bear: %+v", card.Effective().Name)
	}
}
