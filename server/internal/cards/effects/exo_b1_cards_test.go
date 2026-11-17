package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exo_b1_cards_test.go — Exodus, batch 1 (#5626). Each test plays the
// real catalog entry and asserts what its text says.

const (
	exoB1AllayOracle           = "664d982a-5cd8-4289-afb8-fa2e083042c9"
	exoB1ShatteringPulseOracle = "77953b7a-813e-45f5-ad7f-ab72a79a0112"
	exoB1ForbidOracle          = "6e26a525-2584-490d-b106-621c1a30e1fd"
	exoB1ReapingOracle         = "b4177445-46d8-4427-a345-d786c820c3af"
	exoB1DeathsDuetOracle      = "28d24add-da3a-4697-8fd2-5106c9225c0e"
	exoB1ErtaiOracle           = "f2902dfd-888f-4b02-b8ad-1056f0059fc3"
	exoB1TreasureTroveOracle   = "3c4a6b34-a9bb-426d-908c-7ae0011591df"
	exoB1KillerWhaleOracle     = "5ace3817-5a98-4eee-b6b9-8c105cea302f"
	exoB1WhiptongueFrogOracle  = "edc961bd-5ec9-47ed-9d34-e26a48c13702"
	exoB1ThalakosScoutOracle   = "c8877da5-7008-4b47-955e-d205220c91e6"
	exoB1EphemeronOracle       = "e4f28144-71c8-4393-b383-047529fe8878"
	exoB1CutthroatOracle       = "5fdde593-54cb-4186-974b-df4383f7efe2"
	exoB1PaladinOracle         = "fd8722e1-9c41-4037-ab7e-47e2aa9858a0"
	exoB1MirriOracle           = "f01efb04-ad2b-46fb-852c-dfb7e523e1b7"
)

func TestExoB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		exoB1AllayOracle:           "Allay",
		exoB1ShatteringPulseOracle: "Shattering Pulse",
		exoB1ForbidOracle:          "Forbid",
		exoB1ReapingOracle:         "Reaping the Rewards",
		exoB1DeathsDuetOracle:      "Death's Duet",
		exoB1ErtaiOracle:           "Ertai, Wizard Adept",
		exoB1TreasureTroveOracle:   "Treasure Trove",
		exoB1KillerWhaleOracle:     "Killer Whale",
		exoB1WhiptongueFrogOracle:  "Whiptongue Frog",
		exoB1ThalakosScoutOracle:   "Thalakos Scout",
		exoB1EphemeronOracle:       "Ephemeron",
		exoB1CutthroatOracle:       "Dauthi Cutthroat",
		exoB1PaladinOracle:         "Paladin en-Vec",
		exoB1MirriOracle:           "Mirri, Cat Warrior",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// exoB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func exoB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}

// exoB1HandCard puts a vanilla card in a player's hand.
func exoB1HandCard(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	return id
}

// exoB1Cast puts a catalog spell in the active player's hand and casts
// it with the given parameters, in a main phase.
func exoB1Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id, g.CastSpell(active.ID, id, params)
}

func exoB1Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return c.Tapped
}

// --- buyback --------------------------------------------------------------------------

func TestExoB1AllayAndShatteringPulseDestroyTheirTarget(t *testing.T) {
	for _, card := range []struct {
		name, oracle, victimType string
		wrongType                string
	}{
		{"Allay", exoB1AllayOracle, "Enchantment", "Artifact"},
		{"Shattering Pulse", exoB1ShatteringPulseOracle, "Artifact", "Enchantment"},
	} {
		for _, tc := range tmpB1BuybackCases {
			t.Run(card.name+" "+tc.name, func(t *testing.T) {
				g := newCatalogGame(t)
				caster, opp := seatsForCombat(g)
				victim := exoB1Push(g, opp.ID, "Victim", card.victimType, "", 0, 0)
				id, err := castWithOptionalCosts(t, g, card.name, "Instant", card.oracle, cardTarget(victim), tc.optional, nil)
				if err != nil {
					t.Fatalf("CastSpell: %v", err)
				}
				passPriorityAroundTable(t, g)
				if _, ok := battlefieldCard(g, victim); ok || !opp.Graveyard.Contains(victim) {
					t.Errorf("the %s was not destroyed", card.victimType)
				}
				tmpB1AssertBuyback(t, caster, id, tc.wantHand)
			})
		}
		t.Run(card.name+" refuses the other type", func(t *testing.T) {
			g := newCatalogGame(t)
			_, opp := seatsForCombat(g)
			wrong := exoB1Push(g, opp.ID, "Wrong", card.wrongType, "", 0, 0)
			if err := castCatalogSpellErr(t, g, card.name, "Instant", card.oracle, cardTarget(wrong)); err == nil {
				t.Errorf("%s accepted a target that is only an %s", card.name, card.wrongType)
			}
		})
	}
}

