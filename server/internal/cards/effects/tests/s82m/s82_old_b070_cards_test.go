package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaldheim batch 070 (S82 OLD, #5505).

const (
	oldB070DreamDevourer = "0ff7440b-b831-40a9-81c4-99d77bf93b8d"
	oldB070Omenpaths     = "9ac01998-6e26-4ba5-b66c-614ec831caef"
	oldB070HaraldsRev    = "bc781549-7b2b-440b-93ae-69677acdfdad"
	oldB070IcehideTroll  = "ce55c9a8-25ba-404c-95b4-c14ba4763081"
	oldB070MaskedVandal  = "210940f1-3c11-4877-bbfc-2429c03f98ee"
	oldB070Ritualist     = "37420ccd-c6c6-4f15-a696-d8986934afb5"
	oldB070GladeWarden   = "438a0a7a-00bc-4d93-8e69-a8a6084523c5"
	oldB070Moritte       = "78699161-fc14-4a44-8a15-0f7c08be0343"
	oldB070Deathknell    = "4658ae19-d203-4c12-b56e-42cd8f0db91d"
	oldB070Rampage       = "49039519-62db-4605-b600-cef6f34c893f"
	oldB070Smashing      = "cb19d2f7-0511-4c66-ba4c-f38516920c20"
	oldB070Skald         = "dc7e259a-d965-4e25-8145-81ab3087f718"
)

func TestOldB070CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB070DreamDevourer: "Dream Devourer", oldB070Omenpaths: "Open the Omenpaths", oldB070HaraldsRev: "King Harald's Revenge",
		oldB070IcehideTroll: "Icehide Troll", oldB070MaskedVandal: "Masked Vandal", oldB070Ritualist: "Gladewalker Ritualist",
		oldB070GladeWarden: "Littjara Glade-Warden", oldB070Moritte: "Moritte of the Frost", oldB070Deathknell: "Deathknell Berserker",
		oldB070Rampage: "Rampage of the Valkyries", oldB070Smashing: "Smashing Success", oldB070Skald: "Warchanter Skald",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB070Attach(g *game.Game, owner, host uuid.UUID, typeLine string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Attachment", TypeLine: typeLine,
		Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

// --- Dream Devourer ------------------------------------------------------

// oldB070Foretells returns the foretell offers (cast cost) the seat has on a card in hand.
func oldB070Foretells(g *game.Game, me *game.Player, id uuid.UUID) []string {
	var out []string
	g.WithWriteLock(func() {
		card, ok := g.LookupCardForEffect(id)
		if !ok {
			return
		}
		for _, sa := range g.SpecialActionsOfferedLocked(me.ID, card, game.ZoneHand) {
			if sa.Kind == game.SpecialActionForetell {
				out = append(out, sa.Cost+"|"+sa.CastCost)
			}
		}
	})
	return out
}

func oldB070HandCard(me *game.Player, name, typeLine, oracle, mana string, p, tough int) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: mana,
		Power: p, Toughness: tough, Owner: me.ID, Controller: me.ID})
	return id
}

func TestOldB070DreamDevourerGivesForetellAtManaCostLessTwoToNonlandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := oldB070HandCard(me, "Bear", "Creature — Bear", "", "{3}{G}", 2, 2)
	gold := oldB070HandCard(me, "Gold", "Creature — Knight", "", "{B}{B}", 2, 2)
	land := oldB070HandCard(me, "Forest", "Basic Land — Forest", "", "", 0, 0)
	free := oldB070HandCard(me, "Free", "Artifact", "", "", 0, 0)
	if got := oldB070Foretells(g, me, bear); len(got) != 0 {
		t.Fatalf("a card had foretell with no Dream Devourer: %v", got)
	}
	effectstest.B12Push(g, me.ID, "Dream Devourer", "Creature — Demon Cleric", oldB070DreamDevourer, 0, 3)
	if got := oldB070Foretells(g, me, bear); len(got) != 1 || got[0] != "{2}|{1}{G}" {
		t.Errorf("Bear offers %v, want foretell {2} with the foretell cost {1}{G}", got)
	}
	if got := oldB070Foretells(g, me, gold); len(got) != 1 || got[0] != "{2}|{B}{B}" {
		t.Errorf("Gold offers %v, want the coloured mana kept: {B}{B}", got)
	}
	if got := oldB070Foretells(g, me, land); len(got) != 0 {
		t.Errorf("a land was offered foretell: %v", got)
	}
	if got := oldB070Foretells(g, me, free); len(got) != 0 {
		t.Errorf("a card with no mana cost was offered foretell: %v", got)
	}
}

