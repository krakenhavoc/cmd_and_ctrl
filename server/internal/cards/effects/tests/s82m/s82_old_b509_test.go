package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b509_test.go — Shards of Alara, "Card draw and library" (S82 OLD
// batch 509, #5588): each card driven through a real cast, activation or loss.

const (
	oldB509GrixisOracle    = "89838d28-6026-417d-9ec3-0081911291fb"
	oldB509AstrolabeOracle = "8d7194f7-1525-4f6e-b974-b947a254f437"
	oldB509BehemothOracle  = "967c509f-dbde-44ee-93fa-a01376f34b99"
	oldB509AngelOracle     = "9e4db334-4698-4986-820f-1daf1bffda83"
	oldB509DemonOracle     = "9f4e4d6e-bb1d-41d6-9184-3cac45197285"
	oldB509WeaverOracle    = "a585186c-d7c2-4f8b-acdc-c5f9ab950398"
	oldB509ThunderOracle   = "b0e1d26e-2fa1-4826-bfe9-01b0bd2c138b"
	oldB509SnipOracle      = "cfdae813-7241-4f6c-99f0-c772af34af91"
	oldB509MirrorOracle    = "e81da4a7-4b03-4107-b7d9-4b635faa615d"
)

func TestS82OldB509AllAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB509GrixisOracle, oldB509AstrolabeOracle, oldB509BehemothOracle, oldB509AngelOracle,
		oldB509DemonOracle, oldB509WeaverOracle, oldB509ThunderOracle, oldB509SnipOracle, oldB509MirrorOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB509Ready(g *game.Game, id uuid.UUID) {
	effectstest.UntapForTest(g, id)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

func TestS82OldB509GrixisBattlemageLootsAndStopsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mage := effectstest.SeedPermanentWithOracle(g, me.ID, "Grixis Battlemage", "Creature — Human Wizard", oldB509GrixisOracle)
	oldB509Ready(g, mage)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("loot: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d after draw then discard", me.Hand.Size(), hand)
	}
	// The second ability needs the mage untapped again.
	oldB509Ready(g, mage)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("can't block: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	card, _ := g.LookupCardForEffect(bear)
	if card.Effective().Restrictions&game.CantBlock == 0 {
		t.Error("the targeted creature can still block")
	}
}

func TestS82OldB509AstrolabeDrawsForAnArtifactAndIsRefusedWithoutOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	astro := effectstest.SeedPermanentWithOracle(g, me.ID, "Etherium Astrolabe", "Artifact", oldB509AstrolabeOracle)
	oldB509Ready(g, astro)
	spare := effectstest.SeedPermanent(g, me.ID, "Spare Rock", "Artifact")
	hand := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, astro, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{spare}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if g.Battlefield.Contains(spare) {
		t.Error("the sacrificed artifact stayed")
	}
	oldB509Ready(g, astro)
	effectstest.FloatMana(t, g, me, "{B}")
	land := effectstest.SeedPermanent(g, me.ID, "Plain Land", "Land")
	if err := g.ActivateCatalogAbility(me.ID, astro, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Error("a land was accepted as the artifact to sacrifice")
	}
}

// oldB509HeraldSetup seeds a Herald, the named target card in the library and
// the sacrificial creatures (name, colors).
func oldB509Herald(t *testing.T, g *game.Game, oracle, name, typeLine, fetched string) (herald, target uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	herald = effectstest.SeedPermanentWithOracle(g, me.ID, name, typeLine, oracle)
	oldB509Ready(g, herald)
	target = uuid.New()
	me.Library.PushTop(game.Card{InstanceID: target, Name: fetched, TypeLine: "Creature — Angel", Owner: me.ID, Controller: me.ID})
	return herald, target
}

func TestS82OldB509BehemothsHeraldFetchesGodsireForThreeColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herald, godsire := oldB509Herald(t, g, oldB509BehemothOracle, "Behemoth's Herald", "Creature — Elf Shaman", "Godsire")
	red := effectstest.SeedColoredCreature(g, me.ID, "Red One", []string{"R"}, 0)
	green := effectstest.SeedColoredCreature(g, me.ID, "Green One", []string{"G"}, 0)
	white := effectstest.SeedColoredCreature(g, me.ID, "White One", []string{"W"}, 0)
	effectstest.FloatMana(t, g, me, "{C}{C}{G}")
	// Two creatures are not enough.
	if err := g.ActivateCatalogAbility(me.ID, herald, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{red, green}}); err == nil {
		t.Fatal("two sacrifices were accepted for three kinds")
	}
	if err := g.ActivateCatalogAbility(me.ID, herald, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{red, green, white}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, godsire)
	if !g.Battlefield.Contains(godsire) {
		t.Error("Godsire is not on the battlefield")
	}
	for _, id := range []uuid.UUID{red, green, white} {
		if g.Battlefield.Contains(id) {
			t.Error("a sacrificed creature stayed")
		}
	}
}

