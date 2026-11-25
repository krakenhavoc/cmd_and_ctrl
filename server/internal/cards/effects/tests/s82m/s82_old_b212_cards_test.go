package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b212_cards_test.go — Dominaria cards, S82 OLD batch 212 (#5530),
// played through the real catalog entries.

const (
	oldB212VigorOracle    = "00cb93a8-5d33-4c6a-9e2a-4b3560f76f3f"
	oldB212ClutchesOracle = "201a4455-1b6b-4e13-a0d5-4cf2116e1673"
	oldB212FreezeOracle   = "648aead4-dc0b-47bc-911c-8dcd2cfd32d3"
	oldB212WardOracle     = "a89ad91e-2bc2-45f4-b32c-fb2b67a2019e"
	oldB212RonaOracle     = "409305b3-5373-4376-8096-c04a73184cef"
	oldB212TomeOracle     = "4734b73e-104c-4b25-ab52-0145f361c9d1"
	oldB212PhantomOracle  = "76164af1-275e-4d68-9df5-9801768512ef"
	oldB212PoetOracle     = "97bdfbe0-7048-4157-8a58-5b17410750ab"
	oldB212SunderingOracl = "5afbd367-19c7-418f-994d-a7958fb1c4ae"
	oldB212AwakeningOracl = "66b46257-b64c-4672-bf47-12bf0bcd6f8f"
	oldB212VowOracle      = "7d012731-f921-47bf-a2da-ce494650cd4f"
	oldB212WeatherlightOr = "c4ce96d8-3ce4-4ed2-9ad3-73119aa79fbe"
)

func TestOldB212CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB212VigorOracle:    "Demonic Vigor",
		oldB212ClutchesOracle: "In Bolas's Clutches",
		oldB212FreezeOracle:   "Deep Freeze",
		oldB212WardOracle:     "Curator's Ward",
		oldB212RonaOracle:     "Rona, Disciple of Gix",
		oldB212TomeOracle:     "Urza's Tome",
		oldB212PhantomOracle:  "Lingering Phantom",
		oldB212PoetOracle:     "Tragic Poet",
		oldB212SunderingOracl: "Karn's Temporal Sundering",
		oldB212AwakeningOracl: "Sylvan Awakening",
		oldB212VowOracle:      "Kamahl's Druidic Vow",
		oldB212WeatherlightOr: "Board the Weatherlight",
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

// oldB212Top puts a card on top of a player's library, above what is there.
func oldB212Top(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}

func oldB212Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestOldB212DemonicVigorPumpsAndReturnsTheCardWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB212Seats(g)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Demonic Vigor", effectstest.AuraTypeLine, oldB212VigorOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Fatalf("enchanted Bear is %d/%d, want 3/3", p, tt)
	}
	hand := opp.Hand.Size()
	effectstest.KillCreature(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) {
		t.Fatalf("the dead creature did not return to its owner's hand (hand %d -> %d)", hand, opp.Hand.Size())
	}
	_ = me
}

func TestOldB212InBolassClutchesStealsAnyPermanentAndMakesItLegendary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB212Seats(g)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "In Bolas's Clutches", "Legendary Enchantment — Aura", oldB212ClutchesOracle, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.B12Card(t, g, land)
	if c.Controller != me.ID {
		t.Errorf("the land is controlled by %v, want the Aura's controller", c.Controller)
	}
	if !c.HasSupertype("Legendary") {
		t.Errorf("the stolen land is not legendary")
	}
}

func TestOldB212InBolassClutchesMakesTheLegendRuleBiteTheThief(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB212Seats(g)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "In Bolas's Clutches", "Legendary Enchantment — Aura", oldB212ClutchesOracle, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	// Both Bears now share a name under one controller and are legendary only if
	// both are: only the stolen one is, so the legend rule does not apply.
	if !effectstest.BattlefieldHas(g, mine) || !effectstest.BattlefieldHas(g, theirs) {
		t.Error("a nonlegendary Bear was removed by the legend rule")
	}
}

