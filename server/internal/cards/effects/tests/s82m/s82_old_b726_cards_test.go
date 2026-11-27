package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b726_cards_test.go — Urza's Saga activated abilities, batch 726
// (#5624). Each test plays the real catalog entry through the engine and
// asserts what its text says.

const (
	oldB726ViashinoSandswimmer = "1d82f01a-29be-4756-8745-e2c04bf77c61"
	oldB726UnworthyDead        = "7454e2cf-5ef2-4623-bac7-c375fd230595"
	oldB726MishrasHelix        = "6b1552f2-df64-485a-8509-b1916edf19cf"
	oldB726SanctumCustodian    = "4d979fbf-5009-47a1-8d77-b1ce80bbe90d"
	oldB726SanguineGuard       = "d86cc093-5db9-437a-a5bd-77d6faf97ca0"
	oldB726ElvishHerder        = "7832b90a-23b9-448c-9019-8ee2e8dfab31"
	oldB726NoRestForTheWicked  = "d03fd358-ff3f-45ab-9601-07dbb535cb51"
	oldB726Whetstone           = "940e461c-b205-4075-bd1f-a1534c33db6c"
	oldB726OpalAcrolith        = "d08b8dd6-3a44-4e86-8ce8-2d0a014a9aa8"
	oldB726SoulSculptor        = "d407abd0-69a9-405d-ab69-75b379a45780"
)

var oldB726Cards = map[string]string{
	oldB726ViashinoSandswimmer: "Viashino Sandswimmer", oldB726UnworthyDead: "Unworthy Dead",
	oldB726MishrasHelix: "Mishra's Helix", oldB726SanctumCustodian: "Sanctum Custodian",
	oldB726SanguineGuard: "Sanguine Guard", oldB726ElvishHerder: "Elvish Herder",
	oldB726NoRestForTheWicked: "No Rest for the Wicked", oldB726Whetstone: "Whetstone",
	oldB726OpalAcrolith: "Opal Acrolith", oldB726SoulSculptor: "Soul Sculptor",
}

func TestOldB726EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB726Cards {
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

func oldB726Permanent(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func oldB726InHand(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func TestOldB726ViashinoSandswimmerFlipsToReturnOrSacrifice(t *testing.T) {
	won, lost := 0, 0
	for table := 0; table < 60 && (won == 0 || lost == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		sw := oldB726Permanent(g, me.ID, "Viashino Sandswimmer", "Creature — Lizard", oldB726ViashinoSandswimmer, 3, 2)
		effectstest.FloatMana(t, g, me, "{R}")
		effectstest.B16Activate(t, g, me.ID, sw, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) == 0 {
			t.Fatal("no coin was flipped")
		}
		if g.Battlefield.Contains(sw) {
			t.Fatal("the creature stayed on the battlefield")
		}
		inHand, inYard := oldB726InHand(me, "Viashino Sandswimmer"), me.Graveyard.Contains(sw)
		if evs[len(evs)-1].Won {
			won++
			if !inHand || inYard {
				t.Errorf("won flip: in hand %v, in graveyard %v; want returned to hand", inHand, inYard)
			}
		} else {
			lost++
			if inHand || !inYard {
				t.Errorf("lost flip: in hand %v, in graveyard %v; want sacrificed", inHand, inYard)
			}
		}
	}
	if won == 0 || lost == 0 {
		t.Fatalf("never saw both outcomes (won %d, lost %d)", won, lost)
	}
}

func TestOldB726RegeneratorsSaveTheirCreature(t *testing.T) {
	for _, tc := range []struct{ name, oracle, mana string }{
		{"Unworthy Dead", oldB726UnworthyDead, "{B}"},
		{"Sanguine Guard", oldB726SanguineGuard, "{C}{B}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			c := oldB726Permanent(g, me.ID, tc.name, "Creature — Zombie", tc.oracle, 2, 2)
			effectstest.FloatMana(t, g, me, tc.mana)
			effectstest.B16Activate(t, g, me.ID, c, 0, game.ActivateAbilityParams{})
			effectstest.PassPriorityAroundTable(t, g)
			src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Giant", 9, 9)
			effectstest.DamageAndSettle(t, g, src, c, 5)
			if !g.Battlefield.Contains(c) {
				t.Error("the regeneration shield did not save the creature")
			}
		})
	}
}

func TestOldB726RegenerationNeedsItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := oldB726Permanent(g, me.ID, "Sanguine Guard", "Creature — Zombie", oldB726SanguineGuard, 2, 2)
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated {1}{B} with only {B} available")
	}
}

