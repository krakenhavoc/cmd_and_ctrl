package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dis_5598_b1_cards_test.go — Dissension batch 1 (#5598). Every test
// plays the real catalog entry through the engine.

const (
	disAssaultZeppelidOracle      = "8049dbff-6b57-454a-a2e6-7e2ae39bb606"
	disAzoriusFirstWingOracle     = "08d02fdb-4b3f-4f55-b5fc-6bb4d2a14b1f"
	disSimicSkySwallowerOracle    = "e5c48c6d-a7bf-45b3-b48c-dc884d49c5f7"
	disEnemyOfTheGuildpactOracle  = "e5cf7b75-b59b-4cac-bfd2-13e452457d74"
	disGnatAlleyCreeperOracle     = "17f930a9-64ee-46f9-8a7c-8877e532c94f"
	disSimicInitiateOracle        = "c7c37369-47f8-4027-a016-a2055307dd24"
	disVigeanHydroponOracle       = "3eab4c00-a8ae-4912-bee1-68376de700a3"
	disAquastrandSpiderOracle     = "cea79fa0-2a76-41f3-846f-4d1a6ff5f629"
	disCytospawnShamblerOracle    = "514c6831-63ac-4fc0-8950-526813591710"
	disPlaxcasterFroglingOracle   = "8c1ea1a5-5966-4d14-8361-ffc6a8d1bd79"
	disVigeanGraftmageOracle      = "4c209165-a159-4cd7-8e99-1e533d3a22a6"
	disSporebackTrollOracle       = "339bddf5-a8fb-46d3-877e-b0441ca8a597"
	disAzoriusHeraldOracle        = "a0476da9-51b1-4cd3-90c4-ad01d0e4c3d6"
	disPlaxmantaOracle            = "36780340-824d-420b-b41b-182f88d944de"
	disPatagiaViperOracle         = "b9985370-d908-431c-ad33-ec532fe07d1f"
	disAuroraEidolonOracle        = "4fc8a155-4eb8-4b86-8ccc-2569630d3bfa"
	disEnigmaEidolonOracle        = "4489a526-0e07-4d37-9e83-628f7eb256c6"
	disEntropicEidolonOracle      = "be5f788a-8a2e-447c-9e65-f9fc197ddfc9"
	disSandstormEidolonOracle     = "84fe1c59-3f7b-4c1c-9bc4-78a185cfcc0c"
	disVerdantEidolonOracle       = "a3acc388-4015-42e4-88ff-b35a91134d12"
	disMinisterOfImpedimentsOracl = "5a51a0ee-bc81-4c42-bb17-a41cea957d9a"
	disSimicRagwormOracle         = "5d8c4172-eb14-47a4-993a-6d68906a26c5"
	disRakdosIckspitterOracle     = "329f6e2e-6c37-4e98-bd45-bd252f1a2bd9"
	disHaazdaExoneratorOracle     = "6df2e8c9-d87e-4f0d-a3d3-b8eee5cdafd0"
	disSkullmeadCauldronOracle    = "b6864a26-9c13-4c6d-b0e0-d5b5ed9864e8"
	disSilkwingScoutOracle        = "4c76db22-1992-448d-be4a-a9095420b8cc"
	disRixMaadiOracle             = "a1bee68d-135b-4e30-8830-48a3315d13a9"
	disOgreGatecrasherOracle      = "0b88bc94-99c4-4be5-b474-a7dad68057b3"
	disWhiptailMolochOracle       = "d57079ae-7760-4a81-895f-dc8990896ed7"
	disProperBurialOracle         = "e5bf77c2-a9b2-4366-8412-e263160403c1"
	disBeaconHawkOracle           = "edf7c98a-c2ed-4ba2-ac08-e83db0d968cb"
	disStompAndHowlOracle         = "7cc809ed-5ce1-40d1-896d-014ca8e2b5a6"
	disWreckingBallOracle         = "0a1b0910-3b0c-4aa3-b47f-8eb069122743"
	disWitsEndOracle              = "97f11aed-7a5c-4a14-8c2c-5c230c1f05b4"
	disDeliriumSkeinsOracle       = "7397036f-8114-47cc-b52f-c532a6845d16"
	disPsychoticFuryOracle        = "1bf982e9-5868-46bf-81e3-f29ada33d169"
	disCacklingFlamesOracle       = "ab48147c-0bd7-4875-8e11-0e26185453dd"
	disBondOfAgonyOracle          = "8e7bd5ae-005a-47d8-90bc-69f48cd8793d"
	disDemonsJesterOracle         = "da67c329-ca12-45fc-a9cd-8fb6f27501a9"
	disRiotSpikesOracle           = "3df676fe-c4f8-41c0-9743-781090bdc547"

	disBearType   = "Creature — Bear"
	disSpiritType = "Creature — Spirit"
)

// disCast casts a card from p's hand with a real, strict payment out of
// the listed mana, in a main phase, and returns it.
func disCast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle string, power, toughness int, mana ...string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(p, mana...)
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
		})
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// disPushWithCounters seeds a permanent carrying n +1/+1 counters.
func disPushWithCounters(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness, n int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Counters: map[string]int{game.CounterPlusOne: n},
		Owner: owner, Controller: owner,
	})
}

