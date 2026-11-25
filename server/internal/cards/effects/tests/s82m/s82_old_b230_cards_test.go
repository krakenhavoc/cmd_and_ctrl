package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b230_cards_test.go — OLD batch 230 (Ixalan, #5532): eleven cards.
// Each test plays the real catalog entry through the engine.

const (
	oldB230CannonsOracle    = "5e7eca9c-a7b8-4b7b-a0a0-e8937530145a"
	oldB230BloodFastOracle  = "be2a4bc4-8af6-48c5-9421-32d26272e71a"
	oldB230LandingOracle    = "f7d8b91b-6541-4d3e-af51-7e000eac69c1"
	oldB230CompassOracle    = "f9085e55-2833-41b7-9100-a35dc04dee93"
	oldB230SacramentOracle  = "851be993-35d7-4566-96a0-3fe84d17979e"
	oldB230VanguardOracle   = "91c380f6-2c41-48b7-9f07-104e8659e397"
	oldB230AshesOracle      = "a75d5b54-5cc9-49c4-8e8c-3bef22d4c01c"
	oldB230DiplomacyOracle  = "ac5585ce-2959-4eaf-af56-5dca8cdfb2d6"
	oldB230ShelterOracle    = "551ac58e-3a27-4ee3-ac91-16873562c0bc"
	oldB230AgroundOracle    = "6ef9bf18-6b09-4398-a502-6867d999ab0c"
	oldB230HijackOracle     = "fd5651f7-52f8-4845-8a1b-32e5d2e9b00e"
	oldB230EarthquakeOracle = "1903200d-0350-48a9-ac7c-b851f7df647f"
)

func TestOldB230EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB230CannonsOracle:          "Vance's Blasting Cannons",
		oldB230CannonsOracle + "#1":   "Spitfire Bastion",
		oldB230BloodFastOracle:        "Arguel's Blood Fast",
		oldB230BloodFastOracle + "#1": "Temple of Aclazotz",
		oldB230LandingOracle:          "Legion's Landing",
		oldB230LandingOracle + "#1":   "Adanto, the First Fort",
		oldB230CompassOracle:          "Thaumatic Compass",
		oldB230CompassOracle + "#1":   "Spires of Orazca",
		oldB230SacramentOracle:        "Sanguine Sacrament",
		oldB230VanguardOracle:         "Adanto Vanguard",
		oldB230AshesOracle:            "Ashes of the Abhorrent",
		oldB230DiplomacyOracle:        "Sword-Point Diplomacy",
		oldB230ShelterOracle:          "Sheltering Light",
		oldB230AgroundOracle:          "Run Aground",
		oldB230HijackOracle:           "Hijack",
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

// oldB230Put puts a transform card onto the battlefield under p, front face up.
func oldB230Put(g *game.Game, p *game.Player, oracle, front, frontType, frontCost, back, backType string) uuid.UUID {
	return effectstest.PushCraftCard(g, p, effectstest.TransformRow(oracle, front, frontType, frontCost, back, backType, "", "", nil))
}

// oldB230Hand puts a card into p's hand and casts it with no payment.
func oldB230Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	if err := g.CastSpell(p.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return c.InstanceID
}

func oldB230Zone(g *game.Game, id uuid.UUID, kind game.ZoneKind) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == kind
}

// --- Vance's Blasting Cannons // Spitfire Bastion ---------------------------

func TestOldB230CannonsExileTheTopCardAtUpkeepAndLetYouCastANonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	oldB230Put(g, me, oldB230CannonsOracle, "Vance's Blasting Cannons", "Legendary Enchantment", "{3}{R}", "Spitfire Bastion", "Legendary Land")
	bears := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: bears, Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB230Zone(g, bears, game.ZoneExile) {
		t.Fatal("the top card was not exiled")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{G}")
	if err := g.CastSpell(me.ID, bears, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("the exiled nonland card could not be cast: %v", err)
	}
}

