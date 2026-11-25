package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b03_test.go — S82 GAP2 batch 3 (#5948). Every test plays the
// real catalog card through the engine.

const (
	gap2B03Titans  = "b011d6b2-8930-4192-aa94-82869faa991b"
	gap2B03Hollow  = "b9e4d34a-ba91-4ad6-b886-daeeda6c9498"
	gap2B03Ian     = "32c9ce8c-a19d-49e4-bc6d-c7182fa00b51"
	gap2B03Phoenix = "52c3fe55-74ee-44cb-b62b-dc6c5dfdc33e"
	gap2B03Benalia = "c15bb7eb-aaaa-4468-9641-8f706d6137e8" // History of Benalia
)

func TestGap2B03CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		gap2B03Titans: "Titan's Presence", gap2B03Hollow: "Hollow One",
		gap2B03Ian: "Ian Chesterton", gap2B03Phoenix: "Detective's Phoenix",
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

// --- Titan's Presence ---------------------------------------------------

func gap2B03HandCreature(p *game.Player, name, cost string, power int, colors ...string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Artifact Creature — Golem", ManaCost: cost,
		Colors: colors, Power: power, Toughness: power, Owner: p.ID, Controller: p.ID})
	return id
}

func gap2B03CastTitans(t *testing.T, g *game.Game, me *game.Player, params game.CastSpellParams) error {
	t.Helper()
	id := effectstest.TrekB39Hand(g, me, "Titan's Presence", "Instant", "{3}", gap2B03Titans, 0, 0)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	return g.CastSpell(me.ID, id, params)
}

func TestGap2B03TitansPresenceExilesACreatureNoStrongerThanTheRevealedCard(t *testing.T) {
	for _, tc := range []struct {
		power  int
		exiled bool
	}{{5, true}, {3, true}, {6, false}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		revealed := gap2B03HandCreature(me, "Golem Five", "{7}", 5)
		victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", tc.power, tc.power)
		if err := gap2B03CastTitans(t, g, me, game.CastSpellParams{Strict: true, RevealIDs: []uuid.UUID{revealed},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
			t.Fatalf("power %d: cast: %v", tc.power, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.InExile(g, victim); got != tc.exiled {
			t.Errorf("power %d against a revealed 5: exiled %v, want %v", tc.power, got, tc.exiled)
		}
		if !me.Hand.Contains(revealed) {
			t.Error("the revealed card left the hand")
		}
	}
}

// The revealed card's power is what it was revealed with: discarding it in
// response does not change the comparison (CR 608.2h).
func TestGap2B03TitansPresenceReadsThePowerAsRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	revealed := gap2B03HandCreature(me, "Golem Four", "{6}", 4)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 4, 4)
	if err := gap2B03CastTitans(t, g, me, game.CastSpellParams{Strict: true, RevealIDs: []uuid.UUID{revealed},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(me.Hand, me.Graveyard, revealed); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, victim) {
		t.Error("the 4-power creature survived a revealed 4-power card that left the hand")
	}
}

func TestGap2B03TitansPresenceRefusesAColoredNoncreatureOrMissingReveal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 1, 1)
	colored := gap2B03HandCreature(me, "Colored", "{3}{W}", 9, "W")
	noncreature := effectstest.HandCardForTest(me, "Rock", "Artifact", "")
	tgt := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}
	for name, params := range map[string]game.CastSpellParams{
		"no reveal":   {Strict: true, Targets: tgt},
		"colored":     {Strict: true, Targets: tgt, RevealIDs: []uuid.UUID{colored}},
		"noncreature": {Strict: true, Targets: tgt, RevealIDs: []uuid.UUID{noncreature}},
	} {
		if err := gap2B03CastTitans(t, g, me, params); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
		effectstest.DrainPool(t, g, me.ID)
	}
}

// --- Hollow One ---------------------------------------------------------

func gap2B03HollowInHand(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.TrekB39Hand(g, me, "Hollow One", "Artifact Creature — Golem", "{5}", gap2B03Hollow, 4, 4)
}

