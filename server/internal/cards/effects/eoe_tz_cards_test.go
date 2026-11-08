package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eoe_tz_cards_test.go — Edge of Eternities cards whose names start with T
// to Z (#4461), the ones the Q–Z batch left. Each test plays the real
// catalog entry and asserts what its text says, and what it refuses.

const (
	eoeTZTannukOracle     = "76cb1912-3791-44b8-b594-b92dab21bafc"
	eoeTZTerrapactOracle  = "c08e0958-8fa2-49c7-ab17-a70dfc1ef33a"
	eoeTZBruntarOracle    = "03f04c2c-1036-497e-8b13-06c61b0b0ba6"
	eoeTZTorpedoOracle    = "9f9ae015-dc59-4ecb-a2bb-668496b48a48"
	eoeTZTractorOracle    = "ed99a0b8-d82d-4c80-af97-cabd333620df"
	eoeTZUnravelOracle    = "1525cbc1-0d95-4f58-99fb-136e54c4d486"
	eoeTZPsionicistOracle = "7f51aede-fb86-4b7e-9027-2ba8be62b567"
	eoeTZScanshipOracle   = "7b4c37dc-8cb0-4870-929e-11c2a45952a2"
	eoeTZGodcoreOracle    = "df08ac72-010f-42f8-beb3-6d645c638e1e"
	eoeTZVaultguardOracle = "e15c78d4-6fb1-4bd3-8219-c66b83fcaba1"
	eoeTZSilencerOracle   = "7d70a2f4-92cb-40c3-9e5a-f6c57d4a64c2"
	eoeTZGunshipOracle    = "c947171b-ed9e-4b83-af45-bd595a8d84ee"
	eoeTZEnhancerOracle   = "4d8ee903-17b3-43c9-9637-ee596955a9f4"
	eoeTZDisplayOracle    = "0ff3afba-7a90-4f62-a1b0-b553ad81c32b"
	eoeTZZookeeperOracle  = "6d4d98a8-a5a3-4cc1-9277-c45dafe40ea0"
)

func TestEoeTZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eoeTZTannukOracle:     "Tannuk, Memorial Ensign",
		eoeTZTerrapactOracle:  "Terrapact Intimidator",
		eoeTZBruntarOracle:    "Territorial Bruntar",
		eoeTZTorpedoOracle:    "Thaumaton Torpedo",
		eoeTZTractorOracle:    "Tractor Beam",
		eoeTZUnravelOracle:    "Unravel",
		eoeTZPsionicistOracle: "Uthros Psionicist",
		eoeTZScanshipOracle:   "Uthros Scanship",
		eoeTZGodcoreOracle:    "Uthros, Titanic Godcore",
		eoeTZVaultguardOracle: "Vaultguard Trooper",
		eoeTZSilencerOracle:   "Virulent Silencer",
		eoeTZGunshipOracle:    "Warmaker Gunship",
		eoeTZEnhancerOracle:   "Weftblade Enhancer",
		eoeTZDisplayOracle:    "Zealous Display",
		eoeTZZookeeperOracle:  "Zookeeper Mechan",
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
	if ac := game.AlternativeCostByKey(eoeTZEnhancerOracle, "warp"); ac == nil || ac.ManaCost != "{2}{W}" {
		t.Errorf("Weftblade Enhancer: warp %+v, want {2}{W}", ac)
	}
}

// --- Tannuk, Memorial Ensign --------------------------------------------------

func TestEoeTZTannukPingsEachLandfallAndDrawsOnTheSecond(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Tannuk, Memorial Ensign", "Legendary Creature — Kavu Pilot", eoeTZTannukOracle, 2, 4)
	start := lifeOfOpponents(g)
	mine := me.Life
	for i, want := range []int{0, 1, 0} {
		hand := me.Hand.Size()
		playLandFromHand(t, g, "Mountain", "")
		passPriorityAroundTable(t, g)
		for j, life := range lifeOfOpponents(g) {
			if life != start[j]-(i+1) {
				t.Errorf("landfall %d: opponent %d at %d life, want %d", i+1, j, life, start[j]-(i+1))
			}
		}
		if me.Life != mine {
			t.Errorf("landfall %d: Tannuk's controller took damage (life %d)", i+1, me.Life)
		}
		if got := me.Hand.Size() - hand; got != want {
			t.Errorf("landfall %d: drew %d cards, want %d", i+1, got, want)
		}
	}
}

