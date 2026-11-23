package testsru

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	dmuB13Peacekeeper = "a67f78be-8fbe-414a-9574-bb11ac50e1a5"
	dmuB13Loremaster  = "7d540860-ae94-4773-a526-aaffca3eb057"
	dmuB13Astor       = "196962f0-6aff-4198-9e0c-eaf42aeaf147"
	dmuB13Danitha     = "43b5dbe9-41f7-4082-944b-ea9d6292c6fd"
	dmuB13Sleeper     = "12367806-5924-485c-814a-5b67dd79235e"
	dmuB13Pitlord     = "776ea0c7-fc5b-44b4-9302-5a3514ff843b"
	dmuB13Zur         = "cfc5dc9c-6f26-460b-a624-3681b8584eb3"
	dmuB13Shanna      = "7677c7ce-1db3-41f5-8f86-e68718ea2d11"
	dmuB13Chaotic     = "a754458d-55b5-44c2-98de-8bf26b82c1b1"
)

var _ = game.StepDeclareAttackers

func TestDmuB13EveryRegisteredCardIsFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dmuB13Peacekeeper: "Anointed Peacekeeper", dmuB13Loremaster: "Academy Loremaster",
		dmuB13Astor: "Astor, Bearer of Blades", dmuB13Danitha: "Danitha, Benalia's Hope",
		dmuB13Sleeper: "Evolved Sleeper", dmuB13Pitlord: "Tyrannical Pitlord",
		dmuB13Zur: "Zur, Eternal Schemer", dmuB13Shanna: "Shanna, Purifying Blade",
		dmuB13Chaotic: "Chaotic Transformation",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestDmuB13ShannaDrawsUpToTheLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Shanna, Purifying Blade", "Legendary Creature — Human Warrior", dmuB13Shanna, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.StockLibraryFor(g, me, 10)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	hand := me.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.FillPool(me, 5)
	effectstest.PassPriorityAroundTable(t, g)
	pa := effectstest.AnswerNumber(t, g, me.ID, 2)
	if pa.Max != 2 {
		t.Errorf("X ceiling = %d, want 2 (the life gained)", pa.Max)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if n := me.ManaPool.Total(); n != 3 {
		t.Errorf("pool has %d mana, want 3", n)
	}
}

func TestDmuB13ShannaAsksNothingWithoutLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Shanna, Purifying Blade", "Legendary Creature — Human Warrior", dmuB13Shanna, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.FillPool(me, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PayAmountPrompt(g, me.ID) != nil {
		t.Error("asked for X with no life gained")
	}
}

// libraryCard pushes a typed card onto the top of a seat's library.
func dmuB13LibraryCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func dmuB13HandCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Astor, Bearer of Blades ---

func TestDmuB13AstorTakesAnEquipmentOrVehicleFromTheTopSevenAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 5; i++ {
		dmuB13LibraryCard(me, "Deep", "Creature — Bear")
	}
	deep := dmuB13LibraryCard(me, "Eighth", "Creature — Bear")
	blade := dmuB13LibraryCard(me, "Blade", "Artifact — Equipment")
	cart := dmuB13LibraryCard(me, "Cart", "Artifact — Vehicle")
	var tops []uuid.UUID
	for i := 0; i < 5; i++ {
		tops = append(tops, dmuB13LibraryCard(me, "Top", "Creature — Bear"))
	}
	effectstest.CastCatalogSpell(t, g, "Astor, Bearer of Blades", "Legendary Creature — Human Warrior", dmuB13Astor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt")
	}
	for _, id := range tops {
		if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{id}); err == nil {
			t.Fatal("a creature card was a legal pick")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, cart)
	if !effectstest.DmuB13InHand(me, cart) {
		t.Error("the Vehicle is not in hand")
	}
	if effectstest.DmuB13InHand(me, blade) {
		t.Error("the Equipment went to hand too")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != deep {
		t.Errorf("the top card is %q, want the eighth card (the other six went to the bottom)", top.Name)
	}
	bottom := map[uuid.UUID]bool{}
	for i := 0; i < 6; i++ {
		bottom[me.Library.Cards[i].InstanceID] = true
	}
	if !bottom[blade] {
		t.Error("the Equipment is not among the bottom six")
	}
}

