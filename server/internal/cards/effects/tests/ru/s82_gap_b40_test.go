package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b40_test.go — Assassin's Creed cards of S82 GAP batch 40 (#5948).
// Each test plays the real catalog entry.

const (
	gapb40FortressOracle   = "5c3f45cb-d2f0-4dac-b124-18e8f6d7e2dc"
	gapb40LaylaOracle      = "e27feb69-3120-4e8d-8ee6-0f87b58e2ae3"
	gapb40SenuOracle       = "bf2dc318-7b51-451e-ac4d-0f1b756dc860"
	gapb40TaxOracle        = "66837faa-fbde-456d-b00f-eb17fbdd04aa"
	gapb40TemplarOracle    = "f9453fe2-fadf-4cd4-8d2c-0eaa0e2d78d6"
	gapb40ConspiracyOracle = "775d284d-0884-4ae3-a1b5-2b2a50a895b4"
	gapb40CoverOracle      = "d78654fc-7d90-40ce-9c36-a211424d1a88"
	gapb40RevelationsOracl = "c7fe1f05-f412-4991-98bd-f6a4b8b62501"
	gapb40BypassOracle     = "ba0d177f-c8e9-4db3-b5cf-40ed2b63e111"
	gapb40BleedingOracle   = "689c9a49-d0f5-494a-9fa2-948cd5517e74"
	gapb40AesirOracle      = "6afe197b-0779-4284-a192-357a3c89b7e3"
)

func TestGapB40CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapb40FortressOracle:   "Escarpment Fortress",
		gapb40LaylaOracle:      "Layla Hassan",
		gapb40SenuOracle:       "Senu, Keen-Eyed Protector",
		gapb40TaxOracle:        "Tax Collector",
		gapb40TemplarOracle:    "Templar Knight",
		gapb40ConspiracyOracle: "Conspiracy",
		gapb40CoverOracle:      "Cover of Darkness",
		gapb40RevelationsOracl: "The Revelations of Ezio",
		gapb40BypassOracle:     "Rooftop Bypass",
		gapb40BleedingOracle:   "Bleeding Effect",
		gapb40AesirOracle:      "The Aesir Escape Valhalla",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

func gapb40Grave(g *game.Game, p *game.Player, name, typeLine, cost string, keywords ...string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Keywords: keywords, Owner: p.ID, Controller: p.ID})
		g.BumpLayerVersionForTest()
	})
	return id
}

// --- Escarpment Fortress ------------------------------------------------------

func TestGapB40EscarpmentFortressPumpsOthersAndDrawsForAWideAttack(t *testing.T) {
	for _, attackers := range []int{1, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		fort := effectstest.B12Push(g, me.ID, "Escarpment Fortress", "Creature — Wall", gapb40FortressOracle, 3, 5)
		for _, kw := range []string{"defender", "reach"} {
			if !effectstest.EffectiveAbilitiesContain(t, g, fort, kw) {
				t.Errorf("the Fortress has no %s", kw)
			}
		}
		a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
		b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		if p, _ := effectstest.EffectivePTOf(t, g, a); p != 3 {
			t.Errorf("my creature's power is %d, want 3", p)
		}
		if p, _ := effectstest.EffectivePTOf(t, g, fort); p != 3 {
			t.Errorf("the Fortress's own power is %d, want 3", p)
		}
		if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 {
			t.Errorf("an opponent's creature got the bonus: power %d", p)
		}
		hand := me.Hand.Size()
		if attackers == 1 {
			effectstest.AttackWith(t, g, opp.ID, a)
		} else {
			effectstest.AttackWith(t, g, opp.ID, a, b)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if attackers == 2 {
			want = hand + 1
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("%d attacker(s): hand %d -> %d, want %d", attackers, hand, got, want)
		}
	}
}

// --- Tax Collector ------------------------------------------------------------

func TestGapB40TaxCollectorTaxTaxesOpponentsSpellsUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastHoldCreature(t, g, "Tax Collector", "Creature — Human Advisor", gapb40TaxOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	price := func(caster uuid.UUID) string {
		base, err := game.ParseCost("{1}{R}")
		if err != nil {
			t.Fatal(err)
		}
		var out string
		g.WithWriteLock(func() {
			card := game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", ManaCost: "{1}{R}", Owner: caster, Controller: caster}
			cost, cerr := g.ApplyCostModifiersForEffect(base, game.CostQuery{Game: g, Card: card, Controller: caster, FromZone: game.ZoneHand})
			if cerr != nil {
				t.Fatalf("ApplyCostModifiers: %v", cerr)
			}
			out = cost.String()
		})
		return out
	}
	if got := price(opp.ID); got != "{2}{R}" {
		t.Errorf("an opponent's spell costs %s, want {2}{R}", got)
	}
	if got := price(me.ID); got != "{1}{R}" {
		t.Errorf("my own spell costs %s, want {1}{R}", got)
	}
	effectstest.PassTurnsTo(t, g, 1)
	if got := price(opp.ID); got != "{2}{R}" {
		t.Errorf("on their own turn their spell costs %s, want {2}{R} (until MY next turn)", got)
	}
	effectstest.PassTurnsTo(t, g, 0)
	if got := price(opp.ID); got != "{1}{R}" {
		t.Errorf("after my next turn begins their spell costs %s, want {1}{R}", got)
	}
}

func TestGapB40TaxCollectorArrestDetainsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Tax Collector", "Creature — Human Advisor", gapb40TaxOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Arrest")
	}
	if effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("my own creature was offered")
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.RestrictionsOf(t, g, theirs)
	if got&game.CantAttackOrBlock != game.CantAttackOrBlock {
		t.Errorf("restrictions = %v, want can't attack or block", got.Names())
	}
}