func gap2B03Cycle(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycling: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// Two cards cycled (each also discarded, counted once) take {4} off: the
// third Hollow One costs {1}.
func TestGap2B03HollowOneCostsTwoLessForEachCardCycledOrDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gap2B03Cycle(t, g, me, gap2B03HollowInHand(g, me))
	gap2B03Cycle(t, g, me, gap2B03HollowInHand(g, me))
	if n := g.TurnTallyFor(me.ID).CardsDiscarded; n != 2 {
		t.Fatalf("%d cards discarded this turn, want 2", n)
	}
	hollow := gap2B03HollowInHand(g, me)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, hollow, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Hollow One for {1} after two cycles: %v", err)
	}
	if left := me.ManaPool.Total(); left != 0 {
		t.Errorf("%d mana left over", left)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldByName(g, "Hollow One") != 1 {
		t.Error("Hollow One did not resolve")
	}
}

func TestGap2B03HollowOneCostsFullWithNothingDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hollow := gap2B03HollowInHand(g, me)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, hollow, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("Hollow One cast for {4} with nothing discarded this turn")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, hollow, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Hollow One for {5}: %v", err)
	}
}

// --- Ian Chesterton -----------------------------------------------------

func gap2B03ResolveAll(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 10 && g.Stack.Size() > 0; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
}

// A Saga spell with replicate paid once is copied once; the copy becomes a
// Saga token, which gets its lore counter and its chapter I.
func TestGap2B03IanGivesSagaSpellsReplicateAtTheirManaCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Ian Chesterton", "Legendary Creature — Human Scientist", gap2B03Ian, 2, 3)
	saga := effectstest.TrekB39Hand(g, me, "History of Benalia", "Enchantment — Saga", "{1}{W}{W}", gap2B03Benalia, 0, 0)
	effectstest.FloatMana(t, g, me, "{W}{W}{W}{W}{C}{C}")
	if err := g.CastSpell(me.ID, saga, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("History of Benalia with replicate once: %v", err)
	}
	if left := me.ManaPool.Total(); left != 0 {
		t.Errorf("%d mana left after paying {1}{W}{W} twice", left)
	}
	gap2B03ResolveAll(t, g)
	if n := effectstest.CountBattlefieldByName(g, "History of Benalia"); n != 2 {
		t.Fatalf("%d History of Benalia on the battlefield, want the card and a token copy", n)
	}
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "History of Benalia" && c.IsToken() {
			tokens++
		}
	}
	if tokens != 1 {
		t.Errorf("%d Saga tokens, want 1", tokens)
	}
	if n := effectstest.CountBattlefieldByName(g, "Knight"); n != 2 {
		t.Errorf("%d Knight tokens, want one from each Saga's chapter I", n)
	}
}

func TestGap2B03IanGrantsOnlyItsControllersSagaSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.B12Push(g, me.ID, "Ian Chesterton", "Legendary Creature — Human Scientist", gap2B03Ian, 2, 3)
	bear := effectstest.TrekB39Hand(g, me, "Test Bear", "Creature — Bear", "{1}{W}", "", 2, 2)
	effectstest.FloatMana(t, g, me, "{W}{W}{W}{W}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err == nil {
		t.Fatal("a non-Saga spell was given replicate")
	}
	effectstest.DrainPool(t, g, me.ID)

	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	opp2 := g2.Seats[(g2.Turn.ActiveSeat+1)%4]
	effectstest.B12Push(g2, opp2.ID, "Ian Chesterton", "Legendary Creature — Human Scientist", gap2B03Ian, 2, 3)
	saga := effectstest.TrekB39Hand(g2, me2, "History of Benalia", "Enchantment — Saga", "{1}{W}{W}", gap2B03Benalia, 0, 0)
	effectstest.FloatMana(t, g2, me2, "{W}{W}{W}{W}{C}{C}")
	if err := g2.CastSpell(me2.ID, saga, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err == nil {
		t.Fatal("an opponent's Ian Chesterton gave my Saga spell replicate")
	}
	_ = opp
}

// --- Detective's Phoenix ------------------------------------------------