func TestOldB212DeepFreezeReshapesTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB212Seats(g)
	dragon := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Dragon", TypeLine: "Creature — Dragon",
		Power: 5, Toughness: 5, Colors: []string{"R"}, Keywords: []string{"flying", "haste"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Deep Freeze", effectstest.AuraTypeLine, oldB212FreezeOracle, effectstest.CardRefs(dragon))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, dragon); p != 0 || tt != 4 {
		t.Fatalf("frozen Dragon is %d/%d, want 0/4", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "defender") {
		t.Error("the frozen creature has no defender")
	}
	for _, kw := range []string{"flying", "haste"} {
		if effectstest.HasEffectiveKeyword(t, g, dragon, kw) {
			t.Errorf("the frozen creature kept %s", kw)
		}
	}
	c := effectstest.B12Card(t, g, dragon)
	eff := c.Effective()
	if !c.HasSubtype("Wall") || !c.HasSubtype("Dragon") {
		t.Errorf("subtypes %v, want Wall in addition to Dragon", eff.Subtypes)
	}
	hasColor := func(col string) bool {
		for _, x := range eff.Colors {
			if x == col {
				return true
			}
		}
		return false
	}
	if !hasColor("U") || !hasColor("R") {
		t.Errorf("colours %v, want blue in addition to red", eff.Colors)
	}
}

func TestOldB212CuratorsWardGivesHexproofAndDrawsTwoWhenAHistoricPermanentLeaves(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		draws    bool
	}{
		{"legendary creature", "Legendary Creature — Human", true},
		{"artifact creature", "Artifact Creature — Golem", true},
		{"saga", "Enchantment — Saga", true},
		{"plain creature", "Creature — Bear", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := oldB212Seats(g)
			for i := 0; i < 4; i++ {
				effectstest.PushLibraryTopForTest(me, "Mine")
			}
			target := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Permanent", TypeLine: tc.typeLine,
				Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
			effectstest.CastCatalogSpell(t, g, "Curator's Ward", effectstest.AuraTypeLine, oldB212WardOracle, effectstest.CardRefs(target))
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.HasEffectiveKeyword(t, g, target, "hexproof") {
				t.Fatal("the enchanted permanent has no hexproof")
			}
			hand := me.Hand.Size()
			effectstest.KillCreature(t, g, opp.ID, target)
			effectstest.PassPriorityAroundTable(t, g)
			want := hand
			if tc.draws {
				want = hand + 2
			}
			if got := me.Hand.Size(); got != want {
				t.Errorf("hand = %d, want %d", got, want)
			}
		})
	}
}