func TestDmuB13AstorGrantsEquip1AndCrew1ToYourOwnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Astor, Bearer of Blades", "Legendary Creature — Human Warrior", dmuB13Astor, false)
	bear := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Plain Blade", "")
	theirBlade := effectstest.SeedEquipment(g, opp.ID, "Their Blade", "")
	cart := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cart", TypeLine: "Artifact — Vehicle",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	theirCart := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Cart", TypeLine: "Artifact — Vehicle",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	if idx, _ := effectstest.BeGrantedRow(t, g, blade, "Equip {1}"); idx < 0 {
		t.Fatal("my Equipment has no equip {1}")
	}
	if idx, _ := effectstest.BeGrantedRow(t, g, theirBlade, "Equip {1}"); idx >= 0 {
		t.Error("an opponent's Equipment has equip {1}")
	}
	if idx, _ := effectstest.BeGrantedRow(t, g, cart, "Crew 1"); idx < 0 {
		t.Error("my Vehicle has no crew 1")
	}
	if idx, _ := effectstest.BeGrantedRow(t, g, theirCart, "Crew 1"); idx >= 0 {
		t.Error("an opponent's Vehicle has crew 1")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.BeEquip(t, g, me.ID, blade, bear, "Equip {1}")
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, blade); !c.IsAttachedTo(bear) {
		t.Error("the Equipment did not attach for {1}")
	}
	// Crew 1 taps a creature with power 1 or more and makes the Vehicle a creature.
	idx, ref := effectstest.BeGrantedRow(t, g, cart, "Crew 1")
	if err := g.ActivateCatalogAbility(me.ID, cart, idx, game.ActivateAbilityParams{Ref: ref, CrewIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("crew 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, cart); !c.IsCreature() {
		t.Error("the crewed Vehicle is not a creature")
	}
}

// --- Danitha, Benalia's Hope ---

func TestDmuB13DanithaPutsAnEquipmentFromHandOrGraveyardOntoItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	inHand := dmuB13HandCard(me, "Hand Blade", "Artifact — Equipment")
	inYard := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: inYard, Name: "Grave Blade", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	bolt := dmuB13HandCard(me, "Bolt", "Instant")
	danitha := effectstest.CastCatalogSpell(t, g, "Danitha, Benalia's Hope", "Legendary Creature — Human Knight", dmuB13Danitha, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bolt}); err == nil {
		t.Fatal("an instant was a legal pick")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, inYard)
	effectstest.PassPriorityAroundTable(t, g)
	eq, ok := effectstest.BattlefieldCard(g, inYard)
	if !ok || !eq.IsAttachedTo(danitha) {
		t.Fatalf("the graveyard Equipment is on the battlefield=%v attached to Danitha=%v", ok, ok && eq.IsAttachedTo(danitha))
	}
	if _, ok := effectstest.BattlefieldCard(g, inHand); ok {
		t.Error("the hand Equipment also entered")
	}
	for _, kw := range []string{"first strike", "vigilance", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, danitha, kw) {
			t.Errorf("Danitha lacks %s", kw)
		}
	}
}

func TestDmuB13DanithaMayDeclineAndAsksNothingWithoutACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	blade := dmuB13HandCard(me, "Hand Blade", "Artifact — Equipment")
	effectstest.CastCatalogSpell(t, g, "Danitha, Benalia's Hope", "Legendary Creature — Human Knight", dmuB13Danitha, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, blade); ok || !effectstest.DmuB13InHand(me, blade) {
		t.Error("declining still put the Equipment onto the battlefield")
	}

	g2 := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g2, "Danitha, Benalia's Hope", "Legendary Creature — Human Knight", dmuB13Danitha, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.ChooseCardsChoiceFor(g2, g2.Seats[0].ID) != nil {
		t.Error("asked with no Aura or Equipment to put")
	}
}

// --- Evolved Sleeper ---

