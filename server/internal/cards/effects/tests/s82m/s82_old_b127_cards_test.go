package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b127_cards_test.go — Theros Beyond Death cards, S82 OLD batch 127
// (#5514), played through the real catalog entries.

const (
	oldB127PriestOracle   = "0a8c1002-e237-42d8-b152-aefecab283d0"
	oldB127RiderOracle    = "580112bd-6804-4981-97ad-0f5e43627b22"
	oldB127SwimmerOracle  = "64ca54ee-1502-4fae-b919-7e4457e641c0"
	oldB127EnemyOracle    = "783f059f-90d0-4148-898b-9e694b22ac0a"
	oldB127WarleaderOracl = "883a2b57-1075-4479-920b-3f7d5da1f444"
	oldB127EidolonOracle  = "ad70144e-16e4-4604-9e6d-cb287f67ddb5"
	oldB127BlessingOracle = "d553ae4d-b2a9-47eb-8c2e-ffab8cc31b26"
	oldB127HopliteOracle  = "e99c58bd-b405-4474-bd32-1a993c9f2979"
	oldB127HeraldOracle   = "ed347609-15f9-4d57-9523-63a33a9dccbc"
	oldB127DevourerOracle = "f747b39a-28a0-46a8-a161-182939556d39"
)

func TestOldB127CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB127PriestOracle:   "Slaughter-Priest of Mogis",
		oldB127RiderOracle:    "Stampede Rider",
		oldB127SwimmerOracle:  "Swimmer in Nightmares",
		oldB127EnemyOracle:    "Enemy of Enlightenment",
		oldB127WarleaderOracl: "Skophos Warleader",
		oldB127EidolonOracle:  "Eidolon of Inspiration",
		oldB127BlessingOracle: "Karametra's Blessing",
		oldB127HopliteOracle:  "Sunlit Hoplite",
		oldB127HeraldOracle:   "Nyx Herald",
		oldB127DevourerOracle: "Devourer of Memory",
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

func oldB127Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// --- Slaughter-Priest of Mogis ------------------------------------------------

func TestOldB127SlaughterPriestGainsFirstStrikeAndGrowsWhenYouSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB127Seats(g)
	priest := effectstest.B12Push(g, me.ID, "Slaughter-Priest of Mogis", "Creature — Minotaur Shaman", oldB127PriestOracle, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	shrine := effectstest.PushCatalogPermanent(g, me.ID, "Shrine", "Enchantment", "", false)
	relic := effectstest.PushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false)
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, priest, "first strike") {
		t.Fatal("the Priest starts with first strike")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{relic}, Strict: true}); err == nil {
		t.Fatal("an artifact paid the Priest's sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{priest}, Strict: true}); err == nil {
		t.Fatal("the Priest sacrificed itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{shrine}, Strict: true}); err != nil {
		t.Fatalf("activate sacrificing an enchantment: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, priest, "first strike") {
		t.Error("the Priest did not gain first strike")
	}
	if p, _ := effectstest.PtOf(t, g, priest); p != 4 {
		t.Errorf("the Priest has power %d after a sacrifice, want 4", p)
	}
	_ = fodder
}

// --- Skophos Warleader -----------------------------------------------------------

func TestOldB127SkophosWarleaderPumpsAndGainsMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB127Seats(g)
	lead := effectstest.B12Push(g, me.ID, "Skophos Warleader", "Creature — Minotaur Warrior", oldB127WarleaderOracl, 4, 5)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, lead, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{lead}, Strict: true}); err == nil {
		t.Fatal("the Warleader sacrificed itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, lead, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, lead); p != 5 {
		t.Errorf("power %d, want 5", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, lead, "menace") {
		t.Error("no menace")
	}
}

// --- Stampede Rider ----------------------------------------------------------------

func TestOldB127StampedeRiderGrowsOnlyWithABigCreature(t *testing.T) {
	for _, big := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB127Seats(g)
		rider := effectstest.B12Push(g, me.ID, "Stampede Rider", "Creature — Satyr", oldB127RiderOracle, 2, 3)
		if big {
			effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
		} else {
			effectstest.B12Creature(g, me.ID, "Small Bear", "Creature — Bear", 3, 3)
		}
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		p, tough := effectstest.PtOf(t, g, rider)
		want := 2
		if big {
			want = 3
		}
		if p != want || tough != want+1 {
			t.Errorf("big creature %v: Rider is %d/%d, want %d/%d", big, p, tough, want, want+1)
		}
	}
}

