package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b232_cards_test.go — OLD batch 232 (Commander 2017, #5533): eight
// cards. Each test plays the real catalog entry through the engine.

const (
	oldB232SorcererOracle   = "01afb7a1-6427-4bb6-935a-613e9bda9de3"
	oldB232DragonOracle     = "86b8a9ef-7cdf-47e9-accb-16fe1eab718f"
	oldB232NivOracle        = "899d58dc-60b2-4013-9f91-c1888c585d66"
	oldB232NivixOracle      = "d04356f1-0e1a-4689-8e54-f88c4c6dd936"
	oldB232TajNarOracle     = "dcbf8d72-3367-43e1-9a71-e674e0ed52d9"
	oldB232PilgrimageOracle = "e83236f8-1c45-4d31-8937-1c76db549955"
	oldB232NinOracle        = "ec439630-e18b-45f6-8034-2ecbea27772f"
	oldB232OratorOracle     = "fbed6f1a-e224-413c-b13e-3555caaef02a"
)

func TestOldB232EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB232SorcererOracle:   "Etherium-Horn Sorcerer",
		oldB232DragonOracle:     "Spellbound Dragon",
		oldB232NivOracle:        "Niv-Mizzet, Dracogenius",
		oldB232NivixOracle:      "Nivix Guildmage",
		oldB232TajNarOracle:     "Taj-Nar Swordsmith",
		oldB232PilgrimageOracle: "Nissa's Pilgrimage",
		oldB232NinOracle:        "Nin, the Pain Artist",
		oldB232OratorOracle:     "Orator of Ojutai",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB232Zone(g *game.Game, id uuid.UUID, kind game.ZoneKind) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == kind
}

// --- Etherium-Horn Sorcerer -------------------------------------------------

func TestOldB232EtheriumHornSorcererCascades(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	elves := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: elves, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Etherium-Horn Sorcerer", TypeLine: "Artifact Creature — Minotaur Wizard Sorcerer", OracleID: oldB232SorcererOracle,
		ManaCost: "{4}{U}{R}", Power: 3, Toughness: 6, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectstest.TriggersOnStackFrom(g, id) != 1 {
		t.Fatal("the cascade trigger is not on the stack above the spell")
	}
	for i := 0; i < 6 && oldB232Zone(g, elves, game.ZoneLibrary); i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if oldB232Zone(g, elves, game.ZoneLibrary) {
		t.Error("cascade left the Elves on the library")
	}
}

func TestOldB232EtheriumHornSorcererReturnsToHandForThreeMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sorcerer := effectstest.DftABPush(g, me.ID, "Etherium-Horn Sorcerer", "Artifact Creature — Minotaur Wizard Sorcerer", oldB232SorcererOracle, 3, 6)
	effectstest.AddPool(t, g, me.ID, "{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, sorcerer, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with two mana")
	}
	effectstest.AddPool(t, g, me.ID, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, sorcerer, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("bounce: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, sorcerer) {
		t.Error("the Sorcerer is not back in hand")
	}
}

// --- Spellbound Dragon ------------------------------------------------------

func TestOldB232SpellboundDragonLootsAndGetsTheDiscardedManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dragon := effectstest.DftABPush(g, me.ID, "Spellbound Dragon", "Creature — Dragon", oldB232DragonOracle, 3, 5)
	me.Hand.Cards = me.Hand.Cards[:0]
	three := effectstest.HandCardFull(me, "Three Drop", "Creature — Bear", "{2}{G}", "", []string{"G"})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Drawn Land", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.DeclareAttack(t, g, opp.ID, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, three)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB232Zone(g, three, game.ZoneGraveyard) {
		t.Fatal("the card was not discarded")
	}
	if p := effectstest.EffectivePower(t, g, dragon); p != 6 {
		t.Errorf("power %d, want 3 + 3 = 6", p)
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	if p := effectstest.EffectivePower(t, g, dragon); p != 3 {
		t.Errorf("after the turn the power is %d, want 3", p)
	}
}

func TestOldB232SpellboundDragonDiscardingALandAddsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dragon := effectstest.DftABPush(g, me.ID, "Spellbound Dragon", "Creature — Dragon", oldB232DragonOracle, 3, 5)
	me.Hand.Cards = me.Hand.Cards[:0]
	land := effectstest.HandCardFull(me, "A Forest", "Basic Land — Forest", "", "", nil)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Drawn Land", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.DeclareAttack(t, g, opp.ID, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, dragon); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
}

// --- Niv-Mizzet, Dracogenius ------------------------------------------------

