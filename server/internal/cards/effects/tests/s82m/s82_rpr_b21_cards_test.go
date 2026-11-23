package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprB21OjutaiOracle    = "654ff673-4630-4068-9bcf-e88a7877eb58"
	rprB21DreamsOracle    = "77acdc13-e1b5-4a6c-9be1-b987e8256f10"
	rprB21EsperOracle     = "6b9cd3d0-4316-4945-b960-12f51052d260"
	rprB21LeapOracle      = "44cf5ef4-2b0f-4d9d-bf87-8835424721d0"
	rprB21GahijiOracle    = "110d5d45-7174-4d1b-9e9c-e1ccd950ee3b"
	rprB21MarauderOracle  = "33f29fcb-c9ea-4d81-afa2-5c9575b4d58e"
	rprB21GrixisOracle    = "743f4488-fef1-4f4d-b745-d2de92423e00"
	rprB21ThiefOracle     = "a200daec-983c-4ea6-ae98-1c963b50fbb1"
	rprB21StomphowlOracle = "de47a1e8-9c69-4af6-9d72-1bdd41352b32"
	rprB21InkfathomOracle = "60b6e96a-63e6-4520-b2dd-b20b7dfbcd57"
	rprB21JenaraOracle    = "d06cd670-7ffc-4295-97d1-0eff042fb6d5"
)

func TestRprB21EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB21OjutaiOracle: "Dragonlord Ojutai", rprB21DreamsOracle: "Drawn from Dreams",
		rprB21EsperOracle: "Esper Panorama", rprB21LeapOracle: "Evolutionary Leap",
		rprB21GahijiOracle: "Gahiji, Honored One", rprB21MarauderOracle: "Graveblade Marauder",
		rprB21GrixisOracle: "Grixis Panorama", rprB21ThiefOracle: "Identity Thief",
		rprB21StomphowlOracle: "Indrik Stomphowler", rprB21InkfathomOracle: "Inkfathom Witch",
		rprB21JenaraOracle: "Jenara, Asura of War",
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

// --- Dragonlord Ojutai ---

func TestRprB21OjutaiHasHexproofOnlyWhileUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Dragonlord Ojutai", "Legendary Creature — Elder Dragon", rprB21OjutaiOracle, 5, 4)
	if !effectstest.HasEffectiveKeyword(t, g, id, "hexproof") {
		t.Fatal("untapped Ojutai lacks hexproof")
	}
	effectstest.StdB62Tap(g, id)
	if effectstest.HasEffectiveKeyword(t, g, id, "hexproof") {
		t.Fatal("tapped Ojutai still has hexproof")
	}
}

func rprB21Creature(g *game.Game, owner uuid.UUID, name string, p, t int) uuid.UUID {
	return effectstest.PushVanillaCreature(g, owner, name, p, t)
}

func TestRprB21OjutaiDigsThreeOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D")
	ojutai := effectstest.B12Push(g, me.ID, "Dragonlord Ojutai", "Legendary Creature — Elder Dragon", rprB21OjutaiOracle, 5, 4)
	handBefore := me.Hand.Size()

	effectstest.AttackWith(t, g, opp.ID, ojutai)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no take prompt")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 3 {
		t.Fatalf("take exactly one of three: %d..%d over %d", pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(ids[1]) || me.Hand.Size() != handBefore+1 {
		t.Fatal("the pick did not reach the hand")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 2 {
		t.Fatalf("want a bottom ordering of the two not taken: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[2], ids[0]}, nil); err != nil {
		t.Fatal(err)
	}
	if top, _ := me.Library.Top(); top.InstanceID != ids[3] {
		t.Fatal("the fourth card should now be on top")
	}
	if got := []uuid.UUID{me.Library.Cards[0].InstanceID, me.Library.Cards[1].InstanceID}; !effectstest.HasID(got, ids[0]) || !effectstest.HasID(got, ids[2]) {
		t.Fatal("the two not taken are not the bottom two")
	}
}

// --- Drawn from Dreams ---

func TestRprB21DrawnFromDreamsTakesTwoOfSevenRestToTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F", "G", "H", "I")
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Drawn from Dreams", "Sorcery", rprB21DreamsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no take prompt")
	}
	if pick.ChooseMin != 2 || pick.ChooseMax != 2 || len(pick.ChooseCards) != 7 {
		t.Fatalf("take exactly two of seven: %d..%d over %d", pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0], ids[6]}); err != nil {
		t.Fatal(err)
	}
	if me.Hand.Size() != handBefore+2 || !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[6]) {
		t.Fatal("the two picks are not in hand")
	}
	if top, _ := me.Library.Top(); top.InstanceID != ids[7] {
		t.Fatal("the card below the look should be on top")
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:5] {
		bottom[c.InstanceID] = true
	}
	for _, id := range ids[1:6] {
		if !bottom[id] {
			t.Fatal("the five not taken are not the bottom five")
		}
	}
}

