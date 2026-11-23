package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b09_cards_test.go — Innistrad Remastered creatures, S82 GAP
// batch 9 (#5948), played through the real catalog entries.

const (
	gapB09TaskmasterOracle = "162e45f5-edb8-4a13-acdc-8980ce5ba8ac"
	gapB09TorturerOracle   = "f578a5e0-77e2-4c35-90b2-6628ed9b218e"
	gapB09GrafRatsOracle   = "852edc00-9d98-427c-9869-7d0c8c121118"
	gapB09HauntedOracle    = "03259490-a789-4f96-8c5b-b410425bdd02"
	gapB09ScavOracle       = "1c29ae09-cc93-4a2a-b432-3c78ae01c1ed"
	gapB09BansheeOracle    = "e2fe1847-4003-41c5-a64b-03afd888b81e"
	gapB09PriestOracle     = "c6064c08-e4e0-49f4-9f0b-55cdecf443af"
	gapB09BloodhallOracle  = "2eed5eb1-ab67-4f23-885d-922131364aba"
	gapB09OliviaOracle     = "dd27d0f2-fe1e-4b63-9ba0-6264ca3c70f8"
	gapB09TrawlerOracle    = "870ca989-0e72-4e74-9570-57e129298f2e"
	gapB09HostOracle       = "6923cf6b-7d3e-4d95-abaf-1df1a04ac7c1"
)

func TestGapB09EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		gapB09TaskmasterOracle: "Demonic Taskmaster", gapB09TorturerOracle: "Falkenrath Torturer",
		gapB09GrafRatsOracle: "Graf Rats", gapB09HauntedOracle: "Haunted Dead",
		gapB09ScavOracle: "Midnight Scavengers", gapB09BansheeOracle: "Morkrut Banshee",
		gapB09PriestOracle: "Skirsdag High Priest", gapB09BloodhallOracle: "Bloodhall Priest",
		gapB09OliviaOracle: "Olivia Voldaren", gapB09TrawlerOracle: "Drunau Corpse Trawler",
		gapB09HostOracle: "Chittering Host",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func gapB09Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- Drunau Corpse Trawler ---

func TestGapB09TrawlerMakesAZombieAndGivesAZombieDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Drunau Corpse Trawler", "Creature — Zombie", gapB09TrawlerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); got != 1 {
		t.Fatalf("%d Zombie tokens, want 1", got)
	}
	trawler := effectstest.BattlefieldIDNamed(g, me.ID, "Drunau Corpse Trawler")
	token := effectstest.BattlefieldIDNamed(g, me.ID, "Zombie")
	if effectstest.HasEffectiveKeyword(t, g, token, "deathtouch") {
		t.Fatal("the token has deathtouch before the ability")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, trawler, 0, effectstest.BroB04Pool("{2}{B}"), gapB09Target(token)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, token, "deathtouch") {
		t.Error("the Zombie did not gain deathtouch")
	}
}

func TestGapB09TrawlerRefusesANonZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trawler := effectstest.B12Push(g, me.ID, "Drunau Corpse Trawler", "Creature — Zombie", gapB09TrawlerOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.HobB4Activate(t, g, me.ID, trawler, 0, effectstest.BroB04Pool("{2}{B}"), gapB09Target(bear)); err == nil {
		t.Error("a Bear is not a legal target")
	}
}

// --- Haunted Dead ---

func TestGapB09HauntedDeadMakesAFlyingSpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Haunted Dead", "Creature — Zombie", gapB09HauntedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	spirit := effectstest.BattlefieldIDNamed(g, me.ID, "Spirit")
	if spirit == uuid.Nil || !effectstest.HasEffectiveKeyword(t, g, spirit, "flying") {
		t.Errorf("no flying Spirit token (id %v)", spirit)
	}
}

func gapB09DeadInYard(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.PushCatalogGraveyardCard(me, "Haunted Dead", "Creature — Zombie", gapB09HauntedOracle, 2, 2)
}