func TestS82OldB509HeraldNeedsOneCreatureForEachColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herald, godsire := oldB509Herald(t, g, oldB509BehemothOracle, "Behemoth's Herald", "Creature — Elf Shaman", "Godsire")
	rg := effectstest.SeedColoredCreature(g, me.ID, "Red-Green One", []string{"R", "G"}, 0)
	white := effectstest.SeedColoredCreature(g, me.ID, "White One", []string{"W"}, 0)
	blue := effectstest.SeedColoredCreature(g, me.ID, "Blue One", []string{"U"}, 0)
	effectstest.FloatMana(t, g, me, "{C}{C}{G}")
	// A red-green creature can be the red or the green one, but not both,
	// and a blue creature is no kind at all.
	if err := g.ActivateCatalogAbility(me.ID, herald, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{rg, white, blue}}); err == nil {
		t.Fatal("a blue creature was accepted in place of the green one")
	}
	if g.Battlefield.Contains(godsire) || !g.Battlefield.Contains(rg) {
		t.Error("a refused activation changed the table")
	}
}

func TestS82OldB509AngelsAndDemonsHeraldsFetchTheirCard(t *testing.T) {
	for _, tc := range []struct {
		oracle, name, typeLine, fetched, pool string
		colors                                [3]string
	}{
		{oldB509AngelOracle, "Angel's Herald", "Creature — Human Cleric", "Empyrial Archangel", "{C}{C}{W}", [3]string{"G", "W", "U"}},
		{oldB509DemonOracle, "Demon's Herald", "Creature — Human Wizard", "Prince of Thralls", "{C}{C}{B}", [3]string{"U", "B", "R"}},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		herald, target := oldB509Herald(t, g, tc.oracle, tc.name, tc.typeLine, tc.fetched)
		var ids []uuid.UUID
		for i, c := range tc.colors {
			ids = append(ids, effectstest.SeedColoredCreature(g, me.ID, "Fodder "+string(rune('A'+i)), []string{c}, 0))
		}
		effectstest.FloatMana(t, g, me, tc.pool)
		if err := g.ActivateCatalogAbility(me.ID, herald, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: ids}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSearchByID(t, g, me.ID, target)
		if !g.Battlefield.Contains(target) {
			t.Errorf("%s: %s is not on the battlefield", tc.name, tc.fetched)
		}
	}
}

func TestS82OldB509JungleWeaverCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.CycleFromHand(t, g, "Jungle Weaver", "Creature — Spider", oldB509WeaverOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled Weaver is not in the graveyard")
	}
	spec, _ := effects.Lookup(oldB509WeaverOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "reach" {
		t.Errorf("keywords %v, want reach", spec.PrintedKeywords)
	}
}

func TestS82OldB509ResoundingThunderDealsThreeCastAndSixCycled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	effectstest.CastSpellWithCost(t, g, "Resounding Thunder", "Instant", oldB509ThunderOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("cast: life %d -> %d, want -3", life, opp.Life)
	}
	_ = me

	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[1]
	life2 := opp2.Life
	id, me2 := effectstest.CycleFromHand(t, g2, "Resounding Thunder", "Instant", oldB509ThunderOracle, "{C}{C}{C}{C}{C}{B}{R}{G}")
	if !me2.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	effectstest.Vow1Pick(t, g2, me2, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp2.ID}})
	effectstest.PassPriorityAroundTable(t, g2)
	if opp2.Life != life2-6 {
		t.Errorf("cycle: life %d -> %d, want -6", life2, opp2.Life)
	}
}

func TestS82OldB509SpellSnipCountersUnlessOnePaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		effectstest.CastFromSeat(t, g, opp, "Spell Snip", "Instant", oldB509SnipOracle, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		if pay {
			effectstest.B06AddMana(me, "C")
		}
		effectstest.PassUntilTaxed(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		_, resolved := effectstest.BattlefieldCard(g, victim)
		if pay && !resolved {
			t.Error("the spell was countered although its controller paid {1}")
		}
		if !pay && (resolved || !me.Graveyard.Contains(victim)) {
			t.Error("the spell was not countered")
		}
	}
}

func TestS82OldB509LichsMirrorReplacesALoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mirror := effectstest.SeedPermanentWithOracle(g, me.ID, "Lich's Mirror", "Artifact", oldB509MirrorOracle)
	own := effectstest.SeedPermanent(g, me.ID, "Own Rock", "Artifact")
	stolen := effectstest.SeedPermanent(g, opp.ID, "Their Rock", "Artifact")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == stolen {
				g.Battlefield.Cards[i].Owner = me.ID // I own it, they control it
			}
		}
	})
	theirs := effectstest.SeedPermanent(g, opp.ID, "Truly Theirs", "Artifact")
	dead := effectstest.PushGraveyardCardTyped(me, "Dead A", "Sorcery")
	g.WithWriteLock(func() { me.Life = -3 })
	g.RunStateChecksForTest()
	if me.Eliminated {
		t.Fatal("the player lost with Lich's Mirror on the battlefield")
	}
	if me.Life != 20 {
		t.Errorf("life %d, want 20", me.Life)
	}
	if me.Hand.Size() != 7 {
		t.Errorf("hand %d, want exactly the seven new cards", me.Hand.Size())
	}
	if me.Graveyard.Contains(dead) {
		t.Error("the graveyard card was not shuffled away")
	}
	for _, id := range []uuid.UUID{mirror, own, stolen} {
		if g.Battlefield.Contains(id) {
			t.Error("a permanent the player owns stayed on the battlefield")
		}
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("a permanent the player does not own was shuffled")
	}
	g.WithWriteLock(func() { me.Life = 0 })
	g.RunStateChecksForTest()
	if !me.Eliminated {
		t.Error("with the Mirror gone, the player did not lose")
	}
}