func gap2B03PhoenixIn(me *game.Player, zone *game.Zone) uuid.UUID {
	id := uuid.New()
	zone.PushTop(game.Card{InstanceID: id, Name: "Detective's Phoenix", TypeLine: "Enchantment Creature — Phoenix",
		ManaCost: "{2}{R}", OracleID: gap2B03Phoenix, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	return id
}

func gap2B03PhoenixTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	host := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	return g, me, host
}

func gap2B03WantBestowed(t *testing.T, g *game.Game, phoenix, host uuid.UUID) {
	t.Helper()
	if effectstest.B12ZoneOf(g, phoenix) != game.ZoneBattlefield {
		t.Fatalf("the Phoenix is in %v", effectstest.B12ZoneOf(g, phoenix))
	}
	if p, tough := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 4 || tough != 4 {
		t.Errorf("enchanted creature is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, host, "flying") || !effectstest.HasEffectiveKeyword(t, g, host, "haste") {
		t.Error("the enchanted creature lacks flying or haste")
	}
}

func TestGap2B03PhoenixBestowsFromHandForRAndEvidenceSix(t *testing.T) {
	g, me, host := gap2B03PhoenixTable(t)
	phoenix := gap2B03PhoenixIn(me, me.Hand)
	gy := effectstest.MkmCEEvidence(me, "{3}", "{2}{R}")
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, phoenix, game.CastSpellParams{
		Strict: true, AlternativeCost: game.BestowKey, AltCostIDs: gy,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}},
	}); err != nil {
		t.Fatalf("bestow for {R} and evidence 6: %v", err)
	}
	if !effectstest.InExile(g, gy[0]) || !effectstest.InExile(g, gy[1]) {
		t.Error("the evidence was not exiled")
	}
	effectstest.WantOneCollection(t, g, me.ID, 6, gy...)
	effectstest.PassPriorityAroundTable(t, g)
	gap2B03WantBestowed(t, g, phoenix, host)
}

func TestGap2B03PhoenixBestowsFromTheGraveyard(t *testing.T) {
	g, me, host := gap2B03PhoenixTable(t)
	phoenix := gap2B03PhoenixIn(me, me.Graveyard)
	gy := effectstest.MkmCEEvidence(me, "{6}")
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, phoenix, game.CastSpellParams{
		Strict: true, FromZone: string(game.ZoneGraveyard), AlternativeCost: effects.BestowGraveyardKey, AltCostIDs: gy,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}},
	}); err != nil {
		t.Fatalf("bestow from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	gap2B03WantBestowed(t, g, phoenix, host)
}

// From the graveyard it is cast only bestowed, and only with enough
// evidence; it can't be its own evidence.
func TestGap2B03PhoenixGraveyardCastRefusals(t *testing.T) {
	g, me, host := gap2B03PhoenixTable(t)
	phoenix := gap2B03PhoenixIn(me, me.Graveyard)
	short := effectstest.MkmCEEvidence(me, "{3}", "{1}{R}")
	tgt := []game.TargetRef{{Kind: game.TargetCard, ID: host}}
	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	for name, params := range map[string]game.CastSpellParams{
		"its mana cost":      {Strict: true, FromZone: string(game.ZoneGraveyard)},
		"evidence of 5":      {Strict: true, FromZone: string(game.ZoneGraveyard), AlternativeCost: effects.BestowGraveyardKey, AltCostIDs: short, Targets: tgt},
		"itself as evidence": {Strict: true, FromZone: string(game.ZoneGraveyard), AlternativeCost: effects.BestowGraveyardKey, AltCostIDs: append([]uuid.UUID{phoenix}, short...), Targets: tgt},
		"the hand offer":     {Strict: true, FromZone: string(game.ZoneGraveyard), AlternativeCost: effects.BestowGraveyardKey + "x", AltCostIDs: short, Targets: tgt},
	} {
		if err := g.CastSpell(me.ID, phoenix, params); err == nil {
			t.Errorf("%s: the graveyard cast was accepted", name)
		}
	}
	if !me.Graveyard.Contains(phoenix) || !me.Graveyard.Contains(short[0]) {
		t.Error("a refused cast moved something")
	}
}

// From hand the graveyard offer is not claimable, and the card still casts
// as a 2/2 flying, haste creature for {2}{R}.
func TestGap2B03PhoenixCastsAsACreatureFromHand(t *testing.T) {
	g, me, _ := gap2B03PhoenixTable(t)
	phoenix := gap2B03PhoenixIn(me, me.Hand)
	gy := effectstest.MkmCEEvidence(me, "{6}")
	effectstest.FloatMana(t, g, me, "{R}{C}{C}")
	if err := g.CastSpell(me.ID, phoenix, game.CastSpellParams{Strict: true, AlternativeCost: effects.BestowGraveyardKey, AltCostIDs: gy}); err == nil {
		t.Fatal("the graveyard bestow offer was claimed from hand")
	}
	if err := g.CastSpell(me.ID, phoenix, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {2}{R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, phoenix, "flying") || !effectstest.HasEffectiveKeyword(t, g, phoenix, "haste") {
		t.Error("the Phoenix lacks flying or haste")
	}
}
