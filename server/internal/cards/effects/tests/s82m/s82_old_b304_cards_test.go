package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b304_cards_test.go — Commander 2015, batch 304 (S82 older sets,
// #5546). Every test plays the real catalog entry through the engine.

var oldB304Oracles = map[string]string{
	"Ohran Viper":               "879fdcd8-62ea-4d39-abd4-44e6d82adf4d",
	"Firemind's Foresight":      "bf5765e0-def0-4ddf-9cd7-083feae71660",
	"Viridian Emissary":         "c93bce5a-48cc-4dc5-930b-1e084c2bae7d",
	"Mirror Match":              "1f0883f8-89c9-4620-9837-72a2909e0599",
	"Necromancer's Covenant":    "9103f01f-e60b-4cda-9102-edc4aef74cc3",
	"Vow of Malice":             "894bc28a-6627-43dd-9650-86300b6fe5e6",
	"Curse of the Nightly Hunt": "cc6f2b79-8fcd-4542-904b-943c7493b8ad",
	"Dawnglare Invoker":         "0d603aef-2b92-43fb-81ba-f89953a198df",
	"Stinkdrinker Daredevil":    "25af5e13-ebe8-4bc4-b8fe-8b9aee59579c",
	"Thief of Blood":            "97d61346-bd53-4eb8-a920-6ae0382eb20d",
	"Noble Quarry":              "d549abd0-fea5-417f-93f9-255cb87ec989",
}

func TestOldB304EveryCardIsRegisteredFull(t *testing.T) {
	for name, oracle := range oldB304Oracles {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not in the catalog", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %d, want Full with none", name, spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// oldB304Cast puts the card in the active seat's hand and casts it strictly,
// free apart from what the params add.
func oldB304Cast(t *testing.T, g *game.Game, name, typeLine string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	effectstest.D5681Spell(g, me, id, name, typeLine, "", oldB304Oracles[name])
	params.Strict = true
	return id, g.CastSpell(me.ID, id, params)
}

func TestOldB304OhranViperDestroysACreatureItDamagedAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	viper := effectstest.B12Push(g, me.ID, "Ohran Viper", "Snow Creature — Snake", oldB304Oracles["Ohran Viper"], 0, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 0, 20)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, viper)
	effectstest.EmitDamageEventForTest(g, viper, wall, 1, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wall) {
		t.Fatal("the wall was destroyed before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the creature the Viper damaged survived end of combat")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a creature the Viper never damaged was destroyed")
	}
}

func TestOldB304OhranViperIgnoresNoncombatDamageToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	viper := effectstest.B12Push(g, me.ID, "Ohran Viper", "Snow Creature — Snake", oldB304Oracles["Ohran Viper"], 0, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 0, 20)
	effectstest.DeclareAttack(t, g, opp.ID, viper)
	effectstest.EmitDamageEventForTest(g, viper, wall, 1, false)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wall) {
		t.Error("noncombat damage marked the creature for destruction")
	}
}

func TestOldB304OhranViperMayDrawOnCombatDamageToAPlayer(t *testing.T) {
	for _, draw := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		viper := effectstest.B12Push(g, me.ID, "Ohran Viper", "Snow Creature — Snake", oldB304Oracles["Ohran Viper"], 1, 3)
		effectstest.SeedLibrary(me, "Draw Me", "Next")
		hand := me.Hand.Size()
		effectstest.DealCombatDamageToPlayer(g, viper, opp.ID, 1)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, draw)
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if draw {
			want++
		}
		if me.Hand.Size() != want {
			t.Errorf("draw=%v: hand %d → %d, want %d", draw, hand, me.Hand.Size(), want)
		}
	}
}

func oldB304LibraryInstant(p *game.Player, name string, mv int) uuid.UUID {
	cost := map[int]string{3: "{2}{R}", 2: "{1}{U}", 1: "{R}", 4: "{3}{U}"}[mv]
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Instant"
	c.ManaCost = cost
	p.Library.PushTop(c)
	return c.InstanceID
}

func TestOldB304FiremindsForesightFindsAThreeThenATwoThenAOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Filler 1", "Filler 2", "Filler 3")
	three := oldB304LibraryInstant(me, "Three", 3)
	two := oldB304LibraryInstant(me, "Two", 2)
	one := oldB304LibraryInstant(me, "One", 1)
	four := oldB304LibraryInstant(me, "Four", 4)
	extraOne := oldB304LibraryInstant(me, "Another One", 1)
	if _, err := oldB304Cast(t, g, "Firemind's Foresight", "Instant", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, three)
	effectstest.AnswerSearchByID(t, g, me.ID, two)
	effectstest.AnswerSearchByID(t, g, me.ID, one)
	for _, id := range []uuid.UUID{three, two, one} {
		if !me.Hand.Contains(id) {
			t.Errorf("a found instant is not in hand")
		}
	}
	for _, id := range []uuid.UUID{four, extraOne} {
		if me.Hand.Contains(id) {
			t.Errorf("an instant of the wrong mana value, or a second one, went to hand")
		}
	}
}

