package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b125_cards_test.go — Commander 2020 cards, S82 OLD batch 125
// (#5513), played through the real catalog entries.

const (
	oldB125PakoOracle      = "680a3e17-5144-4e97-bc59-9f5fc3ad6514"
	oldB125HaldanOracle    = "6a0b78fb-3838-47d9-8d1a-b6b0c0e8927d"
	oldB125YannikOracle    = "81ab8947-8046-4944-ae58-006a6eefa08a"
	oldB125BountyOracle    = "056b5d53-f2fd-4e83-9fd3-d8f0d0be64b9"
	oldB125ChaplainOracle  = "1e575a82-e128-4637-87cd-f075654bc41b"
	oldB125PegasusOracle   = "1f6fd5d9-d123-4986-a18b-206cc2dfb3f1"
	oldB125DjinnOracle     = "d7dcf1be-2779-419d-8c90-03800f9e7ea2"
	oldB125TitanOracle     = "490df8a0-182a-40b8-a36e-f2666fa7c7fa"
	oldB125VigilanteOracle = "9610176a-dcd6-4117-89db-45f56b16cdd6"
	oldB125MetamorphOracle = "68f34b1c-e06b-4859-b7c0-62148341e38c"
	oldB125FumikoOracle    = "cba58942-b98c-45aa-8651-7dc666e31c93"
)

func TestOldB125CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB125PakoOracle:      "Pako, Arcane Retriever",
		oldB125HaldanOracle:    "Haldan, Avid Arcanist",
		oldB125YannikOracle:    "Yannik, Scavenging Sentinel",
		oldB125BountyOracle:    "Bounty Agent",
		oldB125ChaplainOracle:  "Devout Chaplain",
		oldB125PegasusOracle:   "Cavalry Pegasus",
		oldB125DjinnOracle:     "Djinn Illuminatus",
		oldB125TitanOracle:     "Titan of Eternal Fire",
		oldB125VigilanteOracle: "Vigilante Justice",
		oldB125MetamorphOracle: "Nascent Metamorph",
		oldB125FumikoOracle:    "Fumiko the Lowblood",
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

func oldB125Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB125Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// --- Bounty Agent ----------------------------------------------------------

func TestOldB125BountyAgentDestroysOnlyALegendaryPermanentAndIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB125Seats(g)
	agent := effectstest.B12Push(g, me.ID, "Bounty Agent", "Creature — Human Soldier", oldB125BountyOracle, 2, 2)
	plain := effectstest.B12Creature(g, opp.ID, "Plain Bear", "Creature — Bear", 2, 2)
	legend := effectstest.B12Creature(g, opp.ID, "Legend", "Legendary Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, agent, 0, game.ActivateAbilityParams{Targets: oldB125Target(plain), Strict: true}); err == nil {
		t.Fatal("Bounty Agent targeted a nonlegendary creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, agent, 0, game.ActivateAbilityParams{Targets: oldB125Target(legend), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, legend); ok {
		t.Error("the legendary creature survived")
	}
	if _, ok := effectstest.BattlefieldCard(g, agent); ok {
		t.Error("Bounty Agent was not sacrificed")
	}
	if _, ok := effectstest.BattlefieldCard(g, plain); !ok {
		t.Error("the nonlegendary creature died")
	}
}

// --- Devout Chaplain -------------------------------------------------------