// --- Swimmer in Nightmares ----------------------------------------------------------

func TestOldB127SwimmerGrowsWithTenCardsInASingleGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB127Seats(g)
	swimmer := effectstest.B12Push(g, me.ID, "Swimmer in Nightmares", "Creature — Nightmare Merfolk", oldB127SwimmerOracle, 1, 4)
	// Real zone moves, so the layers see each arrival.
	fill := func(p *game.Player, n int) {
		for i := 0; i < n; i++ {
			id := uuid.New()
			p.Hand.PushTop(game.Card{InstanceID: id, Name: "Spent", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
			if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: p.ID}, id); err != nil {
				t.Fatalf("move to graveyard: %v", err)
			}
		}
	}
	fill(me, 5)
	fill(opp, 5)
	g.RunStateChecksForTest()
	if p, _ := effectstest.PtOf(t, g, swimmer); p != 1 {
		t.Errorf("power %d with ten cards split over two graveyards, want 1", p)
	}
	fill(opp, 5)
	g.RunStateChecksForTest()
	if p, _ := effectstest.PtOf(t, g, swimmer); p != 4 {
		t.Errorf("power %d with ten cards in one graveyard, want 4", p)
	}
}

func TestOldB127SwimmerCantBeBlockedWhileYouControlAnAshiok(t *testing.T) {
	for _, ashiok := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB127Seats(g)
		swimmer := effectstest.B12Push(g, me.ID, "Swimmer in Nightmares", "Creature — Nightmare Merfolk", oldB127SwimmerOracle, 1, 4)
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
		if ashiok {
			effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ashiok, Test", TypeLine: "Legendary Planeswalker — Ashiok",
				Counters: map[string]int{game.CounterLoyalty: 5}, Owner: me.ID, Controller: me.ID})
		}
		effectstest.BrAttack(t, g, swimmer)
		err := effectstest.ReqBlock(t, g, swimmer, blocker)
		if ashiok {
			effectstest.BrRefusal(t, err, game.BlockReasonCantBeBlocked)
		} else if err != nil {
			t.Errorf("without an Ashiok the Swimmer should be blockable: %v", err)
		}
	}
}

// --- Enemy of Enlightenment ---------------------------------------------------------

func TestOldB127EnemyShrinksPerCardInOpponentsHandsAndMakesEveryoneDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB127Seats(g)
	enemy := effectstest.B12Push(g, me.ID, "Enemy of Enlightenment", "Enchantment Creature — Demon", oldB127EnemyOracle, 5, 5)
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Hand.Cards = nil
		}
		for i := 0; i < 2; i++ {
			opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		}
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mine", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	g.RunStateChecksForTest()
	if p, tough := effectstest.PtOf(t, g, enemy); p != 3 || tough != 3 {
		t.Errorf("Enemy is %d/%d with two cards in opponents' hands, want 3/3", p, tough)
	}
	drawn := effectstest.TopOfLibrary(opp, "Drawn", "Sorcery", "{1}")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneLibrary, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, drawn); err != nil {
		t.Fatalf("draw: %v", err)
	}
	g.RunStateChecksForTest()
	if p, _ := effectstest.PtOf(t, g, enemy); p != 2 {
		t.Errorf("Enemy power %d after an opponent draws, want 2", p)
	}
}

// One card in each hand keeps the Enemy (-1/-1 per card in opponents' hands) alive.
func TestOldB127EnemyUpkeepMakesEachPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.B12Push(g, me.ID, "Enemy of Enlightenment", "Enchantment Creature — Demon", oldB127EnemyOracle, 5, 5)
	for _, p := range g.Seats {
		effectstest.D2dSetHand(p, 1)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if effectstest.DiscardChoiceFor(g, p.ID) == nil {
			t.Errorf("seat %s was not asked to discard", p.ID)
			continue
		}
		effectstest.AnswerDiscard(t, g, p.ID, p.Hand.Cards[0].InstanceID)
		if p.Hand.Size() != 0 {
			t.Errorf("hand %d after the discard, want 0", p.Hand.Size())
		}
	}
}

// --- Eidolon of Inspiration and Nyx Herald -------------------------------------------

func TestOldB127EidolonPumpsATargetOnYourCombatOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB127Seats(g)
	effectstest.B12Push(g, me.ID, "Eidolon of Inspiration", "Enchantment Creature — Spirit", oldB127EidolonOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 4 {
		t.Errorf("target has power %d, want 4", p)
	}
}