func TestOldB230CannonsLeaveAnExiledLandInExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	oldB230Put(g, me, oldB230CannonsOracle, "Vance's Blasting Cannons", "Legendary Enchantment", "{3}{R}", "Spitfire Bastion", "Legendary Land")
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB230Zone(g, forest, game.ZoneExile) {
		t.Fatal("the land was not exiled")
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, forest, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("an exiled land was cast")
	}
}

func TestOldB230CannonsTransformOnTheThirdSpellOfATurn(t *testing.T) {
	for _, spells := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		oldB230Put(g, me, oldB230CannonsOracle, "Vance's Blasting Cannons", "Legendary Enchantment", "{3}{R}", "Spitfire Bastion", "Legendary Land")
		for i := 0; i < spells; i++ {
			oldB230Cast(t, g, me, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
			if i == 2 {
				effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		_, transformed := oldB230Named(g, "Spitfire Bastion")
		if transformed != (spells == 3) {
			t.Errorf("after %d spells transformed=%v", spells, transformed)
		}
	}
}

func oldB230Named(g *game.Game, name string) (game.Card, bool) {
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				out, ok = c, true
			}
		}
	})
	return out, ok
}

func TestOldB230SpitfireBastionDealsThreeDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id := oldB230Put(g, me, oldB230CannonsOracle, "Vance's Blasting Cannons", "Legendary Enchantment", "{3}{R}", "Spitfire Bastion", "Legendary Land")
	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("transform: %v", err)
	}
	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want 3 damage", life, opp.Life)
	}
}

// --- Arguel's Blood Fast // Temple of Aclazotz ------------------------------

func TestOldB230BloodFastDrawsForTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := oldB230Put(g, me, oldB230BloodFastOracle, "Arguel's Blood Fast", "Legendary Enchantment", "{1}{B}", "Temple of Aclazotz", "Legendary Land")
	effectstest.FloatMana(t, g, me, "{C}{B}")
	life, hand := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d → %d, hand %d → %d; want -2 life and a card", life, me.Life, hand, me.Hand.Size())
	}
}

func TestOldB230BloodFastTransformsAtUpkeepOnlyAtFiveLifeOrLess(t *testing.T) {
	for _, life := range []int{6, 5} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[1]
		me.Life = life
		oldB230Put(g, me, oldB230BloodFastOracle, "Arguel's Blood Fast", "Legendary Enchantment", "{1}{B}", "Temple of Aclazotz", "Legendary Land")
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		if life == 5 {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if _, ok := oldB230Named(g, "Temple of Aclazotz"); ok != (life == 5) {
			t.Errorf("life %d: transformed=%v", life, ok)
		}
	}
}

func TestOldB230TempleOfAclazotzGainsLifeEqualToToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := oldB230Put(g, me, oldB230BloodFastOracle, "Arguel's Blood Fast", "Legendary Enchantment", "{1}{B}", "Temple of Aclazotz", "Legendary Land")
	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("transform: %v", err)
	}
	wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 5)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{wall}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5", life, me.Life)
	}
}

// --- Legion's Landing // Adanto, the First Fort ------------------------------

func TestOldB230LegionsLandingMakesAVampireAndTransformsOnAThreeCreatureAttack(t *testing.T) {
	for _, n := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		oldB230Put(g, me, oldB230LandingOracle, "Legion's Landing", "Legendary Enchantment", "{W}", "Adanto, the First Fort", "Legendary Land")
		var attackers []uuid.UUID
		for i := 0; i < n; i++ {
			attackers = append(attackers, effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 1, 1))
		}
		effectstest.DeclareAttack(t, g, opp.ID, attackers...)
		effectstest.PassPriorityAroundTable(t, g)
		if _, ok := oldB230Named(g, "Adanto, the First Fort"); ok != (n == 3) {
			t.Errorf("%d attackers: transformed=%v", n, ok)
		}
	}
}