func TestOldB070DreamDevourerForetellsAndCastsForTheReducedCostAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	devourer := effectstest.B12Push(g, me.ID, "Dream Devourer", "Creature — Demon Cleric", oldB070DreamDevourer, 0, 3)
	bear := oldB070HandCard(me, "Bear", "Creature — Bear", "", "{3}{G}", 2, 2)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, bear, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Fatal("the foretold card is not in exile")
	}
	if got := effectstest.EffectivePower(t, g, devourer); got != 2 {
		t.Errorf("Dream Devourer has power %d after a foretell, want 2", got)
	}
	// A later turn: the foretell cost is {1}{G}, not {3}{G}.
	g.WithWriteLock(func() { g.Turn.Seq++ })
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyForetell}); err == nil {
		t.Fatal("cast the foretold card for {G}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyForetell}); err != nil {
		t.Fatalf("cast the foretold card for {1}{G}: %v", err)
	}
}

func TestOldB070DreamDevourerLeavesAPrintedForetellAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Dream Devourer", "Creature — Demon Cleric", oldB070DreamDevourer, 0, 3)
	saw := oldB070HandCard(me, "Saw It Coming", "Instant", effectstest.SawItComingOracle, "{1}{U}{U}", 0, 0)
	if got := oldB070Foretells(g, me, saw); len(got) != 1 || got[0] != "{2}|{1}{U}" {
		t.Errorf("Saw It Coming offers %v, want only its own foretell cost {1}{U}", got)
	}
}

func TestOldB070DreamDevourerGrantsNothingForAnOpponentsHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Dream Devourer", "Creature — Demon Cleric", oldB070DreamDevourer, 0, 3)
	theirs := oldB070HandCard(opp, "Their Bear", "Creature — Bear", "", "{3}{G}", 2, 2)
	if got := oldB070Foretells(g, me, theirs); len(got) != 0 {
		t.Errorf("a card in an opponent's hand was offered foretell: %v", got)
	}
	if got := oldB070Foretells(g, opp, theirs); len(got) != 0 {
		t.Errorf("the opponent was offered foretell by my Devourer: %v", got)
	}
}

// --- Open the Omenpaths ----------------------------------------------------

func oldB070PickColor(t *testing.T, g *game.Game, chooser uuid.UUID, color string) {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceColor, chooser)
	if c == nil {
		t.Fatal("no colour prompt")
	}
	if err := g.ResolveColorChoice(c.ID, chooser, color); err != nil {
		t.Fatalf("ResolveColorChoice(%s): %v", color, err)
	}
}

func TestOldB070OpenTheOmenpathsAddsTwoAndTwoRestrictedToCreatureAndEnchantmentSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpellWithModes(t, g, "Open the Omenpaths", "Instant", oldB070Omenpaths, []int{0})
	effectstest.PassPriorityAroundTable(t, g)
	oldB070PickColor(t, g, me.ID, "R")
	// The second prompt does not offer red.
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceColor, me.ID)
	if c == nil {
		t.Fatal("no second colour prompt")
	}
	if err := g.ResolveColorChoice(c.ID, me.ID, "R"); err == nil {
		t.Fatal("the second colour could be the first")
	}
	oldB070PickColor(t, g, me.ID, "G")
	r, gr := 0, 0
	for _, m := range me.ManaPool.Units() {
		switch m.Color {
		case "R":
			r++
		case "G":
			gr++
		}
	}
	if r != 2 || gr != 2 {
		t.Fatalf("pool has %d red and %d green, want 2 and 2", r, gr)
	}
	// Only a creature or enchantment spell can use it: a sorcery cannot.
	for _, m := range me.ManaPool.Units() {
		if len(m.Restrictions) == 0 {
			t.Fatalf("mana %v carries no restriction", m)
		}
	}
}

func TestOldB070OpenTheOmenpathsPumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpellWithModes(t, g, "Open the Omenpaths", "Instant", oldB070Omenpaths, []int{1})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, mine); got != 3 {
		t.Errorf("my Bear has power %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("their Bear has power %d, want 2", got)
	}
}

// --- King Harald's Revenge -------------------------------------------------