func TestOldB304FiremindsForesightMayFindNothingAtEachStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Filler 1", "Filler 2")
	hand := me.Hand.Size()
	if _, err := oldB304Cast(t, g, "Firemind's Foresight", "Instant", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		if effectstest.SearchChoiceFor(g, me.ID) == nil {
			break
		}
		effectstest.AnswerSearchFailToFind(t, g, me.ID)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want no change", hand, me.Hand.Size())
	}
}

func TestOldB304ViridianEmissarySearchesForATappedBasicWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := game.NewCard("Forest", me.ID)
	forest.TypeLine = "Basic Land — Forest"
	me.Library.PushTop(forest)
	rock := game.NewCard("Not A Land", me.ID)
	rock.TypeLine = "Artifact"
	me.Library.PushTop(rock)
	emissary := effectstest.B12Push(g, me.ID, "Viridian Emissary", "Creature — Phyrexian Elf Scout", oldB304Oracles["Viridian Emissary"], 2, 1)
	effectstest.KillCreature(t, g, me.ID, emissary)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest.InstanceID) {
		t.Fatal("the basic land is not on the battlefield")
	}
	if !effectstest.B16Tapped(t, g, forest.InstanceID) {
		t.Error("the basic land entered untapped")
	}
}

func TestOldB304MirrorMatchBlocksEveryAttackerWithACopyThatLeavesAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	atk, me, third := g.Seats[0], g.Seats[1], g.Seats[2]
	a := effectstest.B12Creature(g, atk.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, atk.ID, "Attacker B", "Creature — Elf", 3, 3)
	c := effectstest.B12Creature(g, atk.ID, "Attacker C", "Creature — Wolf", 4, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for id, defender := range map[uuid.UUID]uuid.UUID{a: me.ID, b: me.ID, c: third.ID} {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.CastInPlace(t, g, me.ID, "Mirror Match", oldB304Oracles["Mirror Match"])
	effectstest.PassPriorityAroundTable(t, g)
	tokens := map[string]uuid.UUID{}
	for _, card := range g.BattlefieldCardsForEffect() {
		if card.IsToken() && card.Controller == me.ID {
			tokens[card.Name] = card.InstanceID
		}
	}
	if len(tokens) != 2 || tokens["Attacker A"] == uuid.Nil || tokens["Attacker B"] == uuid.Nil {
		t.Fatalf("want one copy of each creature attacking me, got %v", tokens)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	for name, tid := range tokens {
		if g.Battlefield.Contains(tid) {
			t.Errorf("the %s copy is still on the battlefield after end of combat", name)
		}
	}
}

func TestOldB304MirrorMatchCanOnlyBeCastDuringDeclareBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	if _, err := oldB304Cast(t, g, "Mirror Match", "Instant", game.CastSpellParams{}); err == nil {
		t.Fatal("cast in the main phase")
	}
}

func TestOldB304NecromancersCovenantExilesTheCreatureCardsAndMakesAZombieEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear1 := effectstest.PushGraveyardPermanent(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	bear2 := effectstest.PushGraveyardPermanent(opp, "Dead Elf", "Creature — Elf", "{G}")
	rock := effectstest.PushGraveyardPermanent(opp, "Dead Rock", "Artifact", "{1}")
	mine := effectstest.PushGraveyardPermanent(me, "My Dead Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastHoldCreature(t, g, "Necromancer's Covenant", "Enchantment", oldB304Oracles["Necromancer's Covenant"], 0, 0)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear1, bear2} {
		if effectstest.ZoneOf(g, id) != game.ZoneExile {
			t.Errorf("a creature card was not exiled")
		}
	}
	if !opp.Graveyard.Contains(rock) || !me.Graveyard.Contains(mine) {
		t.Error("a noncreature card or another player's graveyard was touched")
	}
	zombies := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == "Zombie" && c.Controller == me.ID {
			zombies++
			if c.Power != 2 || c.Toughness != 2 || !c.HasColor("B") {
				t.Errorf("Zombie is %d/%d", c.Power, c.Toughness)
			}
			if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "lifelink") {
				t.Error("a Zombie I control has no lifelink")
			}
		}
	}
	if zombies != 2 {
		t.Errorf("%d Zombies, want 2", zombies)
	}
}

func TestOldB304VowOfMaliceGivesPlusTwoIntimidateAndBarsAttackingYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if _, err := oldB304Cast(t, g, "Vow of Malice", "Enchantment — Aura", game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	live := effectstest.JmpB4Live(t, g, bear)
	if live.CurrentPower() != 4 || live.CurrentToughness() != 4 {
		t.Errorf("enchanted creature is %d/%d, want 4/4", live.CurrentPower(), live.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "intimidate") {
		t.Error("no intimidate")
	}
	g.Turn.ActiveSeat = 1
	g.Turn.Step = game.StepBeginCombat
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, me.ID); err == nil {
		t.Error("the enchanted creature attacked its enchanter")
	}
}