// disGraveyardCard puts a catalog card into p's graveyard.
func disGraveyardCard(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id
}

// disSetTapped taps a permanent directly.
func disSetTapped(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = true })
}

// --- every card is in the catalog ----------------------------------------

func TestDis5598B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		disAssaultZeppelidOracle, disAzoriusFirstWingOracle, disSimicSkySwallowerOracle,
		disEnemyOfTheGuildpactOracle, disGnatAlleyCreeperOracle, disSimicInitiateOracle, disVigeanHydroponOracle,
		disAquastrandSpiderOracle, disCytospawnShamblerOracle, disPlaxcasterFroglingOracle, disVigeanGraftmageOracle,
		disSporebackTrollOracle, disAzoriusHeraldOracle, disPlaxmantaOracle, disPatagiaViperOracle,
		disAuroraEidolonOracle, disEnigmaEidolonOracle, disEntropicEidolonOracle, disSandstormEidolonOracle,
		disVerdantEidolonOracle, disMinisterOfImpedimentsOracl, disSimicRagwormOracle, disRakdosIckspitterOracle,
		disHaazdaExoneratorOracle, disSkullmeadCauldronOracle, disSilkwingScoutOracle, disRixMaadiOracle,
		disOgreGatecrasherOracle, disWhiptailMolochOracle, disProperBurialOracle, disBeaconHawkOracle,
		disStompAndHowlOracle, disWreckingBallOracle, disWitsEndOracle, disDeliriumSkeinsOracle,
		disPsychoticFuryOracle, disCacklingFlamesOracle, disBondOfAgonyOracle, disDemonsJesterOracle,
		disRiotSpikesOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- printed keywords -----------------------------------------------------------

