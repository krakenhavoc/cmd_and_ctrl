package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b680_cards_test.go — S82 OLD batch 680 (#5618, Nemesis: artifacts,
// life, other creatures, counters, graveyard recursion, auras, lands). Each
// test plays the real catalog entry.

const (
	oldB680FlintGolem  = "0993c5f1-1e8f-4871-9904-6c5a00405f36"
	oldB680Automaton   = "e4951b4e-ba68-495c-8062-e89250d02cf3"
	oldB680NobleStand  = "1aa1e34c-b2ff-4750-9e73-5707a8f49bfb"
	oldB680Cutter      = "fbd71691-d0fd-4842-913b-fe6096d509a3"
	oldB680Mammoth     = "29ef65bf-1066-44a5-a2eb-e3670b076647"
	oldB680Woodworm    = "75dcdf05-b392-4a96-98ae-5ec4cd425eee"
	oldB680Mossdog     = "95cc3652-b78b-473b-84cf-0dd3a7006219"
	oldB680Inhibitor   = "b8944199-9d33-4126-b529-b52e9f0116ff"
	oldB680LinSivvi    = "6c3847b6-9bfb-49b5-8f1a-bdb40e6900f1"
	oldB680Bracers     = "ba71eb8a-f9f8-4608-88d8-963a1770d8d4"
	oldB680RathsEdge   = "e10e84a7-d564-487a-ac64-5a001a45ee90"
	oldB680Woodripper  = "d0653734-9eb7-48c6-9042-65ffdd50fd2a"
	oldB680Invigorate  = "9264c198-76c7-4ac0-9334-37c691de1b54"
	oldB680BearType    = "Creature — Bear"
	oldB680ForestType  = "Basic Land — Forest"
	oldB680RebelType   = "Creature — Human Rebel"
	oldB680GolemType   = "Artifact Creature — Golem"
	oldB680ArtifactTyp = "Artifact"
)

func TestOldB680EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB680FlintGolem: "Flint Golem",
		oldB680Automaton:  "Complex Automaton",
		oldB680NobleStand: "Noble Stand",
		oldB680Cutter:     "Skyshroud Cutter",
		oldB680Mammoth:    "Wild Mammoth",
		oldB680Woodworm:   "Coiling Woodworm",
		oldB680Mossdog:    "Mossdog",
		oldB680Inhibitor:  "Parallax Inhibitor",
		oldB680LinSivvi:   "Lin Sivvi, Defiant Hero",
		oldB680Bracers:    "Treetop Bracers",
		oldB680RathsEdge:  "Rath's Edge",
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

// --- Flint Golem ----------------------------------------------------------------

func TestOldB680FlintGolemMillsTheDefenderWhenBlocked(t *testing.T) {
	for _, blockers := range []int{0, 1, 2} {
		t.Run(string(rune('0'+blockers))+" blockers", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			golem := effectstest.PushCatalogPermanent(g, me.ID, "Flint Golem", oldB680GolemType, oldB680FlintGolem, false)
			blocks := map[uuid.UUID]uuid.UUID{}
			for i := 0; i < blockers; i++ {
				blocks[effectstest.B12Push(g, opp.ID, "Blocker", oldB680BearType, "", 2, 2)] = golem
			}
			other := len(g.Seats[2].Library.Cards)
			lib := len(opp.Library.Cards)
			effectstest.Pr7bAttack(t, g, opp.ID, golem)
			effectstest.Pr7bBlock(t, g, blocks)
			effectstest.PassPriorityAroundTable(t, g)
			want := lib
			if blockers > 0 {
				want = lib - 3
			}
			if got := len(opp.Library.Cards); got != want {
				t.Errorf("defender's library %d -> %d, want %d", lib, got, want)
			}
			if len(g.Seats[2].Library.Cards) != other || len(me.Library.Cards) == 0 {
				t.Error("a player other than the defender milled")
			}
		})
	}
}

// --- Complex Automaton ---------------------------------------------------------