// --- Panoramas ---

func TestRprB21PanoramasFetchTheirThreeBasics(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		want, no     []string
	}{
		{"Esper Panorama", rprB21EsperOracle, []string{"Plains", "Island", "Swamp"}, []string{"Mountain", "Forest"}},
		{"Grixis Panorama", rprB21GrixisOracle, []string{"Island", "Swamp", "Mountain"}, []string{"Plains", "Forest"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			land := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			var basics []game.Card
			for _, n := range append(append([]string{}, tc.want...), tc.no...) {
				basics = append(basics, game.Card{Name: n, TypeLine: "Basic Land — " + n})
			}
			effectstest.SeedSearchLibrary(me, basics...)
			if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("tap for mana: %v", err)
			}
			if got := effectstest.SortedPool(me); len(got) != 1 || got[0] != "C" {
				t.Fatalf("pool %v, want [C]", got)
			}
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == land {
						g.Battlefield.Cards[i].Tapped = false
					}
				}
			})
			effectstest.B06AddMana(me, "C")
			if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatal(err)
			}
			if g.Battlefield.Contains(land) {
				t.Fatal("the Panorama was not sacrificed as a cost")
			}
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			for _, n := range tc.want {
				if effectstest.SearchOptionNamed(g, c, n) == uuid.Nil {
					t.Errorf("%s not offered", n)
				}
			}
			for _, n := range tc.no {
				if effectstest.SearchOptionNamed(g, c, n) != uuid.Nil {
					t.Errorf("%s offered", n)
				}
			}
			effectstest.AnswerSearchByID(t, g, me.ID, effectstest.SearchOptionNamed(g, c, tc.want[0]))
			if !rprB21OnBattlefieldTapped(g, tc.want[0]) {
				t.Fatalf("%s did not enter tapped", tc.want[0])
			}
		})
	}
}

// --- Evolutionary Leap ---

func TestRprB21EvolutionaryLeapFindsACreatureAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	leap := effectstest.PushCatalogPermanent(g, me.ID, "Evolutionary Leap", "Enchantment", rprB21LeapOracle, false)
	fodder := rprB21Creature(g, me.ID, "Fodder", 1, 1)
	effectstest.SeedLibraryTop(me, "Tail", "Sorcery", "")
	hit := effectstest.SeedLibraryTop(me, "Hit", "Creature — Bear", "")
	miss2 := effectstest.SeedLibraryTop(me, "Miss2", "Sorcery", "")
	miss1 := effectstest.SeedLibraryTop(me, "Miss1", "Sorcery", "")
	size := me.Library.Size()
	handBefore := me.Hand.Size()

	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, leap, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without sacrificing a creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, leap, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(hit) || me.Hand.Size() != handBefore+1 {
		t.Fatal("the creature card did not reach the hand")
	}
	if me.Library.Size() != size-1 {
		t.Fatalf("library %d, want %d", me.Library.Size(), size-1)
	}
	if top, _ := me.Library.Top(); top.InstanceID == miss1 || top.InstanceID == miss2 {
		t.Fatal("the revealed non-creatures were left on top")
	}
}

func TestRprB21EvolutionaryLeapWithNoCreatureInLibraryBottomsEverything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	leap := effectstest.PushCatalogPermanent(g, me.ID, "Evolutionary Leap", "Enchantment", rprB21LeapOracle, false)
	fodder := rprB21Creature(g, me.ID, "Fodder", 1, 1)
	me.Library.Cards = nil
	effectstest.SeedLibrary(me, "A", "B", "C")
	handBefore := me.Hand.Size()
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, leap, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore || me.Library.Size() != 3 {
		t.Fatalf("hand %d→%d, library %d", handBefore, me.Hand.Size(), me.Library.Size())
	}
}

func TestRprB21EvolutionaryLeapNeedsItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	leap := effectstest.PushCatalogPermanent(g, me.ID, "Evolutionary Leap", "Enchantment", rprB21LeapOracle, false)
	fodder := rprB21Creature(g, me.ID, "Fodder", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, leap, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("activated with no mana")
	}
	if !g.Battlefield.Contains(fodder) {
		t.Fatal("a refused activation paid the sacrifice")
	}
}

// --- Gahiji, Honored One ---

func TestRprB21GahijiPumpsAttackersAtAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Gahiji, Honored One", "Legendary Creature — Beast", rprB21GahijiOracle, 4, 4)
	a := rprB21Creature(g, me.ID, "Bear A", 2, 2)
	b := rprB21Creature(g, me.ID, "Bear B", 2, 2)
	home := rprB21Creature(g, me.ID, "Bear at home", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, _ := effectstest.EffectivePTOf(t, g, id); p != 4 {
			t.Errorf("attacker power %d, want 4", p)
		}
	}
	if p, _ := effectstest.EffectivePTOf(t, g, home); p != 2 {
		t.Errorf("a creature that did not attack has power %d", p)
	}
}

func TestRprB21GahijiPumpsAnAttackAtAnOpponentsPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Gahiji, Honored One", "Legendary Creature — Beast", rprB21GahijiOracle, 4, 4)
	pw := effectstest.PushWalkerForTest(g, opp.ID, "Their Walker", "", 3)
	a := rprB21Creature(g, me.ID, "Bear A", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a, Target: pw}}); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, a); p != 4 {
		t.Errorf("attacker at a planeswalker has power %d, want 4", p)
	}
}

func TestRprB21GahijiIgnoresAttacksAtItsOwnController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Seat 1 controls Gahiji, so seat 0's attack at seat 1 is not an attack
	// on one of Gahiji's controller's opponents.
	effectstest.B12Push(g, opp.ID, "Gahiji, Honored One", "Legendary Creature — Beast", rprB21GahijiOracle, 4, 4)
	a := rprB21Creature(g, me.ID, "Bear A", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, a); p != 2 {
		t.Errorf("attacker at Gahiji's controller has power %d, want 2", p)
	}
}

// --- Graveblade Marauder ---

func TestRprB21GravebladeMarauderDrainsByCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	m := effectstest.B12Push(g, me.ID, "Graveblade Marauder", "Creature — Human Warrior", rprB21MarauderOracle, 1, 4)
	if !effectstest.HasEffectiveKeyword(t, g, m, "deathtouch") {
		t.Fatal("no deathtouch")
	}
	for _, c := range []game.Card{
		{Name: "Bear 1", TypeLine: "Creature — Bear"}, {Name: "Bear 2", TypeLine: "Creature — Bear"},
		{Name: "Bolt", TypeLine: "Instant"},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		me.Graveyard.PushTop(c)
	}
	// Their graveyard does not count.
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	effectstest.AttackWith(t, g, opp.ID, m)
	effectstest.PassPriorityAroundTable(t, g)
	if want := 40 - 1 - 2; opp.Life != want {
		t.Fatalf("opponent life %d, want %d", opp.Life, want)
	}
}

func TestRprB21GravebladeMarauderWithAnEmptyGraveyardOnlyDealsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	m := effectstest.B12Push(g, me.ID, "Graveblade Marauder", "Creature — Human Warrior", rprB21MarauderOracle, 1, 4)
	effectstest.AttackWith(t, g, opp.ID, m)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Fatalf("opponent life %d, want 39", opp.Life)
	}
}

// --- Identity Thief ---

func TestRprB21IdentityThiefExilesCopiesAndReturnsTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thief := effectstest.B12Push(g, me.ID, "Identity Thief", "Creature — Shapeshifter", rprB21ThiefOracle, 0, 3)
	victim := effectstest.B12Push(g, opp.ID, "Their Giant", "Creature — Giant", "", 6, 6)
	effectstest.DeclareAttack(t, g, opp.ID, thief)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature was not exiled")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, thief); p != 6 || tough != 6 {
		t.Fatalf("Identity Thief is %d/%d, want 6/6", p, tough)
	}
	// The card comes back, under its owner's control, at the next end step.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	back := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Giant" && c.Controller == opp.ID {
			back = true
		}
	}
	if !back {
		t.Fatal("the exiled creature did not return under its owner's control")
	}
}

