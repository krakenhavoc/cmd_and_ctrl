package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b146_cards_test.go — Throne of Eldraine, S82 older sets batch 146
// (#5516): Irencrag Feat, Covetous Urge, Festive Funeral, Specter's Shriek,
// Blow Your House Down and Reaper of Night // Harvest Fear.

const (
	eld146Feat    = "2d550e17-7340-4b11-8b25-1d4f1a58f82f"
	eld146Urge    = "6f52008f-3365-4cf9-9cae-8e761d657902"
	eld146Funeral = "85d7c8d4-6f99-4b1d-9e0e-ff41bbeee682"
	eld146Shriek  = "8c9164b5-10ed-4cad-af33-7d4b5d3d6bf2"
	eld146Blow    = "a44e8d98-3bc8-42ce-b09f-a3b947d9ce53"
	eld146Reaper  = "d40e5a76-6c9a-47eb-b412-d3056cf36b3f"
)

func TestS82OldB146EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eld146Feat: "Irencrag Feat", eld146Urge: "Covetous Urge", eld146Funeral: "Festive Funeral",
		eld146Shriek: "Specter's Shriek", eld146Blow: "Blow Your House Down",
		eld146Reaper: "Reaper of Night", eld146Reaper + "#1": "Harvest Fear",
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

// --- Irencrag Feat ---

func TestS82OldB146IrencragFeatAddsSevenRedAndAllowsExactlyOneMoreSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.CastCatalogSpell(t, g, "Irencrag Feat", "Sorcery", eld146Feat, nil)
	effectstest.PassPriorityAroundTable(t, g)
	red := 0
	for _, tok := range me.ManaPool.Units() {
		if tok.Color == "R" {
			red++
		}
	}
	if red != 7 {
		t.Errorf("%d red mana in the pool, want 7", red)
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Spell One", "Instant", "", nil); err != nil {
		t.Fatalf("the one more spell was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Spell Two", "Instant", "", nil); err == nil {
		t.Error("a second spell after the Feat was allowed")
	}
}

func TestS82OldB146IrencragFeatCountsSpellsAlreadyCastThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.CastCatalogSpell(t, g, "Earlier", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Irencrag Feat", "Sorcery", eld146Feat, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Spell One", "Instant", "", nil); err != nil {
		t.Fatalf("the one more spell was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Spell Two", "Instant", "", nil); err == nil {
		t.Error("a further spell was allowed")
	}
}

// --- Covetous Urge ---

func TestS82OldB146CovetousUrgeTakesANonlandCardFromGraveyardOrHandAndLetsYouCastIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	hand := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt())
	opp.Graveyard.Cards = nil
	landGrave, spellGrave := uuid.New(), uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: landGrave, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: spellGrave, Name: "Dark Ritual", TypeLine: "Instant", ManaCost: "{B}", Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastAt(t, g, "Covetous Urge", "Sorcery", eld146Urge, effectstest.PbPlayer(opp.ID))
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, []uuid.UUID{hand[1], spellGrave}) {
		t.Fatalf("options = %v, want the Bolt and the graveyard Ritual only", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{spellGrave}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	g.SettleResolution()
	if !g.Exile.Contains(spellGrave) {
		t.Fatal("the chosen card was not exiled")
	}
	g.Turn.Step = game.StepPrecombatMain
	effectstest.FloatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, spellGrave, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled card with blue mana: %v", err)
	}
}

func TestS82OldB146CovetousUrgeCardFromHandCanBeCastAndOthersCannot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	hand := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt())
	opp.Graveyard.Cards = nil
	effectstest.CastAt(t, g, "Covetous Urge", "Sorcery", eld146Urge, effectstest.PbPlayer(opp.ID))
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{hand[1]}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	g.SettleResolution()
	g.Turn.Step = game.StepPrecombatMain
	effectstest.FloatForTest(g, me, "UU")
	if err := g.CastSpell(me.ID, hand[1], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled Bolt with blue mana: %v", err)
	}
	if !opp.Hand.Contains(hand[0]) {
		t.Error("the land left their hand")
	}
}

// --- Festive Funeral ---

func TestS82OldB146FestiveFuneralShrinksByTheGraveyardCount(t *testing.T) {
	for _, tc := range []struct {
		graveyard int
		dies      bool
	}{{2, false}, {3, true}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		me.Graveyard.Cards = nil
		for i := 0; i < tc.graveyard; i++ {
			me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		}
		bear := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 3, 3)
		effectstest.CastCatalogSpell(t, g, "Festive Funeral", "Instant", eld146Funeral, effectstest.CardTarget(bear))
		effectstest.PassPriorityAroundTable(t, g)
		alive := effectstest.AppInZone(g, bear, game.ZoneBattlefield)
		if alive == tc.dies {
			t.Errorf("graveyard %d: bear alive = %v, want %v", tc.graveyard, alive, !tc.dies)
		}
		if alive {
			if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 1 || tough != 1 {
				t.Errorf("graveyard %d: the Bear is %d/%d, want 1/1", tc.graveyard, p, tough)
			}
		}
	}
}

// --- Specter's Shriek ---

func eld146Shriek145(t *testing.T) (*game.Game, *game.Player, *game.Player, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	ids := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt(),
		game.Card{Name: "Doom Blade", TypeLine: "Instant", ManaCost: "{1}{B}", Colors: []string{"B"}},
		game.Card{Name: "Wurm Relic", TypeLine: "Artifact", ManaCost: "{2}"})
	return g, me, opp, ids
}

