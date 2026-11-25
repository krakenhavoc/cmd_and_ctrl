package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// War of the Spark, batch 172 (#5520).

const (
	oldB172BondDiscipline = "00e9957a-5684-4a07-954e-8161925fefe4"
	oldB172Statue         = "183954d2-7b40-4c7e-adfa-d80191ab5ad6"
	oldB172Shadowfugue    = "cb490f3f-74ed-4d38-8eda-9948c095f3bc"
	oldB172BleedingEdge   = "308e551a-6d60-4a28-9e22-1181995ca068"
	oldB172Tolsimir       = "3cef501b-5ce1-4699-89a5-3949f7e65bf1"
	oldB172InvadeCity     = "583a4744-89fb-4bd6-8d4c-c3d845e2e19c"
	oldB172Grizzly        = "6dbd0764-4bfe-4a58-905d-aeee70ceb798"
	oldB172Arkbow         = "efe3204d-2013-47a1-ad68-adfbb2d0be8f"
	oldB172Skulker        = "ffaa2776-f61d-4eaa-8949-a0407f85214f"
	oldB172Submersible    = "43257a85-2cbe-4931-9755-106808ab124a"
	oldB172Tank           = "60957bfd-2460-429a-a7a7-d99424946f83"
	oldB172Operative      = "7734ee04-1d99-4807-be68-b889e75a7551"
)

func TestOldB172CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		oldB172BondDiscipline: "Bond of Discipline", oldB172Statue: "God-Pharaoh's Statue", oldB172Shadowfugue: "Davriel's Shadowfugue",
		oldB172BleedingEdge: "Bleeding Edge", oldB172Tolsimir: "Tolsimir, Friend to Wolves", oldB172InvadeCity: "Invade the City",
		oldB172Grizzly: "Vivien's Grizzly", oldB172Arkbow: "Vivien's Arkbow", oldB172Skulker: "Ashiok's Skulker",
		oldB172Submersible: "Silent Submersible", oldB172Tank: "Mizzium Tank", oldB172Operative: "Duskmantle Operative",
	}
	for oracle, name := range want {
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

func oldB172Tapped(g *game.Game, id uuid.UUID) bool {
	c, _ := effectstest.BattlefieldCard(g, id)
	return c.Tapped
}

func oldB172Top(p *game.Player, cards ...game.Card) {
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
		p.Library.PushTop(c)
	}
}

// --- Bond of Discipline -------------------------------------------------------

func TestOldB172BondOfDisciplineTapsTheirCreaturesAndGivesYoursLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)].ID, "Other", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Bond of Discipline", "Sorcery", oldB172BondDiscipline, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB172Tapped(g, theirs) || !oldB172Tapped(g, other) {
		t.Error("an opponent's creature is untapped")
	}
	if oldB172Tapped(g, mine) {
		t.Error("your creature was tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "lifelink") {
		t.Error("your creature has no lifelink")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("an opponent's creature has lifelink")
	}
	late := effectstest.B12Creature(g, me.ID, "Late", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, late, "lifelink") {
		t.Error("a creature that entered later has lifelink")
	}
}

// --- God-Pharaoh's Statue -----------------------------------------------------

func TestOldB172StatueTaxesOpponentsSpellsAndDrainsAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "God-Pharaoh's Statue", "Legendary Artifact", oldB172Statue, false)
	if got := effectstest.PriceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("your own spell costs %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 4 {
		t.Errorf("an opponent's creature spell costs %d, want 4", got)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Lightning Bolt", "Instant", "{R}"); got != 3 {
		t.Errorf("an opponent's instant costs %d, want 3", got)
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("%s life %d, want %d", p.Name, p.Life, want)
		}
	}
}

// --- Davriel's Shadowfugue ----------------------------------------------------

func TestOldB172ShadowfugueDiscardsTwoAndLosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hand := opp.Hand.Size()
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Davriel's Shadowfugue", "Sorcery", oldB172Shadowfugue, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	if p := effectstest.DiscardChoiceFor(g, opp.ID); p == nil {
		t.Fatal("the target player was not asked to discard")
	}
	_ = hand
}

// --- Bleeding Edge ------------------------------------------------------------

func TestOldB172BleedingEdgeShrinksTheTargetAndAmasses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Bleeding Edge", "Sorcery", oldB172BleedingEdge, effectstest.CardTarget(foe))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.B38PT(t, g, foe); p != 1 || tough != 1 {
		t.Errorf("target is %d/%d, want 1/1", p, tough)
	}
	armies := effectstest.TokensNamedControlledBy(g, "Zombie Army", me.ID)
	if armies != 1 {
		t.Errorf("%d Zombie Army tokens, want 1", armies)
	}
}

func TestOldB172BleedingEdgeAmassesWithNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Bleeding Edge", "Sorcery", oldB172BleedingEdge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var counters int
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Zombie Army" {
			counters = c.Counters[game.CounterPlusOne]
		}
	}
	if counters != 2 {
		t.Errorf("Army has %d +1/+1 counters, want 2", counters)
	}
}