func TestExoB1ForbidCountersWithAndWithoutItsDiscardBuyback(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster, opp := seatsForCombat(g)
			bolt := b10OpponentCastsBolt(t, g, opp, playerTarget(caster.ID))
			life := caster.Life
			params := game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}
			var pile []uuid.UUID
			if tc.wantHand {
				pile = []uuid.UUID{exoB1HandCard(caster, "Fodder A"), exoB1HandCard(caster, "Fodder B")}
				params.OptionalCosts = tc.optional
				params.DiscardIDs = pile
			}
			id, err := exoB1Cast(t, g, "Forbid", "Instant", exoB1ForbidOracle, params)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			for _, d := range pile {
				if !caster.Graveyard.Contains(d) {
					t.Errorf("the buyback discard %s is not in the graveyard", d)
				}
			}
			passPriorityAroundTable(t, g)
			if caster.Life != life || !opp.Graveyard.Contains(bolt) {
				t.Errorf("the Bolt was not countered: life %d → %d", life, caster.Life)
			}
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
		})
	}
}

func TestExoB1ForbidRefusesBuybackWithFewerThanTwoCards(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := seatsForCombat(g)
	bolt := b10OpponentCastsBolt(t, g, opp, playerTarget(caster.ID))
	one := exoB1HandCard(caster, "Lone card")
	if _, err := exoB1Cast(t, g, "Forbid", "Instant", exoB1ForbidOracle, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
		OptionalCosts: []int{0},
		DiscardIDs:    []uuid.UUID{one},
	}); err == nil {
		t.Error("Forbid's buyback was paid with one discard")
	}
}

func TestExoB1ReapingTheRewardsGainsTwoLife(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			var sacs []uuid.UUID
			var land uuid.UUID
			if tc.wantHand {
				land = seedLand(g, caster.ID, "Plains", "Basic Land — Plains", "")
				sacs = []uuid.UUID{land}
			}
			life := caster.Life
			id, err := castWithOptionalCosts(t, g, "Reaping the Rewards", "Instant", exoB1ReapingOracle, nil, tc.optional, sacs)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			if tc.wantHand && !caster.Graveyard.Contains(land) {
				t.Error("the buyback land was not sacrificed")
			}
			passPriorityAroundTable(t, g)
			if got := caster.Life - life; got != 2 {
				t.Errorf("gained %d life, want 2", got)
			}
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
		})
	}
}

// --- Death's Duet ------------------------------------------------------------------------

func TestExoB1DeathsDuetReturnsTwoCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	a := pushGraveyardCardTyped(caster, "Bear A", "Creature — Bear")
	b := pushGraveyardCardTyped(caster, "Bear B", "Creature — Bear")
	if err := castCatalogSpellErr(t, g, "Death's Duet", "Sorcery", exoB1DeathsDuetOracle, cardTargets(a, b)); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !caster.Hand.Contains(a) || !caster.Hand.Contains(b) {
		t.Errorf("hand has A=%v B=%v, want both creature cards back", caster.Hand.Contains(a), caster.Hand.Contains(b))
	}
}

