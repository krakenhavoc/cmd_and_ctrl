package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/h"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	midB16Kyler       = "726cd041-5d0b-436c-bced-9335f56c0b0d"
	midB16Heronblade  = "e3e255e0-ab04-4488-bdc1-bb1542f51e16"
	midB16Katilda     = "6c728edd-ab67-44ce-b19d-aa98dfce68bb"
	midB16Beastmaster = "df35bf19-4cbf-4624-925d-ceca6bff596e"
	midB16Eloise      = "65e65eab-bc3f-4eae-8361-7f0a65fc1673"
	midB16Geistcatch  = "175db9f1-0188-4640-9b54-6b461c5440aa"
	midB16Rescuer     = "f687a746-4807-41d5-848d-4053ddef6a76"
	midB16Savior      = "14bc9e20-a6ad-4eca-8c5c-5459ce078d9a"
	midB16Rotbelly    = "8f0738ce-1f07-422c-9895-8ac3c65f9054"
)

func TestMidB16CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		midB16Kyler: "Kyler, Sigardian Emissary", midB16Heronblade: "Heronblade Elite",
		midB16Katilda: "Katilda, Dawnhart Prime", midB16Beastmaster: "Somberwald Beastmaster",
		midB16Eloise: "Eloise, Nephalia Sleuth", midB16Geistcatch: "Prowling Geistcatcher",
		midB16Rescuer: "Moorland Rescuer", midB16Savior: "Sigardian Savior",
		midB16Rotbelly: "Ravenous Rotbelly",
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

func midB16Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.AvB1Card(t, g, id).Counters[game.CounterPlusOne]
}