func TestOldB070KingHaraldsRevengePumpsByCreaturesYouControlGivesTrampleAndMustBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	target := effectstest.B12Creature(g, me.ID, "Target", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "King Harald's Revenge", "Sorcery", oldB070HaraldsRev, effectstest.CardRefs(target))
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tough := effectstest.EffectivePTOf(t, g, target); pw != 4 || tough != 4 {
		t.Errorf("target is %d/%d, want 4/4 (+2/+2 for two creatures)", pw, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, target, "trample") {
		t.Error("the target did not gain trample")
	}
	found := false
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAddBlockRequirement && m.Text == string(game.BlockRequirementMustBeBlocked) {
				found = true
			}
		}
	}
	if !found {
		t.Error("the target has no must-be-blocked requirement")
	}
}

// --- Icehide Troll ---------------------------------------------------------

func TestOldB070IcehideTrollPumpsGainsIndestructibleAndTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	troll := effectstest.BlbB1Push(g, me.ID, "Icehide Troll", "Snow Creature — Troll Warrior", oldB070IcehideTroll, 2, 3)
	effectstest.AdvanceToMain(t, g)
	// Two plain Forests make no snow mana.
	effectstest.BlbB1Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.BlbB1Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, troll, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("{S}{S} paid with nonsnow lands")
	}
	effectstest.BlbB1Push(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "", 0, 0)
	effectstest.BlbB1Push(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, troll, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tough := effectstest.EffectivePTOf(t, g, troll); pw != 4 || tough != 3 {
		t.Errorf("Troll is %d/%d, want 4/3", pw, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, troll, "indestructible") {
		t.Error("the Troll did not gain indestructible")
	}
	if c, _ := effectstest.BattlefieldCard(g, troll); !c.Tapped {
		t.Error("the Troll was not tapped")
	}
}

// --- Masked Vandal ---------------------------------------------------------

func TestOldB070MaskedVandalExilesAnOpposingArtifactOrEnchantmentForACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	spell := effectstest.PushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Masked Vandal", "Creature — Shapeshifter", oldB070MaskedVandal, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, mine) {
		t.Fatal("my own artifact was offered")
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no exile prompt")
	}
	if len(p.ChooseCards) != 1 || p.ChooseCards[0] != dead {
		t.Fatalf("offered %v, want only the creature card (not %v)", p.ChooseCards, spell)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead) || !g.Exile.Contains(rock) {
		t.Error("the creature card and the artifact were not both exiled")
	}
}

func TestOldB070MaskedVandalDoesNothingWhenYouDeclineTheExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ench := effectstest.PushCatalogPermanent(g, opp.ID, "Their Ward", "Enchantment", "", false)
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Masked Vandal", "Creature — Shapeshifter", oldB070MaskedVandal, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no exile prompt")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(ench) || g.Exile.Contains(dead) {
		t.Error("something was exiled although the creature card was kept")
	}
	if !g.Battlefield.Contains(ench) {
		t.Error("the enchantment left the battlefield")
	}
}

// --- Gladewalker Ritualist -------------------------------------------------

func TestOldB070GladewalkerRitualistDrawsWhenAnotherOneEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Creature — Bear", "{G}")
	}
	effectstest.B12Push(g, me.ID, "Gladewalker Ritualist", "Creature — Shapeshifter", oldB070Ritualist, 3, 3)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	_ = other
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bear Two", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("another creature entering drew a card: hand %d, want %d", me.Hand.Size(), hand)
	}
	effectstest.CastCatalogSpell(t, g, "Gladewalker Ritualist", "Creature — Shapeshifter", oldB070Ritualist, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d after a second Ritualist entered, want %d (one draw)", me.Hand.Size(), hand+1)
	}
}

func TestOldB070GladewalkerRitualistDoesNotDrawForItselfOrAnOpponentsRitualist(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Creature — Bear", "{G}")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Gladewalker Ritualist", "Creature — Shapeshifter", oldB070Ritualist, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a lone Ritualist drew a card: hand %d, want %d", me.Hand.Size(), hand)
	}
	effectstest.B12Push(g, g.Seats[1].ID, "Gladewalker Ritualist", "Creature — Shapeshifter", oldB070Ritualist, 3, 3)
	if me.Hand.Size() != hand {
		t.Error("an opponent's Ritualist drew me a card")
	}
}

// --- Littjara Glade-Warden -------------------------------------------------