func TestOldB304CurseOfTheNightlyHuntForcesTheCursedPlayersCreaturesToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	curse := uuid.New()
	effectstest.D5681Spell(g, me, curse, "Curse of the Nightly Hunt", "Enchantment — Aura Curse", "", oldB304Oracles["Curse of the Nightly Hunt"])
	if err := g.CastSpell(me.ID, curse, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// My own turn: the Curse says nothing about my creatures, so I may stay home.
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("my creature was required to attack: %v", err)
	}
	// The enchanted player's turn: their creature must attack.
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) {
		t.Fatalf("pass with the cursed player's creature home = %v, want an attack requirement refusal", err)
	}
	if err := g.DeclareAttacker(theirs, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("pass after attacking: %v", err)
	}
}

func TestOldB304DawnglareInvokerTapsEveryCreatureOfTheTargetPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	invoker := effectstest.B12Push(g, me.ID, "Dawnglare Invoker", "Creature — Kor Wizard", oldB304Oracles["Dawnglare Invoker"], 2, 1)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, invoker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, a) || !effectstest.B16Tapped(t, g, b) {
		t.Error("an opposing creature is untapped")
	}
	if effectstest.B16Tapped(t, g, mine) {
		t.Error("my creature was tapped")
	}
}

func TestOldB304StinkdrinkerDaredevilDiscountsOnlyMyGiantSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Stinkdrinker Daredevil", "Creature — Goblin Rogue", oldB304Oracles["Stinkdrinker Daredevil"], false)
	if got := effectstest.PriceInHand(t, g, me, "Hill Giant", "Creature — Giant", "{3}{R}"); got != 2 {
		t.Errorf("a {3}{R} Giant costs %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Red Giant", "Creature — Giant", "{R}{R}"); got != 2 {
		t.Errorf("a {R}{R} Giant costs %d, want 2 (coloured pips untouched)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Plain Bear", "Creature — Bear", "{3}{R}"); got != 4 {
		t.Errorf("a non-Giant costs %d, want 4", got)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Their Giant", "Creature — Giant", "{3}{R}"); got != 4 {
		t.Errorf("an opponent's Giant costs %d, want 4", got)
	}
}

func TestOldB304ThiefOfBloodRemovesEveryCounterAndEntersWithThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	walker := effectstest.PushWalkerForCounterCost(g, opp.ID, 3)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 2) })
	effectstest.CastHoldCreature(t, g, "Thief of Blood", "Creature — Vampire", oldB304Oracles["Thief of Blood"], 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOneCounters(g, bear) != 0 {
		t.Error("the Bear kept its counters")
	}
	if effectstest.LoyaltyCount(g, walker) != 0 {
		t.Errorf("the planeswalker kept %d loyalty counters", effectstest.LoyaltyCount(g, walker))
	}
	var thief uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Thief of Blood" {
			thief = c.InstanceID
		}
	}
	if thief == uuid.Nil {
		t.Fatal("the Thief is not on the battlefield")
	}
	if got := effectstest.PlusOneCounters(g, thief); got != 5 {
		t.Errorf("the Thief entered with %d +1/+1 counters, want 5 (2 + 3)", got)
	}
}

func TestOldB304ThiefOfBloodWithNoCountersEntersAsAOneOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastHoldCreature(t, g, "Thief of Blood", "Creature — Vampire", oldB304Oracles["Thief of Blood"], 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Thief of Blood" {
			if effectstest.PlusOneCounters(g, c.InstanceID) != 0 {
				t.Error("entered with counters though none were removed")
			}
			return
		}
	}
	t.Fatal("the Thief is not on the battlefield")
}

func TestOldB304NobleQuarryLuresOnItselfAsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	quarry := effectstest.ReqCreature(g, me.ID, "Noble Quarry", oldB304Oracles["Noble Quarry"], 1, 1, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Blocker A", "", 3, 4)
	b := effectstest.ReqCreature(g, opp.ID, "Blocker B", "", 3, 4)
	effectstest.ReqToBlockers(t, g, quarry)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Blocker A must block Noble Quarry if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := effectstest.ReqBlock(t, g, quarry, a, b); err != nil {
		t.Fatal(err)
	}
}

func TestOldB304NobleQuarryBestowedPumpsAndLuresOnItsHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	effectstest.AdvanceToMain(t, g)
	host := effectstest.ReqCreature(g, me.ID, "Host", "", 2, 2, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Blocker A", "", 3, 4)
	effectstest.FloatMana(t, g, me, "{G}{G}{G}{G}{G}{G}")
	id := uuid.New()
	effectstest.D5681Spell(g, me, id, "Noble Quarry", "Enchantment Creature — Unicorn", "", oldB304Oracles["Noble Quarry"])
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "bestow", Targets: effectstest.CardTarget(host)}); err != nil {
		t.Fatalf("bestow cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	live := effectstest.JmpB4Live(t, g, host)
	if live.CurrentPower() != 3 || live.CurrentToughness() != 3 {
		t.Errorf("the host is %d/%d, want 3/3", live.CurrentPower(), live.CurrentToughness())
	}
	effectstest.ReqToBlockers(t, g, host)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Blocker A must block Host if able (Noble Quarry)." {
		t.Errorf("sentence = %q", got)
	}
	_ = a
}