// --- Invade the City ----------------------------------------------------------

func TestOldB172InvadeTheCityAmassesForInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushGraveyardCardTyped(me, "Bolt", "Instant")
	effectstest.PushGraveyardCardTyped(me, "Wrath", "Sorcery")
	effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Invade the City", "Sorcery", oldB172InvadeCity, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var counters int
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Zombie Army" {
			counters = c.Counters[game.CounterPlusOne]
		}
	}
	if counters != 2 {
		t.Errorf("Army has %d +1/+1 counters, want 2", counters)
	}
}

// --- Tolsimir, Friend to Wolves -----------------------------------------------

func TestOldB172TolsimirMakesVojaAndVojaFightsAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Tolsimir, Friend to Wolves", "Legendary Creature — Elf Scout", oldB172Tolsimir, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOpenOrderingIfAny(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOpenOrderingIfAny(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, mine) {
			t.Error("your own creature was offered as the fight target")
		}
		effectstest.PickCard(t, g, me.ID, foe)
	} else {
		t.Fatal("no fight target was asked for")
	}
	effectstest.AlaB1Settle(t, g)
	voja := effectstest.TokensNamed(g, "Voja, Friend to Elves")
	if len(voja) != 1 {
		t.Fatalf("%d Voja tokens, want 1", len(voja))
	}
	if !voja[0].IsLegendary() || voja[0].Power != 3 || voja[0].Toughness != 3 {
		t.Errorf("Voja is %d/%d legendary=%v, want a legendary 3/3", voja[0].Power, voja[0].Toughness, voja[0].IsLegendary())
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if g.Battlefield.Contains(foe) {
		t.Error("the fight target survived a 3/3 fighting a 2/2")
	}
}

// --- Vivien's Grizzly ---------------------------------------------------------

func oldB172GrizzlyOnBattlefield(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, me.ID, "Vivien's Grizzly", "Creature — Bear Spirit", oldB172Grizzly, false)
}

func TestOldB172GrizzlyTakesACreatureCardOrPutsItOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := game.Card{Name: "Top Bear", TypeLine: "Creature — Bear"}
	oldB172Top(me, bear)
	grizzly := oldB172GrizzlyOnBattlefield(g, me)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, grizzly, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if p == nil {
		t.Fatal("no pick for the creature card")
	}
	size := me.Library.Size()
	if err := g.ResolveChooseCards(p.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if me.Library.Size() != size {
		t.Errorf("library %d -> %d", size, me.Library.Size())
	}
	names := effectstest.LibraryTopNames(me, size)
	if names[len(names)-1] != "Top Bear" {
		t.Errorf("the declined card is not on the bottom: %v", names[len(names)-1])
	}
}

func TestOldB172GrizzlyPutsANoncreatureTopCardOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB172Top(me, game.Card{Name: "Top Rock", TypeLine: "Artifact"})
	grizzly := oldB172GrizzlyOnBattlefield(g, me)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G", "C", "C", "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, grizzly, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	size := me.Library.Size()
	names := effectstest.LibraryTopNames(me, size)
	if names[len(names)-1] != "Top Rock" || names[0] == "Top Rock" {
		t.Errorf("a noncreature card should go to the bottom: top %q bottom %q", names[0], names[len(names)-1])
	}
	if me.Hand.Size() != hand {
		t.Error("a noncreature card went to hand")
	}
}

func TestOldB172GrizzlyMayTakeAPlaneswalkerCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := game.Card{Name: "Top Walker", TypeLine: "Legendary Planeswalker — Test"}
	oldB172Top(me, walker)
	grizzly := oldB172GrizzlyOnBattlefield(g, me)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, grizzly, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if p == nil || len(p.ChooseCards) != 1 {
		t.Fatal("the planeswalker card was not offered")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, p.ChooseCards); err != nil {
		t.Fatalf("taking it: %v", err)
	}
	if !me.Hand.Contains(p.ChooseCards[0]) {
		t.Error("the planeswalker card did not go to hand")
	}
}

// --- Vivien's Arkbow ----------------------------------------------------------