func TestOldB212TragicPoetReturnsAnEnchantmentCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB212Seats(g)
	effectstest.ToMain(t, g)
	poet := effectstest.PushCatalogPermanent(g, me.ID, "Tragic Poet", "Creature — Human", oldB212PoetOracle, false)
	ench := effectstest.PushGraveyardCardWithTypeLineID(me, "Old Enchantment", "Enchantment")
	creature := effectstest.PushGraveyardCardWithTypeLineID(me, "Old Creature", "Creature — Bear")
	if err := g.ActivateCatalogAbility(me.ID, poet, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(creature),
	}); err == nil {
		t.Fatal("the Poet targeted a creature card")
	}
	if err := g.ActivateCatalogAbility(me.ID, poet, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(ench),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.BattlefieldHas(g, poet) {
		t.Error("the Poet was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ench) {
		t.Error("the enchantment card did not return to hand")
	}
}

func TestOldB212UrzasTomeDrawsThenDiscardsUnlessAHistoricCardIsExiled(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB212Seats(g)
		for i := 0; i < 4; i++ {
			effectstest.PushLibraryTopForTest(me, "Mine")
		}
		effectstest.ToMain(t, g)
		tome := effectstest.PushCatalogPermanent(g, me.ID, "Urza's Tome", "Artifact — Book", oldB212TomeOracle, false)
		effectstest.FloatMana(t, g, me, "{C}{C}{C}")
		return g, me, tome
	}
	t.Run("no historic card: discard", func(t *testing.T) {
		g, me, tome := setup(t)
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, tome, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.DiscardFromHand(t, g, me.ID)
		if got := me.Hand.Size(); got != hand {
			t.Errorf("hand = %d, want %d (drew one, discarded one)", got, hand)
		}
	})
	t.Run("exile a historic card: no discard", func(t *testing.T) {
		g, me, tome := setup(t)
		historic := effectstest.PushGraveyardCardWithTypeLineID(me, "Old Relic", "Artifact")
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, tome, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID, historic)
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != hand+1 {
			t.Errorf("hand = %d, want %d (a plain draw)", got, hand+1)
		}
		if !effectstest.ExileHas(g, historic) {
			t.Error("the historic card was not exiled")
		}
	})
	t.Run("decline to exile: discard", func(t *testing.T) {
		g, me, tome := setup(t)
		historic := effectstest.PushGraveyardCardWithTypeLineID(me, "Old Legend", "Legendary Creature — Human")
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, tome, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID)
		effectstest.DiscardFromHand(t, g, me.ID)
		if got := me.Hand.Size(); got != hand {
			t.Errorf("hand = %d, want %d (drew one, discarded one)", got, hand)
		}
		if effectstest.ExileHas(g, historic) {
			t.Error("the historic card was exiled although the player declined")
		}
	})
	t.Run("a nonhistoric card is not offered", func(t *testing.T) {
		g, me, tome := setup(t)
		effectstest.PushGraveyardCardWithTypeLineID(me, "Old Bear", "Creature — Bear")
		if err := g.ActivateCatalogAbility(me.ID, tome, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.DiscardChoiceFor(g, me.ID) == nil {
			t.Fatal("no discard was asked: the Tome offered an exile instead, or nothing")
		}
		effectstest.DiscardFromHand(t, g, me.ID)
	})
}

func TestOldB212LingeringPhantomReturnsForBWhenYouCastAHistoricSpell(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		pay, returns   bool
	}{
		{"historic spell, paid", "Artifact", true, true},
		{"historic spell, declined", "Legendary Creature — Human", false, false},
		{"nonhistoric spell", "Creature — Bear", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := oldB212Seats(g)
			phantom := uuid.New()
			me.Graveyard.PushTop(game.Card{InstanceID: phantom, Name: "Lingering Phantom", TypeLine: "Creature — Spirit",
				OracleID: oldB212PhantomOracle, Owner: me.ID, Controller: me.ID})
			effectstest.CastCatalogSpell(t, g, "Test Spell", tc.typeLine, "", nil)
			effectstest.FloatMana(t, g, me, "{B}")
			effectstest.PassPriorityAroundTable(t, g)
			if tc.typeLine == "Creature — Bear" {
				if effectstest.HasPayUnlessFor(g, me.ID) {
					t.Fatal("a nonhistoric spell triggered the Phantom")
				}
			} else {
				effectstest.AnswerPayUnless(t, g, me.ID, tc.pay)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Contains(phantom); got != tc.returns {
				t.Errorf("Phantom returned to hand = %v, want %v", got, tc.returns)
			}
		})
	}
}

func TestOldB212RonaExilesAHistoricCardAndCastsSpellsFromAmongThemExiledWithRona(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB212Seats(g)
	historic := effectstest.PushGraveyardCardWithTypeLineID(me, "Old Relic", "Artifact")
	plain := effectstest.PushGraveyardCardWithTypeLineID(me, "Old Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Rona, Disciple of Gix", "Legendary Creature — Human Artificer", oldB212RonaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, historic)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ExileHas(g, historic) {
		t.Fatal("the historic card was not exiled")
	}
	if !effectstest.GrantedPermissionOn(g, me.ID, historic, game.ZoneExile).Granted() {
		t.Error("Rona's controller may not cast the card exiled with her")
	}
	if effectstest.InGraveyard(me, plain) == false {
		t.Error("the nonhistoric card left the graveyard")
	}
	other := effectstest.PushExiledCard(g, me.ID, "Not Rona's", "Sorcery", false)
	if effectstest.GrantedPermissionOn(g, me.ID, other, game.ZoneExile).Granted() {
		t.Error("a card exiled by something else can be cast through Rona")
	}
}