func TestOldB232NivMizzetPingsAndMayDrawWhenItHitsAPlayer(t *testing.T) {
	for _, draw := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		niv := effectstest.DftABPush(g, me.ID, "Niv-Mizzet, Dracogenius", "Legendary Creature — Dragon Wizard", oldB232NivOracle, 5, 5)
		bear := effectstest.DftABPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
		effectstest.AddPool(t, g, me.ID, "{U}{R}{U}{R}")
		// A ping at a creature does not trigger the draw.
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, niv, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
			t.Fatalf("ping the Bear: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != hand {
			t.Errorf("a ping at a creature drew a card")
		}
		life := opp.Life
		if err := g.ActivateCatalogAbility(me.ID, niv, 0, game.ActivateAbilityParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
			t.Fatalf("ping the opponent: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, draw)
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != life-1 {
			t.Errorf("opponent life %d → %d, want -1", life, opp.Life)
		}
		want := hand
		if draw {
			want++
		}
		if me.Hand.Size() != want {
			t.Errorf("draw=%v: hand %d → %d, want %d", draw, hand, me.Hand.Size(), want)
		}
	}
}

// --- Nivix Guildmage --------------------------------------------------------

func TestOldB232NivixGuildmageLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gm := effectstest.DftABPush(g, me.ID, "Nivix Guildmage", "Creature — Human Wizard", oldB232NivixOracle, 2, 2)
	effectstest.AddPool(t, g, me.ID, "{C}{U}{R}")
	gy := me.Graveyard.Size()
	if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("loot: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrDiscard(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != gy+1 {
		t.Errorf("graveyard %d → %d, want one discarded", gy, me.Graveyard.Size())
	}
}

func TestOldB232NivixGuildmageCopiesYourSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gm := effectstest.DftABPush(g, me.ID, "Nivix Guildmage", "Creature — Human Wizard", oldB232NivixOracle, 2, 2)
	bolt := effectstest.CastCopyTarget(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	life := opp.Life
	effectstest.AddPool(t, g, me.ID, "{C}{C}{U}{R}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bolt)}); err != nil {
		t.Fatalf("copy the Bolt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.CopyPromptFor(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 6 {
		t.Errorf("the opponent lost %d, want 6 (the spell and its copy)", got)
	}
}

// --- Taj-Nar Swordsmith -----------------------------------------------------

// oldB232Pay answers the payment prompt of a "you may pay {X}" if one is open.
func oldB232Pay(t *testing.T, g *game.Game, me *game.Player, pay bool) {
	t.Helper()
	if p := effectstest.PayUnlessFor(g, me.ID); p != nil {
		if err := g.ResolvePayUnless(p.ID, me.ID, pay); err != nil {
			t.Fatalf("pay: %v", err)
		}
	}
}

func TestOldB232TajNarSwordsmithFindsAnEquipmentForX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cheap := uuid.New()
	pricey := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: pricey, Name: "Big Sword", TypeLine: "Artifact — Equipment", ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: cheap, Name: "Small Sword", TypeLine: "Artifact — Equipment", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}")
	effectstest.DftABCast(t, g, me, game.Card{Name: "Taj-Nar Swordsmith", TypeLine: "Creature — Cat Soldier", OracleID: oldB232TajNarOracle,
		Power: 2, Toughness: 3})
	effectstest.AnswerNumber(t, g, me.ID, 3)
	oldB232Pay(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		if effectstest.HasID(c.SearchCards, pricey) {
			t.Error("an Equipment with mana value 4 is offered for X = 3")
		}
		effectstest.AnswerSearchNamed(t, g, me.ID, "Small Sword")
		effectstest.PassPriorityAroundTable(t, g)
	} else {
		t.Fatal("no search prompt after paying X")
	}
	if !oldB232Zone(g, cheap, game.ZoneBattlefield) {
		t.Error("the Equipment did not enter the battlefield")
	}
	if oldB232Zone(g, pricey, game.ZoneBattlefield) {
		t.Error("the expensive Equipment entered the battlefield")
	}
}

func TestOldB232TajNarSwordsmithPayingNothingSearchesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sword := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: sword, Name: "Small Sword", TypeLine: "Artifact — Equipment", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}")
	effectstest.DftABCast(t, g, me, game.Card{Name: "Taj-Nar Swordsmith", TypeLine: "Creature — Cat Soldier", OracleID: oldB232TajNarOracle,
		Power: 2, Toughness: 3})
	effectstest.AnswerNumber(t, g, me.ID, 0)
	oldB232Pay(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil && effectstest.HasID(c.SearchCards, sword) {
		// An Equipment of mana value 2 is not legal for X = 0.
		t.Error("an Equipment with mana value 2 is offered for X = 0")
	}
	if oldB232Zone(g, sword, game.ZoneBattlefield) {
		t.Error("an Equipment entered the battlefield for X = 0")
	}
}

// --- Nissa's Pilgrimage -----------------------------------------------------

func oldB232Forests(me *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		ids = append(ids, id)
	}
	return ids
}

func oldB232CastPilgrimage(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Nissa's Pilgrimage", TypeLine: "Sorcery", OracleID: oldB232PilgrimageOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for i := 0; i < 6; i++ {
		if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
			n := min(max(c.SearchMax, 1), len(c.SearchCards))
			if err := g.ResolveSearchLibrary(c.ID, me.ID, c.SearchCards[:n]); err != nil {
				t.Fatalf("search: %v", err)
			}
			continue
		}
		if g.Stack.Size() == 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