// --- Terrapact Intimidator ----------------------------------------------------

func TestEoeTZTerrapactIntimidatorLetsTheOpponentChoose(t *testing.T) {
	for _, tc := range []struct {
		name              string
		yes               bool
		landers, counters int
	}{{"landers", true, 2, 0}, {"counters", false, 0, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			kavu := avB1Cast(t, g, me, game.Card{Name: "Terrapact Intimidator", TypeLine: "Creature — Kavu Scout", ManaCost: "{1}{R}",
				OracleID: eoeTZTerrapactOracle, Power: 2, Toughness: 1}, game.CastSpellParams{})
			b04WaitForPick(t, g, me.ID)
			pickPlayer(t, g, me.ID, opp.ID)
			passPriorityAroundTable(t, g)
			answerMayChoice(t, g, opp.ID, tc.yes)
			passPriorityAroundTable(t, g)
			if got := eoeCDCount(g, me.ID, "Lander"); got != tc.landers {
				t.Errorf("Landers %d, want %d", got, tc.landers)
			}
			if got := eoeCDCount(g, opp.ID, "Lander"); got != 0 {
				t.Errorf("the opponent made %d Landers", got)
			}
			if got := countersOn(g, kavu, game.CounterPlusOne); got != tc.counters {
				t.Errorf("+1/+1 counters %d, want %d", got, tc.counters)
			}
		})
	}
}

// --- Territorial Bruntar ------------------------------------------------------

func TestEoeTZTerritorialBruntarExilesToANonlandCastableThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bruntar := b12Push(g, me.ID, "Territorial Bruntar", "Creature — Beast", eoeTZBruntarOracle, 6, 6)
	if !hasEffectiveKeyword(t, g, bruntar, "reach") {
		t.Error("Territorial Bruntar has reach")
	}
	ids := seedSearchLibrary(me,
		game.Card{Name: "Exiled Land", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	playLandFromHand(t, g, "Mountain", "")
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || !g.Exile.Contains(ids[1]) || !me.Library.Contains(ids[2]) {
		t.Fatal("exile until a nonland card: the land and the Shock, not the third card")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(ids[0]); perm != nil {
		t.Error("the land exiled on the way may be played; only \"that card\" may be cast")
	}
	perm := exiledPermission(g, ids[1])
	if perm == nil || perm.Player != me.ID || !permissionLive(g, perm, me.ID) {
		t.Fatalf("%+v: the nonland card may be cast this turn", perm)
	}
	if !perm.CastOnly || perm.Cost != "" {
		t.Errorf("a cast-only permission at the card's own cost, got cast-only %v, cost %q", perm.CastOnly, perm.Cost)
	}
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if permissionLive(g, exiledPermission(g, ids[1]), me.ID) {
		t.Error("the permission outlasted the turn")
	}
}

// --- Thaumaton Torpedo --------------------------------------------------------

func TestEoeTZThaumatonTorpedoCostsSixWithoutASpacecraftAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	torpedo := b12Push(g, me.ID, "Thaumaton Torpedo", "Artifact", eoeTZTorpedoOracle, 0, 0)
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, torpedo, 0, game.ActivateAbilityParams{Targets: cardRefs(land)}); err == nil {
		t.Fatal("a land is not a nonland permanent")
	}
	floatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, torpedo, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(relic)}); err == nil {
		t.Fatal("{3} paid the full {6} without a Spacecraft attack")
	}
	floatForTest(g, me, "CCC")
	b16Activate(t, g, me.ID, torpedo, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(relic)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || g.Battlefield.Contains(torpedo) {
		t.Error("the target is destroyed and the Torpedo sacrificed")
	}
}

func TestEoeTZThaumatonTorpedoCostsThreeAfterASpacecraftAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	torpedo := b12Push(g, me.ID, "Thaumaton Torpedo", "Artifact", eoeTZTorpedoOracle, 0, 0)
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	ship := b12Creature(g, me.ID, "My Spacecraft", "Artifact Creature — Spacecraft", 2, 2)
	attackIntoBlocks(t, g, opp.ID, ship)
	advanceToStepInTurn(t, g, game.StepPostcombatMain)
	floatForTest(g, me, "CCC")
	b16Activate(t, g, me.ID, torpedo, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(relic)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("after a Spacecraft attacked, {3} pays for the destroy")
	}
}

