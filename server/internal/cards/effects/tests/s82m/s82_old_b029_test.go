package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OLD batch 29 (S82, #5501): Modern Horizons 2 — Battle Plan, Landscaper
// Colos, Sojourner's Companion, Asmoranomardicadaistinaculdacar, Yusri,
// Sterling Grove, Moderation and Fast // Furious.

const (
	oldB029BattlePlan   = "09b4fba6-d943-49e4-9f9e-a74fcbabe83f"
	oldB029Colos        = "144bef8b-3583-4e98-9285-a1965e7e17f8"
	oldB029Companion    = "1cbfe1f2-776b-429d-9a60-8da32058a927"
	oldB029Asmo         = "1f74e26f-5cf3-4b0e-bcc6-f858599474fc"
	oldB029Yusri        = "2c0a0a7d-b12f-47c8-a6a6-d046f999ec6b"
	oldB029SterlingGrov = "2c275a85-5a15-46cd-a6e7-add63f9b853d"
	oldB029Moderation   = "57f94fcb-a224-410b-a38e-3ed8cca7d0c7"
	oldB029FastFurious  = "62411ced-843e-4b63-bdf6-dafb2ac27047"
)

func TestOldB029CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB029BattlePlan:         "Battle Plan",
		oldB029Colos:              "Landscaper Colos",
		oldB029Companion:          "Sojourner's Companion",
		oldB029Asmo:               "Asmoranomardicadaistinaculdacar",
		oldB029Yusri:              "Yusri, Fortune's Flame",
		oldB029SterlingGrov:       "Sterling Grove",
		oldB029Moderation:         "Moderation",
		oldB029FastFurious:        "Fast",
		oldB029FastFurious + "#1": "Furious",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

func oldB029Library(p *game.Player, n int) {
	p.Library.Cards = nil
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

// --- Battle Plan ---------------------------------------------------------

func TestOldB029BattlePlanPumpsATargetCreatureAtBeginningOfCombatOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Battle Plan", "Enchantment", oldB029BattlePlan, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Their Bear", 2, 2)
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.GaLayered(t, g, bear).CurrentPower(); got != 4 {
		t.Errorf("the target has power %d, want 4", got)
	}
	if got := effectstest.GaLayered(t, g, theirs).CurrentPower(); got != 2 {
		t.Errorf("an opposing creature has power %d, want 2", got)
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	if got := effectstest.GaLayered(t, g, bear).CurrentPower(); got != 2 {
		t.Errorf("after the turn the target has power %d, want 2", got)
	}
}

func TestOldB029BattlePlanBasicLandcyclingFindsOnlyABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Battle Plan", "Enchantment", oldB029BattlePlan)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{C}{C}", nil); err == nil {
		t.Fatal("landcycling {1}{R} was paid with colourless mana only")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{R}{C}", nil); err != nil {
		t.Fatalf("landcycling: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || !me.Graveyard.Contains(id) || me.Hand.Contains(ids[0]) {
		t.Errorf("Forest in hand %v, Battle Plan in graveyard %v", me.Hand.Contains(ids[1]), me.Graveyard.Contains(id))
	}
}

// --- Landscaper Colos ------------------------------------------------------

func TestOldB029LandscaperColosPutsAnOpponentsGraveyardCardOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB029Library(opp, 5)
	mine := effectstest.PushGraveyardPermanent(me, "My Card", "Creature — Bear", "{1}")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Card", "Creature — Bear", "{1}")
	effectstest.EnterFromHand(t, g, me.ID, "Landscaper Colos", "Creature — Goat Beast", oldB029Colos)
	if effectstest.LatestPickTarget(g, me.ID) == nil && effectstest.LatestTriggerPromptFor(g, me.ID) == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(theirs) {
		t.Fatal("the card is still in the opponent's graveyard")
	}
	n := len(opp.Library.Cards)
	if n != 6 || opp.Library.Cards[n-1].InstanceID != theirs && opp.Library.Cards[0].InstanceID != theirs {
		t.Errorf("the card is not at an end of the opponent's library (size %d)", n)
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard card was moved")
	}
}

// --- Sojourner's Companion -----------------------------------------------

func oldB029Price(t *testing.T, g *game.Game, caster uuid.UUID, card game.Card) string {
	t.Helper()
	card.Owner, card.Controller = caster, caster
	base, err := game.ParseCost(card.ManaCost)
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	var out game.ParsedCost
	g.WithWriteLock(func() {
		var err error
		out, err = g.ApplyCostModifiersForEffect(base, game.CostQuery{Game: g, Card: card, Controller: caster, FromZone: game.ZoneHand})
		if err != nil {
			t.Fatalf("ApplyCostModifiers: %v", err)
		}
	})
	return out.String()
}

func TestOldB029SojournersCompanionHasAffinityForArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	companion := game.Card{InstanceID: uuid.New(), Name: "Sojourner's Companion", TypeLine: "Artifact Creature — Salamander", ManaCost: "{7}", OracleID: oldB029Companion}
	if got := oldB029Price(t, g, me.ID, companion); got != "{7}" {
		t.Fatalf("with no artifacts it costs %q, want {7}", got)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Other Rock", "Artifact Creature — Golem", "", false)
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Their Rock", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	if got := oldB029Price(t, g, me.ID, companion); got != "{5}" {
		t.Errorf("with two artifacts of mine it costs %q, want {5}", got)
	}
}

