package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b477_cards_test.go — Archenemy, batch 477 (#5580, S82 OLD). Every
// test plays the real catalog entry through the engine.

const (
	oldB477Metallurgeon = "09373283-a5c1-4d8e-ad25-19cd55b1e869"
	oldB477Shinen       = "7b9b3347-a374-40fd-a12a-73187225194e"
	oldB477Mistvein     = "1499af11-274f-4f8d-b0b6-20e1693ebab7"
	oldB477Fieldmist    = "a292ca5b-2472-4b02-a26b-d1b5b717a34b"
	oldB477Sabotage     = "25fe48be-95c4-4011-8c60-4628fb5ecfcb"
	oldB477Sanctum      = "643ddf81-bbd3-4398-b5a7-1169c6e2b081"
	oldB477Abunas       = "10064324-34a0-47eb-a58e-01db10234ed9"
	oldB477March        = "51092634-308e-4779-aa51-182715dbc734"
	oldB477Architects   = "79d290ea-c095-4135-9711-f3ce089dae43"
	oldB477Dragon       = "a4fa3052-d1ec-4442-a832-15ba36c500ee"
)

func TestOldB477CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB477Metallurgeon: "Metallurgeon", oldB477Shinen: "Shinen of Life's Roar", oldB477Mistvein: "Mistvein Borderpost",
		oldB477Fieldmist: "Fieldmist Borderpost", oldB477Sabotage: "Gleeful Sabotage", oldB477Sanctum: "Synod Sanctum",
		oldB477Abunas: "Leonin Abunas", oldB477March: "March of the Machines", oldB477Architects: "Architects of Will",
		oldB477Dragon: "Two-Headed Dragon",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func TestOldB477BorderpostIsCastByPayingOneAndBouncingABasicLand(t *testing.T) {
	for _, tc := range []struct{ oracle, name, cost string }{
		{oldB477Mistvein, "Mistvein Borderpost", "{1}{U}{B}"},
		{oldB477Fieldmist, "Fieldmist Borderpost", "{1}{W}{U}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.ToMainForCost(t, g)
			island := effectstest.FreeSpellPermanent(g, me.ID, "Island", "Basic Land — Island")
			post := effectstest.HandCardFull(me, tc.name, "Artifact", tc.cost, tc.oracle, nil)
			effectstest.FloatMana(t, g, me, "{C}")
			if err := g.CastSpell(me.ID, post, game.CastSpellParams{Strict: true, AlternativeCost: "return", AltCostIDs: []uuid.UUID{island}}); err != nil {
				t.Fatalf("cast for {1} and a basic land: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(post) {
				t.Fatal("the Borderpost did not resolve")
			}
			if !effectstest.IsTapped(g, post) {
				t.Error("the Borderpost entered untapped")
			}
			if g.Battlefield.Contains(island) || !me.Hand.Contains(island) {
				t.Error("the basic land was not returned to its owner's hand")
			}
		})
	}
}

func TestOldB477BorderpostRefusesANonbasicLandAndAFreeCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	shock := effectstest.FreeSpellPermanent(g, me.ID, "Steam Vents", "Land — Island Mountain")
	post := effectstest.HandCardFull(me, "Mistvein Borderpost", "Artifact", "{1}{U}{B}", oldB477Mistvein, nil)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, post, game.CastSpellParams{Strict: true, AlternativeCost: "return", AltCostIDs: []uuid.UUID{shock}}); err == nil {
		t.Error("a nonbasic land paid the alternative cost")
	}
	island := effectstest.FreeSpellPermanent(g, me.ID, "Island", "Basic Land — Island")
	g.Seats[0].ManaPool = game.ManaPool{}
	if err := g.CastSpell(me.ID, post, game.CastSpellParams{Strict: true, AlternativeCost: "return", AltCostIDs: []uuid.UUID{island}}); err == nil {
		t.Error("the {1} was not charged: the alternative cost cast for free")
	}
}

func TestOldB477BorderpostTapsForEitherColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	post := effectstest.PushCatalogPermanent(g, me.ID, "Mistvein Borderpost", "Artifact", oldB477Mistvein, false)
	for idx, color := range []string{"U", "B"} {
		effectstest.UntapForTest(g, post)
		if err := g.ActivateManaAbility(me.ID, post, idx, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("tap for {%s}: %v", color, err)
		}
	}
}

