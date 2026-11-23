package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exert_cost_cards_test.go — ADR 0130 §4, PR 4: the exert cost
// component on activated and mana abilities, and its eight cards.

const (
	stewardOfSolidarityOracle  = "dcaf3c6f-06e2-4762-82aa-625113e375a1"
	prideSovereignOracle       = "a131c32d-2b4f-4ee4-aab2-ab05ab010978"
	basriTomorrowsOracle       = "d8e1e9ba-708c-4f54-abc3-817004a2f2ac"
	ferventPaincasterOracle    = "1d569df1-23cf-4e01-8ef9-a1a8b815f11e"
	hopeTenderOracle           = "6f377da9-7d7b-4407-8846-75a12b4dd72d"
	angelOfCondemnationOracle  = "2317a33f-665e-4cbc-bb5c-16b912ac2eef"
	oasisRitualistOracle       = "de89ef8f-ef6a-4f95-9c38-5debc06e1d74"
	arenaOfGloryOracle         = "63dfe794-5f56-41ec-9883-5523b41cc3e0"
	exertCostCreatureTypeLine  = "Creature — Human Warrior"
	exertCostRitualistTypeLine = "Creature — Snake Druid"
)

// exertSkips counts the plain next-untap markers on `id` naming `player`.
func exertSkips(g *game.Game, id, player uuid.UUID) int {
	c := effectstest.FindBattlefieldCardForTest(g, id)
	if c == nil {
		return 0
	}
	n := 0
	for _, s := range c.NextUntapSkips {
		if s.While == nil && s.Player == player {
			n++
		}
	}
	return n
}

// costExertEvents returns the EventExerts for `id` and whether any of
// them named an attack target.
func costExertEvents(g *game.Game, id uuid.UUID) (n int, targeted bool) {
	for _, ev := range g.Events {
		if ev.Kind == game.EventExert && ev.CardID == id {
			n++
			if ev.Target != uuid.Nil {
				targeted = true
			}
		}
	}
	return n, targeted
}

// Every card is registered and Full (Arena of Glory since #3844, when
// its rider stopped depending on the payment mode), and the exert sits
// on the right row.
func TestExertCostCardsDeclareTheirCosts(t *testing.T) {
	activated := map[string]int{
		stewardOfSolidarityOracle: 0, prideSovereignOracle: 0, basriTomorrowsOracle: 0,
		ferventPaincasterOracle: 1, hopeTenderOracle: 1, angelOfCondemnationOracle: 1,
	}
	for oracle, row := range activated {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
		for i, ab := range spec.Activated {
			if ab.Cost.Exert != (i == row) {
				t.Errorf("%s row %d: exert %v", spec.Name, i, ab.Cost.Exert)
			}
		}
	}
	for oracle, full := range map[string]bool{oasisRitualistOracle: true, arenaOfGloryOracle: true} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if full != (spec.Completeness == effects.CompletenessFull) {
			t.Errorf("%s: completeness %v", spec.Name, spec.Completeness)
		}
		if len(spec.ManaAbilities) != 2 || spec.ManaAbilities[0].Cost.Exert || !spec.ManaAbilities[1].Cost.Exert {
			t.Errorf("%s: the second mana row, and only it, exerts", spec.Name)
		}
	}
}

// Steward of Solidarity: the activation taps and exerts it (one marker
// keyed to the activator, an EventExert with no attack target), and
// makes a 1/1 Warrior with vigilance.
func TestStewardOfSolidarityExertsAndMakesAWarrior(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	steward := effectstest.PushCatalogPermanent(g, me.ID, "Steward of Solidarity", exertCostCreatureTypeLine, stewardOfSolidarityOracle, false)
	effectstest.P7Activate(t, g, me, steward, 0, game.ActivateAbilityParams{})
	if !effectstest.IsTapped(g, steward) {
		t.Error("the Steward did not tap")
	}
	if exertSkips(g, steward, me.ID) != 1 {
		t.Errorf("markers naming the activator = %d, want 1", exertSkips(g, steward, me.ID))
	}
	n, targeted := costExertEvents(g, steward)
	if n != 1 || targeted {
		t.Errorf("exert events = %d (targeted %v), want one with no attack target", n, targeted)
	}
	if !g.ExertedThisTurn(steward) {
		t.Error("the tally does not record the exert")
	}
	var warrior *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Warrior" && c.Controller == me.ID {
			warrior = c
		}
	}
	if warrior == nil {
		t.Fatal("no Warrior token")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, warrior.InstanceID, "vigilance") {
		t.Error("the Warrior has no vigilance")
	}
}