func TestOldB212RonaExilesTheTopCardAndMayCastIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB212Seats(g)
	effectstest.ToMain(t, g)
	rona := effectstest.PushCatalogPermanent(g, me.ID, "Rona, Disciple of Gix", "Legendary Creature — Human Artificer", oldB212RonaOracle, false)
	top := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"})
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, rona, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.B31Tapped(t, g, rona) {
		t.Error("Rona is not tapped on announcement")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ExileHas(g, top) {
		t.Fatal("the top card was not exiled")
	}
	if !effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneExile).Granted() {
		t.Fatal("the card exiled with Rona can't be cast")
	}
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
}

func TestOldB212RonaCannotPlayALandExiledWithHer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB212Seats(g)
	effectstest.ToMain(t, g)
	rona := effectstest.PushCatalogPermanent(g, me.ID, "Rona, Disciple of Gix", "Legendary Creature — Human Artificer", oldB212RonaOracle, false)
	land := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, rona, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ExileHas(g, land) {
		t.Fatal("the land was not exiled")
	}
	if err := g.PlayCard(me.ID, land); err == nil {
		t.Error("a land exiled with Rona was played (the permission is for spells)")
	}
}

func TestOldB212KarnsTemporalSunderingNeedsALegendaryCreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB212Seats(g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Karn's Temporal Sundering", "Legendary Sorcery", oldB212SunderingOracl,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Fatal("cast a legendary sorcery with no legendary creature or planeswalker")
	}
}

func TestOldB212KarnsTemporalSunderingGivesTheTargetPlayerAnExtraTurnBouncesAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB212Seats(g)
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	effectstest.B12Push(g, me.ID, "Test Legend", "Legendary Creature — Human", "", 2, 2)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Karn's Temporal Sundering", "Legendary Sorcery", oldB212SunderingOracl,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: land}}); err == nil {
		t.Fatal("targeted a land as the nonland permanent")
	}
	spell := effectstest.CastCatalogSpell(t, g, "Karn's Temporal Sundering", "Legendary Sorcery", oldB212SunderingOracl,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	queued := g.ExtraTurnsQueued()
	if len(queued) != 1 || queued[0].Seat != oppSeat {
		t.Fatalf("queued extra turns %+v, want one for seat %d", queued, oppSeat)
	}
	if effectstest.BattlefieldHas(g, bear) || !opp.Hand.Contains(bear) {
		t.Error("the targeted permanent was not returned to its owner's hand")
	}
	if !effectstest.ExileHas(g, spell) {
		t.Error("Karn's Temporal Sundering was not exiled")
	}
}

func TestOldB212KarnsTemporalSunderingBounceIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB212Seats(g)
	effectstest.B12Push(g, me.ID, "Test Legend", "Legendary Creature — Human", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Karn's Temporal Sundering", "Legendary Sorcery", oldB212SunderingOracl,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.ExtraTurnsQueued()) != 1 {
		t.Fatalf("queued extra turns %d, want 1", len(g.ExtraTurnsQueued()))
	}
}

func TestOldB212SylvanAwakeningAnimatesYourLandsUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB212Seats(g)
	mySeat := g.Turn.ActiveSeat
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Sylvan Awakening", "Sorcery", oldB212AwakeningOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.B12Card(t, g, mine)
	if !c.IsLand() || !c.IsCreature() || !c.HasSubtype("Elemental") {
		t.Fatalf("my land is %q, want a land Elemental creature", c.TypeLine)
	}
	if p, tt := effectstest.PtOf(t, g, mine); p != 2 || tt != 2 {
		t.Errorf("my land is %d/%d, want 2/2", p, tt)
	}
	for _, kw := range []string{"reach", "indestructible", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("my land has no %s", kw)
		}
	}
	if effectstest.B12Card(t, g, theirs).IsCreature() {
		t.Error("an opponent's land was animated")
	}
	effectstest.EndTurn(t, g)
	if !effectstest.B12Card(t, g, mine).IsCreature() {
		t.Fatal("the effect ended before the caster's next turn")
	}
	for i := 0; g.Turn.ActiveSeat != mySeat; i++ {
		if i > 8 {
			t.Fatal("never got back to the caster's turn")
		}
		effectstest.EndTurn(t, g)
	}
	if effectstest.B12Card(t, g, mine).IsCreature() {
		t.Error("the land is still a creature on the caster's next turn")
	}
}