func TestOldB680AutomatonReturnsToHandOnlyWithSevenPermanents(t *testing.T) {
	for _, tc := range []struct {
		others int
		bounce bool
	}{{5, false}, {6, true}, {9, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		auto := effectstest.PushCatalogPermanent(g, me.ID, "Complex Automaton", oldB680GolemType, oldB680Automaton, false)
		for i := 0; i < tc.others; i++ {
			effectstest.FreeSpellPermanent(g, me.ID, "Rock", oldB680ArtifactTyp)
		}
		effectstest.FreeSpellPermanent(g, g.Seats[1].ID, "Their Rock", oldB680ArtifactTyp)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if got := !g.Battlefield.Contains(auto) && me.Hand.Contains(auto); got != tc.bounce {
			t.Errorf("%d other permanents: returned to hand = %v, want %v", tc.others, got, tc.bounce)
		}
	}
}

func TestOldB680AutomatonDoesNothingIfPermanentsFellBelowSevenBeforeResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	auto := effectstest.PushCatalogPermanent(g, me.ID, "Complex Automaton", oldB680GolemType, oldB680Automaton, false)
	var rocks []uuid.UUID
	for i := 0; i < 6; i++ {
		rocks = append(rocks, effectstest.FreeSpellPermanent(g, me.ID, "Rock", oldB680ArtifactTyp))
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if len(g.StackMeta) == 0 {
		t.Fatal("the trigger did not go on the stack with seven permanents")
	}
	_ = g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, rocks[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(auto) {
		t.Error("the Automaton returned although the intervening if no longer held")
	}
}

// --- Noble Stand ---------------------------------------------------------------

func TestOldB680NobleStandGainsTwoPerBlockingCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Noble Stand", "Enchantment", oldB680NobleStand, false)
	atk := effectstest.B12Push(g, me.ID, "Attacker", oldB680BearType, "", 2, 2)
	b1 := effectstest.B12Push(g, opp.ID, "Blocker One", oldB680BearType, "", 2, 2)
	b2 := effectstest.B12Push(g, opp.ID, "Blocker Two", oldB680BearType, "", 2, 2)
	life, mine := opp.Life, me.Life
	effectstest.Pr7bAttack(t, g, opp.ID, atk)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{b1: atk, b2: atk})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+4 || me.Life != mine {
		t.Errorf("life: controller %d (was %d), attacker %d (was %d); want +4 for two blockers", opp.Life, life, me.Life, mine)
	}
}

func TestOldB680NobleStandIgnoresAnOpponentsBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Noble Stand", "Enchantment", oldB680NobleStand, false)
	atk := effectstest.B12Push(g, me.ID, "Attacker", oldB680BearType, "", 2, 2)
	b1 := effectstest.B12Push(g, opp.ID, "Blocker", oldB680BearType, "", 2, 2)
	life := me.Life
	effectstest.Pr7bAttack(t, g, opp.ID, atk)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{b1: atk})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("an opponent's blocker gained life: %d -> %d", life, me.Life)
	}
}

// --- Skyshroud Cutter ----------------------------------------------------------