func TestOldB477MetallurgeonRegeneratesAnArtifactAndNothingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	smith := effectstest.PushCatalogPermanent(g, me.ID, "Metallurgeon", "Artifact Creature — Human Artificer", oldB477Metallurgeon, false)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, smith, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("Metallurgeon targeted a nonartifact creature")
	}
	effectstest.B16Activate(t, g, me.ID, smith, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(relic)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Gleeful Sabotage", "Sorcery", oldB477Sabotage, effectstest.CardTarget(relic))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(relic) {
		t.Error("the regenerated artifact was destroyed")
	}
	if !effectstest.IsTapped(g, relic) {
		t.Error("the regenerated artifact is untapped")
	}
}

func TestOldB477GleefulSabotageDestroysAnArtifactOrEnchantmentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Gleeful Sabotage", "Sorcery", oldB477Sabotage, effectstest.CardTarget(bear)); err == nil {
		t.Error("Gleeful Sabotage targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Gleeful Sabotage", "Sorcery", oldB477Sabotage, effectstest.CardTarget(relic))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact survived")
	}
}

func TestOldB477GleefulSabotageConspireCopiesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Permanent(g, opp.ID, "Relic A", "Artifact")
	b := effectstest.B12Permanent(g, opp.ID, "Relic B", "Artifact")
	mk := func(name string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Elf",
			ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	}
	tapA, tapB := mk("Tapper A"), mk("Tapper B")
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Gleeful Sabotage", TypeLine: "Sorcery", OracleID: oldB477Sabotage,
		ManaCost: "{1}{G}", Colors: []string{"G"}},
		game.CastSpellParams{
			Targets:       effectstest.CardTarget(a),
			OptionalCosts: []int{0},
			TeamworkIDs:   []uuid.UUID{tapA, tapB},
		})
	if !effectstest.IsTapped(g, tapA) || !effectstest.IsTapped(g, tapB) {
		t.Fatal("conspire did not tap the two creatures")
	}
	for i := 0; i < 64; i++ {
		if effectstest.StackFullyEmpty(g) && effectstest.LatestPickTarget(g, me.ID) == nil {
			break
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: b}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Errorf("the original or the copy (retargeted) did not destroy its artifact: A %v, B %v", g.Battlefield.Contains(a), g.Battlefield.Contains(b))
	}
}

func TestOldB477LeoninAbunasGivesYourArtifactsHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	if effectstest.HasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Fatal("an artifact had hexproof before the Abunas")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Leonin Abunas", "Creature — Cat Cleric", oldB477Abunas, false)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("your artifact lacks hexproof")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "hexproof") {
		t.Error("an opponent's artifact has hexproof")
	}
}

func TestOldB477MarchOfTheMachinesMakesNoncreatureArtifactsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mineID := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Three Relic", TypeLine: "Artifact", ManaCost: "{3}"})
	theirID := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Five Relic", TypeLine: "Artifact", ManaCost: "{2}{U}{U}{B}"})
	golemID := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{6}", Power: 1, Toughness: 1})
	enchantID := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Aura", TypeLine: "Enchantment", ManaCost: "{2}"})
	effectstest.PushCatalogPermanent(g, me.ID, "March of the Machines", "Enchantment", oldB477March, false)
	for id, want := range map[uuid.UUID]int{mineID: 3, theirID: 5} {
		c := effectstest.GaLayered(t, g, id)
		if !c.IsCreature() {
			t.Errorf("%s is not a creature", c.Name)
		}
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != want || tough != want {
			t.Errorf("%s is %d/%d, want %d/%d", c.Name, p, tough, want, want)
		}
	}
	if p, tough := effectstest.EffectivePTOf(t, g, golemID); p != 1 || tough != 1 {
		t.Errorf("an artifact creature is %d/%d, want its own 1/1", p, tough)
	}
	if effectstest.GaLayered(t, g, enchantID).IsCreature() {
		t.Error("a nonartifact enchantment became a creature")
	}
}

func TestOldB477ArchitectsOfWillLooksAtTheTargetPlayersTopThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := effectstest.SeedLibrary(opp, "A", "B", "C", "D")
	effectstest.ToMainForCost(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{U}{B}")
	id := effectstest.HandCardFull(me, "Architects of Will", "Artifact Creature — Human Wizard", "{2}{U}{B}", oldB477Architects, []string{"U", "B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if ask == nil || len(ask.ScryCards) != 3 || ask.LibraryPlacement != game.LibraryPlaceTop {
		t.Fatalf("order prompt %+v, want the opponent's top three to put back on top", ask)
	}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, nil, []uuid.UUID{lib[2], lib[0], lib[1]}); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.LibraryTopNames(opp, 1); len(got) != 1 || got[0] != "C" {
		t.Errorf("the opponent's top card %v, want C", got)
	}
}