// --- Tractor Beam -------------------------------------------------------------

func TestEoeTZTractorBeamStealsATappedSpacecraftThatStaysTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := b12Permanent(g, opp.ID, "Their Ship", "Artifact — Spacecraft")
	castCatalogSpell(t, g, "Tractor Beam", "Enchantment — Aura", eoeTZTractorOracle, cardRefs(ship))
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, ship)
	if c.Controller != me.ID || !c.Tapped {
		t.Fatalf("the enchanted Spacecraft: controller %v (want me), tapped %v", c.Controller, c.Tapped)
	}
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	if c := layeredCard(t, g, ship); c.Controller != me.ID || !c.Tapped {
		t.Errorf("after its controller's untap step: controller %v, tapped %v", c.Controller, c.Tapped)
	}
}

func TestEoeTZTractorBeamCannotEnchantALand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := b12Permanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	if err := castCatalogSpellErr(t, g, "Tractor Beam", "Enchantment — Aura", eoeTZTractorOracle, cardRefs(land)); err == nil {
		t.Error("a land is neither a creature nor a Spacecraft")
	}
}

// --- Unravel ------------------------------------------------------------------

// eoeTZOpponentCastsBolt has the opponent cast a {2}{R/P} instant at me,
// paying the Phyrexian symbol with life when `phyrexian` is set.
func eoeTZOpponentCastsBolt(t *testing.T, g *game.Game, phyrexian bool) uuid.UUID {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	params := game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}
	if phyrexian {
		floatForTest(g, opp, "CC")
		params.PhyrexianLife = 1
	} else {
		floatForTest(g, opp, "CCR")
	}
	victim := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{
			InstanceID: victim, Name: "Phyrexian Bolt", TypeLine: "Instant", ManaCost: "{2}{R/P}",
			Owner: opp.ID, Controller: opp.ID,
		})
	})
	if err := g.CastSpell(opp.ID, victim, params); err != nil {
		t.Fatalf("opponent casts: %v", err)
	}
	return victim
}

func TestEoeTZUnravelDrawsWhenLessManaThanItsValueWasSpent(t *testing.T) {
	for _, tc := range []struct {
		name      string
		phyrexian bool
		drawn     int
	}{{"paid with life", true, 1}, {"paid in full", false, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			victim := eoeTZOpponentCastsBolt(t, g, tc.phyrexian)
			castCatalogSpell(t, g, "Unravel", "Instant", eoeTZUnravelOracle, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
			hand := me.Hand.Size()
			passPriorityAroundTable(t, g)
			if !opp.Graveyard.Contains(victim) {
				t.Fatal("the spell was not countered")
			}
			if got := me.Hand.Size() - hand; got != tc.drawn {
				t.Errorf("drew %d cards, want %d", got, tc.drawn)
			}
		})
	}
}

// --- Uthros Psionicist --------------------------------------------------------

func TestEoeTZUthrosPsionicistDiscountsOnlyTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Uthros Psionicist", "Creature — Jellyfish Scientist", eoeTZPsionicistOracle, false)
	if got := priceInHand(t, g, me, "Cultivate", "Sorcery", "{3}{G}"); got != 4 {
		t.Errorf("first spell: %d, want 4 (undiscounted)", got)
	}
	b24CastNoncreature(t, g, "First", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if got := priceInHand(t, g, me, "Cultivate", "Sorcery", "{3}{G}"); got != 2 {
		t.Errorf("second spell: %d, want 2 ({2} less)", got)
	}
	b24CastNoncreature(t, g, "Second", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if got := priceInHand(t, g, me, "Cultivate", "Sorcery", "{3}{G}"); got != 4 {
		t.Errorf("third spell: %d, want 4 (undiscounted)", got)
	}
}

// --- Uthros Scanship ----------------------------------------------------------

func TestEoeTZUthrosScanshipLootsTwoAndFliesAtEight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ship := avB1Cast(t, g, me, game.Card{Name: "Uthros Scanship", TypeLine: "Artifact — Spacecraft", ManaCost: "{3}{U}",
		OracleID: eoeTZScanshipOracle, Power: 4, Toughness: 4}, game.CastSpellParams{})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("draw two then discard one: hand changed by %d, want 1", got)
	}
	if err := g.AddCounter(ship, game.CounterCharge, 7); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("below 8 charge counters it is not a creature")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at8 := layeredCard(t, g, ship)
	if !at8.IsCreature() || !game.HasKeyword(&at8, "flying") || at8.CurrentPower() != 4 || at8.CurrentToughness() != 4 {
		t.Errorf("at 8+ it is a 4/4 flying creature, got %d/%d", at8.CurrentPower(), at8.CurrentToughness())
	}
}