func TestRprB21IdentityThiefDeclinedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thief := effectstest.B12Push(g, me.ID, "Identity Thief", "Creature — Shapeshifter", rprB21ThiefOracle, 0, 3)
	victim := effectstest.B12Push(g, opp.ID, "Their Giant", "Creature — Giant", "", 6, 6)
	effectstest.DeclareAttack(t, g, opp.ID, thief)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("a declined trigger exiled the creature")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, thief); p != 0 || tough != 3 {
		t.Fatalf("Identity Thief is %d/%d, want 0/3", p, tough)
	}
}

func TestRprB21IdentityThiefCannotTargetItselfOrATokenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thief := effectstest.B12Push(g, me.ID, "Identity Thief", "Creature — Shapeshifter", rprB21ThiefOracle, 0, 3)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID})
	effectstest.DeclareAttack(t, g, opp.ID, thief)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		for _, id := range p.PickTargetCards {
			if id == thief || id == token {
				t.Fatalf("an illegal target is offered: %s", id)
			}
		}
	}
}

// --- Indrik Stomphowler ---

func TestRprB21IndrikStomphowlerDestroysAnArtifactOrEnchantment(t *testing.T) {
	for _, typeLine := range []string{"Artifact", "Enchantment"} {
		t.Run(typeLine, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			rock := effectstest.PushCatalogPermanent(g, opp.ID, "Their "+typeLine, typeLine, "", false)
			bear := rprB21Creature(g, opp.ID, "Their Bear", 2, 2)
			effectstest.CastHoldCreature(t, g, "Indrik Stomphowler", "Creature — Beast", rprB21StomphowlOracle, 4, 4)
			effectstest.B04WaitForPick(t, g, me.ID)
			if p := effectstest.LatestPickTarget(g, me.ID); effectstest.HasID(p.PickTargetCards, bear) {
				t.Fatal("a creature is a legal target")
			}
			effectstest.PickCard(t, g, me.ID, rock)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(rock) {
				t.Fatalf("the %s survived", typeLine)
			}
			if !g.Battlefield.Contains(bear) {
				t.Fatal("the bear was destroyed")
			}
		})
	}
}

func TestRprB21IndrikStomphowlerWithNoTargetDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := rprB21Creature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Indrik Stomphowler", "Creature — Beast", rprB21StomphowlOracle, 4, 4)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was offered with no artifact or enchantment in play")
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the bear was destroyed")
	}
}

// --- Inkfathom Witch ---

func TestRprB21InkfathomWitchSetsUnblockedCreaturesToFourOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	witch := effectstest.B12Push(g, me.ID, "Inkfathom Witch", "Creature — Merfolk Wizard", rprB21InkfathomOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, witch, "fear") {
		t.Fatal("no fear")
	}
	unblocked := rprB21Creature(g, me.ID, "Unblocked", 3, 3)
	blocked := rprB21Creature(g, me.ID, "Blocked", 2, 5)
	blocker := rprB21Creature(g, opp.ID, "Blocker", 1, 1)
	idle := rprB21Creature(g, me.ID, "Idle", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, unblocked, blocked)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, blocked); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.B06AddMana(me, "C", "C", "U", "B")
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, unblocked); p != 4 || tough != 1 {
		t.Errorf("unblocked creature is %d/%d, want 4/1", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, blocked); p != 2 || tough != 5 {
		t.Errorf("blocked creature is %d/%d, want 2/5", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, idle); p != 2 || tough != 2 {
		t.Errorf("non-attacking creature is %d/%d, want 2/2", p, tough)
	}
}

func TestRprB21InkfathomWitchNeedsItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	witch := effectstest.B12Push(g, me.ID, "Inkfathom Witch", "Creature — Merfolk Wizard", rprB21InkfathomOracle, 1, 1)
	effectstest.B06AddMana(me, "C", "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {B}")
	}
}

// --- Jenara, Asura of War ---

func TestRprB21JenaraGrowsForOneWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	j := effectstest.B12Push(g, me.ID, "Jenara, Asura of War", "Legendary Creature — Angel", rprB21JenaraOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, j, "flying") {
		t.Fatal("no flying")
	}
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, j, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {W}")
	}
	effectstest.B06AddMana(me, "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, j, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePowerWithCounters(t, g, j); p != 4 {
		t.Fatalf("Jenara has power %d, want 4", p)
	}
}

func rprB21OnBattlefieldTapped(g *game.Game, name string) bool {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Tapped {
			return true
		}
	}
	return false
}