func TestDis5598KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Assault Zeppelid", "Creature — Beast", disAssaultZeppelidOracle, []string{"flying", "trample"}},
		{"Azorius First-Wing", "Creature — Griffin", disAzoriusFirstWingOracle, []string{"flying", "protection from enchantments"}},
		{"Simic Sky Swallower", "Creature — Leviathan", disSimicSkySwallowerOracle, []string{"flying", "trample", "shroud"}},
		{"Enemy of the Guildpact", disSpiritType, disEnemyOfTheGuildpactOracle, []string{"protection from multicolored"}},
		{"Silkwing Scout", "Creature — Faerie Scout", disSilkwingScoutOracle, []string{"flying"}},
		{"Beacon Hawk", "Creature — Bird", disBeaconHawkOracle, []string{"flying"}},
		{"Demon's Jester", "Creature — Imp", disDemonsJesterOracle, []string{"flying"}},
		{"Patagia Viper", "Creature — Snake", disPatagiaViperOracle, []string{"flying"}},
		{"Plaxmanta", "Creature — Beast", disPlaxmantaOracle, []string{"flash"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

func TestDis5598GnatAlleyCreeperCantBeBlockedByFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creeper := effectstest.B12Push(g, me.ID, "Gnat Alley Creeper", "Creature — Human Rogue", disGnatAlleyCreeperOracle, 3, 1)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	walker := effectstest.B12Creature(g, opp.ID, "Walker", disBearType, 2, 2)
	effectstest.BrAttack(t, g, creeper)
	if err := g.DeclareBlocker(flyer, creeper); err == nil {
		t.Error("a creature with flying blocked the Creeper")
	}
	if err := g.DeclareBlocker(walker, creeper); err != nil {
		t.Errorf("a creature without flying can't block the Creeper: %v", err)
	}
}

// --- graft -------------------------------------------------------------------------

func TestDis5598GraftCreaturesEnterWithTheirCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Simic Initiate", "Creature — Human Mutant", disSimicInitiateOracle, 1},
		{"Vigean Hydropon", "Creature — Plant Mutant", disVigeanHydroponOracle, 5},
		{"Aquastrand Spider", "Creature — Spider Mutant", disAquastrandSpiderOracle, 2},
		{"Cytospawn Shambler", "Creature — Elemental Mutant", disCytospawnShamblerOracle, 6},
		{"Plaxcaster Frogling", "Creature — Frog Mutant", disPlaxcasterFroglingOracle, 3},
		{"Vigean Graftmage", "Creature — Vedalken Wizard Mutant", disVigeanGraftmageOracle, 2},
		{"Sporeback Troll", "Creature — Troll Mutant", disSporebackTrollOracle, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			id := effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s did not resolve", tc.name)
			}
			if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != tc.n {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.n)
			}
		})
	}
}

func TestDis5598SimicInitiateGraftsOntoANewCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	initiate := effectstest.CastCatalogSpell(t, g, "Simic Initiate", "Creature — Human Mutant", disSimicInitiateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.CastCatalogSpell(t, g, "Bear", disBearType, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, initiate, game.CounterPlusOne) != 0 || effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Errorf("graft: initiate %d, bear %d; want 0 and 1",
			effectstest.CountersOn(g, initiate, game.CounterPlusOne), effectstest.CountersOn(g, bear, game.CounterPlusOne))
	}
}

func TestDis5598VigeanHydroponCantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	hydropon := disPushWithCounters(g, me.ID, "Vigean Hydropon", "Creature — Plant Mutant", disVigeanHydroponOracle, 0, 0, 5)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	if err := g.DeclareAttacker(hydropon, opp.ID); err == nil {
		t.Error("Vigean Hydropon attacked")
	}
}

// The three "gains <keyword>" rows reach only a creature with a +1/+1
// counter on it.
func TestDis5598GraftGrantRowsNeedACounteredTarget(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
		mana                            []string
	}{
		{"Aquastrand Spider", "Creature — Spider Mutant", disAquastrandSpiderOracle, "reach", []string{"G"}},
		{"Cytospawn Shambler", "Creature — Elemental Mutant", disCytospawnShamblerOracle, "trample", []string{"G"}},
		{"Plaxcaster Frogling", "Creature — Frog Mutant", disPlaxcasterFroglingOracle, "shroud", []string{"C", "C"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			src := disPushWithCounters(g, me.ID, tc.name, tc.typeLine, tc.oracle, 0, 0, 2)
			plain := effectstest.B12Creature(g, me.ID, "Plain Bear", disBearType, 2, 2)
			countered := disPushWithCounters(g, me.ID, "Counter Bear", disBearType, "", 2, 2, 1)
			effectstest.ApaMana(me, tc.mana...)
			if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(plain)}); err == nil {
				t.Fatal("a creature without a +1/+1 counter was a legal target")
			}
			effectstest.ApaActivate(t, g, me, src, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(countered)})
			if !effectstest.HasEffectiveKeyword(t, g, countered, tc.keyword) {
				t.Errorf("the target has no %s", tc.keyword)
			}
		})
	}
}