func TestOldB070LittjaraGladeWardenPutsTwoCountersForAnExiledCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	warden := effectstest.B12Push(g, me.ID, "Littjara Glade-Warden", "Creature — Shapeshifter", oldB070GladeWarden, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spell := effectstest.PushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{G}")
	target := effectstest.CardRefs(bear)
	if err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Strict: true, Targets: target, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a sorcery paid \"exile a creature card\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Strict: true, Targets: target, ExileIDs: []uuid.UUID{dead}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters on the Bear, want 2", n)
	}
	if !g.Exile.Contains(dead) {
		t.Error("the creature card was not exiled")
	}
}

func TestOldB070LittjaraGladeWardenIsSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	warden := effectstest.B12Push(g, me.ID, "Littjara Glade-Warden", "Creature — Shapeshifter", oldB070GladeWarden, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{G}")
	// With something on the stack it is not the main phase with an empty stack.
	effectstest.CastCatalogSpell(t, g, "Spell", "Instant", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear), ExileIDs: []uuid.UUID{dead}}); err == nil {
		t.Fatal("activated at instant speed")
	}
}

// --- Moritte of the Frost --------------------------------------------------

func TestOldB070MoritteCopiesAPermanentAsLegendarySnowWithTwoCountersAndChangeling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bears := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id := effectstest.CastCatalogSpell(t, g, "Moritte of the Frost", "Legendary Snow Creature — Shapeshifter", oldB070Moritte, nil)
	effectstest.ResolveWithCopyChoice(t, g, bears)
	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Grizzly Bears" {
		t.Fatalf("name %q, want a copy of Grizzly Bears", got.Name)
	}
	for _, w := range []string{"Legendary", "Snow", "Creature"} {
		if !got.HasSupertype(w) && !got.HasCardType(w) {
			t.Errorf("the copy is not %s: %q", w, got.TypeLine)
		}
	}
	if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}
	if pw, tough := got.CurrentPower(), got.CurrentToughness(); pw != 4 || tough != 4 {
		t.Errorf("the copy is %d/%d, want 4/4", pw, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "changeling") {
		t.Error("the copy has no changeling")
	}
}

func TestOldB070MoritteCopiesANoncreaturePermanentWithoutCountersOrChangeling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID})
	id := effectstest.CastCatalogSpell(t, g, "Moritte of the Frost", "Legendary Snow Creature — Shapeshifter", oldB070Moritte, nil)
	effectstest.ResolveWithCopyChoice(t, g, rock)
	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Sol Ring" || !got.HasSupertype("Legendary") || !got.HasSupertype("Snow") {
		t.Errorf("got %q (%s), want a legendary snow Sol Ring", got.Name, got.TypeLine)
	}
	if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 0 {
		t.Errorf("%d +1/+1 counters on an artifact copy, want 0", n)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, id, "changeling") {
		t.Error("a noncreature copy has changeling")
	}
}

func TestOldB070MoritteOffersOnlyYourOwnPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedCopyableCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.SeedCopyableCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Moritte of the Frost", "Legendary Snow Creature — Shapeshifter", oldB070Moritte, nil)
	for i := 0; i < 8 && effectstest.CopyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := effectstest.CopyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if len(pc.CopyOptions) != 1 || pc.CopyOptions[0] != mine {
		t.Errorf("CopyOptions = %v, want only my permanent %s", pc.CopyOptions, mine)
	}
}

// --- Deathknell Berserker --------------------------------------------------

func TestOldB070DeathknellBerserkerMakesAZombieOnlyWhenItsPowerWasThreeOrMore(t *testing.T) {
	for _, c := range []struct {
		name  string
		power int
		want  int
	}{{"power 2", 2, 0}, {"power 3", 3, 1}, {"power 5", 5, 1}} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.B12Push(g, me.ID, "Deathknell Berserker", "Creature — Elf Berserker", oldB070Deathknell, c.power, 2)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
			g.RunStateChecksForTest()
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie Berserker"); n != c.want {
				t.Errorf("%d Zombie Berserker tokens, want %d", n, c.want)
			}
		})
	}
}

func TestOldB070DeathknellBerserkerReadsAPumpedPowerAsItDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Deathknell Berserker", "Creature — Elf Berserker", oldB070Deathknell, 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterPlusOne, 1) })
	g.RunStateChecksForTest()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie Berserker"); n != 1 {
		t.Errorf("%d Zombie Berserker tokens for a 3-power Berserker, want 1", n)
	}
}