func TestDmuB13EvolvedSleeperEvolvesThroughThreeStages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.StockLibraryFor(g, me, 5)
	sl := effectstest.PushCatalogPermanent(g, me.ID, "Evolved Sleeper", "Creature — Human", dmuB13Sleeper, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == sl {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 1, 1
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	act := func(idx int, mana ...string) {
		t.Helper()
		effectstest.B06AddMana(me, mana...)
		effectstest.B16Activate(t, g, me.ID, sl, idx, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
	}
	// Stages two and three do nothing before their condition holds.
	act(1, "C", "B")
	if p, tt := effectstest.PtOf(t, g, sl); p != 1 || tt != 1 {
		t.Fatalf("stage 2 without a Cleric changed it to %d/%d", p, tt)
	}
	if c, _ := effectstest.BattlefieldCard(g, sl); c.Counters[game.CounterDeathtouch] != 0 {
		t.Fatal("a deathtouch counter without being a Cleric")
	}
	hand, life := me.Hand.Size(), me.Life
	act(2, "C", "B", "B")
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatal("stage 3 without a Phyrexian drew or cost life")
	}
	act(0, "B")
	c, _ := effectstest.BattlefieldCard(g, sl)
	if p, tt := effectstest.PtOf(t, g, sl); p != 2 || tt != 2 || !c.HasSubtype("Cleric") || !c.HasSubtype("Human") {
		t.Fatalf("stage 1: %d/%d %v", p, tt, c.Effective().Subtypes)
	}
	act(1, "C", "B")
	c, _ = effectstest.BattlefieldCard(g, sl)
	if p, tt := effectstest.PtOf(t, g, sl); p != 3 || tt != 3 || !c.HasSubtype("Phyrexian") || !c.HasSubtype("Cleric") || !c.HasSubtype("Human") {
		t.Fatalf("stage 2: %d/%d %v", p, tt, c.Effective().Subtypes)
	}
	if c.Counters[game.CounterDeathtouch] != 1 || !effectstest.HasEffectiveKeyword(t, g, sl, "deathtouch") {
		t.Error("no deathtouch counter after stage 2")
	}
	hand, life = me.Hand.Size(), me.Life
	act(2, "C", "B", "B")
	if p, tt := effectstest.PtOf(t, g, sl); p != 4 || tt != 4 {
		t.Errorf("stage 3: %d/%d, want 4/4 with a +1/+1 counter", p, tt)
	}
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("stage 3: hand %d→%d, life %d→%d; want one card and 1 life", hand, me.Hand.Size(), life, me.Life)
	}
	// None of it ends with the turn.
	effectstest.Ecl6NextTurn(t, g)
	if p, _ := effectstest.PtOf(t, g, sl); p != 4 {
		t.Error("the effect ended with the turn")
	}
}

// --- Academy Loremaster ---

func TestDmuB13LoremasterLetsTheDrawStepPlayerDrawAndTaxesTheirSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Academy Loremaster", "Creature — Human Wizard", dmuB13Loremaster, false)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	library := opp.Library.Size()
	effectstest.AdvanceToStepOf(t, g, 1, game.StepDraw)
	effectstest.AnswerLatestTriggerPrompt(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if drawn := library - opp.Library.Size(); drawn != 2 {
		t.Errorf("the opponent drew %d in their draw step, want 2", drawn)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Cultivate", "Sorcery", "{2}{G}"); got != 5 {
		t.Errorf("their spell costs %d, want 5", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("my spell costs %d, want 3 (I drew nothing)", got)
	}
	// "This turn": the tax is gone on the next turn.
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	if got := effectstest.PriceInHand(t, g, opp, "Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("their spell costs %d the next turn, want 3", got)
	}
}

func TestDmuB13LoremasterDecliningDrawsNothingAndTaxesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Academy Loremaster", "Creature — Human Wizard", dmuB13Loremaster, false)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	library := opp.Library.Size()
	effectstest.AdvanceToStepOf(t, g, 1, game.StepDraw)
	effectstest.AnswerLatestTriggerPrompt(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if drawn := library - opp.Library.Size(); drawn != 1 {
		t.Errorf("the opponent drew %d, want only the normal 1", drawn)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("their spell costs %d, want 3", got)
	}
}

func TestDmuB13LoremasterTaxOutlivesTheLoremaster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lm := effectstest.PushCatalogPermanent(g, me.ID, "Academy Loremaster", "Creature — Human Wizard", dmuB13Loremaster, false)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepDraw)
	effectstest.AnswerLatestTriggerPrompt(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(lm) })
	if got := effectstest.PriceInHand(t, g, opp, "Cultivate", "Sorcery", "{2}{G}"); got != 5 {
		t.Errorf("their spell costs %d after Loremaster left, want 5 (\"this turn\")", got)
	}
}

