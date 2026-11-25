package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b303_cards_test.go — Commander 2015, batch 303 (S82 older sets,
// #5546). Every test plays the real catalog entry through the engine.

var oldB303Oracles = map[string]string{
	"Arbiter of Knollridge": "62705a41-7c4d-4e6c-85b0-36a9d27ab5ae",
	"Act of Aggression":     "9d7624ea-ffff-470b-928b-bbac66ddbd1c",
	"Underworld Coinsmith":  "ae06e0e5-8a42-4d1f-a629-94152e5f1fe6",
	"Victory's Herald":      "e410f80f-cf67-46ff-aeb3-da0204237d66",
	"Breath of Darigaaz":    "1954994f-17bf-4ea5-af72-60f9bfcb6569",
	"Primal Growth":         "3cbdf37b-7fe3-4791-b33c-8591158b0ce5",
	"Desolation Giant":      "c23aaf6d-151e-481a-a345-ca00c14940d1",
	"Treasury Thrull":       "34161ec4-5d99-45e6-b43c-c8d28958660a",
	"Dawn to Dusk":          "c7a98058-b40d-4d74-9b97-6f04adf3c684",
	"Diabolic Servitude":    "e275cbfb-3d25-4189-84f0-3162eadc7d65",
}

func TestOldB303EveryCardIsRegisteredFull(t *testing.T) {
	for name, oracle := range oldB303Oracles {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not in the catalog", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %d, want Full with none", name, spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// oldB303Cast puts the card in the active seat's hand and casts it strictly
// (no mana cost on the hand card, so the spell is free apart from what the
// params add).
func oldB303Cast(t *testing.T, g *game.Game, name, typeLine string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	effectstest.D5681Spell(g, me, id, name, typeLine, "", oldB303Oracles[name])
	params.Strict = true
	return id, g.CastSpell(me.ID, id, params)
}

func oldB303Damage(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.DamageMarked
			}
		}
	})
	return n
}

func TestOldB303ArbiterOfKnollridgeRaisesEveryLifeTotalToTheHighest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	lives := []int{25, 10, 35, 20}
	for i, p := range g.Seats {
		p.Life = lives[i%len(lives)]
	}
	effectstest.CastHoldCreature(t, g, "Arbiter of Knollridge", "Creature — Giant Wizard", oldB303Oracles["Arbiter of Knollridge"], 5, 5)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != 35 {
			t.Errorf("seat %d is at %d, want 35", i, p.Life)
		}
	}
}

func TestOldB303ActOfAggressionStealsUntapsAndGivesHasteForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(theirs) })
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Fatal("setup: the creature is not tapped")
	}
	if _, err := oldB303Cast(t, g, "Act of Aggression", "Instant", game.CastSpellParams{Targets: effectstest.CardTarget(mine)}); err == nil {
		t.Fatal("targeted my own creature")
	}
	if _, err := oldB303Cast(t, g, "Act of Aggression", "Instant", game.CastSpellParams{Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	live := effectstest.JmpB4Live(t, g, theirs)
	if live.Controller != me.ID {
		t.Errorf("controller %v, want me", live.Controller)
	}
	if live.Tapped {
		t.Error("the stolen creature is still tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("the stolen creature has no haste")
	}
	effectstest.EndTurn(t, g)
	if live := effectstest.JmpB4Live(t, g, theirs); live.Controller != opp.ID {
		t.Errorf("after the turn the controller is %v, want the opponent back", live.Controller)
	}
}

func TestOldB303UnderworldCoinsmithGainsLifeForItselfAndOtherEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Underworld Coinsmith", "Enchantment Creature — Human Cleric", oldB303Oracles["Underworld Coinsmith"], 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("life %d → %d on its own entry, want +1", life, me.Life)
	}
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Some Enchantment", TypeLine: "Enchantment"})
	if me.Life != life+2 {
		t.Errorf("life %d after another enchantment entered, want %d", me.Life, life+2)
	}
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Some Rock", TypeLine: "Artifact"})
	if me.Life != life+2 {
		t.Errorf("a nonenchantment gained life: %d", me.Life)
	}
}

func TestOldB303UnderworldCoinsmithDrainsEachOpponentForALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	coin := effectstest.B12Push(g, me.ID, "Underworld Coinsmith", "Enchantment Creature — Human Cleric", oldB303Oracles["Underworld Coinsmith"], 2, 2)
	effectstest.FloatMana(t, g, me, "{W}{B}")
	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	effectstest.B16Activate(t, g, me.ID, coin, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		// Each opponent loses 1; I paid 1 life as a cost and gained nothing.
		if want := life[p.ID] - 1; p.Life != want {
			t.Errorf("a player is at %d, want %d", p.Life, want)
		}
	}
}