// --- Uthros, Titanic Godcore --------------------------------------------------

func TestEoeTZUthrosGodcoreTapsForEachArtifactAtTwelve(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	uthros := b12Push(g, me.ID, "Uthros, Titanic Godcore", "Land — Planet", eoeTZGodcoreOracle, 0, 0)
	b12Permanent(g, me.ID, "Relic", "Artifact")
	b12Permanent(g, me.ID, "Bauble", "Artifact")
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	toMainForCost(t, g)

	sweepFAddMana(g, me.ID, "{U}")
	if err := g.ActivateManaAbility(me.ID, uthros, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the 12+ ability activated with no charge counters")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(uthros, game.CounterCharge, 12) })
	activateManaFor(t, g, me.ID, uthros, 1, game.ManaAbilityParams{})
	if got := me.ManaPool.CountColor("U"); got != 2 {
		t.Errorf("pool has %d {U} after {U}, {T}: want two for two artifacts", got)
	}
}

// --- Vaultguard Trooper -------------------------------------------------------

func TestEoeTZVaultguardTrooperDiscardsTheHandToDrawTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trooper := b12Push(g, me.ID, "Vaultguard Trooper", "Creature — Kavu Soldier", eoeTZVaultguardOracle, 5, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{trooper, bear} {
		eoeCDSet(g, id, func(c *game.Card) { c.Tapped = true })
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	hand, graveyard := me.Hand.Size(), me.Graveyard.Size()
	if hand == 0 {
		t.Fatal("the test needs cards in hand")
	}
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand %d, want 2", got)
	}
	if got := me.Graveyard.Size() - graveyard; got != hand {
		t.Errorf("discarded %d cards, want the whole hand of %d", got, hand)
	}
}

func TestEoeTZVaultguardTrooperNeedsTwoTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trooper := b12Push(g, me.ID, "Vaultguard Trooper", "Creature — Kavu Soldier", eoeTZVaultguardOracle, 5, 5)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	eoeCDSet(g, trooper, func(c *game.Card) { c.Tapped = true })
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if c := latestConfirmFor(g, me.ID); c != nil {
		t.Errorf("one tapped creature asked to discard: %+v", c)
	}
}

// --- Virulent Silencer --------------------------------------------------------

func TestEoeTZVirulentSilencerPoisonsForEachNontokenArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	silencer := b12Push(g, me.ID, "Virulent Silencer", "Artifact Creature — Robot Assassin", eoeTZSilencerOracle, 2, 3)
	golem := b12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	thopter := b12Creature(g, me.ID, "Thopter", "Token Artifact Creature — Thopter", 1, 1)
	attackIntoBlocks(t, g, them.ID, silencer, golem, bear, thopter)
	advanceToStepInTurn(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := them.Counters[game.CounterPoison]; got != 4 {
		t.Errorf("poison counters %d, want 4 (two nontoken artifact creatures connected)", got)
	}
	if got := me.Counters[game.CounterPoison]; got != 0 {
		t.Errorf("the controller got %d poison counters", got)
	}
}

// --- Warmaker Gunship ---------------------------------------------------------