func TestOldB230LegionsLandingEntersWithALifelinkVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	row := effectstest.TransformRow(oldB230LandingOracle, "Legion's Landing", "Legendary Enchantment", "{W}", "Adanto, the First Fort", "Legendary Land", "", "", []string{"W"})
	effectstest.ImportToHand(row, me)
	var id uuid.UUID
	for _, c := range me.Hand.Cards {
		if c.OracleID == oldB230LandingOracle {
			id = c.InstanceID
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	tok, ok := oldB230Named(g, "Vampire")
	if !ok || !game.HasKeyword(&tok, "lifelink") || tok.Power != 1 || tok.Toughness != 1 {
		t.Errorf("Vampire token = %+v, ok=%v; want a 1/1 lifelinker", tok.Name, ok)
	}
}

func TestOldB230AdantoMakesAVampireForThreeMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := oldB230Put(g, me, oldB230LandingOracle, "Legion's Landing", "Legendary Enchantment", "{W}", "Adanto, the First Fort", "Legendary Land")
	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("transform: %v", err)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := oldB230Named(g, "Vampire"); !ok {
		t.Error("no Vampire token")
	}
}

// --- Thaumatic Compass // Spires of Orazca ----------------------------------

func TestOldB230CompassFindsABasicLandAndTransformsWithSevenLands(t *testing.T) {
	for _, lands := range []int{6, 7} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		oldB230Put(g, me, oldB230CompassOracle, "Thaumatic Compass", "Artifact", "{2}", "Spires of Orazca", "Land")
		for i := 0; i < lands; i++ {
			effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if _, ok := oldB230Named(g, "Spires of Orazca"); ok != (lands == 7) {
			t.Errorf("%d lands: transformed=%v", lands, ok)
		}
	}
}

func TestOldB230CompassSearchesForABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := oldB230Put(g, me, oldB230CompassOracle, "Thaumatic Compass", "Artifact", "{2}", "Spires of Orazca", "Land")
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, forest) {
		t.Error("the Forest is not in hand")
	}
}

func TestOldB230SpiresUntapsAndRemovesAnAttackerFromCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spires := oldB230Put(g, opp, oldB230CompassOracle, "Thaumatic Compass", "Artifact", "{2}", "Spires of Orazca", "Land")
	if err := g.TransformPermanentForEffect(spires); err != nil {
		t.Fatalf("transform: %v", err)
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	if err := g.ActivateCatalogAbility(opp.ID, spires, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.DftABCardRef(attacker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.B12Card(t, g, attacker)
	if c.AttackingTarget != uuid.Nil || c.Tapped {
		t.Errorf("attacking=%v tapped=%v, want removed from combat and untapped", c.AttackingTarget != uuid.Nil, c.Tapped)
	}
}

// --- Sanguine Sacrament -----------------------------------------------------

func TestOldB230SanguineSacramentGainsTwiceXAndGoesToTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	id := oldB230Cast(t, g, me, game.Card{Name: "Sanguine Sacrament", TypeLine: "Instant", OracleID: oldB230SacramentOracle}, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d → %d, want +6", life, me.Life)
	}
	if !oldB230Zone(g, id, game.ZoneLibrary) || me.Library.Cards[0].InstanceID != id {
		t.Errorf("the spell is not on the bottom of the library (zone library=%v)", oldB230Zone(g, id, game.ZoneLibrary))
	}
	if oldB230Zone(g, id, game.ZoneGraveyard) {
		t.Error("the spell also went to the graveyard")
	}
}

// --- Adanto Vanguard --------------------------------------------------------

func TestOldB230AdantoVanguardGetsPlusTwoWhileAttackingAndPaysLifeForIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	v := effectstest.DftABPush(g, me.ID, "Adanto Vanguard", "Creature — Vampire Soldier", oldB230VanguardOracle, 1, 1)
	if p := effectstest.EffectivePower(t, g, v); p != 1 {
		t.Errorf("not attacking: power %d, want 1", p)
	}
	effectstest.DeclareAttack(t, g, opp.ID, v)
	if p := effectstest.EffectivePower(t, g, v); p != 3 {
		t.Errorf("attacking: power %d, want 3", p)
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, v, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-4 {
		t.Errorf("life %d → %d, want -4", life, me.Life)
	}
	if c := effectstest.B12Card(t, g, v); !game.HasKeyword(&c, "indestructible") {
		t.Error("the Vanguard is not indestructible")
	}
	me.Life = 3
	if err := g.ActivateCatalogAbility(me.ID, v, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("4 life was paid with 3")
	}
}

// --- Ashes of the Abhorrent -------------------------------------------------

func TestOldB230AshesGainsLifeForEveryCreatureThatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.DftABPush(g, me.ID, "Ashes of the Abhorrent", "Enchantment", oldB230AshesOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, mine)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 || opp.Life != 40 {
		t.Errorf("my life %d → %d (want +2), opponent %d", life, me.Life, opp.Life)
	}
}