// --- Bleeding Effect ----------------------------------------------------------

func TestGapB40BleedingEffectGivesTheKeywordsOfCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Bleeding Effect", OracleID: gapb40BleedingOracle, TypeLine: "Enchantment"})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	gapb40Grave(g, me, "Dead Bird", "Creature — Bird", "{1}", "flying")
	gapb40Grave(g, me, "Dead Vampire", "Creature — Vampire", "{1}", "lifelink", "vigilance")
	gapb40Grave(g, me, "Dead Wurm", "Sorcery", "{1}", "trample") // not a creature card
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"flying", "lifelink", "vigilance"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("my creature did not gain %s", kw)
		}
	}
	for _, kw := range []string{"trample", "reach", "menace", "first strike"} {
		if effectstest.EffectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("my creature gained %s, which no creature card in the graveyard has", kw)
		}
	}
}

// --- Rooftop Bypass -----------------------------------------------------------

func TestGapB40RooftopBypassCountsNontokenCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Rooftop Bypass", OracleID: gapb40BypassOracle, TypeLine: "Enchantment"})
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	token := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1})
	effectstest.AttackWith(t, g, opp.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapb39Assassins40(g, me.ID); got != 0 {
		t.Fatalf("a token connecting made %d Assassins", got)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.PushPermanent(g2, me2.ID, game.Card{Name: "Rooftop Bypass", OracleID: gapb40BypassOracle, TypeLine: "Enchantment"})
	a2 := effectstest.B12Creature(g2, me2.ID, "A", "Creature — Bear", 2, 2)
	b2 := effectstest.B12Creature(g2, me2.ID, "B", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g2, opp2.ID, a2, b2)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := gapb39Assassins40(g2, me2.ID); got != 1 {
		t.Errorf("two nontoken creatures connecting made %d Assassins, want 1 (one trigger)", got)
	}
	_, _ = a, b
}

func gapb39Assassins40(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Assassin" && c.IsToken() {
			n++
		}
	}
	return n
}

// --- Conspiracy and Cover of Darkness -----------------------------------------

func gapb40NameTribe(g *game.Game, id uuid.UUID, tribe string) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].NamedTribe = tribe
			}
		}
		g.BumpLayerVersionForTest()
	})
}

func TestGapB40ConspiracyMakesYourCreaturesTheChosenTypeAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	conspiracy := effectstest.B12Push(g, me.ID, "Conspiracy", "Enchantment", gapb40ConspiracyOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Artifact Creature — Bear Warrior", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	inHand := gapb40Hand(me, "Hand Elf", "Creature — Elf Druid", "{1}{G}")
	spell := gapb40Hand(me, "Hand Land", "Land", "")
	if subs := effectstest.EffectiveSubtypes(t, g, mine); !effects.HasFold(subs, "Bear") {
		t.Fatalf("before a type is chosen my creature is %v, want Bear Warrior unchanged", subs)
	}
	gapb40NameTribe(g, conspiracy, "Zombie")
	subs := effectstest.EffectiveSubtypes(t, g, mine)
	if len(subs) != 1 || !effects.HasFold(subs, "Zombie") {
		t.Errorf("my creature is %v, want only Zombie (the other creature types are lost)", subs)
	}
	if types := effectstest.EffectiveTypes(t, g, mine); !effects.HasFold(types, "Artifact") || !effects.HasFold(types, "Creature") {
		t.Errorf("my creature's types are %v, want its card types kept", types)
	}
	if effects.HasFold(effectstest.EffectiveSubtypes(t, g, theirs), "Zombie") {
		t.Error("an opponent's creature became a Zombie")
	}
	handSubs := gapb37HandSubtypes40(t, g, inHand)
	if len(handSubs) != 1 || !effects.HasFold(handSubs, "Zombie") {
		t.Errorf("a creature card in hand is %v, want only Zombie", handSubs)
	}
	if effects.HasFold(gapb37HandSubtypes40(t, g, spell), "Zombie") {
		t.Error("a land card in hand became a Zombie")
	}
}