func TestMidB16KylerGrowsOnAnotherHumanAndPumpsOtherHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kyler := effectstest.B12Push(g, me.ID, "Kyler, Sigardian Emissary", "Legendary Creature — Human Cleric", midB16Kyler, 2, 2)
	human := effectstest.B12Creature(g, me.ID, "Plain Human", "Creature — Human Soldier", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Their Human", "Creature — Human", 1, 1)
	if p := effectstest.AvB1Card(t, g, human).PowerForComparison(); p != 1 {
		t.Fatalf("no counters yet, Human power %d, want 1", p)
	}
	effectstest.AvB1OpponentCreatureEnters(t, g, opp.ID, "Creature — Human")
	effectstest.AvB1ResolveCreature(t, g, me, "Bear Two", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB16Counters(t, g, kyler); n != 0 {
		t.Fatalf("a non-Human or an opponent's Human added a counter: %d", n)
	}
	effectstest.AvB1ResolveCreature(t, g, me, "New Human", "Creature — Human Knight", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB16Counters(t, g, kyler); n != 1 {
		t.Fatalf("Kyler counters %d, want 1", n)
	}
	for _, tc := range []struct {
		id           uuid.UUID
		name         string
		wantP, wantT int
	}{{kyler, "Kyler", 3, 3}, {human, "my other Human", 2, 2}, {bear, "my Bear", 2, 2}, {foe, "their Human", 1, 1}} {
		_ = tc.wantT
		c := effectstest.AvB1Card(t, g, tc.id)
		if p := c.PowerForComparison(); p != tc.wantP {
			t.Errorf("%s power = %d, want %d", tc.name, p, tc.wantP)
		}
	}
	// Any counter on Kyler counts, not only +1/+1.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(kyler, "charge", 2) })
	if p := effectstest.AvB1Card(t, g, human).PowerForComparison(); p != 4 {
		t.Errorf("with a +1/+1 and two other counters on Kyler, a Human has power %d, want 4", p)
	}
}

func TestMidB16HeronbladeEliteGrowsAndTapsForItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elite := effectstest.B12Push(g, me.ID, "Heronblade Elite", "Creature — Human Warrior", midB16Heronblade, 1, 1)
	if !effectstest.HasKeywordOf(g, elite, "vigilance") {
		t.Error("no vigilance")
	}
	effectstest.AvB1ResolveCreature(t, g, me, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB16Counters(t, g, elite); n != 0 {
		t.Fatalf("a Bear added a counter: %d", n)
	}
	effectstest.AvB1ResolveCreature(t, g, me, "Human", "Creature — Human", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AvB1ResolveCreature(t, g, me, "Human Two", "Creature — Human", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB16Counters(t, g, elite); n != 2 {
		t.Fatalf("counters %d, want 2", n)
	}
	var produced game.ManaAmount
	g.WithWriteLock(func() { produced = effects.ProducedOneColor(cardsh.HeronbladeEliteX)(g, me.ID, elite) })
	if produced.N != 3 || !produced.OneColor {
		t.Errorf("produces %+v, want three mana of one colour", produced)
	}
	if err := g.ActivateManaAbility(me.ID, elite, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID)
	if pick == nil {
		t.Fatal("no colour prompt")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	total := 0
	for _, m := range me.ManaPool {
		total += m.N()
	}
	if total != 3 {
		t.Errorf("pool holds %d mana, want 3", total)
	}
	for _, m := range me.ManaPool {
		if m.Color != "U" {
			t.Errorf("a token is %q, want all {U}", m.Color)
		}
	}
}

func TestMidB16KatildaGivesHumansAColourReadingManaAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	katilda := effectstest.B12Push(g, me.ID, "Katilda, Dawnhart Prime", "Legendary Creature — Human Warlock", midB16Katilda, 1, 1)
	g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Colors = []string{"G", "W"}
	human := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Human", TypeLine: "Creature — Human", Colors: []string{"R"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	colorless := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Colorless Human", TypeLine: "Creature — Human", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	g.ReadSnapshot(func() {
		for _, tc := range []struct {
			id   uuid.UUID
			name string
			want int
		}{{human, "a Human", 1}, {katilda, "Katilda", 1}, {bear, "a Bear", 0}, {colorless, "a colorless Human", 1}} {
			c, _ := g.LookupCardForEffect(tc.id)
			if got := len(game.ManaAbilitiesOf(&c)); got != tc.want {
				t.Errorf("%s has %d mana abilities, want %d", tc.name, got, tc.want)
			}
		}
	})
	var abs []game.ManaAbilityShape
	var origins game.AbilityOrigins
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(human)
		abs, origins = game.ManaAbilitiesWithOrigins(c)
	})
	if len(abs) != 1 {
		t.Fatalf("Human mana abilities %+v", abs)
	}
	if err := g.ActivateManaAbility(me.ID, human, 0, game.ManaAbilityParams{Ref: origins.Ref(0)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "R" {
		t.Errorf("pool %+v, want one {R}", me.ManaPool)
	}
}

func TestMidB16KatildaPutsACounterOnEachCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	katilda := effectstest.B12Push(g, me.ID, "Katilda, Dawnhart Prime", "Legendary Creature — Human Warlock", midB16Katilda, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	if err := effectstest.BroB01Activate(t, g, me, katilda, 0, "{C}{C}{C}{C}{G}"); err == nil {
		t.Fatal("activated with five mana")
	}
	if err := effectstest.BroB01Activate(t, g, me, katilda, 0, "{G}{W}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if midB16Counters(t, g, katilda) != 1 || midB16Counters(t, g, bear) != 1 || midB16Counters(t, g, foe) != 0 {
		t.Errorf("counters katilda %d, bear %d, foe %d; want 1, 1, 0",
			midB16Counters(t, g, katilda), midB16Counters(t, g, bear), midB16Counters(t, g, foe))
	}
}

func TestMidB16BeastmasterMakesThreeTokensWithDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	foeToken := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Token", TypeLine: "Token Creature — Bear", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	master := effectstest.AvB1ResolveCreature(t, g, me, "Somberwald Beastmaster", "Creature — Human Ranger", midB16Beastmaster)
	effectstest.PassPriorityAroundTable(t, g)
	got := map[string]int{}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() {
			got[c.Name+" "+itoaPT(effectivePT(t, g, c.InstanceID))]++
			if !effectstest.HasKeywordOf(g, c.InstanceID, "deathtouch") {
				t.Errorf("token %s has no deathtouch", c.Name)
			}
		}
	}
	if got["Wolf 2/2"] != 1 || got["Beast 3/3"] != 1 || got["Beast 4/4"] != 1 {
		t.Errorf("tokens = %v, want a 2/2 Wolf, a 3/3 Beast and a 4/4 Beast", got)
	}
	if effectstest.HasKeywordOf(g, bear, "deathtouch") || effectstest.HasKeywordOf(g, master, "deathtouch") || effectstest.HasKeywordOf(g, foeToken, "deathtouch") {
		t.Error("a nontoken or an opponent's token has deathtouch")
	}
}

func itoaPT(p, tough int) string {
	return string(rune('0'+p)) + "/" + string(rune('0'+tough))
}

func effectivePT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePTOf(t, g, id)
}

func midB16Clues(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.HasSubtype("Clue") {
			n++
		}
	}
	return n
}