func TestOldB125DevoutChaplainTapsTwoHumansToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB125Seats(g)
	chaplain := effectstest.B12Push(g, me.ID, "Devout Chaplain", "Creature — Human Cleric", oldB125ChaplainOracle, 2, 2)
	h1 := effectstest.B12Creature(g, me.ID, "Human One", "Creature — Human", 1, 1)
	h2 := effectstest.B12Creature(g, me.ID, "Human Two", "Creature — Human", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	relic := effectstest.PushCatalogPermanent(g, opp.ID, "Relic", "Artifact", "", false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, chaplain, 0, game.ActivateAbilityParams{Targets: oldB125Target(relic), TapIDs: []uuid.UUID{h1, bear}, Strict: true}); err == nil {
		t.Fatal("a non-Human paid for the Chaplain")
	}
	if err := g.ActivateCatalogAbility(me.ID, chaplain, 0, game.ActivateAbilityParams{Targets: oldB125Target(relic), TapIDs: []uuid.UUID{h1, h2}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.E2Zone(g, relic) != game.ZoneExile {
		t.Errorf("the artifact is in %q, want exile", effectstest.E2Zone(g, relic))
	}
	for _, id := range []uuid.UUID{chaplain, h1, h2} {
		if _, c := effectstest.CardWhere(g, me, id); !c.Tapped {
			t.Errorf("%s is untapped after paying", c.Name)
		}
	}
}

// --- Cavalry Pegasus -------------------------------------------------------

func TestOldB125CavalryPegasusGivesAttackingHumansFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB125Seats(g)
	pegasus := effectstest.B12Push(g, me.ID, "Cavalry Pegasus", "Creature — Pegasus", oldB125PegasusOracle, 1, 1)
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human Soldier", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Human", "Creature — Human", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, pegasus, human, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, human, "flying") {
		t.Error("an attacking Human did not gain flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("an attacking Bear gained flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, idle, "flying") {
		t.Error("a Human that did not attack gained flying")
	}
}

// --- Djinn Illuminatus -----------------------------------------------------

func TestOldB125DjinnGivesInstantsAndSorceriesReplicateAtTheirCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Djinn Illuminatus", "Creature — Djinn", oldB125DjinnOracle, 3, 5)
	sorcery := effectstest.TrekB39Hand(g, me, "Test Sorcery", "Sorcery", "{1}{B}", "", 0, 0)
	bear := effectstest.TrekB39Hand(g, me, "Test Bear", "Creature — Bear", "{1}{B}", "", 2, 2)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}{C}{C}{C}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err == nil {
		t.Fatal("a creature spell was offered replicate")
	}
	// {1}{B} plus two replicates of {1}{B}.
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true, OptionalCosts: []int{0, 0}}); err != nil {
		t.Fatalf("a sorcery with replicate twice: %v", err)
	}
	if left := me.ManaPool.Total(); left != 0 {
		t.Errorf("%d mana left after paying {1}{B} three times", left)
	}
}

func TestOldB125DjinnOnlyGrantsItsControllersSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, opp.ID, "Djinn Illuminatus", "Creature — Djinn", oldB125DjinnOracle, 3, 5)
	sorcery := effectstest.TrekB39Hand(g, me, "Test Sorcery", "Sorcery", "{1}{B}", "", 0, 0)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}{C}{C}{C}")
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err == nil {
		t.Fatal("an opponent's Djinn gave my sorcery replicate")
	}
}

// --- Titan of Eternal Fire -------------------------------------------------

func TestOldB125TitanGivesHumansYouControlAPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB125Seats(g)
	effectstest.B12Push(g, me.ID, "Titan of Eternal Fire", "Creature — Giant", oldB125TitanOracle, 5, 6)
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Human", "Creature — Human", 1, 1)
	effectstest.AdvanceToMain(t, g)
	before := opp.Life
	if err := effectstest.HobB4Activate(t, g, me.ID, human, 0, "{R}", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("a Human did not get the ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d → %d, want 1 damage", before, opp.Life)
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, bear, 0, "{R}", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Error("a Bear got the ping")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, theirs, 0, "{R}", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Error("an opponent's Human got the ping")
	}
}

// --- Vigilante Justice -----------------------------------------------------

func TestOldB125VigilanteJusticePingsForEachHumanOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB125Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Vigilante Justice", "Enchantment", oldB125VigilanteOracle, false)
	effectstest.EnterFromHand(t, g, me.ID, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a non-Human entering triggered Vigilante Justice")
	}
	effectstest.EnterFromHand(t, g, me.ID, "Human", "Creature — Human", "")
	effectstest.PassPriorityAroundTable(t, g)
	before := opp.Life
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d → %d, want 1 damage", before, opp.Life)
	}
	// An opponent's Human does not trigger it.
	effectstest.EnterFromHand(t, g, opp.ID, "Their Human", "Creature — Human", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's Human triggered Vigilante Justice")
	}
}

// --- Fumiko the Lowblood ---------------------------------------------------

func TestOldB125FumikoForcesOpponentsToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB125Seats(g)
	fumiko := effectstest.B12Push(g, me.ID, "Fumiko the Lowblood", "Legendary Creature — Human Samurai", oldB125FumikoOracle, 3, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	required := func(id uuid.UUID) (n int) {
		g.ReadSnapshot(func() { n = len(effectstest.BattlefieldCardFor(g, id).Effective().AttackRequirements) })
		return n
	}
	if required(theirs) == 0 {
		t.Error("an opponent's creature has no attack requirement")
	}
	if required(mine) != 0 || required(fumiko) != 0 {
		t.Error("Fumiko's controller's creatures must attack")
	}
}

func TestOldB125FumikoBushidoCountsAttackingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB125Seats(g)
	fumiko := effectstest.B12Push(g, me.ID, "Fumiko the Lowblood", "Legendary Creature — Human Samurai", oldB125FumikoOracle, 3, 2)
	a := effectstest.B12Creature(g, opp.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Attacker B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == a || id == b {
				g.Battlefield.Cards[i].AttackingTarget = me.ID
			}
		}
		g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: me.ID, Source: a, CardID: fumiko, Target: fumiko})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, fumiko); p != 5 || tough != 4 {
		t.Errorf("Fumiko is %d/%d after being blocked by two attackers' worth, want 5/4 (bushido 2)", p, tough)
	}
}

// --- Nascent Metamorph -----------------------------------------------------

func TestOldB125NascentMetamorphCopiesTheRevealedCreatureAndBottomsTheRun(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB125Seats(g)
	meta := effectstest.B12Push(g, me.ID, "Nascent Metamorph", "Creature — Shapeshifter", oldB125MetamorphOracle, 1, 1)
	// Top to bottom: a land, a sorcery, the creature.
	creature := effectstest.TopOfLibrary(opp, "Big Bear", "Creature — Bear", "{2}{G}")
	g.WithWriteLock(func() {
		for i := range opp.Library.Cards {
			if opp.Library.Cards[i].InstanceID == creature {
				opp.Library.Cards[i].Power, opp.Library.Cards[i].Toughness = 4, 4
			}
		}
	})
	sorcery := effectstest.TopOfLibrary(opp, "Plain Sorcery", "Sorcery", "{1}")
	land := effectstest.TopOfLibrary(opp, "Plain Island", "Basic Land — Island", "")
	effectstest.DeclareAttack(t, g, opp.ID, meta)
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, meta); p != 4 || tough != 4 {
		t.Errorf("Nascent Metamorph is %d/%d, want a 4/4 copy of Big Bear", p, tough)
	}
	n := len(opp.Library.Cards)
	if n < 3 {
		t.Fatalf("library has %d cards", n)
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range opp.Library.Cards[:3] {
		bottom[c.InstanceID] = true
	}
	for _, id := range []uuid.UUID{creature, sorcery, land} {
		if !bottom[id] {
			t.Errorf("revealed card %s is not on the bottom of the library", id)
		}
	}
}

// --- Pako, Arcane Retriever and Haldan, Avid Arcanist ----------------------

func TestOldB125PakoExilesEachTopCardWithAFetchCounterAndGrowsPerNoncreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB125Seats(g)
	pako := effectstest.B12Push(g, me.ID, "Pako, Arcane Retriever", "Legendary Creature — Elemental Dog", oldB125PakoOracle, 3, 3)
	tops := map[uuid.UUID]uuid.UUID{}
	for i, p := range g.Seats {
		if i%2 == 0 {
			tops[p.ID] = effectstest.TopOfLibrary(p, "Top of "+p.Name, "Sorcery", "{B}")
		} else {
			tops[p.ID] = effectstest.TopOfLibrary(p, "Top of "+p.Name, "Creature — Bear", "{G}")
		}
	}
	effectstest.DeclareAttack(t, g, opp.ID, pako)
	effectstest.PassPriorityAroundTable(t, g)
	noncreature := 0
	for _, p := range g.Seats {
		c := effectstest.CardInExile(t, g, tops[p.ID])
		if c.Counters[effects.FetchCounter] != 1 {
			t.Errorf("%s's top card: counters %v, want a fetch counter", p.Name, c.Counters)
		}
		if !c.IsCreature() {
			noncreature++
		}
	}
	if got := effectstest.B12Card(t, g, pako).Counters["+1/+1"]; got != noncreature || noncreature == 0 {
		t.Errorf("Pako has %d +1/+1 counters, want %d (one per noncreature card)", got, noncreature)
	}
}