func TestOldB232NissasPilgrimageFindsTwoForestsOneOntoTheBattlefieldTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	forests := oldB232Forests(me, 4)
	hand := me.Hand.Size()
	oldB232CastPilgrimage(t, g, me)
	onBf, inHand := 0, 0
	for _, id := range forests {
		switch {
		case oldB232Zone(g, id, game.ZoneBattlefield):
			onBf++
			if tapped, _ := effectstest.BattlefieldCardTapped(g, id); !tapped {
				t.Error("the Forest entered untapped")
			}
		case effectstest.HandHasID(me, id):
			inHand++
		}
	}
	if onBf != 1 || inHand != 1 {
		t.Errorf("%d on the battlefield and %d in hand, want 1 and 1", onBf, inHand)
	}
	_ = hand
}

func TestOldB232NissasPilgrimageFindsThreeWithSpellMastery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushGraveyardPermanent(me, "Old Bolt", "Instant", "{R}")
	effectstest.PushGraveyardPermanent(me, "Old Wrath", "Sorcery", "{2}{W}{W}")
	forests := oldB232Forests(me, 5)
	oldB232CastPilgrimage(t, g, me)
	onBf, inHand := 0, 0
	for _, id := range forests {
		switch {
		case oldB232Zone(g, id, game.ZoneBattlefield):
			onBf++
		case effectstest.HandHasID(me, id):
			inHand++
		}
	}
	if onBf != 1 || inHand != 2 {
		t.Errorf("%d on the battlefield and %d in hand, want 1 and 2", onBf, inHand)
	}
}

// --- Nin, the Pain Artist ---------------------------------------------------

func TestOldB232NinDealsXAndTheCreaturesControllerDrawsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	nin := effectstest.DftABPush(g, me.ID, "Nin, the Pain Artist", "Legendary Creature — Vedalken Wizard", oldB232NinOracle, 1, 1)
	elk := effectstest.DftABPush(g, opp.ID, "Elk", "Creature — Elk", "", 3, 3)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{U}{R}")
	hand, mine := opp.Hand.Size(), me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, nin, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(elk)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+2 || me.Hand.Size() != mine {
		t.Errorf("opponent hand %d → %d (want +2), mine %d → %d (want unchanged)", hand, opp.Hand.Size(), mine, me.Hand.Size())
	}
	if d := effectstest.Pr6Marked(g, elk); d != 2 {
		t.Errorf("the Elk has %d damage, want 2", d)
	}
}

func TestOldB232NinStillDrawsWhenTheCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	nin := effectstest.DftABPush(g, me.ID, "Nin, the Pain Artist", "Legendary Creature — Vedalken Wizard", oldB232NinOracle, 1, 1)
	elk := effectstest.DftABPush(g, opp.ID, "Elk", "Creature — Elk", "", 2, 2)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{U}{R}")
	hand := opp.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, nin, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.CardRefs(elk)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if opp.Hand.Size() != hand+3 {
		t.Errorf("opponent hand %d → %d, want +3 after the Elk died", hand, opp.Hand.Size())
	}
	if oldB232Zone(g, elk, game.ZoneBattlefield) {
		t.Error("the Elk survived 3 damage")
	}
}

// --- Orator of Ojutai -------------------------------------------------------

func TestOldB232OratorDrawsOnlyWithADragonRevealedOrControlled(t *testing.T) {
	for _, tc := range []string{"nothing", "controlled", "revealed", "controlled but it left in response"} {
		t.Run(tc, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			params := game.CastSpellParams{}
			var dragonOnBf uuid.UUID
			switch tc {
			case "controlled", "controlled but it left in response":
				dragonOnBf = effectstest.B12Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4)
			case "revealed":
				dragon := effectstest.PutInHand(me, game.Card{Name: "Hand Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4})
				params = game.CastSpellParams{OptionalCosts: []int{0}, RevealIDs: []uuid.UUID{dragon}}
			}
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Orator of Ojutai", TypeLine: "Creature — Bird Monk", OracleID: oldB232OratorOracle,
				ManaCost: "{1}{W}", Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID})
			hand := me.Hand.Size() - 1 // the Orator itself leaves the hand
			if err := g.CastSpell(me.ID, id, params); err != nil {
				t.Fatalf("cast: %v", err)
			}
			if dragonOnBf != uuid.Nil && tc != "controlled" {
				effectstest.KillCreature(t, g, me.ID, dragonOnBf)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want := hand
			if tc != "nothing" {
				want++
			}
			if me.Hand.Size() != want {
				t.Errorf("%s: hand %d, want %d", tc, me.Hand.Size(), want)
			}
		})
	}
}

func TestOldB232OratorPutOntoTheBattlefieldDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4)
	hand := me.Hand.Size()
	effectstest.DftABPush(g, me.ID, "Orator of Ojutai", "Creature — Bird Monk", oldB232OratorOracle, 0, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d: a creature that was not cast drew a card", hand, me.Hand.Size())
	}
}