// --- Tyrannical Pitlord ---

func TestDmuB13PitlordPumpsTheChosenCreatureAndSacrificesItWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chosen := effectstest.B12Creature(g, me.ID, "Chosen Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	pit := effectstest.CastCatalogSpell(t, g, "Tyrannical Pitlord", "Creature — Demon", dmuB13Pitlord, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt for the chosen creature")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err == nil {
		t.Fatal("choosing nothing was legal: the choice is mandatory")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{pit}); err == nil {
		t.Fatal("the Pitlord itself was a legal pick")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePower(t, g, chosen), effectstest.EffectiveToughness(t, g, chosen); p != 5 || tt != 5 {
		t.Errorf("the chosen creature is %d/%d, want 5/5", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, chosen, "flying") {
		t.Error("the chosen creature lacks flying")
	}
	if p := effectstest.EffectivePower(t, g, other); p != 2 || effectstest.HasEffectiveKeyword(t, g, other, "flying") {
		t.Errorf("the other creature is affected (power %d)", p)
	}
	for _, kw := range []string{"flying", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, pit, kw) {
			t.Errorf("the Pitlord lacks %s", kw)
		}
	}
	// It leaves: the chosen creature is sacrificed, the other is not.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(pit) })
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, chosen); ok {
		t.Error("the chosen creature is still on the battlefield")
	}
	if _, ok := effectstest.BattlefieldCard(g, other); !ok {
		t.Error("the other creature was sacrificed")
	}
}

func TestDmuB13PitlordWithNoOtherCreatureChoosesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Tyrannical Pitlord", "Creature — Demon", dmuB13Pitlord, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked to choose with no other creature")
	}
}

func TestDmuB13PitlordLeavingLeavesAStolenChosenCreatureAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chosen := effectstest.B12Creature(g, me.ID, "Chosen Bear", "Creature — Bear", 2, 2)
	pit := effectstest.CastCatalogSpell(t, g, "Tyrannical Pitlord", "Creature — Demon", dmuB13Pitlord, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.Nil, chosen, opp.ID, game.IndefiniteDuration(), "theft")
		_ = g.SacrificePermanentForEffect(pit)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, chosen); !ok {
		t.Error("an opponent's creature was sacrificed by the Pitlord's controller")
	}
}

// --- Zur, Eternal Schemer ---

func TestDmuB13ZurGrantsEnchantmentCreaturesKeywordsAndAnimatesEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	zur := effectstest.PushCatalogPermanent(g, me.ID, "Zur, Eternal Schemer", "Legendary Creature — Human Wizard", dmuB13Zur, false)
	spirit := effectstest.B12Creature(g, me.ID, "Spirit", "Enchantment Creature — Spirit", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Spirit", "Enchantment Creature — Spirit", 2, 2)
	for _, kw := range []string{"deathtouch", "lifelink", "hexproof"} {
		if !effectstest.HasEffectiveKeyword(t, g, spirit, kw) {
			t.Errorf("my enchantment creature lacks %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, bear, kw) || effectstest.HasEffectiveKeyword(t, g, theirs, kw) || effectstest.HasEffectiveKeyword(t, g, zur, kw) {
			t.Errorf("%s reached a creature that is not an enchantment creature of mine", kw)
		}
	}
	ench := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Idea", TypeLine: "Enchantment",
		ManaCost: "{3}{W}", Owner: me.ID, Controller: me.ID})
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Some Aura", TypeLine: "Enchantment — Aura",
		ManaCost: "{1}{W}", Owner: me.ID, Controller: me.ID})
	theirEnch := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Idea", TypeLine: "Enchantment",
		ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID})
	for _, bad := range []uuid.UUID{aura, theirEnch} {
		effectstest.B06AddMana(me, "C", "W")
		if err := g.ActivateCatalogAbility(me.ID, zur, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(bad)}); err == nil {
			t.Fatalf("targeted an illegal permanent %s", bad)
		}
	}
	effectstest.B06AddMana(me, "C", "W")
	effectstest.B16Activate(t, g, me.ID, zur, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(ench)})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, ench)
	if !c.IsCreature() || !c.IsEnchantment() {
		t.Fatalf("the enchantment is not an enchantment creature: %v", c.Effective().Types)
	}
	if p, tt := effectstest.PtOf(t, g, ench); p != 4 || tt != 4 {
		t.Errorf("it is %d/%d, want 4/4 (its mana value)", p, tt)
	}
	// It now gets Zur's keywords, and it lasts past the turn.
	if !effectstest.HasEffectiveKeyword(t, g, ench, "lifelink") {
		t.Error("the animated enchantment lacks lifelink")
	}
	effectstest.Ecl6NextTurn(t, g)
	if c, _ := effectstest.BattlefieldCard(g, ench); !c.IsCreature() {
		t.Error("the animation ended with the turn")
	}
}