func TestOldB125HaldanPlaysFetchCardsYouExiledWithAnyColorAndRefusesCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB125Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Haldan, Avid Arcanist", "Legendary Creature — Human Wizard", oldB125HaldanOracle, false)
	mark := func(owner *game.Player, by uuid.UUID, name, typeLine, cost string, counters int) uuid.UUID {
		id := effectstest.Stashed(g, owner.ID, name, typeLine, cost, 0)
		g.WithWriteLock(func() {
			for i := range g.Exile.Cards {
				if g.Exile.Cards[i].InstanceID == id {
					g.Exile.Cards[i].ExiledByAbilityOf = by
					if counters > 0 {
						g.Exile.Cards[i].Counters = map[string]int{effects.FetchCounter: counters}
					}
				}
			}
		})
		return id
	}
	spell := mark(opp, me.ID, "Their Sorcery", "Sorcery", "{G}", 1)
	creature := mark(opp, me.ID, "Their Bear", "Creature — Bear", "{G}", 1)
	unmarked := mark(opp, me.ID, "Unmarked Sorcery", "Sorcery", "{G}", 0)
	theirs := mark(opp, opp.ID, "Opposing Fetch Sorcery", "Sorcery", "{G}", 1)
	for _, id := range []uuid.UUID{creature, unmarked, theirs} {
		effectstest.PayAnyType(t, g, me, "{B}")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
			t.Errorf("Haldan allowed casting %s", id)
		}
	}
	g.Turn.Step = game.StepPrecombatMain
	effectstest.PayAnyType(t, g, me, "{B}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast a fetch-countered sorcery with black mana: %v", err)
	}
}

// --- Yannik, Scavenging Sentinel -------------------------------------------

func TestOldB125YannikExilesAnotherCreatureAndDistributesItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB125Seats(g)
	big := effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 3, 3)
	small := effectstest.B12Creature(g, me.ID, "Small Bear", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	yannik := oldB125EnterYannik(t, g, me)
	if legs := effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {big}}); legs == 0 {
		t.Fatal("Yannik asked for no creature to exile")
	}
	if effectstest.E2Zone(g, big) != game.ZoneExile {
		t.Fatalf("the chosen creature is in %q, want exile", effectstest.E2Zone(g, big))
	}
	effectstest.PassPriorityAroundTable(t, g)
	// "When you do": 3 counters split between Yannik and the small bear.
	effectstest.PickTargetRefs(t, g, me.ID, map[uuid.UUID]int{yannik: 1, small: 2},
		game.TargetRef{Kind: game.TargetCard, ID: yannik}, game.TargetRef{Kind: game.TargetCard, ID: small})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Card(t, g, small).Counters["+1/+1"]; got != 2 {
		t.Errorf("Small Bear has %d counters, want 2", got)
	}
	if got := effectstest.B12Card(t, g, yannik).Counters["+1/+1"]; got != 1 {
		t.Errorf("Yannik has %d counters, want 1", got)
	}
}

// Yannik leaving brings the exiled creature back.
func TestOldB125YannikReturnsTheCreatureWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB125Seats(g)
	big := effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	yannik := oldB125EnterYannik(t, g, me)
	effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {big}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTargetRefs(t, g, me.ID, map[uuid.UUID]int{yannik: 3}, game.TargetRef{Kind: game.TargetCard, ID: yannik})
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 8; i++ {
		c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
		if c == nil {
			break
		}
		if err := g.ResolveConfirm(c.ID, c.Chooser, false); err != nil {
			t.Fatalf("decline: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(yannik) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Big Bear", me.ID); n != 1 {
		t.Errorf("after Yannik left, %d Big Bears under their owner's control, want 1", n)
	}
}

// oldB125EnterYannik enters Yannik and settles his two entry triggers: the
// partner search (aimed at its controller, who finds nothing to search for)
// and the exile, leaving the exile's pick open.
func oldB125EnterYannik(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	yannik := effectstest.EnterFromHand(t, g, me.ID, "Yannik, Scavenging Sentinel", "Legendary Creature — Hyena Beast", oldB125YannikOracle)
	for i := 0; i < 12; i++ {
		if p := effectstest.TriggerOrderPromptFor(g, me.ID); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, me.ID, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			effectstest.PickPlayerTarget(t, g, me.ID, me.ID)
			continue
		}
		if c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm); c != nil {
			if err := g.ResolveConfirm(c.ID, c.Chooser, false); err != nil {
				t.Fatalf("decline the partner search: %v", err)
			}
			continue
		}
		open := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOwnPermanents {
				open = true
			}
		}
		if open || effectstest.StackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return yannik
}
