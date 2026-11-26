package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b342_test.go — Commander 2014 cards, S82 OLD batch 342 (#5553),
// played through the real catalog entries.

const (
	oldB342Predator   = "647c1339-9508-4bb9-8e77-ebc3dfdf4069"
	oldB342Whirlwind  = "d3946df7-cb24-47c0-a8b0-1bdf95d99926"
	oldB342Dregs      = "8126e60b-cf3d-4bf0-bfb5-f1e914d3e20e"
	oldB342Armistice  = "da103316-2a85-4de9-8531-ac2cd2859d6f"
	oldB342Mesa       = "8776d14f-a5a2-4ef3-98d0-e1148e579e4d"
	oldB342Hexmage    = "fcb2a95b-cef0-4b29-8b2b-538f380aff2b"
	oldB342Warcaller  = "8959a147-75b1-4f4a-a699-cb354ff55e59"
	oldB342Titania    = "a9e72fe7-c4b3-421a-8b2f-bb9448b04a44"
	oldB342Resonance  = "94846856-fe68-4ada-8c34-a54de568e803"
	oldB342Feud       = "e4add901-3f2b-4f92-bb26-7f599e288802"
	oldB342Drove      = "3ba81fbe-2c53-495b-9a2b-f124475bd71b"
	oldB342Reflection = "5bbcfb73-cccc-47b0-81ae-17eb7c33ff5e"
)

func TestOldB342EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB342Predator: "Predator, Flagship", oldB342Whirlwind: "Whirlwind", oldB342Dregs: "Dregs of Sorrow",
		oldB342Armistice: "Armistice", oldB342Mesa: "Sacred Mesa", oldB342Hexmage: "Vampire Hexmage",
		oldB342Warcaller: "Joraga Warcaller", oldB342Titania: "Titania's Chosen", oldB342Resonance: "Impact Resonance",
		oldB342Feud: "Bitter Feud", oldB342Drove: "Drove of Elves", oldB342Reflection: "Infinite Reflection",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestOldB342PredatorFlagshipGrantsFlyingAndDestroysFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ship := effectstest.MirB1Push(g, me.ID, "Predator, Flagship", "Legendary Artifact", oldB342Predator, 0, 0)
	ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 2)
	flier := effectstest.MirB1Push(g, opp.ID, "Flier", "Creature — Bird", "", 2, 2, "flying")
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{C}{C}")
	if err := effectstest.MirB1Activate(g, me.ID, ship, 1, effectstest.CardTarget(ground)...); err == nil {
		t.Fatal("a creature without flying was accepted by the destroy ability")
	}
	if err := effectstest.MirB1Activate(g, me.ID, ship, 0, effectstest.CardTarget(ground)...); err != nil {
		t.Fatalf("grant flying: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, ground, "flying") {
		t.Fatal("the target did not gain flying")
	}
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{C}{C}")
	if err := effectstest.MirB1Activate(g, me.ID, ship, 1, effectstest.CardTarget(flier)...); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	if !effectstest.TappedOnBattlefield(t, g, ship) {
		t.Error("the {T} cost was not paid")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) {
		t.Error("the flier survived")
	}
}

func TestOldB342WhirlwindDestroysEveryFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.MirB1Push(g, me.ID, "My Bird", "Creature — Bird", "", 1, 1, "flying")
	theirs := effectstest.MirB1Push(g, opp.ID, "Their Bird", "Creature — Bird", "", 1, 1, "flying")
	ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 2)
	effectstest.P1g2Cast(t, g, "Whirlwind", "Sorcery", oldB342Whirlwind, "{2}{G}{G}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a flier survived")
	}
	if !g.Battlefield.Contains(ground) {
		t.Error("a ground creature died")
	}
}

func TestOldB342DregsOfSorrowDestroysXNonblackCreaturesAndDrawsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Zombie", TypeLine: "Creature — Zombie",
		Power: 2, Toughness: 2, Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{C}{C}{B}{C}{C}")
	if err := g.CastSpell(me.ID, effectstest.HandCardFull(me, "Dregs of Sorrow", "Sorcery", "{X}{4}{B}", oldB342Dregs, []string{"B"}),
		game.CastSpellParams{Strict: true, XValue: 1, Targets: effectstest.CardTargets(black)}); err == nil {
		t.Fatal("a black creature was accepted")
	}
	handBefore := me.Hand.Size()
	effectstest.P1g2Cast(t, g, "Dregs of Sorrow", "Sorcery", oldB342Dregs, "{X}{4}{B}",
		game.CastSpellParams{XValue: 2, Targets: effectstest.CardTargets(a, b)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || !g.Battlefield.Contains(black) {
		t.Error("the wrong creatures were destroyed")
	}
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand grew by %d, want 2 (two drawn, the spell cast left)", got)
	}
}