func TestOldB303VictorysHeraldGivesEveryAttackerFlyingAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herald := effectstest.B12Push(g, me.ID, "Victory's Herald", "Creature — Angel", oldB303Oracles["Victory's Herald"], 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	stay := effectstest.B12Creature(g, me.ID, "Stay Home", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, herald, bear)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{herald, bear} {
		for _, kw := range []string{"flying", "lifelink"} {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("an attacker lacks %s", kw)
			}
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, stay, "lifelink") {
		t.Error("a creature that did not attack gained lifelink")
	}
}

func TestOldB303BreathOfDarigaazHitsGroundCreaturesAndPlayersForOneOrFour(t *testing.T) {
	for _, tc := range []struct {
		kicked bool
		want   int
	}{{false, 1}, {true, 4}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[0], g.Seats[1]
		ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 9)
		flyer := effectstest.B30Flyer(g, opp.ID, "Flyer", "Creature — Bird", 2, 9)
		lives := map[uuid.UUID]int{}
		for _, p := range g.Seats {
			lives[p.ID] = p.Life
		}
		params := game.CastSpellParams{}
		if tc.kicked {
			effectstest.FloatMana(t, g, me, "{C}{C}")
			params.OptionalCosts = []int{0}
		}
		if _, err := oldB303Cast(t, g, "Breath of Darigaaz", "Sorcery", params); err != nil {
			t.Fatalf("kicked=%v cast: %v", tc.kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := oldB303Damage(g, ground); got != tc.want {
			t.Errorf("kicked=%v: the ground creature has %d damage, want %d", tc.kicked, got, tc.want)
		}
		if got := oldB303Damage(g, flyer); got != 0 {
			t.Errorf("kicked=%v: the flyer has %d damage", tc.kicked, got)
		}
		for _, p := range g.Seats {
			if p.Life != lives[p.ID]-tc.want {
				t.Errorf("kicked=%v: a player is at %d, want %d", tc.kicked, p.Life, lives[p.ID]-tc.want)
			}
		}
	}
}

func oldB303Basics(p *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		c := game.NewCard("Forest", p.ID)
		c.TypeLine = "Basic Land — Forest"
		p.Library.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func TestOldB303PrimalGrowthFetchesOneBasicOrTwoWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[0]
		basics := oldB303Basics(me, 3)
		fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
		params := game.CastSpellParams{}
		if kicked {
			params.OptionalCosts = []int{0}
			params.SacrificeIDs = []uuid.UUID{fodder}
		}
		if _, err := oldB303Cast(t, g, "Primal Growth", "Sorcery", params); err != nil {
			t.Fatalf("kicked=%v cast: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSearchByID(t, g, me.ID, basics[:map[bool]int{false: 1, true: 2}[kicked]]...)
		effectstest.PassPriorityAroundTable(t, g)
		got := 0
		for _, id := range basics {
			if g.Battlefield.Contains(id) {
				got++
				if effectstest.B16Tapped(t, g, id) {
					t.Errorf("kicked=%v: a land entered tapped", kicked)
				}
			}
		}
		if want := map[bool]int{false: 1, true: 2}[kicked]; got != want {
			t.Errorf("kicked=%v: %d lands on the battlefield, want %d", kicked, got, want)
		}
		if kicked && g.Battlefield.Contains(fodder) {
			t.Error("the kicker's sacrifice was not paid")
		}
	}
}

func TestOldB303PrimalGrowthRefusesAKickerSacrificeOfNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	if _, err := oldB303Cast(t, g, "Primal Growth", "Sorcery", game.CastSpellParams{OptionalCosts: []int{0}}); err == nil {
		t.Fatal("kicked with no creature sacrificed")
	}
}

func TestOldB303DesolationGiantDestroysYourOthersOrEveryOtherWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		params := game.CastSpellParams{}
		if kicked {
			effectstest.FloatMana(t, g, me, "{W}{W}")
			params.OptionalCosts = []int{0}
		}
		id, err := oldB303Cast(t, g, "Desolation Giant", "Creature — Giant", params)
		if err != nil {
			t.Fatalf("kicked=%v cast: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(mine) {
			t.Errorf("kicked=%v: my creature survived", kicked)
		}
		if got := g.Battlefield.Contains(theirs); got == kicked {
			t.Errorf("kicked=%v: their creature on the battlefield = %v", kicked, got)
		}
		if !g.Battlefield.Contains(id) {
			t.Errorf("kicked=%v: the Giant destroyed itself", kicked)
		}
	}
}