func TestOldB127NyxHeraldTargetsOnlyEnchantedOrEnchantmentCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB127Seats(g)
	effectstest.B12Push(g, me.ID, "Nyx Herald", "Enchantment Creature — Centaur Shaman", oldB127HeraldOracle, 2, 3)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	enchanted := effectstest.B12Creature(g, me.ID, "Enchanted Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Test Aura", "Enchantment — Aura", "", false)
	effectstest.AttachForTest(t, g, aura, enchanted)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, plain) || !effectstest.HasID(pick.PickTargetCards, enchanted) {
		t.Fatalf("targets offered %v, want the enchanted creature and not the plain one", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, enchanted)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, enchanted); p != 3 || tough != 3 {
		t.Errorf("target is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, enchanted, "trample") {
		t.Error("target did not gain trample")
	}
}

// --- Sunlit Hoplite -------------------------------------------------------------------

func TestOldB127SunlitHopliteHasFirstStrikeOnlyOnYourTurnAndGrowsWithElspeth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB127Seats(g)
	hoplite := effectstest.B12Push(g, me.ID, "Sunlit Hoplite", "Creature — Human Soldier", oldB127HopliteOracle, 2, 1)
	theirs := effectstest.B12Push(g, opp.ID, "Sunlit Hoplite", "Creature — Human Soldier", oldB127HopliteOracle, 2, 1)
	if !effectstest.HasEffectiveKeyword(t, g, hoplite, "first strike") {
		t.Error("no first strike on its controller's turn")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "first strike") {
		t.Error("first strike during the opponent's turn")
	}
	if p, _ := effectstest.PtOf(t, g, hoplite); p != 2 {
		t.Errorf("power %d without an Elspeth, want 2", p)
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Elspeth, Test", TypeLine: "Legendary Planeswalker — Elspeth",
		Counters: map[string]int{game.CounterLoyalty: 5}, Owner: me.ID, Controller: me.ID})
	g.RunStateChecksForTest()
	if p, _ := effectstest.PtOf(t, g, hoplite); p != 3 {
		t.Errorf("power %d with an Elspeth, want 3", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Hoplite is %d, want 2 (Elspeth is not theirs)", p)
	}
}

// --- Karametra's Blessing ----------------------------------------------------------------

func TestOldB127KarametrasBlessingGrantsProtectionOnlyToEnchantedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB127Seats(g)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	enchCreature := effectstest.B12Creature(g, me.ID, "Nyx Bear", "Enchantment Creature — Bear", 2, 2)
	enchanted := effectstest.B12Creature(g, me.ID, "Enchanted Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Test Aura", "Enchantment — Aura", "", false)
	effectstest.AttachForTest(t, g, aura, enchanted)
	effectstest.AdvanceToMain(t, g)
	for _, c := range []struct {
		id   uuid.UUID
		want bool
	}{{plain, false}, {enchCreature, true}, {enchanted, true}} {
		effectstest.CastCatalogSpell(t, g, "Karametra's Blessing", "Instant", oldB127BlessingOracle, []game.TargetRef{{Kind: game.TargetCard, ID: c.id}})
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, c.id); p != 4 || tough != 4 {
			t.Errorf("target is %d/%d, want 4/4", p, tough)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, c.id, "hexproof") && effectstest.HasEffectiveKeyword(t, g, c.id, "indestructible"); got != c.want {
			t.Errorf("hexproof and indestructible = %v, want %v", got, c.want)
		}
	}
}

// --- Devourer of Memory ------------------------------------------------------------------

func TestOldB127DevourerGrowsAndEvadesWhenYourLibraryFeedsYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB127Seats(g)
	devourer := effectstest.B12Push(g, me.ID, "Devourer of Memory", "Creature — Nightmare", oldB127DevourerOracle, 2, 1)
	effectstest.AdvanceToMain(t, g)
	// An opponent milling does not trigger it.
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, devourer); p != 2 {
		t.Fatalf("an opponent's mill pumped the Devourer to %d", p)
	}
	effectstest.FloatMana(t, g, me, "{C}{U}{B}")
	if err := g.ActivateCatalogAbility(me.ID, devourer, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("mill a card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, devourer); p != 3 || tough != 2 {
		t.Errorf("Devourer is %d/%d after milling itself, want 3/2", p, tough)
	}
	if r := effectstest.BattlefieldCardFor(g, devourer).Effective().Restrictions; r&game.CantBeBlocked == 0 {
		t.Error("the Devourer can be blocked")
	}
}