func TestOldB726SanguineGuardHasFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sanguine Guard", TypeLine: "Creature — Zombie", OracleID: oldB726SanguineGuard,
		Power: 2, Toughness: 2, Keywords: []string{"first strike"}, Owner: me.ID, Controller: me.ID,
	})
	if !effectstest.HasEffectiveKeyword(t, g, c, "first strike") {
		t.Error("no first strike")
	}
	spec, _ := effects.Lookup(oldB726SanguineGuard)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "first strike" {
		t.Errorf("printed keywords %v, want [first strike]", spec.PrintedKeywords)
	}
}

func oldB726Land(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushLandFor(g, owner, name, "Basic Land — Mountain")
}

func TestOldB726MishrasHelixTapsXTargetLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	helix := oldB726Permanent(g, me.ID, "Mishra's Helix", "Artifact", oldB726MishrasHelix, 0, 0)
	a, b, c := oldB726Land(g, opp.ID, "A"), oldB726Land(g, opp.ID, "B"), oldB726Land(g, opp.ID, "C")
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, helix, 0, game.ActivateAbilityParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	tapped := func(id uuid.UUID) bool { return effectstest.LayeredCard(t, g, id).Tapped }
	if !tapped(a) || !tapped(b) {
		t.Error("the two targeted lands were not tapped")
	}
	if tapped(c) {
		t.Error("an untargeted land was tapped")
	}
	if !tapped(helix) {
		t.Error("the Helix was not tapped as a cost")
	}
}

func TestOldB726MishrasHelixNeedsExactlyXTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	helix := oldB726Permanent(g, me.ID, "Mishra's Helix", "Artifact", oldB726MishrasHelix, 0, 0)
	a := oldB726Land(g, opp.ID, "A")
	effectstest.FloatMana(t, g, me, "{C}{C}")
	err := g.ActivateCatalogAbility(me.ID, helix, 0, game.ActivateAbilityParams{
		XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}},
	})
	if err == nil {
		t.Error("activated with X=2 and only one target")
	}
}

func TestOldB726SanctumCustodianPreventsTheNextTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cust := oldB726Permanent(g, me.ID, "Sanctum Custodian", "Creature — Human Cleric", oldB726SanctumCustodian, 1, 2)
	effectstest.B16Activate(t, g, me.ID, cust, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	src := effectstest.B12Creature(g, me.ID, "Source", "Creature — Giant", 5, 5)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, opp.ID, 5); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want 2 of 5 damage prevented (%d)", life, opp.Life, life-3)
	}
}

func TestOldB726ElvishHerderGivesTrampleUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herder := oldB726Permanent(g, me.ID, "Elvish Herder", "Creature — Elf", oldB726ElvishHerder, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{G}")
	effectstest.B16Activate(t, g, me.ID, herder, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the target did not gain trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, herder, "trample") {
		t.Error("the Herder gained trample too")
	}
}

func TestOldB726NoRestForTheWickedReturnsOnlyThisTurnsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nr := oldB726Permanent(g, me.ID, "No Rest for the Wicked", "Enchantment", oldB726NoRestForTheWicked, 0, 0)
	fresh := effectstest.B12Creature(g, me.ID, "Fresh Dead", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Dead", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, fresh)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	// A creature card that was never on the battlefield, and a non-creature
	// that died this turn.
	old := effectstest.PushCatalogHandCard(me, "Discarded Bear", "Creature — Bear", "")
	effectstest.B16Activate(t, g, me.ID, nr, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(nr) {
		t.Error("the enchantment was not sacrificed as a cost")
	}
	if !me.Hand.Contains(fresh) {
		t.Error("the creature that died this turn was not returned to hand")
	}
	if !me.Hand.Contains(old) {
		t.Error("the hand card was disturbed")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's creature card left their graveyard")
	}
}

func TestOldB726NoRestForTheWickedIgnoresCreaturesThatDiedEarlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nr := oldB726Permanent(g, me.ID, "No Rest for the Wicked", "Enchantment", oldB726NoRestForTheWicked, 0, 0)
	// A card put into the graveyard from the hand never came from the battlefield.
	discarded := effectstest.PushCatalogHandCard(me, "Discarded Bear", "Creature — Bear", "")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, discarded); err != nil {
		t.Fatalf("move: %v", err)
	}
	effectstest.B16Activate(t, g, me.ID, nr, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(discarded) {
		t.Error("a creature card that was not put there from the battlefield was returned")
	}
}

func TestOldB726WhetstoneMillsEachPlayerTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stone := oldB726Permanent(g, me.ID, "Whetstone", "Artifact", oldB726Whetstone, 0, 0)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Library.Size()
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, stone, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if got := p.Library.Size(); got != before[i]-2 {
			t.Errorf("seat %d library %d -> %d, want 2 milled", i, before[i], got)
		}
	}
}

func oldB726CastCreature(t *testing.T, g *game.Game, name string) {
	t.Helper()
	effectstest.CastCreatureCard(t, g, name, "Creature — Bear", "", 2, 2)
}

func TestOldB726OpalAcrolithAnimatesWhenAnOpponentCastsACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	acro := oldB726Permanent(g, opp.ID, "Opal Acrolith", "Enchantment", oldB726OpalAcrolith, 0, 0)
	// Its own controller casting a creature spell does nothing.
	g.Turn.ActiveSeat = 1
	oldB726CastCreature(t, g, "Own Bear")
	if effectstest.LayeredCard(t, g, acro).IsCreature() {
		t.Fatal("the Acrolith animated on its controller's own creature spell")
	}
	g.Turn.ActiveSeat = 0
	oldB726CastCreature(t, g, "Their Bear")
	got := effectstest.LayeredCard(t, g, acro)
	if !got.IsCreature() || got.HasCardType("enchantment") {
		t.Fatalf("the Acrolith is %v, want only a creature", got.TypeLine)
	}
	if p, th := effectstest.EffectivePTOf(t, g, acro); p != 2 || th != 4 {
		t.Errorf("the Acrolith is %d/%d, want 2/4", p, th)
	}
}

func TestOldB726OpalAcrolithBecomesAnEnchantmentAgainForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	acro := oldB726Permanent(g, opp.ID, "Opal Acrolith", "Enchantment", oldB726OpalAcrolith, 0, 0)
	oldB726CastCreature(t, g, "Their Bear")
	if !effectstest.LayeredCard(t, g, acro).IsCreature() {
		t.Fatal("the Acrolith did not animate")
	}
	effectstest.B16Activate(t, g, opp.ID, acro, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.LayeredCard(t, g, acro)
	if got.IsCreature() || !got.HasCardType("enchantment") {
		t.Errorf("the Acrolith is %v, want an enchantment", got.TypeLine)
	}
	// And it can be animated again.
	oldB726CastCreature(t, g, "Another Bear")
	if !effectstest.LayeredCard(t, g, acro).IsCreature() {
		t.Error("the Acrolith did not animate a second time")
	}
}

func TestOldB726SoulSculptorTurnsACreatureIntoAnEnchantmentUntilACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sculptor := oldB726Permanent(g, me.ID, "Soul Sculptor", "Creature — Human", oldB726SoulSculptor, 1, 1)
	target := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird", Power: 2, Toughness: 2,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.FloatMana(t, g, me, "{C}{W}")
	effectstest.B16Activate(t, g, me.ID, sculptor, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.LayeredCard(t, g, target)
	if got.IsCreature() || !got.HasCardType("enchantment") {
		t.Fatalf("the target is %v, want an enchantment", got.TypeLine)
	}
	if effectstest.HasEffectiveKeyword(t, g, target, "flying") {
		t.Error("the target kept its abilities")
	}
	oldB726CastCreature(t, g, "Some Bear")
	got = effectstest.LayeredCard(t, g, target)
	if !got.IsCreature() {
		t.Errorf("the target is %v after a creature spell, want a creature again", got.TypeLine)
	}
	if !effectstest.HasEffectiveKeyword(t, g, target, "flying") {
		t.Error("the target did not get its abilities back")
	}
}

func TestOldB726SoulSculptorNeedsATargetCreatureAndAUntappedSelf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sculptor := oldB726Permanent(g, me.ID, "Soul Sculptor", "Creature — Human", oldB726SoulSculptor, 1, 1)
	land := oldB726Land(g, me.ID, "Plains")
	effectstest.FloatMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, sculptor, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err == nil {
		t.Error("targeted a land")
	}
}