// gapb37HandSubtypes40 is a card's subtypes with the layers applied, wherever it is.
func gapb37HandSubtypes40(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	var out []string
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("card %s not found", id)
		}
		out = c.Effective().Subtypes
	})
	return out
}

func TestGapB40CoverOfDarknessGivesFearToEveryCreatureOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cover := effectstest.B12Push(g, me.ID, "Cover of Darkness", "Enchantment", gapb40CoverOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectstest.EffectiveAbilitiesContain(t, g, mine, "fear") {
		t.Error("fear before a type is chosen")
	}
	gapb40NameTribe(g, cover, "Goblin")
	for _, id := range []uuid.UUID{mine, theirs} {
		if !effectstest.EffectiveAbilitiesContain(t, g, id, "fear") {
			t.Error("a Goblin has no fear")
		}
	}
	if effectstest.EffectiveAbilitiesContain(t, g, other, "fear") {
		t.Error("a non-Goblin has fear")
	}
}

// --- Layla Hassan -------------------------------------------------------------

func TestGapB40LaylaReturnsAHistoricCardWhenItEntersAndWhenAssassinsConnect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := gapb40Grave(g, me, "Old Relic", "Artifact", "{1}")
	legend := gapb40Grave(g, me, "Old Legend", "Legendary Creature — Human", "{2}")
	gapb40Grave(g, me, "Plain Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastHoldCreature(t, g, "Layla Hassan", "Legendary Creature — Human Assassin", gapb40LaylaOracle, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt when Layla enters")
	}
	if len(p.PickTargetCards) != 2 || !effectstest.HasID(p.PickTargetCards, relic) || !effectstest.HasID(p.PickTargetCards, legend) {
		t.Errorf("targets %v, want only the two historic cards", p.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(relic); z == nil || z.Kind != game.ZoneHand {
		t.Error("the historic card did not return to hand")
	}
	// Two Assassins connecting are one trigger.
	a := effectstest.B12Creature(g, me.ID, "Assassin A", "Creature — Human Assassin", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Assassin B", "Creature — Human Assassin", 1, 1)
	effectstest.AttackWith(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	p = effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt when Assassins connect")
	}
	effectstest.AnswerPickTarget(t, g, legend)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(legend); z == nil || z.Kind != game.ZoneHand {
		t.Error("the second historic card did not return to hand")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("two Assassins connecting triggered Layla twice")
	}
}

// --- Templar Knight -----------------------------------------------------------

func TestGapB40TemplarKnightFetchesALegendaryArtifactWithFiveAttackingKnights(t *testing.T) {
	for _, knights := range []int{4, 5} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		var ids []uuid.UUID
		for i := 0; i < knights; i++ {
			ids = append(ids, effectstest.PushPermanent(g, me.ID, game.Card{Name: "Templar Knight", OracleID: gapb40TemplarOracle, TypeLine: "Creature — Human Knight", Power: 3, Toughness: 1, Keywords: []string{"vigilance"}}))
		}
		relic := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: relic, Name: "Grand Relic", TypeLine: "Legendary Artifact", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Plain Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		for _, id := range ids {
			if err := g.DeclareAttacker(id, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
		}
		effectstest.LockInAttacks(t, g)
		effectstest.FloatForTest(g, me, "W")
		err := g.ActivateCatalogAbility(me.ID, ids[0], 0, game.ActivateAbilityParams{Strict: true, TapIDs: ids})
		if knights < 5 {
			if err == nil {
				t.Errorf("%d Knights: the ability was activated", knights)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%d Knights: activate: %v", knights, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSearchNamed(t, g, me.ID, "Grand Relic")
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(relic) {
			t.Error("the legendary artifact was not put onto the battlefield")
		}
		for _, id := range ids {
			if c, ok := effectstest.BattlefieldCard(g, id); !ok || !c.Tapped {
				t.Error("a Knight was not tapped to pay")
			}
		}
	}
}

// --- Senu, Keen-Eyed Protector ------------------------------------------------

func TestGapB40SenuExilesItselfForLifeAndScry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	senu := effectstest.B12Push(g, me.ID, "Senu, Keen-Eyed Protector", "Legendary Creature — Bird Scout", gapb40SenuOracle, 2, 1)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, senu, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	if z := g.FindCardZoneForEffect(senu); z == nil || z.Kind != game.ZoneExile {
		t.Error("Senu was not exiled")
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil {
		t.Error("no scry 2 prompt")
	}
}

func TestGapB40SenuEntersAttackingWhenALegendaryCreatureIsNotBlocked(t *testing.T) {
	for _, blocked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		senu := uuid.New()
		g.WithWriteLock(func() {
			g.Exile.PushTop(game.Card{
				InstanceID: senu, Name: "Senu, Keen-Eyed Protector", OracleID: gapb40SenuOracle,
				TypeLine: "Legendary Creature — Bird Scout", Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
			})
		})
		legend := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Legend", TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
		var wall uuid.UUID
		if blocked {
			wall = effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(legend, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if blocked {
			if err := g.DeclareBlocker(wall, legend); err != nil {
				t.Fatalf("DeclareBlocker: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		answerOptionPickIfOpen(t, g, me.ID, 0)
		effectstest.PassPriorityAroundTable(t, g)
		// A card put onto the battlefield from exile is a new object: find it by name.
		var c game.Card
		onBattlefield := false
		for _, bc := range g.Battlefield.Cards {
			if bc.Name == "Senu, Keen-Eyed Protector" {
				c, onBattlefield = bc, true
			}
		}
		if blocked {
			if onBattlefield {
				t.Error("Senu entered although the legendary creature was blocked")
			}
			continue
		}
		if !onBattlefield {
			t.Fatal("Senu did not enter")
		}
		if c.AttackingTarget == uuid.Nil {
			t.Error("Senu entered but is not attacking")
		}
	}
}

func answerOptionPickIfOpen(t *testing.T, g *game.Game, who uuid.UUID, index int) {
	t.Helper()
	if ch := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, who); ch != nil {
		if err := g.ResolveOptionPick(ch.ID, who, index); err != nil {
			t.Fatalf("ResolveOptionPick: %v", err)
		}
	}
}

// --- The Revelations of Ezio --------------------------------------------------

func TestGapB40RevelationsOfEzioChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped Ogre", "Creature — Ogre", 3, 3)
	idle := effectstest.B12Creature(g, opp.ID, "Idle Ogre", "Creature — Ogre", 3, 3)
	gapb38TapLike(g, tapped)
	dead := gapb40Grave(g, me, "Fallen Assassin", "Creature — Human Assassin", "{2}")
	saga := effectstest.CastCatalogSpell(t, g, "The Revelations of Ezio", "Enchantment — Saga", gapb40RevelationsOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no chapter I target prompt")
	}
	if effectstest.HasID(p.PickTargetCards, idle) || !effectstest.HasID(p.PickTargetCards, tapped) {
		t.Errorf("chapter I targets %v, want only the tapped creature", p.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, tapped)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("chapter I did not destroy the tapped creature")
	}
	// Chapter II: an Assassin that attacks this turn gets a counter; one
	// that doesn't, and a non-Assassin, don't.
	assassin := effectstest.B12Creature(g, me.ID, "Assassin", "Creature — Human Assassin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, saga, game.CounterLore, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AttackWith(t, g, opp.ID, assassin, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, assassin, game.CounterPlusOne); got != 1 {
		t.Errorf("the attacking Assassin has %d +1/+1 counters, want 1", got)
	}
	if got := effectstest.CounterOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("a non-Assassin got %d counters", got)
	}
	// Chapter III.
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, saga, game.CounterLore, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("chapter III did not return the Assassin card")
	}
	if got := effectstest.CounterOn(g, dead, game.CounterPlusOne); got != 1 {
		t.Errorf("the returned Assassin has %d +1/+1 counters, want 1", got)
	}
}

func gapb38TapLike(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
}

// --- The Aesir Escape Valhalla ------------------------------------------------

func TestGapB40AesirEscapeValhallaChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gapb40Grave(g, me, "Cheap Thing", "Artifact", "{1}")
	big := gapb40Grave(g, me, "Big Thing", "Creature — Giant", "{3}{G}")
	gapb40Grave(g, me, "Plain Spell", "Sorcery", "{1}")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	saga := effectstest.CastCatalogSpell(t, g, "The Aesir Escape Valhalla", "Enchantment — Saga", gapb40AesirOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("chapter I did not ask for a card")
	}
	for _, id := range pick.ChooseCards {
		if c, _ := g.LookupCardForEffect(id); c.Name == "Plain Spell" {
			t.Error("a nonpermanent card was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{big}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 (the exiled card's mana value)", life, me.Life)
	}
	if z := g.FindCardZoneForEffect(big); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the chosen card was not exiled")
	}
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, saga, game.CounterLore, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, bear, game.CounterPlusOne); got != 4 {
		t.Errorf("the creature has %d +1/+1 counters, want 4", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, saga, game.CounterLore, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga stayed on the battlefield after chapter III")
	}
	for _, id := range []uuid.UUID{saga, big} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneHand {
			t.Errorf("%v is not in the hand after chapter III", id)
		}
	}
}

func gapb40Hand(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}
