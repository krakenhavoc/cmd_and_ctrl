package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set MAT, batch 1 (#5476): March of the Machine: The Aftermath.

const (
	matCoppercoatOracle  = "9f6fc734-4935-42d9-93a0-1b5ad8d8d277"
	matThopteryxOracle   = "c27e2d6f-498c-42d8-aa1a-ac6402bf3f9a"
	matImmersionOracle   = "aa6c4079-72e1-4b2e-bd26-9399298c6105"
	matRuinsOracle       = "d1f10cca-8dfa-4ea5-b227-4446cd8514a8"
	matSamutOracle       = "178b3fdf-dbee-4245-a811-b21ff8a24f66"
	matAnimistsOracle    = "4940783f-2534-41c1-b614-5dabb0dff405"
	matBlotOutOracle     = "9609cd5c-b537-4cb2-9dde-4b12795531f7"
	matFilterOutOracle   = "8c95d823-6bbd-4484-8f38-b1bb20403085"
	matOathswornOracle   = "ff3936bb-3281-4260-84de-1f65dbfbe993"
	matMarkovBaronOracle = "bafc2e11-5f1c-46f1-a852-2330e099b8e2"
)

// matPush puts a creature on the battlefield (not entered this turn, not
// summoning sick).
func matPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// matEntered records that the permanent entered the battlefield this turn.
func matEntered(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id}) })
}

// matWardPrompt casts Doom Blade at victim as opp and reports whether the
// opponent is asked to pay a ward cost; the spell is then settled either way.
func matWardPrompt(t *testing.T, g *game.Game, opp *game.Player, victim uuid.UUID) bool {
	t.Helper()
	castAtWardedCreature(t, g, opp, victim)
	asked := false
	for i := 0; i < 8; i++ {
		if hasPayUnlessFor(g, opp.ID) {
			asked = true
			answerPayUnless(t, g, opp.ID, false)
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	return asked
}

// --- Coppercoat Vanguard ------------------------------------------------

func TestCoppercoatVanguardPumpsAndWardsOtherHumansYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	vanguard := matPush(g, me.ID, "Coppercoat Vanguard", "Creature — Human Soldier", matCoppercoatOracle, 2, 2)
	human := matPush(g, me.ID, "Peasant", "Creature — Human Peasant", "", 1, 1)
	bear := matPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := matPush(g, opp.ID, "Their Human", "Creature — Human Soldier", "", 1, 1)

	if got := effectivePower(t, g, human); got != 2 {
		t.Errorf("another Human's power = %d, want 2 (+1/+0)", got)
	}
	if got := effectiveToughness(t, g, human); got != 1 {
		t.Errorf("another Human's toughness = %d, want 1", got)
	}
	if got := effectivePower(t, g, vanguard); got != 2 {
		t.Errorf("the Vanguard's power = %d, want 2 (\"other\" excludes it)", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("a Bear's power = %d, want 2", got)
	}
	if got := effectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opponent's Human's power = %d, want 1 (\"you control\")", got)
	}

	if !matWardPrompt(t, g, opp, human) {
		t.Error("another Human you control has ward {1}: the opponent should be asked to pay")
	}
	if !g.Battlefield.Contains(human) {
		t.Error("a declined ward counters the spell: the Human should live")
	}
}

func TestCoppercoatVanguardGivesNoWardToItselfNonHumansOrOpposingHumans(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	vanguard := matPush(g, me.ID, "Coppercoat Vanguard", "Creature — Human Soldier", matCoppercoatOracle, 2, 2)
	bear := matPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := matPush(g, opp.ID, "Their Human", "Creature — Human Soldier", "", 1, 1)
	for name, id := range map[string]uuid.UUID{"the Vanguard": vanguard, "a Bear": bear, "an opposing Human": theirs} {
		if matWardPrompt(t, g, opp, id) {
			t.Errorf("%s has no ward, but the opponent was asked to pay", name)
		}
	}
}

// --- Gold-Forged Thopteryx ----------------------------------------------

func TestGoldForgedThopteryxWardsEachLegendaryPermanentYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	matPush(g, me.ID, "Gold-Forged Thopteryx", "Artifact Creature — Dinosaur Thopter", matThopteryxOracle, 1, 3)
	legend := seedLegendaryCreature(g, me.ID, "Legend")
	plain := matPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirLegend := seedLegendaryCreature(g, opp.ID, "Their Legend")

	if !matWardPrompt(t, g, opp, legend) {
		t.Error("a legendary permanent you control has ward {2}")
	}
	if matWardPrompt(t, g, opp, plain) {
		t.Error("a nonlegendary creature has no ward")
	}
	if matWardPrompt(t, g, opp, theirLegend) {
		t.Error("an opponent's legendary creature has no ward from your Thopteryx")
	}
}

func TestGoldForgedThopteryxWardCostsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	matPush(g, me.ID, "Gold-Forged Thopteryx", "Artifact Creature — Dinosaur Thopter", matThopteryxOracle, 1, 3)
	legend := seedLegendaryCreature(g, me.ID, "Legend")
	castAtWardedCreature(t, g, opp, legend)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward prompt")
	}
	if pc := answerPayUnless(t, g, opp.ID, false); pc.PayCost != "{2}" {
		t.Errorf("ward cost = %q, want {2}", pc.PayCost)
	}
}

// --- Leyline Immersion --------------------------------------------------

func TestLeylineImmersionEnchantsOnlyALegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := matPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	legend := seedLegendaryCreature(g, me.ID, "Legend")
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Leyline Immersion", TypeLine: "Enchantment — Aura",
		ManaCost: "{3}{G}", OracleID: matImmersionOracle, Owner: me.ID, Controller: me.ID})
	aura := me.Hand.Cards[len(me.Hand.Cards)-1].InstanceID
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("Leyline Immersion enchanted a nonlegendary creature")
	}
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{ForceCast: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: legend}}}); err != nil {
		t.Fatalf("cast on a legendary creature: %v", err)
	}
}

func TestLeylineImmersionGivesWardAndRestrictedFiveManaToTheHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	legend := seedLegendaryCreature(g, me.ID, "Legend")
	other := matPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Leyline Immersion", "Enchantment — Aura", matImmersionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: legend}})
	passPriorityAroundTable(t, g)

	if got := manaAbilityProduced(t, g, legend); len(got) != 1 {
		t.Fatalf("the enchanted creature's mana abilities = %v, want the one granted", got)
	}
	if got := manaAbilityProduced(t, g, other); len(got) != 0 {
		t.Errorf("an unenchanted creature has mana abilities %v", got)
	}
	tapForMana(t, g, me.ID, legend)
	// Five independent colour picks, answered differently.
	for i, color := range []string{"W", "U", "B", "R", "G"} {
		c := riderLatestManaPick(g, me.ID)
		if c == nil {
			t.Fatalf("ran out of picks before %s: %d picks, want 5", color, i)
		}
		if err := g.ResolveManaChoice(c.ID, me.ID, color); err != nil {
			t.Fatalf("ResolveManaChoice(%s): %v", color, err)
		}
	}
	if me.ManaPool.Total() != 5 {
		t.Fatalf("the pool holds %d mana, want 5", me.ManaPool.Total())
	}
	for _, u := range me.ManaPool.Units() {
		if len(u.Restrictions) == 0 {
			t.Errorf("mana %+v carries no spend restriction", u)
		}
	}
	// It is the creature's {T}: tapped now.
	if !tapCostTapped(g, legend) {
		t.Error("the {T} did not tap the creature")
	}

	// Ward {2} on the enchanted creature.
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if !matWardPrompt(t, g, opp, legend) {
		t.Error("the enchanted creature has ward {2}")
	}
	if matWardPrompt(t, g, opp, other) {
		t.Error("an unenchanted creature has no ward")
	}
}

// --- Drannith Ruins -----------------------------------------------------

func matRuins(g *game.Game, me *game.Player) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Drannith Ruins", TypeLine: "Land", OracleID: matRuinsOracle,
		Owner: me.ID, Controller: me.ID,
	})
}

func matActivateRuins(g *game.Game, me *game.Player, ruins, target uuid.UUID, mana string) error {
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			return err
		}
	}
	return g.ActivateCatalogAbility(me.ID, ruins, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}, Strict: true})
}