func TestOldB342ArmisticeDrawsAndGivesAnOpponentLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	armistice := effectstest.MirB1Push(g, me.ID, "Armistice", "Enchantment", oldB342Armistice, 0, 0)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{W}{W}")
	if err := effectstest.MirB1Activate(g, me.ID, armistice, 0, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("the controller was accepted as the target opponent")
	}
	hand, life := me.Hand.Size(), opp.Life
	if err := effectstest.MirB1Activate(g, me.ID, armistice, 0, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+1)
	}
	if opp.Life != life+3 {
		t.Errorf("opponent life = %d, want %d", opp.Life, life+3)
	}
}

func TestOldB342SacredMesaMakesPegasusesAndTaxesYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mesa := effectstest.MirB1Push(g, me.ID, "Sacred Mesa", "Enchantment", oldB342Mesa, 0, 0)
	effectstest.DcMana(t, g, me.ID, "{C}{W}")
	if err := effectstest.MirB1Activate(g, me.ID, mesa, 0); err != nil {
		t.Fatalf("make a Pegasus: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pegasi := oldB342Named(g, "Pegasus", me.ID)
	if len(pegasi) != 1 || pegasi[0].Power != 1 || pegasi[0].Toughness != 1 || !effectstest.HasKeywordOnBattlefield(t, g, pegasi[0].InstanceID, "flying") {
		t.Fatalf("Pegasus tokens = %+v, want one 1/1 flier", pegasi)
	}
	// With a Pegasus the controller chooses; giving it up keeps the Mesa.
	effectstest.BeginUpkeep(t, g, me.ID)
	c := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatal("no choice between the Pegasus and the Mesa")
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, pegasi[0].InstanceID)
	if !g.Battlefield.Contains(mesa) || g.Battlefield.Contains(pegasi[0].InstanceID) {
		t.Error("the Pegasus should have been sacrificed and the Mesa kept")
	}
}

func TestOldB342SacredMesaIsSacrificedWithNoPegasus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mesa := effectstest.MirB1Push(g, me.ID, "Sacred Mesa", "Enchantment", oldB342Mesa, 0, 0)
	effectstest.BeginUpkeep(t, g, me.ID)
	if g.Battlefield.Contains(mesa) {
		t.Error("the Mesa survived an upkeep with no Pegasus")
	}
}

func TestOldB342VampireHexmageRemovesEveryCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hex := effectstest.MirB1Push(g, me.ID, "Vampire Hexmage", "Creature — Vampire Shaman", oldB342Hexmage, 2, 1, "first strike")
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 4, game.CounterPlusOne: 2}, Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.MirB1Activate(g, me.ID, hex, 0, effectstest.CardTarget(walker)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(hex) {
		t.Error("the Hexmage was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if l, p := effectstest.B12Counter(t, g, walker, game.CounterLoyalty), effectstest.B12Counter(t, g, walker, game.CounterPlusOne); l != 0 || p != 0 {
		t.Errorf("counters left: loyalty %d, +1/+1 %d", l, p)
	}
}

func oldB342Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
	p.Hand.PushTop(c)
	if err := g.CastSpell(p.ID, c.InstanceID, params); err != nil {
		t.Fatalf("CastSpell %s: %v", c.Name, err)
	}
	return c.InstanceID
}

func TestOldB342JoragaWarcallerPumpsOtherElvesPerCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirElf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	id := oldB342Cast(t, g, me, game.Card{Name: "Joraga Warcaller", TypeLine: "Creature — Elf Warrior", OracleID: oldB342Warcaller,
		ManaCost: "{G}", Power: 1, Toughness: 1, Colors: []string{"G"}}, game.CastSpellParams{OptionalCosts: []int{0, 0}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, id, game.CounterPlusOne); got != 2 {
		t.Fatalf("counters = %d, want 2", got)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, id); got != 3 {
		t.Errorf("Warcaller power = %d, want 3 (it does not pump itself beyond its counters)", got)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, elf); got != 3 {
		t.Errorf("my Elf power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("my Bear power = %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, theirElf); got != 1 {
		t.Errorf("their Elf power = %d, want 1", got)
	}
}

func TestOldB342TitaniasChosenGrowsOnAnyGreenSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	titania := effectstest.MirB1Push(g, me.ID, "Titania's Chosen", "Creature — Elf Archer", oldB342Titania, 1, 1)
	oldB342Cast(t, g, me, game.Card{Name: "Red Spell", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, titania, game.CounterPlusOne); got != 0 {
		t.Fatalf("a red spell gave %d counters", got)
	}
	oldB342Cast(t, g, me, game.Card{Name: "Green Spell", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, titania, game.CounterPlusOne); got != 1 {
		t.Errorf("after my green spell: %d counters, want 1", got)
	}
	// An opponent's green spell counts too (it is cast at instant speed here).
	oldB342Cast(t, g, opp, game.Card{Name: "Green Trick", TypeLine: "Instant", ManaCost: "{G}", Colors: []string{"G"}}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, titania, game.CounterPlusOne); got != 2 {
		t.Errorf("after an opponent's green spell: %d counters, want 2", got)
	}
}

func TestOldB342DroveOfElvesCountsYourGreenPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	green := func(owner uuid.UUID, name, typeLine string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
			Colors: []string{"G"}, Owner: owner, Controller: owner})
	}
	drove := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Drove of Elves", TypeLine: "Creature — Elf",
		OracleID: oldB342Drove, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID, Keywords: []string{"hexproof"}})
	green(me.ID, "Mine A", "Enchantment")
	green(me.ID, "Mine B", "Creature — Bear")
	green(opp.ID, "Theirs", "Creature — Bear")
	effectstest.B12Permanent(g, me.ID, "Colorless", "Artifact")
	if p, tough := effectstest.EffectivePower(t, g, drove), effectstest.EffectiveToughness(t, g, drove); p != 3 || tough != 3 {
		t.Errorf("Drove of Elves = %d/%d, want 3/3", p, tough)
	}
}