func TestOldB680CutterIsFreeOnlyWithAForestAndGivesEachOtherPlayerFiveLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cast := func() error {
		effectstest.AdvanceToMain(t, g)
		id := effectstest.HandCardOf(g, me, "Skyshroud Cutter", "Creature — Beast", "{3}{G}", oldB680Cutter)
		return g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "skyshroud-cutter", Strict: true})
	}
	if err := cast(); err == nil {
		t.Fatal("the free cast was accepted with no Forest")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", oldB680ForestType, "", false)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	if err := cast(); err != nil {
		t.Fatalf("free cast with a Forest: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[p.ID] + 5
		if i == 0 {
			want = before[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
	if effectstest.FindBattlefieldByName(g, "Skyshroud Cutter") == uuid.Nil {
		t.Error("the Cutter did not enter")
	}
}

// --- Wild Mammoth ---------------------------------------------------------------

func oldB680MammothTable(t *testing.T, mine, theirs int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mammoth := effectstest.PushCatalogPermanent(g, me.ID, "Wild Mammoth", "Creature — Elephant", oldB680Mammoth, false)
	for i := 0; i < mine; i++ {
		effectstest.B12Push(g, me.ID, "Mine", oldB680BearType, "", 1, 1)
	}
	for i := 0; i < theirs; i++ {
		effectstest.B12Push(g, opp.ID, "Theirs", oldB680BearType, "", 1, 1)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	return g, mammoth
}

func TestOldB680MammothGoesToThePlayerWithTheMostCreatures(t *testing.T) {
	g, mammoth := oldB680MammothTable(t, 1, 3) // me: Mammoth + 1 = 2, opponent 3
	c, ok := g.LookupCardForEffect(mammoth)
	if !ok || c.Controller != g.Seats[1].ID {
		t.Errorf("the Mammoth is controlled by %v, want the opponent with the most creatures", c.Controller)
	}
}

func TestOldB680MammothStaysOnATieOrWithTheMost(t *testing.T) {
	for name, tc := range map[string][2]int{"tie": {2, 3}, "controller has most": {3, 1}} {
		t.Run(name, func(t *testing.T) {
			g, mammoth := oldB680MammothTable(t, tc[0], tc[1])
			c, _ := g.LookupCardForEffect(mammoth)
			if c.Controller != g.Seats[0].ID {
				t.Error("control of the Mammoth changed")
			}
		})
	}
}

// --- Coiling Woodworm ----------------------------------------------------------

func TestOldB680WoodwormCountsEveryForestOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	worm := effectstest.PushCatalogPermanent(g, me.ID, "Coiling Woodworm", "Creature — Insect Worm", oldB680Woodworm, false)
	if p, tough := effectstest.PtOf(t, g, worm); p != 0 || tough != 1 {
		t.Fatalf("with no Forest it is %d/%d, want 0/1", p, tough)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", oldB680ForestType, "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Forest", oldB680ForestType, "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	if p, tough := effectstest.PtOf(t, g, worm); p != 2 || tough != 1 {
		t.Errorf("with two Forests it is %d/%d, want 2/1", p, tough)
	}
}

// --- Mossdog -----------------------------------------------------------------------

func TestOldB680MossdogGrowsOnlyWhenAnOpponentTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dog := effectstest.PushCatalogPermanent(g, me.ID, "Mossdog", "Creature — Plant Dog", oldB680Mossdog, false)
	effectstest.AppCastAsSeat(t, g, me, "Invigorate", "Instant", oldB680Invigorate, effectstest.CardRefs(dog))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.Pr8Counters(g, dog, game.CounterPlusOne); n != 0 {
		t.Fatalf("its own controller's spell gave it %d counters", n)
	}
	effectstest.AppCastAsSeat(t, g, opp, "Invigorate", "Instant", oldB680Invigorate, effectstest.CardRefs(dog))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.Pr8Counters(g, dog, game.CounterPlusOne); n != 1 {
		t.Errorf("an opponent's spell gave it %d counters, want 1", n)
	}
}

// --- Parallax Inhibitor ------------------------------------------------------------

func TestOldB680InhibitorFadesOnlyYourFadingPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inhibitor := effectstest.PushCatalogPermanent(g, me.ID, "Parallax Inhibitor", oldB680ArtifactTyp, oldB680Inhibitor, false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Woodripper", "Creature — Treefolk", oldB680Woodripper, false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Woodripper", "Creature — Treefolk", oldB680Woodripper, false)
	plain := effectstest.B12Push(g, me.ID, "Bear", oldB680BearType, "", 2, 2)
	effectstest.DcMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, inhibitor, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(inhibitor) {
		t.Error("the Inhibitor was not sacrificed")
	}
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want int
	}{"mine": {mine, 1}, "opponent's": {theirs, 0}, "no fading": {plain, 0}} {
		if n := effectstest.Pr8Counters(g, tc.id, game.CounterFade); n != tc.want {
			t.Errorf("%s: %d fade counters, want %d", name, n, tc.want)
		}
	}
}

// --- Lin Sivvi, Defiant Hero -------------------------------------------------------

func TestOldB680LinSivviSearchesARebelWithinX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Edge Rebel", TypeLine: oldB680RebelType, ManaCost: "{1}{W}"},
		game.Card{Name: "Big Rebel", TypeLine: oldB680RebelType, ManaCost: "{2}{W}"},
		game.Card{Name: "Not Rebel", TypeLine: "Creature — Human Soldier", ManaCost: "{W}"},
	)
	lin := effectstest.MmqB1Push(g, me.ID, "Lin Sivvi, Defiant Hero", "Legendary Creature — Human Rebel", oldB680LinSivvi, 1, 3, "W")
	effectstest.DcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, lin, 0, game.ActivateAbilityParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("activate with X=2: %v", err)
	}
	if !effectstest.B16Tapped(t, g, lin) {
		t.Error("Lin Sivvi did not tap")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the search names a quality, so it asks")
	}
	for _, name := range []string{"Big Rebel", "Not Rebel"} {
		if effectstest.SearchOptionNamed(g, c, name) != uuid.Nil {
			t.Errorf("%s was offered", name)
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Edge Rebel")
	if effectstest.FindBattlefieldByName(g, "Edge Rebel") == uuid.Nil {
		t.Error("the Rebel with mana value 2 did not enter the battlefield")
	}
}

func TestOldB680LinSivviTucksARebelFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lin := effectstest.MmqB1Push(g, me.ID, "Lin Sivvi, Defiant Hero", "Legendary Creature — Human Rebel", oldB680LinSivvi, 1, 3, "W")
	rebel := effectstest.PushGraveyardCardTyped(me, "Dead Rebel", oldB680RebelType)
	soldier := effectstest.PushGraveyardCardTyped(me, "Dead Soldier", "Creature — Human Soldier")
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, lin, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(soldier)}); err == nil {
		t.Fatal("a non-Rebel card was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, lin, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(rebel)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(rebel) {
		t.Error("the Rebel stayed in the graveyard")
	}
	if n := len(me.Library.Cards); n == 0 || me.Library.Cards[0].InstanceID != rebel {
		t.Error("the Rebel is not on the bottom of the library")
	}
	if effectstest.B16Tapped(t, g, lin) {
		t.Error("the second ability tapped Lin Sivvi")
	}
}