func TestS82OldB146SpectersShriekNonblackCardCostsYouACardFromHand(t *testing.T) {
	for _, pickIdx := range []int{1, 3} { // Lightning Bolt (red), the colourless artifact
		g, me, opp, ids := eld146Shriek145(t)
		mine := effectstest.HandCardFull(me, "Mine", "Sorcery", "{1}", "", nil)
		effectstest.CastAt(t, g, "Specter's Shriek", "Sorcery", eld146Shriek, effectstest.PbPlayer(opp.ID))
		pick := effectstest.OpenVariant(t, g)
		if !pick.PickOptional {
			t.Error("the pick is not optional")
		}
		if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[pickIdx]}); err != nil {
			t.Fatalf("resolve: %v", err)
		}
		g.SettleResolution()
		if !g.Exile.Contains(ids[pickIdx]) || effectstest.RvDiscards(g, ids[pickIdx]) != 0 {
			t.Errorf("pick %d: the card was not exiled (as an exile, not a discard)", pickIdx)
		}
		ask := effectstest.ChooseCardsChoiceFor(g, me.ID)
		if ask == nil {
			t.Fatalf("pick %d: no exile-from-your-hand prompt", pickIdx)
		}
		if err := g.ResolveChooseCards(ask.ID, me.ID, []uuid.UUID{mine}); err != nil {
			t.Fatalf("exile from hand: %v", err)
		}
		if !g.Exile.Contains(mine) {
			t.Errorf("pick %d: the caster's card was not exiled", pickIdx)
		}
	}
}

func TestS82OldB146SpectersShriekBlackCardOrNoPickCostsNothing(t *testing.T) {
	for _, pickIdx := range []int{2, -1} { // Doom Blade (black), no pick
		g, me, opp, ids := eld146Shriek145(t)
		mine := effectstest.HandCardFull(me, "Mine", "Sorcery", "{1}", "", nil)
		effectstest.CastAt(t, g, "Specter's Shriek", "Sorcery", eld146Shriek, effectstest.PbPlayer(opp.ID))
		pick := effectstest.OpenVariant(t, g)
		var picked []uuid.UUID
		if pickIdx >= 0 {
			picked = []uuid.UUID{ids[pickIdx]}
		}
		if err := g.ResolvePendingChoice(pick.ID, me.ID, picked); err != nil {
			t.Fatalf("resolve: %v", err)
		}
		g.SettleResolution()
		if pickIdx >= 0 && !g.Exile.Contains(ids[pickIdx]) {
			t.Error("the black card was not exiled")
		}
		if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
			t.Errorf("pick %d: the caster was asked to exile a card", pickIdx)
		}
		if !me.Hand.Contains(mine) {
			t.Errorf("pick %d: the caster's card left their hand", pickIdx)
		}
	}
}

// --- Blow Your House Down ---

func TestS82OldB146BlowYourHouseDownStopsBlocksAndDestroysWalls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	wall := effectstest.B12Push(g, them.ID, "Stone Wall", "Creature — Wall", "", 0, 4)
	bear := effectstest.B12Push(g, them.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	other := effectstest.B12Push(g, them.ID, "Their Other", "Creature — Bear", "", 2, 2)
	attacker := effectstest.B12Push(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Blow Your House Down", "Sorcery", eld146Blow, effectstest.CardRefs(wall, bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AppInZone(g, wall, game.ZoneBattlefield) {
		t.Error("the Wall was not destroyed")
	}
	if !effectstest.AppInZone(g, bear, game.ZoneBattlefield) {
		t.Fatal("a non-Wall target was destroyed")
	}
	effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
	if effectstest.BlockRefusal(t, g, bear, attacker) == "" {
		t.Error("the targeted creature blocked")
	}
	if got := effectstest.BlockRefusal(t, g, other, attacker); got != "" {
		t.Errorf("an untargeted creature was refused: %q", got)
	}
}

// --- Reaper of Night // Harvest Fear ---

func TestS82OldB146ReaperOfNightGainsFlyingOnlyAgainstAShortHand(t *testing.T) {
	for _, tc := range []struct {
		hand   int
		flying bool
	}{{2, true}, {3, false}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, them := effectstest.SeatsForCombat(g)
		reaper := effectstest.B12Push(g, me.ID, "Reaper of Night", "Creature — Specter", eld146Reaper, 4, 5)
		them.Hand.Cards = nil
		for i := 0; i < tc.hand; i++ {
			them.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: them.ID, Controller: them.ID})
		}
		effectstest.DeclareAttack(t, g, them.ID, reaper)
		effectstest.PassPriorityAroundTable(t, g)
		has := effectstest.HasKeywordOnBattlefield(t, g, reaper, "flying")
		if has != tc.flying {
			t.Errorf("opponent hand %d: flying = %v, want %v", tc.hand, has, tc.flying)
		}
	}
}

func TestS82OldB146HarvestFearMakesTheOpponentDiscardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	ids := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbDivination())
	effectstest.CastCatalogSpell(t, g, "Harvest Fear", "Sorcery — Adventure", eld146Reaper+"#1", []game.TargetRef{effectstest.PbPlayer(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, ids[0], ids[2])
	if opp.Hand.Size() != 1 || !opp.Hand.Contains(ids[1]) {
		t.Errorf("the opponent holds %d cards, want only the Bolt", opp.Hand.Size())
	}
}