func TestOldB230AshesStopsGraveyardCastsAndAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dragon := effectstest.PushGraveyardPermanent(me, "Earthquake Dragon", "Creature — Elemental Dragon", "{14}{G}")
	setGraveOracle(me, dragon, oldB230EarthquakeOracle)
	effectstest.FloatMana(t, g, me, "{C}{C}{G}")
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("without Ashes the graveyard ability should work: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, dragon) {
		t.Fatal("setup: the Dragon did not return")
	}
	// Back in the graveyard with Ashes in play, the same ability is refused.
	effectstest.DftABPush(g, me.ID, "Ashes of the Abhorrent", "Enchantment", oldB230AshesOracle, 0, 0)
	g.WithWriteLock(func() {
		for i, c := range me.Hand.Cards {
			if c.InstanceID == dragon {
				me.Hand.Cards = append(me.Hand.Cards[:i], me.Hand.Cards[i+1:]...)
				me.Graveyard.PushTop(c)
				break
			}
		}
	})
	land2 := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.FloatMana(t, g, me, "{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land2}}); err == nil {
		t.Error("an ability of a card in a graveyard was activated with Ashes in play")
	}
}

func setGraveOracle(p *game.Player, id uuid.UUID, oracle string) {
	for i := range p.Graveyard.Cards {
		if p.Graveyard.Cards[i].InstanceID == id {
			p.Graveyard.Cards[i].OracleID = oracle
		}
	}
}

// --- Sword-Point Diplomacy --------------------------------------------------

// oldB230Offers answers every open any-player link: `pays` says whether the
// opponent asked pays 3 life. It returns how many links were asked per seat.
func oldB230Offers(t *testing.T, g *game.Game, pays func(chooser uuid.UUID, link int) bool) map[uuid.UUID]int {
	t.Helper()
	asked := map[uuid.UUID]int{}
	for i := 0; i < 20; i++ {
		link := effectstest.AppOpenLink(g)
		if link == nil {
			effectstest.PassPriorityAroundTable(t, g)
			if link = effectstest.AppOpenLink(g); link == nil {
				return asked
			}
		}
		asked[link.Chooser]++
		if err := g.ResolvePayUnless(link.ID, link.Chooser, pays(link.Chooser, i)); err != nil {
			t.Fatalf("answer: %v", err)
		}
	}
	t.Fatal("the offers never ended")
	return nil
}

func oldB230Top3(me *game.Player) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Card " + string(rune('A'+i)), TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		ids = append(ids, id)
	}
	// PushTop leaves the last pushed on top: top card first.
	return []uuid.UUID{ids[2], ids[1], ids[0]}
}