func TestDis5598VigeanGraftmageUntapsACounteredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	mage := disPushWithCounters(g, me.ID, "Vigean Graftmage", "Creature — Vedalken Wizard Mutant", disVigeanGraftmageOracle, 0, 0, 2)
	countered := disPushWithCounters(g, me.ID, "Counter Bear", disBearType, "", 2, 2, 1)
	disSetTapped(g, countered)
	effectstest.ApaMana(me, "U", "C")
	effectstest.ApaActivate(t, g, me, mage, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(countered)})
	if effectstest.HmlTapped(t, g, countered) {
		t.Error("the target is still tapped")
	}
}

func TestDis5598SporebackTrollRegeneratesACounteredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	troll := disPushWithCounters(g, me.ID, "Sporeback Troll", "Creature — Troll Mutant", disSporebackTrollOracle, 0, 0, 2)
	effectstest.ApaMana(me, "G", "C")
	effectstest.ApaActivate(t, g, me, troll, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(troll)})
	effectstest.B27Kill(g, troll)
	if !g.Battlefield.Contains(troll) {
		t.Error("the shielded Troll was destroyed")
	}
}

// --- sacrifice unless {C} was spent -------------------------------------------------

func TestDis5598SacrificeUnlessTheColorWasSpent(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, oracle string
		power, toughness             int
		withColor, withoutColor      []string
	}{
		{"Azorius Herald", disSpiritType, "{2}{W}", disAzoriusHeraldOracle, 2, 1, []string{"W", "U", "C"}, []string{"W", "C", "C"}},
		{"Plaxmanta", "Creature — Beast", "{1}{U}", disPlaxmantaOracle, 2, 2, []string{"U", "G"}, []string{"U", "C"}},
		{"Patagia Viper", "Creature — Snake", "{3}{G}", disPatagiaViperOracle, 2, 1, []string{"G", "U", "C", "C"}, []string{"G", "C", "C", "C"}},
	} {
		for _, paid := range []bool{true, false} {
			label := tc.name + " without the colour"
			mana := tc.withoutColor
			if paid {
				label = tc.name + " with the colour"
				mana = tc.withColor
			}
			t.Run(label, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				id := disCast(t, g, me, tc.name, tc.typeLine, tc.cost, tc.oracle, tc.power, tc.toughness, mana...)
				effectstest.B35SettleTriggerOrders(t, g, me.ID)
				if got := g.Battlefield.Contains(id); got != paid {
					t.Errorf("on the battlefield = %v, want %v", got, paid)
				}
				if !paid && !me.Graveyard.Contains(id) {
					t.Error("the sacrificed creature is not in the graveyard")
				}
			})
		}
	}
}

func TestDis5598AzoriusHeraldGainsFourAndCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	disCast(t, g, me, "Azorius Herald", disSpiritType, "{2}{W}", disAzoriusHeraldOracle, 2, 1, "W", "U", "C")
	effectstest.B35SettleTriggerOrders(t, g, me.ID)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}

	g2 := effectstest.NewCatalogGame(t)
	attacker, opp := g2.Seats[0], g2.Seats[1]
	herald := effectstest.B12Push(g2, attacker.ID, "Azorius Herald", disSpiritType, disAzoriusHeraldOracle, 2, 1)
	blocker := effectstest.B12Creature(g2, opp.ID, "Blocker", disBearType, 2, 2)
	effectstest.BrAttack(t, g2, herald)
	if err := g2.DeclareBlocker(blocker, herald); err == nil {
		t.Error("Azorius Herald was blocked")
	}
}

func TestDis5598PlaxmantaGivesYourCreaturesShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", disBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", disBearType, 2, 2)
	disCast(t, g, me, "Plaxmanta", "Creature — Beast", "{1}{U}", disPlaxmantaOracle, 2, 2, "U", "G")
	effectstest.B35SettleTriggerOrders(t, g, me.ID)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "shroud") {
		t.Error("your creature has no shroud")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "shroud") {
		t.Error("an opponent's creature has shroud")
	}
}