func TestMidB16EloiseInvestigatesWhenAnotherCreatureOfYoursDiesAndSurveilsOnATokenSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	eloise := effectstest.B12Push(g, me.ID, "Eloise, Nephalia Sleuth", "Legendary Creature — Human Rogue", midB16Eloise, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(foe) })
	effectstest.PassPriorityAroundTable(t, g)
	if midB16Clues(g, me.ID) != 0 {
		t.Fatal("an opponent's creature dying investigated")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if midB16Clues(g, me.ID) != 1 {
		t.Fatalf("clues %d, want 1", midB16Clues(g, me.ID))
	}
	// Sacrificing the Clue is sacrificing a token: surveil 1.
	var clue uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Clue") {
			clue = c.InstanceID
		}
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(clue) })
	effectstest.PassPriorityAroundTable(t, g)
	if sc := effectstest.SurveilChoiceFor(g, me.ID); sc == nil {
		t.Error("sacrificing a token did not surveil")
	}
	// A nontoken sacrifice does not.
	g.PendingChoices = nil
	nt := effectstest.B12Creature(g, me.ID, "Goat", "Creature — Goat", 1, 1)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(nt) })
	effectstest.PassPriorityAroundTable(t, g)
	if sc := effectstest.SurveilChoiceFor(g, me.ID); sc != nil {
		t.Error("a nontoken sacrifice surveilled")
	}
	// Eloise dying herself does not investigate.
	before := midB16Clues(g, me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(eloise) })
	effectstest.PassPriorityAroundTable(t, g)
	if midB16Clues(g, me.ID) != before {
		t.Error("Eloise's own death investigated")
	}
}

func TestMidB16GeistcatcherExilesSacrificedCreaturesAndGivesThemBackWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	catcher := effectstest.B12Push(g, me.ID, "Prowling Geistcatcher", "Creature — Human Rogue", midB16Geistcatch, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	land := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(land) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(land) || !me.Graveyard.Contains(land) {
		t.Fatal("a sacrificed noncreature was exiled")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Fatal("the sacrificed Bear was not exiled")
	}
	if n := midB16Counters(t, g, catcher); n != 0 {
		t.Fatalf("a nontoken added a counter: %d", n)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(token) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := midB16Counters(t, g, catcher); n != 1 {
		t.Fatalf("counters %d after a token was sacrificed, want 1", n)
	}
	// A creature that dies another way is not exiled.
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(other) {
		t.Error("a destroyed creature was exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(catcher) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(bear) {
		t.Fatal("the exiled Bear did not return when the Geistcatcher left")
	}
	if effectstest.AvB1Named(g, me.ID, "Bear") == uuid.Nil {
		t.Error("the Bear is not on the battlefield under my control")
	}
}

func midB16Grave(me *game.Player, name string, power int) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: power, Toughness: power, Owner: me.ID, Controller: me.ID})
	return id
}

func TestMidB16MoorlandRescuerReturnsCardsUpToItsPowerThenExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rescuer := effectstest.B12Push(g, me.ID, "Moorland Rescuer", "Creature — Human Knight", midB16Rescuer, 4, 4)
	a, b := midB16Grave(me, "Two A", 2), midB16Grave(me, "Two B", 2)
	c, big := midB16Grave(me, "Three", 3), midB16Grave(me, "Five", 5)
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rescuer) })
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.PendingChooseCards(t, g, me.ID)
	for _, id := range pick.ChooseCards {
		if id == big || id == spell || id == rescuer {
			t.Errorf("%s was offered", id)
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, c}); err == nil {
		t.Fatal("total power 5 was accepted for X = 4")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("total power 4 refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) || g.Battlefield.Contains(c) {
		t.Error("the two 2-power cards should be on the battlefield and the 3-power card should not")
	}
	if !g.Exile.Contains(rescuer) || me.Graveyard.Contains(rescuer) {
		t.Error("the Rescuer's card was not exiled")
	}
}

func TestMidB16MoorlandRescuerExilesItselfEvenWhenNothingReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rescuer := effectstest.B12Push(g, me.ID, "Moorland Rescuer", "Creature — Human Knight", midB16Rescuer, 4, 4)
	midB16Grave(me, "Five", 5)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rescuer) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(rescuer) {
		t.Error("the Rescuer's card was not exiled")
	}
	// Picking none still exiles it.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	r2 := effectstest.B12Push(g2, me2.ID, "Moorland Rescuer", "Creature — Human Knight", midB16Rescuer, 4, 4)
	z := midB16Grave(me2, "One", 1)
	g2.WithWriteLock(func() { _ = g2.DestroyPermanentForEffect(r2) })
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerChooseCards(t, g2, me2.ID)
	effectstest.PassPriorityAroundTable(t, g2)
	if !g2.Exile.Contains(r2) || !me2.Graveyard.Contains(z) {
		t.Error("picking none should exile the Rescuer and leave the graveyard alone")
	}
}