func TestEoeTZWarmakerGunshipDealsItsArtifactCountAndFliesAtSix(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Permanent(g, me.ID, "Relic", "Artifact")
	b12Permanent(g, me.ID, "Bauble", "Artifact")
	b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 2, 4)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ship := avB1Cast(t, g, me, game.Card{Name: "Warmaker Gunship", TypeLine: "Artifact — Spacecraft", ManaCost: "{2}{R}",
		OracleID: eoeTZGunshipOracle, Power: 4, Toughness: 3}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil && hasID(p.PickTargetCards, mine) {
		t.Error("the controller's own creature is offered")
	}
	pickIfAsked(t, g, me.ID, giant)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, giant); !ok || c.DamageMarked != 3 {
		t.Errorf("the Giant has %d damage, want 3 (two artifacts and the Gunship)", c.DamageMarked)
	}
	if err := g.AddCounter(ship, game.CounterCharge, 6); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at6 := layeredCard(t, g, ship)
	if !at6.IsCreature() || !game.HasKeyword(&at6, "flying") || at6.CurrentPower() != 4 || at6.CurrentToughness() != 3 {
		t.Errorf("at 6+ it is a 4/3 flying creature, got %d/%d", at6.CurrentPower(), at6.CurrentToughness())
	}
}

// --- Weftblade Enhancer -------------------------------------------------------

func TestEoeTZWeftbladeEnhancerCountersUpToTwoAndWarpsOut(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "My Ox", "Creature — Ox", 2, 2)
	enhancer := avB1Cast(t, g, me, game.Card{Name: "Weftblade Enhancer", TypeLine: "Creature — Drix Artificer", ManaCost: "{5}{W}",
		OracleID: eoeTZEnhancerOracle, Power: 3, Toughness: 4}, game.CastSpellParams{AlternativeCost: "warp"})
	passPriorityAroundTable(t, g)
	pickIfAsked(t, g, me.ID, bear, theirs)
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{bear: 1, theirs: 1, other: 0} {
		if got := countersOn(g, id, game.CounterPlusOne); got != want {
			t.Errorf("%s: %d +1/+1 counters, want %d", id, got, want)
		}
	}
	advanceThroughEndStep(t, g)
	if !g.Exile.Contains(enhancer) {
		t.Error("the warped Enhancer is exiled at the end step")
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the counter stays on the Bear: %d", got)
	}
}

// --- Zealous Display ----------------------------------------------------------

func TestEoeTZZealousDisplayPumpsWithoutUntappingOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	eoeCDSet(g, bear, func(c *game.Card) { c.Tapped = true })
	castCatalogSpell(t, g, "Zealous Display", "Instant", eoeTZDisplayOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := eoeCDPT(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("my Bear is %d/%d, want 4/2", p, tough)
	}
	if p, _ := eoeCDPT(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature was pumped to power %d", p)
	}
	if !b16Tapped(t, g, bear) {
		t.Error("cast on its controller's own turn, it untaps nothing")
	}
}

func TestEoeTZZealousDisplayUntapsThoseCreaturesOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Active Bear", "Creature — Bear", 2, 2)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	eoeCDSet(g, attacker, func(c *game.Card) { c.Tapped = true })
	eoeCDSet(g, bear, func(c *game.Card) { c.Tapped = true })
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Zealous Display", TypeLine: "Instant", OracleID: eoeTZDisplayOracle,
			Owner: opp.ID, Controller: opp.ID})
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("the opponent casts Zealous Display: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := eoeCDPT(t, g, bear); p != 4 {
		t.Errorf("the caster's Bear has power %d, want 4", p)
	}
	if b16Tapped(t, g, bear) {
		t.Error("cast on another player's turn, it untaps the caster's creatures")
	}
	if !b16Tapped(t, g, attacker) {
		t.Error("the active player's creature was untapped")
	}
}

// --- Zookeeper Mechan ---------------------------------------------------------

func TestEoeTZZookeeperMechanTapsForRedAndPumpsAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mechan := b12Push(g, me.ID, "Zookeeper Mechan", "Artifact Creature — Robot", eoeTZZookeeperOracle, 1, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	activateManaFor(t, g, me.ID, mechan, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.CountColor("R"); got != 1 {
		t.Fatalf("pool has %d {R}, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, mechan, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err == nil {
		t.Fatal("an opponent's creature is not a creature you control")
	}
	floatForTest(g, me, "CCCCCC")
	b16Activate(t, g, me.ID, mechan, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := eoeCDPT(t, g, bear); p != 6 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 6/2", p, tough)
	}
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, mechan, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Error("activated outside a main phase")
	}
}