func TestOldB342InfiniteReflectionCopiesYourOtherNontokenCreaturesAndFutureOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, me.ID, "Original", "Creature — Wizard", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.AuraCast(t, g, "Infinite Reflection", oldB342Reflection, host)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Card(t, g, bear).Name; got != "Original" {
		t.Errorf("my Bear is %q, want a copy of Original", got)
	}
	if got := effectstest.B12Card(t, g, token).Name; got != "Goblin" {
		t.Errorf("my token is %q, tokens must not be copied", got)
	}
	newcomer := effectstest.B12PlayFromHand(t, g, "Newcomer", "Creature — Elf", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(newcomer) {
		t.Fatal("the new creature did not enter")
	}
	if got := effectstest.B12Card(t, g, newcomer).Name; got != "Original" {
		t.Errorf("a nontoken creature entering is %q, want a copy of the enchanted creature", got)
	}
}

func TestOldB342BitterFeudDoublesDamageBetweenTheChosenPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Bitter Feud", "Enchantment", oldB342Feud, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChosenPlayer(t, g, me.ID, a.ID)
	effectstest.AnswerChosenPlayer(t, g, me.ID, b.ID)
	srcA := effectstest.B12Creature(g, a.ID, "A's Source", "Creature — Bear", 2, 2)
	srcMe := effectstest.B12Creature(g, me.ID, "My Source", "Creature — Bear", 2, 2)
	hurt := func(source, victim uuid.UUID, n int) int {
		before := victim2life(g, victim)
		oldB342DealTo(t, g, source, victim, n)
		return before - victim2life(g, victim)
	}
	if got := hurt(srcA, b.ID, 3); got != 6 {
		t.Errorf("a chosen player's source hit the other chosen player for %d, want 6", got)
	}
	if got := hurt(srcA, c.ID, 3); got != 3 {
		t.Errorf("damage to an unchosen player was %d, want 3", got)
	}
	if got := hurt(srcMe, a.ID, 3); got != 3 {
		t.Errorf("an unchosen player's source dealt %d to a chosen player, want 3", got)
	}
	bBear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 20, Owner: b.ID, Controller: b.ID})
	effectstest.DealDamageTo(t, g, srcA, bBear, 3)
	if got := effectstest.B12Card(t, g, bBear).DamageMarked; got != 6 {
		t.Errorf("damage to the other chosen player's creature = %d, want 6", got)
	}
}

func victim2life(g *game.Game, id uuid.UUID) int {
	for _, p := range g.Seats {
		if p.ID == id {
			return p.Life
		}
	}
	return 0
}

func TestOldB342ImpactResonanceDividesTheGreatestDamageOfTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	src := effectstest.B12Creature(g, me.ID, "Source", "Creature — Bear", 2, 2)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Wall", 0, 4)
	// Nothing has been dealt yet: X is 0, so no creature can be targeted.
	id := effectstest.HandCardFull(me, "Impact Resonance", "Instant", "{1}{R}", oldB342Resonance, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(small), Distribution: map[uuid.UUID]int{small: 1}}); err == nil {
		t.Fatal("cast with X = 0")
	}
	oldB342DealTo(t, g, src, opp.ID, 2)
	oldB342DealTo(t, g, src, g.Seats[2].ID, 3)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(small, big), Distribution: map[uuid.UUID]int{small: 1, big: 1}}); err == nil {
		t.Fatal("a division of 2 was accepted for X = 3")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(small, big), Distribution: map[uuid.UUID]int{small: 1, big: 2}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the Small creature took 1 and survived")
	}
	if got := effectstest.B12Card(t, g, big).DamageMarked; got != 2 {
		t.Errorf("Big has %d damage, want 2", got)
	}
}

func oldB342Named(g *game.Game, name string, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}

func oldB342DealTo(t *testing.T, g *game.Game, source, victim uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, victim, n); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
}