func TestGapB09HauntedDeadReturnsTappedForTwoDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := gapB09DeadInYard(g, me)
	a := effectstest.PushCatalogHandCard(me, "Fodder A", "Sorcery", "")
	b := effectstest.PushCatalogHandCard(me, "Fodder B", "Sorcery", "")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, effectstest.BroB04Pool("{1}{B}")); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{a, b}, Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	zone, c := effectstest.CardWhere(g, me, id)
	if zone != "battlefield" || !c.Tapped {
		t.Errorf("Haunted Dead is in %s (tapped %v), want the battlefield tapped", zone, c.Tapped)
	}
	if me.Hand.Contains(a) || me.Hand.Contains(b) {
		t.Error("the two cards were not discarded")
	}
}

func TestGapB09HauntedDeadCantPayWithOneCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	for me.Hand.Size() > 1 {
		g.WithWriteLock(func() { me.Hand.Cards = me.Hand.Cards[:len(me.Hand.Cards)-1] })
	}
	id := gapB09DeadInYard(g, me)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, effectstest.BroB04Pool("{1}{B}"), nil); err == nil {
		t.Error("two cards can't be discarded from a hand of one")
	}
}

// --- Demonic Taskmaster ---

func TestGapB09TaskmasterFliesAndEatsAnotherCreatureEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	tm := effectstest.B12Push(g, me.ID, "Demonic Taskmaster", "Creature — Demon", gapB09TaskmasterOracle, 4, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AssertKeywords(t, g, tm, "flying")
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if zone, _ := effectstest.CardWhere(g, me, fodder); zone == "battlefield" {
		t.Error("the other creature was not sacrificed")
	}
	if zone, _ := effectstest.CardWhere(g, me, tm); zone != "battlefield" {
		t.Errorf("the Taskmaster is in %s, want the battlefield", zone)
	}
}

func TestGapB09TaskmasterAloneSacrificesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	tm := effectstest.B12Push(g, me.ID, "Demonic Taskmaster", "Creature — Demon", gapB09TaskmasterOracle, 4, 3)
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if zone, _ := effectstest.CardWhere(g, me, tm); zone != "battlefield" {
		t.Errorf("the lone Taskmaster is in %s, want the battlefield", zone)
	}
}

// --- Falkenrath Torturer ---

func gapB09Sacrifice(t *testing.T, g *game.Game, me *game.Player, src, fodder uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB09TorturerGrowsOnlyWhenAHumanIsSacrificed(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		counters int
	}{{"Creature — Human Soldier", 1}, {"Creature — Bear", 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		tort := effectstest.B12Push(g, me.ID, "Falkenrath Torturer", "Creature — Vampire", gapB09TorturerOracle, 2, 1)
		fodder := effectstest.B12Creature(g, me.ID, "Fodder", tc.typeLine, 1, 1)
		effectstest.AdvanceToMain(t, g)
		gapB09Sacrifice(t, g, me, tort, fodder)
		if !effectstest.HasEffectiveKeyword(t, g, tort, "flying") {
			t.Errorf("%s: no flying", tc.typeLine)
		}
		if got := effectstest.LayeredCard(t, g, tort).CurrentPower(); got != 2+tc.counters {
			t.Errorf("%s: power %d, want %d", tc.typeLine, got, 2+tc.counters)
		}
	}
}

// --- Morkrut Banshee ---

func TestGapB09BansheeShrinksAfterADeathOnly(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 4)
		if died {
			spare := effectstest.B12Creature(g, opp.ID, "Spare", "Creature — Bear", 1, 1)
			effectstest.KillCreature(t, g, opp.ID, spare)
			g.RunStateChecksForTest()
		}
		effectstest.CastCatalogSpell(t, g, "Morkrut Banshee", "Creature — Spirit", gapB09BansheeOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			effectstest.PickTriggerTarget(t, g, me.ID, victim)
			effectstest.PassPriorityAroundTable(t, g)
		}
		zone, _ := effectstest.CardWhere(g, opp, victim)
		if died && zone == "battlefield" {
			t.Error("morbid: the 3/4 survived -4/-4")
		}
		if !died && zone != "battlefield" {
			t.Errorf("no morbid: the Victim is in %s", zone)
		}
	}
}

// --- Midnight Scavengers ---

func TestGapB09ScavengersReturnsACheapCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cheap := effectstest.GraveyardCardOf(g, me, "Cheap Bear", "Creature — Bear", "{2}", "")
	dear := effectstest.GraveyardCardOf(g, me, "Dear Bear", "Creature — Bear", "{4}", "")
	effectstest.CastHoldCreature(t, g, "Midnight Scavengers", "Creature — Human Rogue", gapB09ScavOracle, 3, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, dear) {
		t.Error("a mana value 4 creature card is offered")
	}
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(cheap) {
		t.Error("the mana value 2 creature card was not returned to hand")
	}
	if me.Hand.Contains(dear) {
		t.Error("the mana value 4 creature card was returned")
	}
}

// --- Skirsdag High Priest ---

func TestGapB09HighPriestMakesADemonAfterADeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	priest := effectstest.B12Push(g, me.ID, "Skirsdag High Priest", "Creature — Human Cleric", gapB09PriestOracle, 1, 2)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	tapIDs := []uuid.UUID{a, b}
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{TapIDs: tapIDs, Strict: true}); err == nil {
		t.Fatal("morbid: activating with no creature dead this turn must be refused")
	}
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 1)
	effectstest.KillCreature(t, g, opp.ID, victim)
	g.RunStateChecksForTest()
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{TapIDs: tapIDs, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	demon := effectstest.BattlefieldIDNamed(g, me.ID, "Demon")
	if demon == uuid.Nil {
		t.Fatal("no Demon token")
	}
	if p, tough := effectstest.EffectivePower(t, g, demon), effectstest.EffectiveToughness(t, g, demon); p != 5 || tough != 5 || !effectstest.HasEffectiveKeyword(t, g, demon, "flying") {
		t.Errorf("Demon is %d/%d, want a 5/5 flier", p, tough)
	}
	for _, id := range append(tapIDs, priest) {
		if _, c := effectstest.CardWhere(g, me, id); !c.Tapped {
			t.Errorf("%s is untapped after paying", c.Name)
		}
	}
}

// --- Bloodhall Priest ---

func gapB09EmptyHand(me *game.Player) {
	me.Hand.Cards = nil
}

func TestGapB09BloodhallPriestPingsOnlyWithAnEmptyHand(t *testing.T) {
	for _, empty := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.AdvanceToMain(t, g)
		effectstest.CastCatalogSpell(t, g, "Bloodhall Priest", "Creature — Vampire Cleric", gapB09BloodhallOracle, nil)
		if empty {
			g.WithWriteLock(func() { gapB09EmptyHand(me) })
		}
		before := opp.Life
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
		}
		want := before
		if empty {
			want = before - 2
		}
		if opp.Life != want {
			t.Errorf("empty hand %v: life %d → %d, want %d", empty, before, opp.Life, want)
		}
	}
}

func TestGapB09BloodhallPriestHasMadness(t *testing.T) {
	spec, ok := effects.Lookup(gapB09BloodhallOracle)
	if !ok || spec.Madness != "{1}{B}{R}" {
		t.Errorf("madness %q", spec.Madness)
	}
}

// --- Olivia Voldaren ---

func TestGapB09OliviaPingsMakesAVampireAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	olivia := effectstest.B12Push(g, me.ID, "Olivia Voldaren", "Legendary Creature — Vampire", gapB09OliviaOracle, 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 3)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.HobB4Activate(t, g, me.ID, olivia, 0, effectstest.BroB04Pool("{1}{R}"), gapB09Target(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, bear); c.DamageMarked != 1 || !effects.HasFold(c.Effective().Subtypes, "Vampire") || !effects.HasFold(c.Effective().Subtypes, "Bear") {
		t.Errorf("Bear: damage %d, subtypes %v; want 1 damage and Bear + Vampire", c.DamageMarked, c.Effective().Subtypes)
	}
	if got := effectstest.LayeredCard(t, g, olivia).CurrentPower(); got != 4 {
		t.Errorf("Olivia power %d, want 4", got)
	}
	// She can't ping herself.
	if err := effectstest.HobB4Activate(t, g, me.ID, olivia, 0, effectstest.BroB04Pool("{1}{R}"), gapB09Target(olivia)); err == nil {
		t.Error("Olivia targeted herself")
	}
}