func TestDrannithRuinsPutsTwoCountersOnANonHumanThatEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ruins := matRuins(g, me)
	fresh := matPush(g, me.ID, "Fresh Bear", "Creature — Bear", "", 2, 2)
	matEntered(g, fresh)

	tapForMana(t, g, me.ID, ruins)
	if poolColors(me)[0] != "C" || me.ManaPool.Total() != 1 {
		t.Fatalf("pool %v, want {C}", poolColors(me))
	}
	// The land is tapped now; a second Ruins pays and activates. One
	// floating mana is not the {2} it costs.
	ruins2 := matRuins(g, me)
	if err := matActivateRuins(g, me, ruins2, fresh, ""); err == nil {
		t.Fatal("an activation paid with one mana was accepted")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := matActivateRuins(g, me, ruins2, fresh, "{C}{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, fresh, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

func TestDrannithRuinsRefusesHumansAndCreaturesThatDidNotEnterThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ruins := matRuins(g, me)
	human := matPush(g, me.ID, "Fresh Human", "Creature — Human Peasant", "", 1, 1)
	matEntered(g, human)
	old := matPush(g, me.ID, "Old Bear", "Creature — Bear", "", 2, 2)
	if err := matActivateRuins(g, me, ruins, human, "{C}{C}"); err == nil {
		t.Error("a Human is not a legal target")
	}
	if err := matActivateRuins(g, me, ruins, old, ""); err == nil {
		t.Error("a creature that did not enter this turn is not a legal target")
	}
}

func TestDrannithRuinsDoesNothingIfTheTargetLeftInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ruins := matRuins(g, me)
	fresh := matPush(g, me.ID, "Fresh Bear", "Creature — Bear", "", 2, 2)
	matEntered(g, fresh)
	if err := matActivateRuins(g, me, ruins, fresh, "{C}{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(fresh); err != nil {
			t.Fatalf("bury: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fresh) {
		t.Error("the creature should be gone")
	}
}

// --- Samut, Vizier of Naktamun ------------------------------------------

func TestSamutDrawsWhenACreatureThatEnteredThisTurnConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	samut := matPush(g, me.ID, "Samut, Vizier of Naktamun", "Legendary Creature — Human Warrior Cleric", matSamutOracle, 2, 3)
	fresh := matPush(g, me.ID, "Fresh Bear", "Creature — Bear", "", 2, 2)
	matEntered(g, fresh)
	old := matPush(g, me.ID, "Old Bear", "Creature — Bear", "", 2, 2)
	theirs := matPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	matEntered(g, theirs)

	hand := len(me.Hand.Cards)
	dealCombatDamageToPlayer(g, old, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Error("a creature that did not enter this turn must not draw a card")
	}
	dealCombatDamageToPlayer(g, theirs, me.ID, 2)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Error("an opponent's creature must not draw a card")
	}
	dealCombatDamageToPlayer(g, fresh, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand = %d, want %d after a new creature connected", len(me.Hand.Cards), hand+1)
	}
	// Samut herself entered this turn too, and draws for her own damage.
	matEntered(g, samut)
	dealCombatDamageToPlayer(g, samut, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+2 {
		t.Errorf("hand = %d, want %d after Samut herself connected", len(me.Hand.Cards), hand+2)
	}
	// Noncombat damage does not count.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: fresh, Target: opp.ID, Amount: 2})
	})
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+2 {
		t.Error("noncombat damage drew a card")
	}
}

func TestSamutHasFirstStrikeVigilanceAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	samut := matPush(g, me.ID, "Samut, Vizier of Naktamun", "Legendary Creature — Human Warrior Cleric", matSamutOracle, 2, 3)
	for _, kw := range []string{"first strike", "vigilance", "haste"} {
		if !effectiveAbilitiesContain(t, g, samut, kw) {
			t.Errorf("Samut lacks %s", kw)
		}
	}
}

// --- Animist's Might ----------------------------------------------------

func matCastAnimists(t *testing.T, g *game.Game, me *game.Player, mana string, mine, theirs uuid.UUID) error {
	t.Helper()
	advanceToMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatal(err)
	}
	_, err := msc17Cast(t, g, me, game.Card{Name: "Animist's Might", TypeLine: "Sorcery", ManaCost: "{2}{G}",
		OracleID: matAnimistsOracle}, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}}})
	return err
}