func TestOldB029SojournersCompanionArtifactLandcyclingFindsAnArtifactLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Sojourner's Companion", "Artifact Creature — Salamander", oldB029Companion)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
		game.Card{Name: "Darksteel Citadel", TypeLine: "Artifact Land"},
	)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("artifact landcycling: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[2]) || !me.Graveyard.Contains(id) {
		t.Errorf("artifact land in hand %v, Companion in graveyard %v", me.Hand.Contains(ids[2]), me.Graveyard.Contains(id))
	}
}

// --- Sterling Grove ----------------------------------------------------

func TestOldB029SterlingGroveGivesShroudToYourOtherEnchantmentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	grove := effectstest.PushCatalogPermanent(g, me.ID, "Sterling Grove", "Enchantment", oldB029SterlingGrov, false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Enchantment", "Enchantment", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Enchantment", "Enchantment", "", false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{{"Sterling Grove", grove, false}, {"my other enchantment", mine, true}, {"their enchantment", theirs, false}, {"my creature", bear, false}} {
		if got := effectstest.HobB10Has(effectstest.GaLayered(t, g, tc.id), "shroud"); got != tc.want {
			t.Errorf("%s has shroud: %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestOldB029SterlingGroveSacrificesToTutorAnEnchantmentToTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	grove := effectstest.PushCatalogPermanent(g, me.ID, "Sterling Grove", "Enchantment", oldB029SterlingGrov, false)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Pacifism", TypeLine: "Enchantment — Aura"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
	)
	if err := effectstest.HobB4Activate(t, g, me.ID, grove, 0, "{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, grove); ok {
		t.Fatal("Sterling Grove was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != ids[1] && me.Library.Cards[0].InstanceID != ids[1] {
		t.Error("the enchantment is not on top of the library")
	}
	if me.Hand.Contains(ids[1]) {
		t.Error("the enchantment went to hand, not the top of the library")
	}
}

// --- Moderation ----------------------------------------------------------

func TestOldB029ModerationLimitsOnlyItsControllerToOneSpellAndDrawsPerSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Moderation", "Enchantment", oldB029Moderation, false)
	oldB029Library(me, 10)
	hand := me.Hand.Size()
	first, err := effectstest.CastFromHandErr(g, me, "Spell One", "Instant", "", "", game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("the first spell: %v", err)
	}
	_ = first
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("casting a spell changed the hand by %d, want +1 (the draw)", got)
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Spell Two", "Instant", "", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a second spell was cast in the same turn")
	}
	card := game.Card{InstanceID: uuid.New(), Name: "Their Spell", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID}
	if err := effectstest.CastGateFor(g, opp.ID, card, game.ZoneHand); err != nil {
		t.Errorf("an opponent is limited by Moderation: %v", err)
	}
}

// --- Fast // Furious -------------------------------------------------------

func TestOldB029FastDiscardsThenDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB029Library(me, 10)
	me.Hand.Cards = nil
	keep := effectstest.HandCardFull(me, "Held", "Sorcery", "{9}", "", nil)
	effectstest.CastCatalogSpell(t, g, "Fast", "Instant", oldB029FastFurious, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, keep)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(keep) {
		t.Error("the held card was not discarded")
	}
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand is %d cards, want the two drawn", got)
	}
}

func TestOldB029FuriousDamagesEachCreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 4, 4)
	bird := effectstest.B30Flyer(g, opp.ID, "Bird", "Creature — Bird", 2, 2)
	cub := effectstest.PushVanillaCreature(g, opp.ID, "Cub", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Furious", "Sorcery", oldB029FastFurious+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, cub); ok {
		t.Error("the 3/3 without flying survived")
	}
	if _, ok := effectstest.BattlefieldCard(g, bird); !ok {
		t.Error("the flier was damaged")
	}
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || c.DamageMarked != 3 {
		t.Errorf("the 4/4 took %d damage, want 3", c.DamageMarked)
	}
}

// --- Asmoranomardicadaistinaculdacar -------------------------------------

func TestOldB029AsmoranomardicadaistinaculdacarCastsOnlyAfterADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Asmoranomardicadaistinaculdacar", TypeLine: "Legendary Creature — Human Wizard",
		Layout: "normal", OracleID: oldB029Asmo, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "B")
	params := game.CastSpellParams{Strict: true, AlternativeCost: "asmoranomardicadaistinaculdacar-discarded"}
	if err := g.CastSpell(me.ID, id, params); err == nil {
		t.Fatal("cast with no discard this turn")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for no mana cost at all")
	}
	effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	effectstest.B30DiscardOne(t, g, me.ID)
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast after discarding, paying {B/R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, id); !ok {
		t.Error("it did not resolve onto the battlefield")
	}
}

func TestOldB029AsmoranomardicadaistinaculdacarFetchesTheCookbookOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "The Underworld Cookbook", TypeLine: "Legendary Artifact"},
		game.Card{Name: "Other", TypeLine: "Artifact"},
	)
	effectstest.EnterFromHand(t, g, me.ID, "Asmoranomardicadaistinaculdacar", "Legendary Creature — Human Wizard", oldB029Asmo)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) || me.Hand.Contains(ids[1]) {
		t.Errorf("Cookbook in hand %v, other card in hand %v", me.Hand.Contains(ids[0]), me.Hand.Contains(ids[1]))
	}
}

func TestOldB029AsmoranomardicadaistinaculdacarSacrificesTwoFoodsForSixToItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	asmo := effectstest.B12Push(g, me.ID, "Asmoranomardicadaistinaculdacar", "Legendary Creature — Human Wizard", oldB029Asmo, 3, 3)
	food1 := effectstest.PushCatalogPermanent(g, me.ID, "Food", "Artifact — Food", "", false)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Giant", 2, 6)
	targets := effectstest.CardTargets(victim)
	activate := func(foods ...uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, asmo, 0, game.ActivateAbilityParams{Targets: targets, SacrificeIDs: foods, Strict: true})
	}
	if err := activate(food1); err == nil {
		t.Fatal("activated with only one Food")
	}
	food2 := effectstest.PushCatalogPermanent(g, me.ID, "Food", "Artifact — Food", "", false)
	if err := activate(food1, food2); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, food1); ok {
		t.Error("a Food was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the 2/6 survived dealing 6 damage to itself")
	}
}

// --- Yusri, Fortune's Flame ----------------------------------------------

// oldB029YusriAttack plays one attack of Yusri with `n` chosen and every
// coin called heads, under the RNG key `key`. It returns the coins won and
// lost, the cards drawn, the life lost and whether a spell paid for with
// nothing can be cast afterwards.
func oldB029YusriAttack(t *testing.T, key byte, n int) (won, lost, drawn, life int, freeCast bool) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	var k [32]byte
	k[0] = key
	g.SetRNGKeyForTest(k)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB029Library(me, 30)
	yusri := effectstest.B12Push(g, me.ID, "Yusri, Fortune's Flame", "Legendary Creature — Efreet", oldB029Yusri, 2, 3)
	effectstest.DeclareAttack(t, g, opp.ID, yusri)
	effectstest.LockInAttacks(t, g)
	hand, startLife := me.Hand.Size(), me.Life
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PayAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no number prompt")
	}
	if pa := c.PayAmount; pa.Min != 1 || pa.Max != 5 {
		t.Fatalf("number prompt %+v, want 1 through 5", *pa)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, n); err != nil {
		t.Fatalf("choose %d: %v", n, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	flip := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCoinCall, me.ID)
	if flip == nil || flip.CoinCount != n {
		t.Fatalf("coin call %+v, want %d coins", flip, n)
	}
	if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
		t.Fatalf("call: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, ev := range g.Events {
		if ev.Kind == game.EventFlipCoin && ev.Actor == me.ID {
			if ev.Won {
				won++
			} else {
				lost++
			}
		}
	}
	drawn, life = me.Hand.Size()-hand, startLife-me.Life
	spell := effectstest.HandCardFull(me, "Pricey Instant", "Instant", "{3}{U}{U}", "", nil)
	// The permission is an alternative cost: it is claimed under the shared
	// granted-cost key, and a plain cast still pays the printed cost.
	if g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}) == nil {
		t.Fatal("a cast that claimed nothing paid no mana")
	}
	freeCast = g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}) == nil
	return
}

func TestOldB029YusriDrawsForWinsAndTakesTwoForLosses(t *testing.T) {
	var sawLoss, sawWin bool
	for key := byte(1); key < 40 && !(sawLoss && sawWin); key++ {
		won, lost, drawn, life, free := oldB029YusriAttack(t, key, 3)
		if won+lost != 3 {
			t.Fatalf("key %d: %d flips, want 3", key, won+lost)
		}
		if drawn != won || life != 2*lost {
			t.Errorf("key %d: won %d lost %d but drew %d and lost %d life", key, won, lost, drawn, life)
		}
		if free {
			t.Errorf("key %d: three flips opened free casting", key)
		}
		sawLoss, sawWin = sawLoss || lost > 0, sawWin || won > 0
	}
	if !sawLoss || !sawWin {
		t.Fatalf("never saw both a win and a loss (win %v, loss %v)", sawWin, sawLoss)
	}
}

func TestOldB029YusriFiveWinsLetYouCastFromHandForFree(t *testing.T) {
	for key := 1; key < 700; key++ {
		won, lost, drawn, life, free := oldB029YusriAttack(t, byte(key), 5)
		if won != 5 {
			if free {
				t.Fatalf("key %d: won %d and could cast for free", key, won)
			}
			continue
		}
		if lost != 0 || drawn != 5 || life != 0 || !free {
			t.Fatalf("five wins: lost %d, drew %d, life lost %d, free cast %v", lost, drawn, life, free)
		}
		return
	}
	t.Skip("no RNG key in range won five flips")
}