// Ruling 7, CR 701.43b: an exert cost can be paid again by a creature
// already exerted this turn, and the second exert adds no marker.
func TestExertCostCanBePaidAgainTheSameTurn(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	steward := effectstest.PushCatalogPermanent(g, me.ID, "Steward of Solidarity", exertCostCreatureTypeLine, stewardOfSolidarityOracle, false)
	effectstest.P7Activate(t, g, me, steward, 0, game.ActivateAbilityParams{})
	effectstest.FindBattlefieldCardForTest(g, steward).Tapped = false // something untapped it
	effectstest.P7Activate(t, g, me, steward, 0, game.ActivateAbilityParams{})
	if n, _ := costExertEvents(g, steward); n != 2 {
		t.Errorf("exert events = %d, want 2", n)
	}
	if got := exertSkips(g, steward, me.ID); got != 1 {
		t.Errorf("markers = %d, want 1: every exert expires at the same untap step", got)
	}
	if got := effectstest.BattlefieldNamed(g, me.ID, "Warrior"); got != 2 {
		t.Errorf("Warriors = %d, want 2", got)
	}
}

// "Whenever you exert a creature" sees an exert paid as a cost, and its
// trigger resolves before the ability whose cost caused it.
func TestResoluteSurvivorsSeesAnExertPaidAsACost(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	survivors := effectstest.PushCatalogPermanent(g, me.ID, "Resolute Survivors", exertCostCreatureTypeLine, resoluteSurvivorsOracle, false)
	steward := effectstest.PushCatalogPermanent(g, me.ID, "Steward of Solidarity", exertCostCreatureTypeLine, stewardOfSolidarityOracle, false)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, steward, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if len(g.StackMeta) != 2 || effectstest.TriggerOnStack(g, survivors) == nil {
		t.Fatalf("stack has %d items, want the ability and the Survivors' payoff", len(g.StackMeta))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	if effectstest.BattlefieldNamed(g, me.ID, "Warrior") != 1 {
		t.Error("the Steward's ability did not resolve")
	}
}

// The enumerator names the exert on the activation's cost.
func TestExertCostMoveNamesTheExert(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	steward := effectstest.PushCatalogPermanent(g, me.ID, "Steward of Solidarity", exertCostCreatureTypeLine, stewardOfSolidarityOracle, false)
	ritualist := effectstest.PushCatalogPermanent(g, me.ID, "Oasis Ritualist", exertCostRitualistTypeLine, oasisRitualistOracle, false)
	var stewardExert, ritualistExert, ritualistPlain bool
	for _, m := range legal.EnumerateFor(g, me.ID) {
		exerts := m.Cost != nil && m.Cost.Exert
		switch {
		case m.Type == "activate_ability" && m.Source == steward:
			stewardExert = exerts
		case m.Type == "activate_mana_ability" && m.Source == ritualist && exerts:
			ritualistExert = true
		case m.Type == "activate_mana_ability" && m.Source == ritualist:
			ritualistPlain = true
		}
	}
	if !stewardExert {
		t.Error("the Steward's activation does not carry cost.exert")
	}
	if !ritualistExert || !ritualistPlain {
		t.Errorf("Ritualist moves: exert %v, plain %v; want both", ritualistExert, ritualistPlain)
	}
}

// Pride Sovereign: +1/+1 per OTHER Cat you control, and its exert
// makes two lifelink Cats that grow it.
func TestPrideSovereignGrowsWithItsCats(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	sov := effectstest.PushCatalogPermanent(g, me.ID, "Pride Sovereign", "Creature — Cat", prideSovereignOracle, false)
	c := effectstest.FindBattlefieldCardForTest(g, sov)
	c.Power, c.Toughness = 2, 2
	effectstest.B31Push(g, opp.ID, "Their Cat", "Creature — Cat", "", "", 1, 1)
	if p, tough := effectstest.EffectivePower(t, g, sov), effectstest.EffectiveToughness(t, g, sov); p != 2 || tough != 2 {
		t.Fatalf("alone: %d/%d, want 2/2 (an opponent's Cat doesn't count)", p, tough)
	}
	effectstest.ApaMana(me, "W")
	effectstest.P7Activate(t, g, me, sov, 0, game.ActivateAbilityParams{})
	if exertSkips(g, sov, me.ID) != 1 {
		t.Error("the Sovereign was not exerted")
	}
	cats := 0
	for _, card := range g.Battlefield.Cards {
		if card.Name == "Cat" && card.Controller == me.ID {
			cats++
			if !effectstest.HasKeywordOnBattlefield(t, g, card.InstanceID, "lifelink") {
				t.Error("a Cat token without lifelink")
			}
		}
	}
	if cats != 2 {
		t.Fatalf("Cat tokens = %d, want 2", cats)
	}
	if p, tough := effectstest.EffectivePower(t, g, sov), effectstest.EffectiveToughness(t, g, sov); p != 4 || tough != 4 {
		t.Errorf("with two Cats: %d/%d, want 4/4", p, tough)
	}
}

// Basri: the exert makes a lifelink Cat; cycling it gives the Cats you
// control hexproof and indestructible until end of turn.
func TestBasriExertsForACatAndCyclesForProtection(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	basri := effectstest.PushCatalogPermanent(g, me.ID, "Basri, Tomorrow's Champion", "Legendary Creature — Human Knight", basriTomorrowsOracle, false)
	effectstest.ApaMana(me, "W")
	effectstest.P7Activate(t, g, me, basri, 0, game.ActivateAbilityParams{})
	if exertSkips(g, basri, me.ID) != 1 || effectstest.BattlefieldNamed(g, me.ID, "Cat") != 1 {
		t.Fatal("Basri's exert did not make a Cat")
	}
	theirCat := effectstest.B31Push(g, opp.ID, "Their Cat", "Creature — Cat", "", "", 1, 1)

	inHand := effectstest.HandSpell(me, "Basri, Tomorrow's Champion", "Legendary Creature — Human Knight", "{W}")
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == inHand {
			me.Hand.Cards[i].OracleID = basriTomorrowsOracle
		}
	}
	effectstest.ApaMana(me, "C", "C", "W")
	effectstest.P7Activate(t, g, me, inHand, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var cat uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cat" && c.Controller == me.ID {
			cat = c.InstanceID
		}
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, cat, "hexproof") || !effectstest.HasKeywordOnBattlefield(t, g, cat, "indestructible") {
		t.Error("my Cat lacks hexproof or indestructible after the cycle")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirCat, "hexproof") {
		t.Error("an opponent's Cat gained hexproof")
	}
}