func TestAnimistsMightCostsTwoLessTargetingALegendaryCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legend := seedLegendaryCreature(g, me.ID, "Legend") // power 2
	victim := matPush(g, opp.ID, "Their Wall", "Creature — Wall", "", 0, 4)
	if err := matCastAnimists(t, g, me, "{G}", legend, victim); err != nil {
		t.Fatalf("with a legendary creature the spell costs {G}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("twice a 2-power creature is 4 damage, which kills a 0/4")
	}
}

func TestAnimistsMightCostsFullPriceForAnOrdinaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := matPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	victim := matPush(g, opp.ID, "Their Wall", "Creature — Wall", "", 0, 4)
	if err := matCastAnimists(t, g, me, "{G}", bear, victim); err == nil {
		t.Fatal("a nonlegendary creature must not get the discount")
	}
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	bear2 := matPush(g2, me2.ID, "Bear", "Creature — Bear", "", 2, 2)
	victim2 := matPush(g2, opp2.ID, "Their Wall", "Creature — Wall", "", 0, 3)
	if err := matCastAnimists(t, g2, me2, "{C}{C}{G}", bear2, victim2); err != nil {
		t.Fatalf("full price: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(victim2) {
		t.Error("4 damage should kill a 0/3")
	}
}

func TestAnimistsMightRefusesYourOwnCreatureAsTheSecondTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	legend := seedLegendaryCreature(g, me.ID, "Legend")
	bear := matPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := matCastAnimists(t, g, me, "{C}{C}{G}", legend, bear); err == nil {
		t.Error("the damaged creature must be one you don't control")
	}
}

func TestAnimistsMightHitsAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := matPush(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker",
		TypeLine: "Legendary Planeswalker — Test", Owner: opp.ID, Controller: opp.ID})
	addCountersForTest(g, walker, game.CounterLoyalty, 4)
	if err := matCastAnimists(t, g, me, "{C}{C}{G}", bear, walker); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("6 damage should remove a 4-loyalty planeswalker")
	}
}

// --- Blot Out -----------------------------------------------------------

func TestBlotOutMakesTheOpponentExileTheirGreatestManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := socB24Perm(g, opp.ID, "Big Beast", "Creature — Beast", "{3}{G}{G}", 5, 5)
	walker := socB24Perm(g, opp.ID, "Tall Walker", "Legendary Planeswalker — Test", "{3}{U}{U}", 0, 0)
	addCountersForTest(g, walker, game.CounterLoyalty, 4)
	small := socB24Perm(g, opp.ID, "Small Bear", "Creature — Bear", "{1}{G}", 2, 2)
	rock := socB24Perm(g, opp.ID, "Big Rock", "Artifact", "{6}", 0, 0)
	mine := socB24Perm(g, me.ID, "Mine Giant", "Creature — Giant", "{7}", 7, 7)
	card := game.Card{Name: "Blot Out", TypeLine: "Instant", ManaCost: "{2}{B}", OracleID: matBlotOutOracle}
	if _, err := msc17Cast(t, g, me, card, game.CastSpellParams{Targets: []game.TargetRef{msc17TargetPlayer(me.ID)}}); err == nil {
		t.Fatal("Blot Out targeted its own caster")
	}
	if _, err := msc17Cast(t, g, me, card, game.CastSpellParams{Targets: []game.TargetRef{msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster was asked to choose: the opponent chooses")
	}
	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the opponent was not asked")
	}
	if !hasID(pick.ChooseCards, big) || !hasID(pick.ChooseCards, walker) || len(pick.ChooseCards) != 2 {
		t.Fatalf("offered %v: only the creature and planeswalker tied at mana value 5", pick.ChooseCards)
	}
	answerChooseCards(t, g, opp.ID, big)
	if !g.Exile.Contains(big) {
		t.Error("the chosen creature is not exiled")
	}
	for name, id := range map[string]uuid.UUID{"the planeswalker": walker, "the smaller creature": small, "the artifact": rock, "the caster's creature": mine} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s left the battlefield", name)
		}
	}
}