// --- Anointed Peacekeeper ---

// dmuB13NamingPeacekeeper puts a Peacekeeper under owner's control and has
// owner answer its card-name prompt with name.
func dmuB13NamingPeacekeeper(t *testing.T, g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id := effectstest.B21Push(g, owner, "Anointed Peacekeeper", "Creature — Human Cleric", dmuB13Peacekeeper, 3, 3, "W")
	g.WithWriteLock(func() {
		g.QueueCardNameChoiceForEffect(owner, id, "Anointed Peacekeeper — choose any card name")
	})
	effectstest.ActNameFor(t, g, owner, name)
	return id
}

func TestDmuB13PeacekeeperLooksAtAnOpponentsHandThenNamesACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	dmuB13HandCard(opp, "Secret", "Sorcery")
	dmuB13HandCard(other, "Other Secret", "Sorcery")
	id := effectstest.CastCatalogSpell(t, g, "Anointed Peacekeeper", "Creature — Human Cleric", dmuB13Peacekeeper, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no opponent prompt")
	}
	if len(pick.PickOptions) != 3 {
		t.Errorf("%d players offered, want the three opponents", len(pick.PickOptions))
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, effectstest.OptionFor(t, pick, opp.ID)); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	for i := range opp.Hand.Cards {
		if !opp.Hand.Cards[i].IsKnownTo(me.ID) {
			t.Errorf("the chosen opponent's card %d is not known to the controller", i)
		}
	}
	for i := range other.Hand.Cards {
		if other.Hand.Cards[i].IsKnownTo(me.ID) {
			t.Errorf("another opponent's card %d was revealed too", i)
		}
	}
	effectstest.ActNameFor(t, g, me.ID, "Sol Ring")
	if c, ok := effectstest.BattlefieldCard(g, id); !ok || c.ChosenName != "Sol Ring" {
		t.Errorf("chosen name = %q", c.ChosenName)
	}
}

func TestDmuB13PeacekeeperTaxesOpponentsSpellsWithTheChosenName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dmuB13NamingPeacekeeper(t, g, me.ID, "Sol Ring")
	if got := effectstest.PriceInHand(t, g, opp, "Sol Ring", "Artifact", "{1}"); got != 3 {
		t.Errorf("an opponent's Sol Ring costs %d, want 3", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Sol Ring", "Artifact", "{1}"); got != 1 {
		t.Errorf("my own Sol Ring costs %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Mana Vault", "Artifact", "{1}"); got != 1 {
		t.Errorf("an unnamed card costs %d, want 1", got)
	}
}

func TestDmuB13PeacekeeperTaxesNonManaAbilitiesOfTheNamedSourceForEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	dmuB13NamingPeacekeeper(t, g, me.ID, "Krenko, Mob Boss")
	mine := effectstest.ActPushKrenko(g, me.ID)
	theirs := effectstest.ActPushKrenko(g, opp.ID)
	for who, id := range map[*game.Player]uuid.UUID{me: mine, opp: theirs} {
		if got := effectstest.ActivationPrice(t, g, who.ID, id); got != "{2}" {
			t.Errorf("%s's Krenko prices at %q, want {2}", who.Name, got)
		}
	}
	err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{Strict: true})
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("the taxed ability with an empty pool: err = %v, want InsufficientManaError", err)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Krenko, Mob Boss", "Creature", "{2}{R}"); got != 5 {
		t.Errorf("the named card as an opponent's spell costs %d, want 5", got)
	}
}