// Fervent Paincaster's exert row pings a creature; its plain row pings a
// player and exerts nothing.
func TestFerventPaincasterPings(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	caster := effectstest.PushCatalogPermanent(g, me.ID, "Fervent Paincaster", "Creature — Human Wizard", ferventPaincasterOracle, false)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	life := opp.Life
	effectstest.P7Activate(t, g, me, caster, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != life-1 || exertSkips(g, caster, me.ID) != 0 {
		t.Fatalf("plain row: life %d (want %d), markers %d (want 0)", opp.Life, life-1, exertSkips(g, caster, me.ID))
	}
	effectstest.FindBattlefieldCardForTest(g, caster).Tapped = false
	effectstest.P7Activate(t, g, me, caster, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if effectstest.DamageMarkedOn(g, bear) != 1 || exertSkips(g, caster, me.ID) != 1 {
		t.Errorf("exert row: damage %d (want 1), markers %d (want 1)", effectstest.DamageMarkedOn(g, bear), exertSkips(g, caster, me.ID))
	}
}

// Hope Tender's exert row untaps two lands.
func TestHopeTenderUntapsTwoLands(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	tender := effectstest.PushCatalogPermanent(g, me.ID, "Hope Tender", "Creature — Human Druid", hopeTenderOracle, false)
	a := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	b := effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	effectstest.FindBattlefieldCardForTest(g, a).Tapped = true
	effectstest.FindBattlefieldCardForTest(g, b).Tapped = true
	effectstest.ApaMana(me, "C")
	effectstest.P7Activate(t, g, me, tender, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}})
	if effectstest.IsTapped(g, a) || effectstest.IsTapped(g, b) {
		t.Error("the lands are still tapped")
	}
	if exertSkips(g, tender, me.ID) != 1 {
		t.Error("Hope Tender was not exerted")
	}
}