func TestExoB1DeathsDuetNeedsTwoCreatureCardsFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := seatsForCombat(g)
	a := pushGraveyardCardTyped(caster, "Bear A", "Creature — Bear")
	rite := pushGraveyardCardTyped(caster, "Rite", "Sorcery")
	theirs := pushGraveyardCardTyped(opp, "Their Bear", "Creature — Bear")
	for name, targets := range map[string][]game.TargetRef{
		"one target":          cardTargets(a),
		"a noncreature card":  cardTargets(a, rite),
		"another graveyard's": cardTargets(a, theirs),
	} {
		if err := castCatalogSpellErr(t, g, "Death's Duet", "Sorcery", exoB1DeathsDuetOracle, targets); err == nil {
			t.Errorf("Death's Duet was cast with %s", name)
		}
	}
}

// --- Ertai, Wizard Adept ------------------------------------------------------------------

func TestExoB1ErtaiCountersTargetSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	ertai := pushCatalogPermanent(g, me.ID, "Ertai, Wizard Adept", "Legendary Creature — Human Wizard", exoB1ErtaiOracle, false)
	bolt := b10OpponentCastsBolt(t, g, opp, playerTarget(me.ID))
	life := me.Life
	dcMana(t, g, me.ID, "{U}{U}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, ertai, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !exoB1Tapped(t, g, ertai) {
		t.Error("Ertai did not tap to activate")
	}
	passPriorityAroundTable(t, g)
	if me.Life != life || !opp.Graveyard.Contains(bolt) {
		t.Errorf("the Bolt was not countered: life %d → %d", life, me.Life)
	}
}

func TestExoB1ErtaiCantActivateWhileSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	ertai := pushCatalogPermanent(g, me.ID, "Ertai, Wizard Adept", "Legendary Creature — Human Wizard", exoB1ErtaiOracle, true)
	bolt := b10OpponentCastsBolt(t, g, opp, playerTarget(me.ID))
	dcMana(t, g, me.ID, "{U}{U}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, ertai, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err == nil {
		t.Error("a summoning-sick Ertai activated its {T} ability")
	}
}

// --- Treasure Trove -----------------------------------------------------------------------

func TestExoB1TreasureTroveDrawsACardEachActivation(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	trove := pushCatalogPermanent(g, me.ID, "Treasure Trove", "Enchantment", exoB1TreasureTroveOracle, false)
	hand := me.Hand.Size()
	for i := 1; i <= 2; i++ {
		dcMana(t, g, me.ID, "{U}{U}{U}{U}")
		if err := g.ActivateCatalogAbility(me.ID, trove, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != hand+i {
			t.Fatalf("hand %d → %d after %d activations, want %d", hand, got, i, hand+i)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, trove, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Treasure Trove was activated with an empty mana pool")
	}
}

// --- flying until end of turn ---------------------------------------------------------------

func TestExoB1KillerWhaleAndWhiptongueFrogGainFlyingUntilEndOfTurn(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		power, toughness       int
	}{
		{"Killer Whale", exoB1KillerWhaleOracle, "Creature — Whale", 3, 5},
		{"Whiptongue Frog", exoB1WhiptongueFrogOracle, "Creature — Frog", 1, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := exoB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			if hasEffectiveKeyword(t, g, id, "flying") {
				t.Fatalf("%s has flying before the activation", tc.name)
			}
			dcMana(t, g, me.ID, "{U}")
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !hasEffectiveKeyword(t, g, id, "flying") {
				t.Fatalf("%s did not gain flying", tc.name)
			}
			dcCleanup(g)
			if hasEffectiveKeyword(t, g, id, "flying") {
				t.Errorf("%s kept flying past the cleanup step", tc.name)
			}
		})
	}
}

// --- discard a card: return this creature ----------------------------------------------------