func TestGapB09OliviaStealsAVampireWhileYouControlHer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	olivia := effectstest.B12Push(g, me.ID, "Olivia Voldaren", "Legendary Creature — Vampire", gapB09OliviaOracle, 3, 3)
	bat := effectstest.B12Creature(g, opp.ID, "Blood Bat", "Creature — Vampire Bat", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.HobB4Activate(t, g, me.ID, olivia, 1, effectstest.BroB04Pool("{3}{B}{B}"), gapB09Target(bear)); err == nil {
		t.Fatal("a Bear is not a Vampire")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, olivia, 1, effectstest.BroB04Pool("{3}{B}{B}"), gapB09Target(bat)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, bat); c.Controller != me.ID {
		t.Fatal("the Vampire was not stolen")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(olivia) })
	g.RunStateChecksForTest()
	if c := effectstest.LayeredCard(t, g, bat); c.Controller != opp.ID {
		t.Error("the Vampire did not go back when Olivia left")
	}
}

// --- Graf Rats, Midnight Scavengers and Chittering Host ---

func gapB09MeldPrint() *game.MeldPrint {
	return &game.MeldPrint{
		ResultOracleID:   gapB09HostOracle,
		ResultScryfallID: uuid.NewString(),
		Result: game.Face{
			Name: "Chittering Host", TypeLine: "Creature — Eldrazi Horror",
			Power: 5, Toughness: 6, Keywords: []string{"haste", "menace"},
		},
	}
}

func gapB09PushPair(g *game.Game, owner uuid.UUID) (rats, scav uuid.UUID) {
	rats = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Graf Rats", OracleID: gapB09GrafRatsOracle, TypeLine: "Creature — Rat",
		ManaCost: "{1}{B}", Power: 2, Toughness: 1, Owner: owner, Controller: owner,
		Layout: game.LayoutMeld, Meld: gapB09MeldPrint(), ScryfallID: uuid.NewString(),
	})
	scav = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Midnight Scavengers", OracleID: gapB09ScavOracle, TypeLine: "Creature — Human Rogue",
		ManaCost: "{4}{B}", Power: 3, Toughness: 3, Owner: owner, Controller: owner,
		Layout: game.LayoutMeld, Meld: gapB09MeldPrint(), ScryfallID: uuid.NewString(),
	})
	return rats, scav
}

func gapB09Host(g *game.Game) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == gapB09HostOracle {
			return c, true
		}
	}
	return game.Card{}, false
}

func TestGapB09GrafRatsMeldsWithScavengersAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	rats, scav := gapB09PushPair(g, me.ID)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	host, ok := gapB09Host(g)
	if !ok || !host.IsMelded() || len(host.MeldedFrom) != 2 {
		t.Fatalf("no melded Chittering Host: %+v", g.Battlefield.Cards)
	}
	for _, id := range []uuid.UUID{rats, scav} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%v is still on the battlefield beside the Host", id)
		}
	}
	effectstest.AssertKeywords(t, g, host.InstanceID, "haste", "menace")
	if mv, _ := host.ParsedManaValue(); mv != 7 {
		t.Errorf("mana value %d, want 7 ({1}{B} + {4}{B})", mv)
	}
	// Its enters trigger: other creatures get +1/+0 and menace.
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("Bear power %d, want 3", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("Bear did not gain menace")
	}
}

func TestGapB09GrafRatsAloneDoesNotMeld(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	rats, scav := gapB09PushPair(g, me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(scav) })
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := gapB09Host(g); ok {
		t.Error("Graf Rats melded with nothing")
	}
	if !g.Battlefield.Contains(rats) {
		t.Error("Graf Rats left the battlefield")
	}
}

func TestGapB09ScavengersHaveNoMeldTriggerOfTheirOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	_, scav := gapB09PushPair(g, me.ID)
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Graf Rats" {
				_ = g.DestroyPermanentForEffect(c.InstanceID)
			}
		}
	})
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scav) {
		t.Error("the Scavengers left the battlefield")
	}
}