func TestBlotOutExilesTheOnlyGreatest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The permanent is not targeted: only the opponent is.
	king := socB24Perm(g, opp.ID, "King", "Creature — Giant", "{4}{R}", 5, 5)
	card := game.Card{Name: "Blot Out", TypeLine: "Instant", ManaCost: "{2}{B}", OracleID: matBlotOutOracle}
	if _, err := msc17Cast(t, g, me, card, game.CastSpellParams{Targets: []game.TargetRef{msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, opp.ID, king)
	if !g.Exile.Contains(king) {
		t.Error("the only greatest creature should be exiled")
	}
}

// --- Filter Out ---------------------------------------------------------

func TestFilterOutReturnsNoncreatureNonlandPermanentsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := socB24Perm(g, me.ID, "Rock", "Artifact", "{2}", 0, 0)
	ench := socB24Perm(g, opp.ID, "Ward", "Enchantment", "{1}{W}", 0, 0)
	walker := socB24Perm(g, opp.ID, "Walker", "Planeswalker — Test", "{3}", 0, 0)
	addCountersForTest(g, walker, game.CounterLoyalty, 3)
	vehicle := socB24Perm(g, opp.ID, "Golem", "Artifact Creature — Golem", "{4}", 3, 3)
	bear := socB24Perm(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	land := socB24Perm(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Treasure",
		TypeLine: "Token Artifact — Treasure", Owner: me.ID, Controller: me.ID})

	castCatalogSpell(t, g, "Filter Out", "Instant", matFilterOutOracle, nil)
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"the artifact": rock} {
		if g.Battlefield.Contains(id) || !me.Hand.Contains(id) {
			t.Errorf("%s should be back in its owner's hand", name)
		}
	}
	if g.Battlefield.Contains(ench) || !opp.Hand.Contains(ench) {
		t.Error("the opponent's enchantment should be in its owner's hand")
	}
	if g.Battlefield.Contains(walker) || !opp.Hand.Contains(walker) {
		t.Error("the planeswalker is a noncreature nonland permanent and should return")
	}
	for name, id := range map[string]uuid.UUID{"the artifact creature": vehicle, "the creature": bear, "the land": land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s should stay on the battlefield", name)
		}
	}
	if g.Battlefield.Contains(token) {
		t.Error("the Treasure token should have left the battlefield")
	}
}

// --- Ayara's Oathsworn --------------------------------------------------

func matOathsworn(g *game.Game, me *game.Player, counters int) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: "Ayara's Oathsworn", TypeLine: "Creature — Human Knight",
		OracleID: matOathswornOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	if counters > 0 {
		c.Counters = map[string]int{game.CounterPlusOne: counters}
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestAyarasOathswornGrowsAndSearchesOnTheFourthCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Alpha", "Beta", "Gamma")
	oath := matOathsworn(g, me, 0)
	if !effectiveAbilitiesContain(t, g, oath, "menace") {
		t.Error("Ayara's Oathsworn has menace")
	}
	advanceToMain(t, g)
	for want := 1; want <= 3; want++ {
		dealCombatDamageToPlayer(g, oath, opp.ID, 2)
		passPriorityAroundTable(t, g)
		if got := counterCount(g, oath, game.CounterPlusOne); got != want {
			t.Fatalf("after hit %d the Oathsworn has %d counters", want, got)
		}
		if searchChoiceFor(g, me.ID) != nil {
			t.Fatalf("searched with only %d counters", want)
		}
	}
	dealCombatDamageToPlayer(g, oath, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, oath, game.CounterPlusOne); got != 4 {
		t.Fatalf("counters = %d, want 4", got)
	}
	answerSearchNamed(t, g, me.ID, "Beta")
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range me.Hand.Cards {
		if c.Name == "Beta" {
			found = true
		}
	}
	if !found {
		t.Error("the searched card should be in hand")
	}
	// Now it has four: it no longer triggers (intervening "if").
	dealCombatDamageToPlayer(g, oath, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, oath, game.CounterPlusOne); got != 4 {
		t.Errorf("a fifth counter was added (%d)", got)
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("searched again at four counters")
	}
}