func TestExoB1ThalakosScoutAndEphemeronReturnToHandForADiscard(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, keyword string
	}{
		{"Thalakos Scout", exoB1ThalakosScoutOracle, "Creature — Thalakos Soldier Scout", "shadow"},
		{"Ephemeron", exoB1EphemeronOracle, "Creature — Illusion", "flying"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := exoB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 1)
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Fatalf("%s does not have %s", tc.name, tc.keyword)
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("the ability was activated with an empty hand")
			}
			fodder := exoB1HandCard(me, "Fodder")
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
				Strict: true, DiscardIDs: []uuid.UUID{fodder},
			}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if !me.Graveyard.Contains(fodder) {
				t.Error("the discarded card is not in the graveyard")
			}
			passPriorityAroundTable(t, g)
			if _, ok := battlefieldCard(g, id); ok || !me.Hand.Contains(id) {
				t.Errorf("%s did not return to its owner's hand", tc.name)
			}
		})
	}
}

// --- Dauthi Cutthroat -----------------------------------------------------------------------

func TestExoB1DauthiCutthroatDestroysOnlyACreatureWithShadow(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := seatsForCombat(g)
	cutthroat := exoB1Push(g, me.ID, "Dauthi Cutthroat", "Creature — Dauthi Minion", exoB1CutthroatOracle, 1, 1, "B")
	if !hasEffectiveKeyword(t, g, cutthroat, "shadow") {
		t.Fatal("Dauthi Cutthroat does not have shadow")
	}
	plain := exoB1Push(g, opp.ID, "Plain Bear", "Creature — Bear", "", 2, 2)
	shadow := tmpB1Shadow(g, opp.ID, "Shadow Bear")
	dcMana(t, g, me.ID, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, cutthroat, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardTarget(plain),
	}); err == nil {
		t.Fatal("Dauthi Cutthroat targeted a creature without shadow")
	}
	if err := g.ActivateCatalogAbility(me.ID, cutthroat, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardTarget(shadow),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, shadow); ok {
		t.Error("the creature with shadow was not destroyed")
	}
	if _, ok := battlefieldCard(g, plain); !ok {
		t.Error("the creature without shadow left the battlefield")
	}
}

// --- keyword creatures ---------------------------------------------------------------------

func TestExoB1PaladinEnVecHasFirstStrikeAndProtectionFromBlackAndRed(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	paladin := exoB1Push(g, me.ID, "Paladin en-Vec", "Creature — Human Knight", exoB1PaladinOracle, 2, 2, "W")
	for _, kw := range []string{"first strike", game.ProtectionFromColor("B"), game.ProtectionFromColor("R")} {
		if !hasEffectiveKeyword(t, g, paladin, kw) {
			t.Errorf("Paladin en-Vec does not have %s", kw)
		}
	}
	black := exoB1Push(g, them.ID, "Black Bear", "Creature — Bear", "", 2, 2, "B")
	red := exoB1Push(g, them.ID, "Red Bear", "Creature — Bear", "", 2, 2, "R")
	green := exoB1Push(g, them.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	attackIntoBlocks(t, g, them.ID, paladin)
	for name, id := range map[string]uuid.UUID{"black": black, "red": red} {
		if blockRefusal(t, g, id, paladin) == "" {
			t.Errorf("a %s creature was allowed to block Paladin en-Vec", name)
		}
	}
	if got := blockRefusal(t, g, green, paladin); got != "" {
		t.Errorf("a green creature was refused: %q", got)
	}
}

func TestExoB1MirriHasFirstStrikeForestwalkAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	mirri := exoB1Push(g, me.ID, "Mirri, Cat Warrior", "Legendary Creature — Cat Warrior", exoB1MirriOracle, 2, 3, "G")
	for _, kw := range []string{"first strike", "forestwalk", "vigilance"} {
		if !hasEffectiveKeyword(t, g, mirri, kw) {
			t.Errorf("Mirri does not have %s", kw)
		}
	}
	bear := exoB1Push(g, them.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2, "G")
	seedLand(g, them.ID, "Forest", "Basic Land — Forest", "")
	attackIntoBlocks(t, g, them.ID, mirri)
	if exoB1Tapped(t, g, mirri) {
		t.Error("Mirri tapped to attack despite vigilance")
	}
	if r := blockRefusal(t, g, bear, mirri); r != game.BlockReasonLandwalk {
		t.Errorf("block refusal = %q, want landwalk while the defender controls a Forest", r)
	}
}