func TestMidB16SigardianSaviorReturnsTwoSmallCreaturesOnlyWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.GraveyardCardOf(g, me, "One Drop", "Creature — Elf", "{G}", "")
	two := effectstest.GraveyardCardOf(g, me, "Two Drop", "Creature — Elf", "{1}{G}", "")
	extra := effectstest.GraveyardCardOf(g, me, "Other Two", "Creature — Elf", "{1}{G}", "")
	three := effectstest.GraveyardCardOf(g, me, "Three Drop", "Creature — Elf", "{2}{G}", "")
	// Not cast: it just arrives.
	effectstest.B12Push(g, me.ID, "Sigardian Savior", "Creature — Angel", midB16Savior, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || g.Battlefield.Contains(one) {
		t.Fatal("a Savior that was not cast returned cards")
	}
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Sigardian Savior", TypeLine: "Creature — Angel", OracleID: midB16Savior, Power: 3, Toughness: 3}, game.CastSpellParams{})
	for i := 0; i < 6 && effectstest.LatestPickTarget(g, me.ID) == nil && g.Stack.Size() > 0; i++ {
		_ = g.PassPriority()
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the cast Savior")
	}
	if effectstest.HasID(p.PickTargetCards, three) {
		t.Error("a 3-mana-value creature was offered")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(one, two, extra)); err == nil {
		t.Error("three targets were accepted")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(one, two)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(one) || !g.Battlefield.Contains(two) || g.Battlefield.Contains(extra) {
		t.Error("the two chosen cards should be on the battlefield and no more")
	}
}

func TestMidB16RotbellySacrificesZombiesAndEachOpponentSacrificesThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	z1 := effectstest.B12Creature(g, me.ID, "Zombie One", "Creature — Zombie", 1, 1)
	z2 := effectstest.B12Creature(g, me.ID, "Zombie Two", "Creature — Zombie", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 1, 1)
	o1 := effectstest.B12Creature(g, opp.ID, "Their One", "Creature — Bear", 1, 1)
	o2 := effectstest.B12Creature(g, opp.ID, "Their Two", "Creature — Bear", 1, 1)
	o3 := effectstest.B12Creature(g, opp.ID, "Their Three", "Creature — Bear", 1, 1)
	effectstest.AvB1ResolveCreature(t, g, me, "Ravenous Rotbelly", "Creature — Zombie Horror", midB16Rotbelly)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no Zombie prompt")
	}
	if effectstest.HasID(pick.ChooseCards, bear) || !effectstest.HasID(pick.ChooseCards, z1) || pick.ChooseMin != 0 || pick.ChooseMax != 3 {
		t.Fatalf("candidates %v bounds %d..%d: only Zombies, up to three", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{z1, z2}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	for i := 0; i < 12 && effectstest.SacrificeChoiceFor(g, opp.ID) == nil; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(z1) || g.Battlefield.Contains(z2) {
		t.Fatal("the Zombies were not sacrificed")
	}
	if c := effectstest.SacrificeChoiceFor(g, opp.ID); c != nil && c.Kind == game.PendingChoiceChooseCards {
		effectstest.AnswerSacrifice(t, g, opp.ID, o1, o2)
	} else {
		effectstest.AnswerSacrifice(t, g, opp.ID, o1)
		effectstest.AnswerSacrifice(t, g, opp.ID, o2)
	}
	effectstest.PassPriorityAroundTable(t, g)
	left := 0
	for _, id := range []uuid.UUID{o1, o2, o3} {
		if g.Battlefield.Contains(id) {
			left++
		}
	}
	if left != 1 {
		t.Errorf("the opponent has %d of three creatures left, want 1", left)
	}
}

func TestMidB16RotbellySacrificingNothingMakesNoOpponentSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	o1 := effectstest.B12Creature(g, opp.ID, "Their One", "Creature — Bear", 1, 1)
	effectstest.AvB1ResolveCreature(t, g, me, "Ravenous Rotbelly", "Creature — Zombie Horror", midB16Rotbelly)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no Zombie prompt")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(o1) || effectstest.SacrificeChoiceFor(g, opp.ID) != nil {
		t.Error("an opponent sacrificed although no Zombie was")
	}
}