// --- Rampage of the Valkyries ----------------------------------------------

func TestOldB070RampageMakesAnAngelAndEdictsWhenAnAngelDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	spare := effectstest.B12Creature(g, opp.ID, "Their Spare", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Rampage of the Valkyries", "Enchantment", oldB070Rampage, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Angel"); n != 1 {
		t.Fatalf("%d Angel tokens, want 1", n)
	}
	var angel uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Angel" && c.Controller == me.ID {
			angel = c.InstanceID
		}
	}
	if pw, tough := effectstest.EffectivePTOf(t, g, angel); pw != 4 || tough != 4 {
		t.Errorf("the Angel is %d/%d, want 4/4", pw, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, angel, "flying") || !effectstest.HasKeywordOnBattlefield(t, g, angel, "vigilance") {
		t.Error("the Angel lacks flying or vigilance")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(angel) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SacrificeChoiceFor(g, me.ID); c != nil {
		t.Fatal("Rampage asked its own controller to sacrifice")
	}
	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the opponent was not asked to sacrifice")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(spare) || !g.Battlefield.Contains(mine) {
		t.Error("only the chosen creature should have been sacrificed")
	}
}

func TestOldB070RampageIgnoresAnOpponentsAngelAndYourNonAngels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.PushCatalogPermanent(g, me.ID, "Rampage of the Valkyries", "Enchantment", oldB070Rampage, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirAngel := effectstest.B12Creature(g, opp.ID, "Their Angel", "Creature — Angel", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirAngel) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SacrificeChoiceFor(g, opp.ID); c != nil {
		t.Error("a non-Angel of mine or an opponent's Angel triggered the edict")
	}
}

// --- Smashing Success ------------------------------------------------------

func TestOldB070SmashingSuccessMakesATreasureOnlyForAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", false)
	effectstest.CastCatalogSpell(t, g, "Smashing Success", "Instant", oldB070Smashing, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("%d Treasures for a land, want 0", n)
	}
	effectstest.CastCatalogSpell(t, g, "Smashing Success", "Instant", oldB070Smashing, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures for an artifact, want 1 (made by the spell's controller)", n)
	}
}

func TestOldB070SmashingSuccessMakesNoTreasureForAnIndestructibleArtifactAndRefusesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	tough := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Darksteel Thing", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"indestructible"}})
	effectstest.CastCatalogSpell(t, g, "Smashing Success", "Instant", oldB070Smashing, effectstest.CardRefs(tough))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tough) {
		t.Fatal("an indestructible artifact was destroyed")
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("%d Treasures although nothing was destroyed, want 0", n)
	}
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Smashing Success", "Instant", oldB070Smashing, effectstest.CardRefs(bear)); err == nil {
		t.Error("targeted a nonartifact creature")
	}
}

// --- Warchanter Skald ------------------------------------------------------

func TestOldB070WarchanterSkaldMakesADwarfWhenTappedWhileEquippedOrEnchanted(t *testing.T) {
	for _, tl := range []string{"Artifact — Equipment", "Enchantment — Aura"} {
		t.Run(tl, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			skald := effectstest.B12Push(g, me.ID, "Warchanter Skald", "Creature — Dwarf Cleric", oldB070Skald, 2, 3)
			oldB070Attach(g, me.ID, skald, tl)
			g.WithWriteLock(func() { _ = g.TapTargetForEffect(skald) })
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CountBattlefieldNamed(g, me.ID, "Dwarf Berserker"); n != 1 {
				t.Errorf("%d Dwarf Berserker tokens, want 1", n)
			}
		})
	}
}

func TestOldB070WarchanterSkaldMakesNothingWhenBareOrWhenTheAttachmentIsGoneAtResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skald := effectstest.B12Push(g, me.ID, "Warchanter Skald", "Creature — Dwarf Cleric", oldB070Skald, 2, 3)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(skald) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Dwarf Berserker"); n != 0 {
		t.Errorf("a bare Skald made %d tokens", n)
	}
	// Equipped when it is tapped, but the Equipment is gone before the trigger resolves.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(skald) })
	gear := oldB070Attach(g, me.ID, skald, "Artifact — Equipment")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(skald) })
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(gear) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Dwarf Berserker"); n != 0 {
		t.Errorf("a Skald whose Equipment left in response made %d tokens, want 0 (intervening if)", n)
	}
}