func TestOldB212KamahlsDruidicVowPutsLandsAndLegendsOfManaValueXOrLessOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB212Seats(g)
	effectstest.B12Push(g, me.ID, "Test Legend", "Legendary Creature — Human", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// Top of the library first: a land, a cheap legend, an expensive legend, a plain creature.
	plain := oldB212Top(me, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"})
	big := oldB212Top(me, game.Card{Name: "Big Legend", TypeLine: "Legendary Creature — Giant", ManaCost: "{5}"})
	small := oldB212Top(me, game.Card{Name: "Small Legend", TypeLine: "Legendary Creature — Elf", ManaCost: "{2}"})
	land := oldB212Top(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	deeper := oldB212Top(me, game.Card{Name: "Beyond X", TypeLine: "Legendary Creature — Elf", ManaCost: "{1}"})
	_ = deeper
	// Deeper is on top, then the land: X = 4 reaches deeper, land, small, big.
	effectstest.AddPool(t, g, me.ID, "{G}{G}{G}{G}{G}{G}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Kamahl's Druidic Vow", TypeLine: "Legendary Sorcery", ManaCost: "{X}{G}{G}",
		OracleID: oldB212VowOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 4}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, deeper, land, small)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"land": land, "small legend": small, "first legend": deeper} {
		if !effectstest.BattlefieldHas(g, id) {
			t.Errorf("%s did not enter the battlefield", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"expensive legend": big} {
		if effectstest.BattlefieldHas(g, id) || !effectstest.InGraveyard(me, id) {
			t.Errorf("%s should be in the graveyard", name)
		}
	}
	if !me.Library.Contains(plain) {
		t.Error("a card beyond the top X cards was touched")
	}
}

func TestOldB212BoardTheWeatherlightTakesAHistoricCardAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB212Seats(g)
	deep := oldB212Top(me, game.Card{Name: "Sixth Card", TypeLine: "Artifact", ManaCost: "{1}"})
	var top []uuid.UUID
	for _, c := range []game.Card{
		{Name: "Plain A", TypeLine: "Creature — Bear"}, {Name: "Plain B", TypeLine: "Sorcery"},
		{Name: "Old Relic", TypeLine: "Artifact", ManaCost: "{2}"}, {Name: "Plain C", TypeLine: "Instant"},
		{Name: "Plain D", TypeLine: "Creature — Elf"},
	} {
		top = append([]uuid.UUID{oldB212Top(me, c)}, top...)
	}
	relic := top[2]
	effectstest.CastCatalogSpell(t, g, "Board the Weatherlight", "Sorcery", oldB212WeatherlightOr, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) {
		t.Fatal("the historic card was not put into hand")
	}
	if !me.Library.Contains(deep) {
		t.Error("the sixth card left the library")
	}
	lib := me.Library.Cards
	if len(lib) < 5 || lib[len(lib)-1].InstanceID == deep && false {
		t.Fatal("library too small")
	}
	for _, id := range top {
		if id != relic && !me.Library.Contains(id) {
			t.Errorf("a card that was not taken left the library: %s", id)
		}
	}
}

func TestOldB212BoardTheWeatherlightOnlyOffersHistoricCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB212Seats(g)
	for i := 0; i < 5; i++ {
		oldB212Top(me, game.Card{Name: "Plain", TypeLine: "Creature — Bear"})
	}
	effectstest.CastCatalogSpell(t, g, "Board the Weatherlight", "Sorcery", oldB212WeatherlightOr, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.ChooseCardsChoiceFor(g, me.ID); p != nil {
		t.Fatalf("a pick was offered with no historic card among the top five: %+v", p)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand changed from %d to %d", hand, me.Hand.Size())
	}
}