func TestOldB477ArchitectsOfWillCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B")
	id := effectstest.PushCatalogHandCard(me, "Architects of Will", "Artifact Creature — Human Wizard", oldB477Architects)
	before := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{U}")
	spec, _ := effects.Lookup(oldB477Architects)
	idx := -1
	for i, ab := range spec.Activated {
		if ab.Cycling {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("no cycling ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle for {U}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(id) || me.Hand.Size() != before {
		t.Errorf("hand %d after cycling (card still there: %v), want %d with the card gone", me.Hand.Size(), me.Hand.Contains(id), before)
	}
}

func TestOldB477TwoHeadedDragonFliesMenacesBlocksTwoAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.PushMonster(g, opp.ID, "Two-Headed Dragon", oldB477Dragon, 4, 4)
	for _, kw := range []string{"flying", "menace"} {
		if !effectstest.HasEffectiveKeyword(t, g, dragon, kw) {
			t.Errorf("no %s", kw)
		}
	}
	if got := effectstest.MbCapacity(t, g, dragon); got != 2 {
		t.Fatalf("block capacity = %d, want 2", got)
	}
	atks := effectstest.MbAttack(t, g, me, opp, 3)
	if err := effectstest.MbBlock(g, dragon, atks[:2]...); err != nil {
		t.Fatalf("blocking two: %v", err)
	}
	if err := effectstest.MbBlock(g, dragon, atks[2]); err == nil {
		t.Error("a third block was accepted")
	}
}

func TestOldB477TwoHeadedDragonPumpsForOneRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dragon := effectstest.PushMonster(g, me.ID, "Two-Headed Dragon", oldB477Dragon, 4, 4)
	effectstest.FloatMana(t, g, me, "{C}{R}{C}{R}")
	for i := 0; i < 2; i++ {
		effectstest.B16Activate(t, g, me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, dragon); p != 8 || tough != 4 {
		t.Errorf("twice pumped is %d/%d, want 8/4", p, tough)
	}
}

func TestOldB477SynodSanctumExilesThenReturnsEverythingItExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sanctum := effectstest.PushCatalogPermanent(g, me.ID, "Synod Sanctum", "Artifact", oldB477Sanctum, false)
	other := effectstest.B12Permanent(g, me.ID, "Other Relic", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, victim := range []uuid.UUID{bear, other} {
		effectstest.FloatMana(t, g, me, "{C}{C}")
		effectstest.B16Activate(t, g, me.ID, sanctum, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim)})
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(victim) {
			t.Fatalf("%s was not exiled", victim)
		}
		effectstest.UntapForTest(g, sanctum)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, sanctum, 1, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sanctum) {
		t.Error("the Sanctum was not sacrificed")
	}
	// A card that changes zones is a new object (CR 400.7): look for them by name.
	if len(effectstest.BattlefieldIDsNamed(g, "Bear")) != 1 || len(effectstest.BattlefieldIDsNamed(g, "Other Relic")) != 1 {
		t.Error("the exiled permanents did not all return to the battlefield")
	}
	if got := len(g.Exile.Cards); got != 0 {
		t.Errorf("%d cards still in exile", got)
	}
}

func TestOldB477ShinenOfLifesRoarLuresAndChannels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shinen := effectstest.PushCatalogPermanent(g, me.ID, "Shinen of Life's Roar", "Creature — Spirit", oldB477Shinen, false)
	a := effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := effectstest.ReqCreature(g, opp.ID, "Wall B", "", 0, 4)
	effectstest.ReqToBlockers(t, g, shinen)
	if err := effectstest.ReqBlock(t, g, shinen, a); err != nil {
		t.Fatal(err)
	}
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, shinen, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both blocking: %v", err)
	}
}

func TestOldB477ShinenOfLifesRoarChannelLuresATargetCreatureThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hand := effectstest.PushCatalogHandCard(me, "Shinen of Life's Roar", "Creature — Spirit", oldB477Shinen)
	runner := effectstest.ReqCreature(g, me.ID, "Runner", "", 2, 2, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := effectstest.ReqCreature(g, opp.ID, "Wall B", "", 0, 4)
	effectstest.FloatMana(t, g, me, "{C}{C}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, hand, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(runner)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(hand) {
		t.Error("the channelled card is still in hand")
	}
	effectstest.ReqToBlockers(t, g, runner)
	if err := effectstest.ReqBlock(t, g, runner, a); err != nil {
		t.Fatal(err)
	}
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, runner, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both blocking: %v", err)
	}
}