func TestOldB303TreasuryThrullReturnsAnArtifactCreatureOrEnchantmentWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thrull := effectstest.PushDiesCreatureForTest(g, me.ID, "Treasury Thrull", oldB303Oracles["Treasury Thrull"], "Creature — Thrull", 4, 4)
	ench := effectstest.PushGraveyardPermanent(me, "Old Enchantment", "Enchantment", "{1}{W}")
	land := effectstest.PushGraveyardPermanent(me, "Old Land", "Land", "")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Rock", "Artifact", "{1}")
	effectstest.DeclareAttack(t, g, opp.ID, thrull)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	if !effectstest.HasID(prompt.PickTargetCards, ench) {
		t.Error("an enchantment card is not a legal target")
	}
	for name, id := range map[string]uuid.UUID{"a land": land, "an opponent's card": theirs} {
		if effectstest.HasID(prompt.PickTargetCards, id) {
			t.Errorf("%s is offered", name)
		}
	}
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ench) || me.Graveyard.Contains(ench) {
		t.Error("the enchantment card did not return to hand")
	}
}

func TestOldB303TreasuryThrullsReturnIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thrull := effectstest.PushDiesCreatureForTest(g, me.ID, "Treasury Thrull", oldB303Oracles["Treasury Thrull"], "Creature — Thrull", 4, 4)
	card := effectstest.PushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	effectstest.DeclareAttack(t, g, opp.ID, thrull)
	effectstest.PickCard(t, g, me.ID, card)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(card) {
		t.Error("the card left the graveyard after the player declined")
	}
}

func TestOldB303DawnToDuskReturnsAndOrDestroysAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	back := effectstest.PushGraveyardPermanent(me, "Old Enchantment", "Enchantment", "{1}{W}")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	both := []game.TargetRef{{Kind: game.TargetCard, ID: back, Mode: 0}, {Kind: game.TargetCard, ID: theirs, Mode: 1}}
	if _, err := oldB303Cast(t, g, "Dawn to Dusk", "Sorcery", game.CastSpellParams{}); err == nil {
		t.Fatal("cast with no mode chosen")
	}
	if _, err := oldB303Cast(t, g, "Dawn to Dusk", "Sorcery", game.CastSpellParams{Modes: []int{0, 1}, Targets: both}); err != nil {
		t.Fatalf("cast both: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(back) {
		t.Error("the enchantment card did not return to hand")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the target enchantment was not destroyed")
	}
}

func TestOldB303DawnToDuskDestroyOnlyLeavesTheGraveyardAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	back := effectstest.PushGraveyardPermanent(me, "Old Enchantment", "Enchantment", "{1}{W}")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	if _, err := oldB303Cast(t, g, "Dawn to Dusk", "Sorcery", game.CastSpellParams{Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs, Mode: 0}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the enchantment was not destroyed")
	}
	if !me.Graveyard.Contains(back) {
		t.Error("the unchosen bullet moved a card")
	}
}

func oldB303Servitude(t *testing.T) (g *game.Game, me *game.Player, card, servitude uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	card = effectstest.PushGraveyardPermanent(me, "Fallen Bear", "Creature — Bear", "{1}{G}")
	servitude = effectstest.CastHoldCreature(t, g, "Diabolic Servitude", "Enchantment", oldB303Oracles["Diabolic Servitude"], 0, 0)
	effectstest.AnswerPickTargetIfOpen(t, g, card)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(card) {
		t.Fatalf("setup: the creature did not return: %+v", g.PendingChoices)
	}
	return g, me, card, servitude
}

func TestOldB303DiabolicServitudeReturnsTheCreatureThenExilesItWhenItDies(t *testing.T) {
	g, me, card, servitude := oldB303Servitude(t)
	effectstest.KillCreature(t, g, me.ID, card)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOf(g, card); z != game.ZoneExile {
		t.Errorf("the dead creature is in %q, want exile", z)
	}
	if !me.Hand.Contains(servitude) || g.Battlefield.Contains(servitude) {
		t.Error("Diabolic Servitude did not return to its owner's hand")
	}
}

func TestOldB303DiabolicServitudeExilesTheCreatureWhenItLeaves(t *testing.T) {
	g, me, card, servitude := oldB303Servitude(t)
	effectstest.KillCreature(t, g, me.ID, servitude)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(card) {
		t.Error("the creature stayed on the battlefield")
	}
	if z := effectstest.ZoneOf(g, card); z != game.ZoneExile {
		t.Errorf("the creature is in %q, want exile", z)
	}
}

func TestOldB303DiabolicServitudeLeavesAnUnlinkedCreatureAlone(t *testing.T) {
	g, me, card, servitude := oldB303Servitude(t)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(servitude) || !g.Battlefield.Contains(card) {
		t.Error("an unrelated creature's death disturbed the Servitude or its creature")
	}
}