// Angel of Condemnation's exert row holds the creature until the Angel
// leaves; it comes back then.
func TestAngelOfCondemnationExilesUntilItLeaves(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	angel := effectstest.PushCatalogPermanent(g, me.ID, "Angel of Condemnation", "Creature — Angel", angelOfCondemnationOracle, false)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	effectstest.ApaMana(me, "C", "C", "W")
	effectstest.P7Activate(t, g, me, angel, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if effectstest.FindBattlefieldCardForTest(g, bear) != nil {
		t.Fatal("the Bear is still on the battlefield")
	}
	if exertSkips(g, angel, me.ID) != 1 {
		t.Error("the Angel was not exerted")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(angel) })
	g.RunStateChecksForTest()
	if effectstest.BattlefieldNamed(g, opp.ID, "Bear") != 1 {
		t.Error("the Bear did not return when the Angel left")
	}
}

// Oasis Ritualist: the exert row adds two mana of one colour and exerts
// it; the auto-tapper pays only with the plain row.
func TestOasisRitualistExertsForTwoManaAndIsNeverAutoExerted(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	rit := effectstest.PushCatalogPermanent(g, me.ID, "Oasis Ritualist", exertCostRitualistTypeLine, oasisRitualistOracle, false)

	two := effectstest.HandSpell(me, "Two Green", "Instant", "{G}{G}")
	if effectstest.CastMove(g, me.ID, two) {
		t.Fatal("{G}{G} is offered: only an exert could pay it, and the auto-tapper never exerts")
	}
	one := effectstest.HandSpell(me, "One Green", "Instant", "{G}")
	if err := g.CastSpell(me.ID, one, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped {G}: %v", err)
	}
	if exertSkips(g, rit, me.ID) != 0 {
		t.Fatal("the auto-tapper exerted the Ritualist")
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.FindBattlefieldCardForTest(g, rit).Tapped = false
	if err := g.ActivateManaAbility(me.ID, rit, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("exert row: %v", err)
	}
	if pick := effectstest.PendingOfKind(g, game.PendingChoiceMana); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
			t.Fatal(err)
		}
	}
	if got := effectstest.PoolColors(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("pool = %v, want {G}{G}", got)
	}
	if exertSkips(g, rit, me.ID) != 1 {
		t.Error("the Ritualist was not exerted")
	}
}

// Arena of Glory: {R}, {T}, exert for {R}{R}. Exerting a land is not
// exerting a creature, so Resolute Survivors does not trigger.
func TestArenaOfGloryExertsForTwoRed(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Resolute Survivors", exertCostCreatureTypeLine, resoluteSurvivorsOracle, false)
	arena := effectstest.PushCatalogPermanent(g, me.ID, "Arena of Glory", "Land", arenaOfGloryOracle, false)
	effectstest.ApaMana(me, "R")
	if err := g.ActivateManaAbility(me.ID, arena, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("exert row: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("pool = %v, want {R}{R}", got)
	}
	if exertSkips(g, arena, me.ID) != 1 || !effectstest.IsTapped(g, arena) {
		t.Error("the Arena was not tapped and exerted")
	}
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Size() != 0 || me.Life != life {
		t.Error("a land's exert triggered \"whenever you exert a creature\"")
	}
}