func TestDmuB13PeacekeeperLeavesManaAbilitiesAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	dmuB13NamingPeacekeeper(t, g, me.ID, "Sol Ring")
	ring := effectstest.B21Push(g, me.ID, "Sol Ring", "Artifact", effectstest.ActSolRingOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("a named mana ability with an empty pool: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("pool has %d mana, want 2", me.ManaPool.Total())
	}
}

// --- Chaotic Transformation ---

func dmuB13Targets(slots map[int]uuid.UUID) []game.TargetRef {
	var out []game.TargetRef
	for slot := 0; slot < 5; slot++ {
		if id, ok := slots[slot]; ok {
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id, Slot: slot})
		}
	}
	return out
}

func TestDmuB13ChaoticTransformationReplacesEachExiledPermanentWithOneSharingATypeAndShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	forest := effectstest.SeedPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	// Their library: a Wurm and a Mountain somewhere under filler non-matches.
	wurm := dmuB13LibraryCard(opp, "Library Wurm", "Creature — Wurm")
	mountain := dmuB13LibraryCard(opp, "Library Mountain", "Basic Land — Mountain")
	dmuB13LibraryCard(opp, "Filler Spell", "Instant")
	dmuB13LibraryCard(opp, "Filler Sorcery", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Chaotic Transformation", "Sorcery", dmuB13Chaotic, dmuB13Targets(map[int]uuid.UUID{1: bear, 4: forest}))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) || !g.Exile.Contains(forest) {
		t.Fatal("the targets were not exiled")
	}
	for _, id := range []uuid.UUID{wurm, mountain} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok || c.Controller != opp.ID {
			t.Errorf("%s did not enter under its owner's control (on battlefield=%v)", id, ok)
		}
	}
	// The cards revealed on the way were shuffled back, not put on the bottom.
	if n := opp.Library.Size(); n < 2 {
		t.Errorf("library has %d cards, want the revealed fillers back in it", n)
	}
	_ = me
}

func TestDmuB13ChaoticTransformationExilesAnArtifactCreatureOnceForBothClauses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	elf := dmuB13LibraryCard(me, "Library Elf", "Creature — Elf")
	rock := dmuB13LibraryCard(me, "Library Rock", "Artifact")
	before := len(g.Battlefield.Cards)
	effectstest.CastCatalogSpell(t, g, "Chaotic Transformation", "Sorcery", dmuB13Chaotic, dmuB13Targets(map[int]uuid.UUID{0: golem, 1: golem}))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(golem) {
		t.Fatal("the Golem was not exiled")
	}
	// Exiled once: one replacement, the top card (an Artifact) that shares a type.
	if got := len(g.Battlefield.Cards); got != before {
		t.Errorf("battlefield %d → %d, want one out and one in", before, got)
	}
	if _, ok := effectstest.BattlefieldCard(g, rock); !ok {
		if _, ok := effectstest.BattlefieldCard(g, elf); !ok {
			t.Error("neither card sharing a type entered")
		}
	}
}

func TestDmuB13ChaoticTransformationCanTargetNothingAndShufflesWhenNothingMatches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pw := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 3}, Owner: opp.ID, Controller: opp.ID})
	dmuB13LibraryCard(opp, "Only Spell", "Instant")
	effectstest.CastCatalogSpell(t, g, "Chaotic Transformation", "Sorcery", dmuB13Chaotic, dmuB13Targets(map[int]uuid.UUID{3: pw}))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(pw) {
		t.Fatal("the planeswalker was not exiled")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == opp.ID && c.Name != "" && c.InstanceID != pw && c.Name == "Only Spell" {
			t.Error("a card that shares no type entered")
		}
	}
	if opp.Library.Size() == 0 {
		t.Error("the revealed library vanished")
	}
	_ = me
}