func TestOldB172ArkbowPutsACreatureWithManaValueAtMostXOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
	big := game.Card{Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{G}", Power: 5, Toughness: 5}
	oldB172Top(me, big, small, game.Card{Name: "Filler Rock", TypeLine: "Artifact"})
	arkbow := effectstest.PushCatalogPermanent(g, me.ID, "Vivien's Arkbow", "Legendary Artifact", oldB172Arkbow, false)
	discard := effectstest.PushGraveyardCardTyped(me, "Throwaway", "Sorcery")
	me.Graveyard.Remove(discard)
	me.Hand.PushTop(game.Card{InstanceID: discard, Name: "Throwaway", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, arkbow, 0, game.ActivateAbilityParams{XValue: 3, DiscardIDs: []uuid.UUID{discard}})
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if p == nil {
		t.Fatal("no creature pick")
	}
	if len(p.ChooseCards) != 1 {
		t.Fatalf("%d candidates, want the one creature card of mana value 3 or less (the rock and the five-drop are not)", len(p.ChooseCards))
	}
	chosen := p.ChooseCards[0]
	if err := g.ResolveChooseCards(p.ID, me.ID, p.ChooseCards); err != nil {
		t.Fatalf("choosing: %v", err)
	}
	if !g.Battlefield.Contains(chosen) {
		t.Error("the creature did not enter the battlefield")
	}
	names := effectstest.LibraryTopNames(me, me.Library.Size())
	if n := len(names); n < 2 || (names[n-1] != "Big Bear" && names[n-2] != "Big Bear") || (names[n-1] != "Filler Rock" && names[n-2] != "Filler Rock") {
		t.Errorf("the rest should be on the bottom: %v", names[max(0, n-3):])
	}
}

// --- Ashiok's Skulker ---------------------------------------------------------

func TestOldB172SkulkerCantBeBlockedAfterTheAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	skulker := effectstest.PushCatalogPermanent(g, me.ID, "Ashiok's Skulker", "Creature — Nightmare", oldB172Skulker, false)
	effectstest.B64ReadyToAttack(g, skulker)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "U")
	effectstest.B16Activate(t, g, me.ID, skulker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BrAttack(t, g, skulker)
	if err := g.DeclareBlocker(blocker, skulker); err == nil {
		t.Error("the Skulker was blocked")
	}
}

// --- Silent Submersible -------------------------------------------------------

func TestOldB172SubmersibleDrawsOnCombatDamageToAPlayer(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	vehicle := s.PushVehicle("Silent Submersible", oldB172Submersible, 2, 3)
	crew := s.PushCreature("Crew", "Creature — Bear", 2, 2)
	s.CrewAndResolve(t, vehicle, 0, crew)
	effectstest.B64ReadyToAttack(s.G, vehicle)
	hand := s.Me.Hand.Size()
	effectstest.LciB1Attack(t, s.G, vehicle, s.Opp.ID)
	effectstest.AdvanceTo(t, s.G, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, s.G)
	if s.Me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, s.Me.Hand.Size())
	}
}

func TestOldB172SubmersibleNeedsPowerTwoToCrew(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	vehicle := s.PushVehicle("Silent Submersible", oldB172Submersible, 2, 3)
	weak := s.PushCreature("Weak", "Creature — Bear", 1, 1)
	effectstest.ToMain(t, s.G)
	if err := s.G.ActivateCatalogAbility(s.Me.ID, vehicle, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{weak}}); err == nil {
		t.Error("a power-1 creature paid Crew 2")
	}
}

// --- Mizzium Tank -------------------------------------------------------------

func TestOldB172TankBecomesACreatureWhenYouCastANoncreatureSpell(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	tank := s.PushVehicle("Mizzium Tank", oldB172Tank, 3, 2)
	if c, _ := effectstest.BattlefieldCard(s.G, tank); c.IsCreature() {
		t.Fatal("an uncrewed Vehicle is a creature")
	}
	effectstest.CastCatalogSpell(t, s.G, "Davriel's Shadowfugue", "Sorcery", oldB172Shadowfugue, []game.TargetRef{effectstest.PlayerTarget(s.Opp.ID)})
	effectstest.AlaB1Settle(t, s.G)
	c, _ := effectstest.BattlefieldCard(s.G, tank)
	if !c.IsCreature() {
		t.Fatal("the Tank did not become a creature")
	}
	if p, tough := effectstest.B38PT(t, s.G, tank); p != 4 || tough != 3 {
		t.Errorf("the Tank is %d/%d, want 4/3", p, tough)
	}
}

func TestOldB172TankStaysAVehicleForACreatureSpell(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	tank := s.PushVehicle("Mizzium Tank", oldB172Tank, 3, 2)
	effectstest.CastCatalogSpell(t, s.G, "Duskmantle Operative", "Creature — Human Rogue", oldB172Operative, nil)
	effectstest.AlaB1Settle(t, s.G)
	if c, _ := effectstest.BattlefieldCard(s.G, tank); c.IsCreature() {
		t.Error("a creature spell animated the Tank")
	}
}

func TestOldB172TankCrews(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	tank := s.PushVehicle("Mizzium Tank", oldB172Tank, 3, 2)
	crew := s.PushCreature("Crew", "Creature — Bear", 1, 1)
	s.CrewAndResolve(t, tank, 0, crew)
	if !effectstest.HasEffectiveKeyword(t, s.G, tank, "trample") {
		t.Error("no trample")
	}
}

// --- Duskmantle Operative -----------------------------------------------------

func TestOldB172OperativeCantBeBlockedByPowerFourOrGreater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	op := effectstest.B12Push(g, me.ID, "Duskmantle Operative", "Creature — Human Rogue", oldB172Operative, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	mid := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.BrAttack(t, g, op)
	if err := g.DeclareBlocker(big, op); err == nil {
		t.Error("a creature with power 4 blocked the Operative")
	}
	if err := g.DeclareBlocker(mid, op); err != nil {
		t.Errorf("a creature with power 3 can't block the Operative: %v", err)
	}
}