func TestAyarasOathswornDoesNotSearchWhenCountersJumpPastFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Alpha", "Beta")
	oath := matOathsworn(g, me, 3)
	pushCatalogPermanent(g, me.ID, "Doubling Season", "Enchantment", doublingSeasonOracle, false)
	advanceToMain(t, g)
	dealCombatDamageToPlayer(g, oath, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, oath, game.CounterPlusOne); got != 5 {
		t.Fatalf("counters = %d, want 5 (doubled)", got)
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("it has five counters, not exactly four: no search")
	}
}

func TestAyarasOathswornIgnoresDamageToAPermanentOrNoncombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oath := matOathsworn(g, me, 0)
	advanceToMain(t, g)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: oath, Target: opp.ID, Amount: 2})
	})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, oath, game.CounterPlusOne); got != 0 {
		t.Errorf("noncombat damage added a counter (%d)", got)
	}
}

// --- Markov Baron -------------------------------------------------------

func TestMarkovBaronPumpsOtherVampiresYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	baron := matPush(g, me.ID, "Markov Baron", "Creature — Vampire Noble", matMarkovBaronOracle, 2, 2)
	vamp := matPush(g, me.ID, "Vamp", "Creature — Vampire", "", 1, 1)
	bear := matPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := matPush(g, opp.ID, "Their Vamp", "Creature — Vampire", "", 1, 1)
	if got := effectivePower(t, g, vamp); got != 2 {
		t.Errorf("another Vampire's power = %d, want 2", got)
	}
	if got := effectiveToughness(t, g, vamp); got != 2 {
		t.Errorf("another Vampire's toughness = %d, want 2", got)
	}
	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{"the Baron": {baron, 2}, "a Bear": {bear, 2}, "an opposing Vampire": {theirs, 1}} {
		if got := effectivePower(t, g, c.id); got != c.want {
			t.Errorf("%s power = %d, want %d", name, got, c.want)
		}
	}
	if !effectiveAbilitiesContain(t, g, baron, "lifelink") {
		t.Error("the Baron has lifelink")
	}
}

func TestMarkovBaronConvokesTwoCreaturesAndABlackMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	soldiers := pushTapCostSoldiers(g, me.ID, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
		t.Fatal(err)
	}
	id, err := castWithTapParams(t, g, "Markov Baron", "Creature — Vampire Noble", "{2}{B}",
		matMarkovBaronOracle, game.CastSpellParams{TapIDs: soldiers, Strict: true})
	if err != nil {
		t.Fatalf("two tapped creatures pay {2} and the floating {B} pays {B}: %v", err)
	}
	for _, s := range soldiers {
		if !tapCostTapped(g, s) {
			t.Error("convoke taps the named creatures")
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Baron did not resolve")
	}
}

func TestMarkovBaronCannotBePaidWithTwoCreaturesAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	soldiers := pushTapCostSoldiers(g, me.ID, 2)
	_, err := castWithTapParams(t, g, "Markov Baron", "Creature — Vampire Noble", "{2}{B}",
		matMarkovBaronOracle, game.CastSpellParams{TapIDs: soldiers, Strict: true})
	if err == nil {
		t.Fatal("two white creatures cannot pay {2}{B}")
	}
}

func TestMarkovBaronHasMadness(t *testing.T) {
	d := newDBOTable(t)
	card := d.handCard(d.victim, "Markov Baron", "Creature — Vampire Noble", "{2}{B}", matMarkovBaronOracle, 2, 2)
	filler := d.filler(d.victim)
	d.theyDiscard(t, card, filler)
	if got := d.zoneOf(card); got != "exile" {
		t.Fatalf("a madness discard is in %q, want exile", got)
	}
	if order := latestChoiceOfKind(d.g, game.PendingChoiceTriggerOrder); order != nil {
		if err := d.g.ResolveTriggerOrder(order.ID, order.Chooser, append([]uuid.UUID(nil), order.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	passPriorityAroundTable(t, d.g)
	offer := keywordCastPromptFor(d.g, d.victim.ID)
	if offer == nil {
		t.Fatal("the madness discard offered no cast")
	}
	answerKeywordCast(t, d.g, offer, false)
	passPriorityAroundTable(t, d.g)
	if got := d.zoneOf(card); got != "graveyard" {
		t.Errorf("a declined madness cast leaves the card in %q, want graveyard", got)
	}
}