func TestDis5598PatagiaViperMakesTwoSnakes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	disCast(t, g, me, "Patagia Viper", "Creature — Snake", "{3}{G}", disPatagiaViperOracle, 2, 1, "G", "C", "C", "C")
	effectstest.B35SettleTriggerOrders(t, g, me.ID)
	if got := effectstest.B27CountNamed(g, me.ID, "Snake"); got != 2 {
		t.Errorf("%d Snakes, want 2 (the tokens come even when the Viper is sacrificed)", got)
	}
}

// --- the Eidolons ---------------------------------------------------------------------

func TestDis5598EidolonsReturnOnAMulticoloredSpell(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Aurora Eidolon", disAuroraEidolonOracle},
		{"Enigma Eidolon", disEnigmaEidolonOracle},
		{"Entropic Eidolon", disEntropicEidolonOracle},
		{"Sandstorm Eidolon", disSandstormEidolonOracle},
		{"Verdant Eidolon", disVerdantEidolonOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			eidolon := disGraveyardCard(me, tc.name, disSpiritType, tc.oracle)
			disCast(t, g, me, "Gold Bear", disBearType, "{G}{U}", "", 2, 2, "G", "U")
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(eidolon) {
				t.Error("the Eidolon did not return to your hand")
			}
		})
	}
}

func TestDis5598EidolonIgnoresAMonocoloredSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eidolon := disGraveyardCard(me, "Enigma Eidolon", disSpiritType, disEnigmaEidolonOracle)
	disCast(t, g, me, "Green Bear", disBearType, "{G}", "", 2, 2, "G")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a monocolored spell offered the return")
	}
	if !me.Graveyard.Contains(eidolon) {
		t.Error("the Eidolon left the graveyard")
	}
}

func TestDis5598AuroraEidolonPreventsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	eidolon := effectstest.B12Push(g, me.ID, "Aurora Eidolon", disSpiritType, disAuroraEidolonOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", disBearType, 2, 5)
	pinger := effectstest.B12Creature(g, me.ID, "Pinger", disBearType, 1, 1)
	effectstest.ApaMana(me, "W")
	effectstest.ApaActivate(t, g, me, eidolon, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if g.Battlefield.Contains(eidolon) {
		t.Error("the Eidolon was not sacrificed")
	}
	effectstest.B27Damage(g, pinger, bear, 4)
	if got := effectstest.FindCardAnywhere(t, g, bear).DamageMarked; got != 1 {
		t.Errorf("the target has %d damage, want 1 (3 of 4 prevented)", got)
	}
}

func TestDis5598EnigmaEidolonMillsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	eidolon := effectstest.B12Push(g, me.ID, "Enigma Eidolon", disSpiritType, disEnigmaEidolonOracle, 2, 2)
	library := opp.Library.Size()
	effectstest.ApaMana(me, "U")
	effectstest.ApaActivate(t, g, me, eidolon, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if got := library - opp.Library.Size(); got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
}

func TestDis5598EntropicEidolonDrainsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	eidolon := effectstest.B12Push(g, me.ID, "Entropic Eidolon", disSpiritType, disEntropicEidolonOracle, 2, 2)
	life, oppLife := me.Life, opp.Life
	effectstest.ApaMana(me, "B")
	effectstest.ApaActivate(t, g, me, eidolon, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if me.Life != life+1 || opp.Life != oppLife-1 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, life+1, oppLife-1)
	}
}

func TestDis5598SandstormEidolonStopsABlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	eidolon := effectstest.B12Push(g, me.ID, "Sandstorm Eidolon", disSpiritType, disSandstormEidolonOracle, 2, 2)
	bear := effectstest.B12Creature(g, them.ID, "Their Bear", disBearType, 2, 2)
	other := effectstest.B12Creature(g, them.ID, "Their Other", disBearType, 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", disBearType, 2, 2)
	effectstest.ApaMana(me, "R")
	effectstest.ApaActivate(t, g, me, eidolon, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
	if effectstest.BlockRefusal(t, g, bear, attacker) == "" {
		t.Error("the target blocked")
	}
	if got := effectstest.BlockRefusal(t, g, other, attacker); got != "" {
		t.Errorf("another creature was refused: %q", got)
	}
}

func TestDis5598VerdantEidolonMakesThreeOfOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	eidolon := effectstest.B12Push(g, me.ID, "Verdant Eidolon", disSpiritType, disVerdantEidolonOracle, 2, 2)
	effectstest.ApaMana(me, "G")
	if err := g.ActivateManaAbility(me.ID, eidolon, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("pool holds %d, want 3", got)
	}
	if g.Battlefield.Contains(eidolon) {
		t.Error("the Eidolon was not sacrificed")
	}
}

// --- activated abilities ---------------------------------------------------------------

func TestDis5598MinisterOfImpedimentsTapsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	minister := effectstest.B12Push(g, me.ID, "Minister of Impediments", "Creature — Human Advisor", disMinisterOfImpedimentsOracl, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", disBearType, 2, 2)
	effectstest.ApaActivate(t, g, me, minister, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !effectstest.HmlTapped(t, g, bear) {
		t.Error("the target is untapped")
	}
	if !effectstest.HmlTapped(t, g, minister) {
		t.Error("the Minister did not tap")
	}
}

func TestDis5598SimicRagwormUntapsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	worm := effectstest.B12Push(g, me.ID, "Simic Ragworm", "Creature — Worm", disSimicRagwormOracle, 3, 3)
	disSetTapped(g, worm)
	effectstest.ApaMana(me, "U")
	effectstest.ApaActivate(t, g, me, worm, 0, game.ActivateAbilityParams{})
	if effectstest.HmlTapped(t, g, worm) {
		t.Error("the Ragworm is still tapped")
	}
}

func TestDis5598RakdosIckspitterPingsAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	spitter := effectstest.B12Push(g, me.ID, "Rakdos Ickspitter", "Creature — Thrull", disRakdosIckspitterOracle, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", disBearType, 2, 3)
	oppLife := opp.Life
	effectstest.ApaActivate(t, g, me, spitter, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if got := effectstest.FindCardAnywhere(t, g, bear).DamageMarked; got != 1 {
		t.Errorf("the target has %d damage, want 1", got)
	}
	if opp.Life != oppLife-1 {
		t.Errorf("the target's controller has %d life, want %d", opp.Life, oppLife-1)
	}
}

func TestDis5598HaazdaExoneratorDestroysAnAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	exonerator := effectstest.B12Push(g, me.ID, "Haazda Exonerator", "Creature — Human Cleric", disHaazdaExoneratorOracle, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", disBearType, 2, 2)
	aura := effectstest.B12Permanent(g, opp.ID, "Some Aura", "Enchantment — Aura")
	effectstest.NcdAttach(t, g, aura, bear)
	plain := effectstest.B12Permanent(g, opp.ID, "Plain Enchantment", "Enchantment")
	if err := g.ActivateCatalogAbility(me.ID, exonerator, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(plain)}); err == nil {
		t.Fatal("a non-Aura enchantment was a legal target")
	}
	effectstest.ApaActivate(t, g, me, exonerator, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(aura)})
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura survived")
	}
	if g.Battlefield.Contains(exonerator) {
		t.Error("the Exonerator was not sacrificed")
	}
}