// --- Treetop Bracers ---------------------------------------------------------------

func TestOldB680BracersPumpsAndOnlyFliersMayBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Push(g, me.ID, "Bear", oldB680BearType, "", 2, 2)
	ground := effectstest.B12Push(g, opp.ID, "Ground Blocker", oldB680BearType, "", 2, 2)
	flier := effectstest.PushCatalogPermanent(g, opp.ID, "Flier", "Creature — Bird", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 1, 1
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Treetop Bracers", "Enchantment — Aura", oldB680Bracers, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("enchanted creature is %d/%d, want 3/3", p, tough)
	}
	effectstest.Pr7bAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(ground, bear); err == nil {
		t.Error("a creature without flying blocked it")
	}
	if err := g.DeclareBlocker(flier, bear); err != nil {
		t.Errorf("a creature with flying could not block it: %v", err)
	}
}

// --- Rath's Edge -------------------------------------------------------------------

func TestOldB680RathsEdgeTapsForColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	edge := effectstest.PushCatalogPermanent(g, me.ID, "Rath's Edge", "Legendary Land", oldB680RathsEdge, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, edge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("mana pool holds %d, want 1", me.ManaPool.Total())
	}
}

func TestOldB680RathsEdgeSacrificesAnotherLandForOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	edge := effectstest.PushCatalogPermanent(g, me.ID, "Rath's Edge", "Legendary Land", oldB680RathsEdge, false)
	forest := effectstest.PushCatalogPermanent(g, me.ID, "Forest", oldB680ForestType, "", false)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}")
	params := game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{forest}, Targets: effectstest.PlayerRefs(opp.ID)}
	if err := g.ActivateCatalogAbility(me.ID, edge, 0, params); err == nil {
		t.Fatal("activated for three mana")
	}
	effectstest.DcMana(t, g, me.ID, "{C}")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, edge, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) || !g.Battlefield.Contains(edge) {
		t.Error("the Forest should be sacrificed and Rath's Edge kept")
	}
	if opp.Life != life-1 {
		t.Errorf("target life %d, want %d", opp.Life, life-1)
	}
}

func TestOldB680RathsEdgeMayTapAndSacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	edge := effectstest.PushCatalogPermanent(g, me.ID, "Rath's Edge", "Legendary Land", oldB680RathsEdge, false)
	bear := effectstest.B12Push(g, opp.ID, "Bear", oldB680BearType, "", 1, 1)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, edge, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{edge}, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate sacrificing itself: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(edge) {
		t.Error("Rath's Edge was not sacrificed")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 1/1 survived the damage from the sacrificed land")
	}
}