func TestOldB230SwordPointDiplomacyGivesYouTheCardsNobodyPaysFor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	top := oldB230Top3(me)
	oppLife := opp.Life
	oldB230Cast(t, g, me, game.Card{Name: "Sword-Point Diplomacy", TypeLine: "Sorcery", OracleID: oldB230DiplomacyOracle}, game.CastSpellParams{})
	// The first opponent in turn order pays for the second card only.
	n := 0
	asked := oldB230Offers(t, g, func(chooser uuid.UUID, _ int) bool {
		n++
		return chooser == opp.ID && n == 4 // links: card 1 (3 opponents), then card 2's first
	})
	if asked[me.ID] != 0 {
		t.Errorf("the caster was asked %d times", asked[me.ID])
	}
	if !effectstest.HandHasID(me, top[0]) || effectstest.HandHasID(me, top[1]) || !effectstest.HandHasID(me, top[2]) {
		t.Errorf("hand has first=%v second=%v third=%v; want first and third only",
			effectstest.HandHasID(me, top[0]), effectstest.HandHasID(me, top[1]), effectstest.HandHasID(me, top[2]))
	}
	if opp.Life != oppLife-3 {
		t.Errorf("opponent life %d → %d, want -3", oppLife, opp.Life)
	}
	if !oldB230Zone(g, top[1], game.ZoneExile) {
		t.Error("the card paid for was not exiled with the rest")
	}
}

func TestOldB230SwordPointDiplomacyTakesAllThreeWhenNobodyPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	top := oldB230Top3(me)
	oldB230Cast(t, g, me, game.Card{Name: "Sword-Point Diplomacy", TypeLine: "Sorcery", OracleID: oldB230DiplomacyOracle}, game.CastSpellParams{})
	oldB230Offers(t, g, func(uuid.UUID, int) bool { return false })
	for i, id := range top {
		if !effectstest.HandHasID(me, id) {
			t.Errorf("card %d is not in hand", i)
		}
	}
}

// --- Sheltering Light -------------------------------------------------------

func TestOldB230ShelteringLightGrantsIndestructibleAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB230Cast(t, g, me, game.Card{Name: "Sheltering Light", TypeLine: "Instant", OracleID: oldB230ShelterOracle},
		game.CastSpellParams{Targets: effectstest.DftABCardRef(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, bear); !game.HasKeyword(&c, "indestructible") {
		t.Error("the Bear is not indestructible")
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil {
		t.Error("no scry prompt")
	}
}

// --- Run Aground ------------------------------------------------------------

func TestOldB230RunAgroundTucksAnArtifactOrCreatureOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	if err := g.CastSpell(me.ID, oldB230HandSpell(me, "Run Aground", "Instant", oldB230AgroundOracle), game.CastSpellParams{Strict: true, Targets: effectstest.DftABCardRef(land)}); err == nil {
		t.Error("a land was targeted")
	}
	for _, id := range []uuid.UUID{rock, bear} {
		oldB230Cast(t, g, me, game.Card{Name: "Run Aground", TypeLine: "Instant", OracleID: oldB230AgroundOracle}, game.CastSpellParams{Targets: effectstest.DftABCardRef(id)})
		effectstest.PassPriorityAroundTable(t, g)
		if !oldB230Zone(g, id, game.ZoneLibrary) || opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID != id {
			t.Errorf("the permanent is not on top of its owner's library")
		}
	}
}

func oldB230HandSpell(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Hijack -----------------------------------------------------------------

func TestOldB230HijackStealsUntapsAndGivesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	oldB230Cast(t, g, me, game.Card{Name: "Hijack", TypeLine: "Sorcery", OracleID: oldB230HijackOracle}, game.CastSpellParams{Targets: effectstest.DftABCardRef(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.B12Card(t, g, bear)
	if c.Controller != me.ID || c.Tapped || !game.HasKeyword(&c, "haste") {
		t.Errorf("controller mine=%v tapped=%v haste=%v; want mine, untapped, haste", c.Controller == me.ID, c.Tapped, game.HasKeyword(&c, "haste"))
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	if c := effectstest.B12Card(t, g, bear); c.Controller != opp.ID {
		t.Error("control did not return at end of turn")
	}
}