func TestDis5598SkullmeadCauldronGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	cauldron := effectstest.PushCatalogPermanent(g, me.ID, "Skullmead Cauldron", "Artifact", disSkullmeadCauldronOracle, false)
	life := me.Life
	effectstest.ApaActivate(t, g, me, cauldron, 0, game.ActivateAbilityParams{})
	if me.Life != life+1 {
		t.Errorf("life %d after the first row, want %d", me.Life, life+1)
	}

	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	me2 := effectstest.Top100ActiveSeat(g2)
	cauldron2 := effectstest.PushCatalogPermanent(g2, me2.ID, "Skullmead Cauldron", "Artifact", disSkullmeadCauldronOracle, false)
	effectstest.D2dSetHand(me2, 1)
	card := me2.Hand.Cards[0].InstanceID
	life2 := me2.Life
	effectstest.ApaActivate(t, g2, me2, cauldron2, 1, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}})
	if me2.Life != life2+3 {
		t.Errorf("life %d after the discard row, want %d", me2.Life, life2+3)
	}
	if me2.Hand.Size() != 0 {
		t.Errorf("hand holds %d after the discard, want 0", me2.Hand.Size())
	}
}

func TestDis5598SilkwingScoutFetchesABasicLandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	scout := effectstest.B12Push(g, me.ID, "Silkwing Scout", "Creature — Faerie Scout", disSilkwingScoutOracle, 2, 1)
	forest := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	effectstest.ApaMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, scout, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("no search prompt")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the Forest is not on the battlefield")
	}
	if !effectstest.HmlTapped(t, g, forest) {
		t.Error("the Forest entered untapped")
	}
	if g.Battlefield.Contains(scout) {
		t.Error("the Scout was not sacrificed")
	}
}

func TestDis5598RixMaadiMakesEachPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	rix := effectstest.B12Push(g, me.ID, "Rix Maadi, Dungeon Palace", "Land", disRixMaadiOracle, 0, 0)
	for _, p := range g.Seats {
		effectstest.D2dSetHand(p, 2)
	}
	effectstest.ApaMana(me, "B", "R", "C")
	effectstest.ApaActivate(t, g, me, rix, 0, game.ActivateAbilityParams{})
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("you owe %d discards, want 1", n)
	}
}

// --- triggers ----------------------------------------------------------------------------

func TestDis5598OgreGatecrasherDestroysADefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wall := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}})
	other := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Other Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}})
	effectstest.CastHoldCreature(t, g, "Ogre Gatecrasher", "Creature — Ogre Rogue", disOgreGatecrasherOracle, 3, 3)
	effectstest.AnswerPickTarget(t, g, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the targeted defender survived")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("the other defender was destroyed")
	}
}

func TestDis5598WhiptailMolochDamagesYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", disBearType, 2, 4)
	effectstest.CastHoldCreature(t, g, "Whiptail Moloch", "Creature — Lizard", disWhiptailMolochOracle, 6, 3)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, bear).DamageMarked; got != 3 {
		t.Errorf("your creature has %d damage, want 3", got)
	}
}

func TestDis5598ProperBurialGainsTheToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Proper Burial", "Enchantment", disProperBurialOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", disBearType, 2, 3)
	life := me.Life
	effectstest.B27Kill(g, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestDis5598BeaconHawkUntapsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hawk := effectstest.B12Push(g, me.ID, "Beacon Hawk", "Creature — Bird", disBeaconHawkOracle, 1, 1)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped Bear", disBearType, 2, 2)
	disSetTapped(g, tapped)
	effectstest.DealCombatDamage(g, me.ID, hawk, opp.ID)
	g.RunStateChecksForTest()
	effectstest.AnswerPickTarget(t, g, tapped)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HmlTapped(t, g, tapped) {
		t.Error("the target is still tapped")
	}
}

func TestDis5598BeaconHawkPumpsItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	hawk := effectstest.B12Push(g, me.ID, "Beacon Hawk", "Creature — Bird", disBeaconHawkOracle, 1, 1)
	effectstest.ApaMana(me, "W")
	effectstest.ApaActivate(t, g, me, hawk, 0, game.ActivateAbilityParams{})
	if got := effectstest.ApaLive(g, hawk).CurrentToughness(); got != 2 {
		t.Errorf("toughness %d, want 2", got)
	}
}

// --- spells --------------------------------------------------------------------------------

func TestDis5598StompAndHowlDestroysBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	glory := effectstest.B12Permanent(g, opp.ID, "Glory", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Stomp and Howl", "Sorcery", disStompAndHowlOracle, effectstest.CardRefs(rock, glory))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(glory) {
		t.Error("a target survived")
	}
}

func TestDis5598WreckingBallDestroysALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	forest := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Wrecking Ball", "Instant", disWreckingBallOracle, effectstest.CardTarget(forest))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) {
		t.Error("the land survived")
	}
}

func TestDis5598WitsEndDiscardsTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.D2dSetHand(opp, 3)
	effectstest.CastCatalogSpell(t, g, "Wit's End", "Sorcery", disWitsEndOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if n := opp.Hand.Size(); n != 0 {
		t.Errorf("the target holds %d cards, want 0", n)
	}
}

func TestDis5598DeliriumSkeinsAsksForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	for _, p := range g.Seats {
		effectstest.D2dSetHand(p, 4)
	}
	effectstest.CastCatalogSpell(t, g, "Delirium Skeins", "Sorcery", disDeliriumSkeinsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, me.ID); n != 3 {
		t.Errorf("you owe %d discards, want 3", n)
	}
}

func TestDis5598PsychoticFuryNeedsAMulticoloredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	mono := effectstest.B12Creature(g, me.ID, "Mono", disBearType, 2, 2)
	effectstest.HmlSetColors(g, mono, "R")
	gold := effectstest.B12Creature(g, me.ID, "Gold", disBearType, 2, 2)
	effectstest.HmlSetColors(g, gold, "R", "G")
	if err := effectstest.CastCatalogSpellErr(t, g, "Psychotic Fury", "Instant", disPsychoticFuryOracle, effectstest.CardTarget(mono)); err == nil {
		t.Fatal("a monocolored creature was a legal target")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Psychotic Fury", "Instant", disPsychoticFuryOracle, effectstest.CardTarget(gold))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, gold, "double strike") {
		t.Error("the target has no double strike")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestDis5598CacklingFlamesHellbent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		hand   int
		damage int
	}{
		{"with cards in hand", 2, 3},
		{"hellbent", 0, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
			effectstest.D2dSetHand(me, tc.hand)
			life := opp.Life
			effectstest.CastCatalogSpell(t, g, "Cackling Flames", "Instant", disCacklingFlamesOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
			effectstest.PassPriorityAroundTable(t, g)
			if opp.Life != life-tc.damage {
				t.Errorf("life %d, want %d", opp.Life, life-tc.damage)
			}
		})
	}
}

func TestDis5598BondOfAgonyDrainsEachOtherPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	effectstest.CastWipe(t, g, "Bond of Agony", "Sorcery", disBondOfAgonyOracle, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != before[p.ID]-3 {
			t.Errorf("%s has %d life, want %d", p.Name, p.Life, before[p.ID]-3)
		}
	}
}

// --- hellbent and Auras -------------------------------------------------------------------------

func TestDis5598DemonsJesterHellbent(t *testing.T) {
	for _, tc := range []struct {
		name             string
		hand             int
		power, toughness int
	}{
		{"with a card in hand", 1, 2, 2},
		{"hellbent", 0, 4, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.D2dSetHand(me, tc.hand)
			jester := effectstest.B12Push(g, me.ID, "Demon's Jester", "Creature — Imp", disDemonsJesterOracle, 2, 2)
			c := effectstest.ApaLive(g, jester)
			if c.CurrentPower() != tc.power || c.CurrentToughness() != tc.toughness {
				t.Errorf("%d/%d, want %d/%d", c.CurrentPower(), c.CurrentToughness(), tc.power, tc.toughness)
			}
		})
	}
}

func TestDis5598RiotSpikesPumpsPowerAndShrinksToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", disBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Riot Spikes", "Enchantment — Aura", disRiotSpikesOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ApaLive(g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 1 {
		t.Errorf("%d/%d, want 4/1", c.CurrentPower(), c.CurrentToughness())
	}
}
