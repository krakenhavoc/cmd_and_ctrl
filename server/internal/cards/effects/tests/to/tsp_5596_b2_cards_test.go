package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tsp_5596_b2_cards_test.go — Time Spiral, batch 2 (#5596, S82): forty-four
// cards built from seams the engine already has. Each test plays the real
// catalog entry.

const (
	tspB2DreadshipReefOracle       = "130a8cf5-1354-4d17-91c8-c073642eb3db"
	tspB2CalciformPoolsOracle      = "2031bc31-81cc-407a-8615-29832f586bbc"
	tspB2SaltcrustedSteppeOracle   = "021e4165-2f02-4bd4-86ca-cb7bf4c9e23d"
	tspB2FungalReachesOracle       = "6f18ea44-3efa-4a45-abc6-86a0627e40f2"
	tspB2SageOfEpityrOracle        = "ef893585-246c-4028-bd2b-14ac51ce29be"
	tspB2AmrouSeekersOracle        = "396efb56-7a54-4d91-b2ce-7f3c62c3238b"
	tspB2AmrouScoutOracle          = "4c80112a-8ff5-4ca4-8737-b8d499d78982"
	tspB2UnyaroBeesOracle          = "7baff5c8-5919-4cbf-961e-962e740ebd8f"
	tspB2SpiketailDrakelingOracle  = "989e390d-f86c-4468-a9ad-5970a571dcc7"
	tspB2TrespasserIlVecOracle     = "90980841-4994-453c-a076-6cdf8d85c809"
	tspB2DrudgeReaversOracle       = "3edc147f-a385-413a-99b6-27d2cf8cf930"
	tspB2BewilderOracle            = "7cc945ed-d0aa-4b1a-a94d-656e03b6b5d3"
	tspB2OrcishCannonadeOracle     = "af224f6f-cedd-498f-a08e-7be0db5a11b7"
	tspB2ErrantDoomsayersOracle    = "82d4d70b-afb9-4ae7-8f30-16bc64dccf3e"
	tspB2SquallLineOracle          = "cdf9fab2-bfdb-4bf3-a06a-b21e07600e23"
	tspB2HauntingHymnOracle        = "b9a0977a-6364-4358-bab1-112898a1b942"
	tspB2CarefulConsiderationOracl = "2b34b361-7b0e-4464-874c-ceb501bc5ecc"
	tspB2FortifyOracle             = "52ea24d4-d33d-40fc-8ffc-09ae625b908f"
	tspB2FeeblenessOracle          = "a1a7757f-7e9b-4b5b-b9df-e98d8918caac"
	tspB2DrifterIlDalOracle        = "dbe8ad46-d862-4ff6-bf17-238fb8341075"
	tspB2IcatianCrierOracle        = "668312dd-7b6e-46bf-9c6a-bfb03dd75b8f"
	tspB2GreenseekerOracle         = "693d8203-4405-4488-9067-cc1cd8041cb4"
	tspB2MagusOfTheDiskOracle      = "3996aa0c-3322-43be-bbaf-7b66836a8ff4"
	tspB2BogardanRagerOracle       = "abbbfc22-18cc-4799-b02c-190173936872"
	tspB2ViashinoBladescoutOracle  = "e23d88c4-8c9c-41cb-b4ff-ebb3f880ab63"
	tspB2TivadarOfThornOracle      = "c2c09308-988b-4d71-83c8-e2a2dc19a0d9"
	tspB2TrompTheDomainsOracle     = "5b5dece6-b1e0-4844-bbf9-3b44008ffa70"
	tspB2WurmcallingOracle         = "8dd0f86d-6f9e-47e3-97dd-3a3fc075ea93"
	tspB2SpellBurstOracle          = "141d9b88-5326-4e3d-8f16-5293246fab88"
	tspB2VolcanicAwakeningOracle   = "2145329d-49df-4f4a-b7d7-3cc2f64639d3"
	tspB2GroundRiftOracle          = "c1f7d451-ca70-403d-81bf-8822876d1a54"
	tspB2ScarwoodTreefolkOracle    = "874b944a-8d08-4143-b477-685087396a88"
	tspB2CelestialCrusaderOracle   = "af06a811-7a57-4f68-9d87-7903e043f36d"
	tspB2TemporalEddyOracle        = "3a35e772-2547-41a3-a2a2-125b48c23aa5"
	tspB2MwonvuliAcidMossOracle    = "53fb8118-f332-407b-add8-1b329e420226"
	tspB2MolderOracle              = "86b98c3e-f568-4bfc-977f-d1be270936e5"
	tspB2VisceridDeepwalkerOracle  = "9dea848b-d758-46d0-829d-b096061fd460"

	tspB2BearType = "Creature — Bear"
)

func TestTspB2CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tspB2DreadshipReefOracle:       "Dreadship Reef",
		tspB2CalciformPoolsOracle:      "Calciform Pools",
		tspB2SaltcrustedSteppeOracle:   "Saltcrusted Steppe",
		tspB2FungalReachesOracle:       "Fungal Reaches",
		tspB2SageOfEpityrOracle:        "Sage of Epityr",
		tspB2AmrouSeekersOracle:        "Amrou Seekers",
		tspB2AmrouScoutOracle:          "Amrou Scout",
		tspB2UnyaroBeesOracle:          "Unyaro Bees",
		tspB2SpiketailDrakelingOracle:  "Spiketail Drakeling",
		tspB2TrespasserIlVecOracle:     "Trespasser il-Vec",
		tspB2DrudgeReaversOracle:       "Drudge Reavers",
		tspB2BewilderOracle:            "Bewilder",
		tspB2OrcishCannonadeOracle:     "Orcish Cannonade",
		tspB2ErrantDoomsayersOracle:    "Errant Doomsayers",
		tspB2SquallLineOracle:          "Squall Line",
		tspB2HauntingHymnOracle:        "Haunting Hymn",
		tspB2CarefulConsiderationOracl: "Careful Consideration",
		tspB2FortifyOracle:             "Fortify",
		tspB2FeeblenessOracle:          "Feebleness",
		tspB2DrifterIlDalOracle:        "Drifter il-Dal",
		tspB2IcatianCrierOracle:        "Icatian Crier",
		tspB2GreenseekerOracle:         "Greenseeker",
		tspB2MagusOfTheDiskOracle:      "Magus of the Disk",
		tspB2BogardanRagerOracle:       "Bogardan Rager",
		tspB2ViashinoBladescoutOracle:  "Viashino Bladescout",
		tspB2TivadarOfThornOracle:      "Tivadar of Thorn",
		tspB2TrompTheDomainsOracle:     "Tromp the Domains",
		tspB2WurmcallingOracle:         "Wurmcalling",
		tspB2SpellBurstOracle:          "Spell Burst",
		tspB2VolcanicAwakeningOracle:   "Volcanic Awakening",
		tspB2GroundRiftOracle:          "Ground Rift",
		tspB2ScarwoodTreefolkOracle:    "Scarwood Treefolk",
		tspB2CelestialCrusaderOracle:   "Celestial Crusader",
		tspB2TemporalEddyOracle:        "Temporal Eddy",
		tspB2MwonvuliAcidMossOracle:    "Mwonvuli Acid-Moss",
		tspB2MolderOracle:              "Molder",
		tspB2VisceridDeepwalkerOracle:  "Viscerid Deepwalker",
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

// --- helpers -----------------------------------------------------------------

// tspB2CastInEndStep casts a catalog instant from the active seat's hand
// in its end step, outside any main phase, and returns the spell.
func tspB2CastInEndStep(t *testing.T, g *game.Game, name, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	id := effectstest.HandCardFull(me, name, "Instant", "", oracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s in the end step: %v", name, err)
	}
	return id
}

// tspB2HandHas reports whether p's hand holds a card named `name`.
func tspB2HandHas(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

// tspB2CountOnBattlefield counts the permanents named `name`.
func tspB2CountOnBattlefield(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

// tspB2Fill gives p `n` spare cards in hand.
func tspB2Fill(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.HandCardFull(p, "Spare", "Sorcery", "{1}", "", nil)
	}
}

// --- storage lands -----------------------------------------------------------

func TestTspB2StorageLandsBankAndPayInTheirTwoColours(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, a, b string
	}{
		{"Dreadship Reef", tspB2DreadshipReefOracle, "U", "B"},
		{"Calciform Pools", tspB2CalciformPoolsOracle, "W", "U"},
		{"Saltcrusted Steppe", tspB2SaltcrustedSteppeOracle, "G", "W"},
		{"Fungal Reaches", tspB2FungalReachesOracle, "R", "G"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			land := effectstest.B12Push(g, me.ID, tc.name, "Land", tc.oracle, 0, 0)
			effectstest.AdvanceToMain(t, g)

			// {1}, {T}: a storage counter, through the stack.
			effectstest.DcMana(t, g, me.ID, "{C}")
			if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("store: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterCount(g, land, game.CounterStorage); got != 1 {
				t.Fatalf("storage counters %d, want 1", got)
			}
			if !effectstest.TappedForTest(t, g, land) {
				t.Error("banking a counter did not tap the land")
			}

			// {1}, Remove X: X mana, each either colour, and no {T}.
			effectstest.AddCountersForTest(g, land, game.CounterStorage, 2)
			effectstest.DcMana(t, g, me.ID, "{C}")
			effectstest.ActivateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{CounterCounts: []int{2}})
			for _, color := range []string{tc.a, tc.b} {
				var pick *game.PendingChoice
				for _, c := range g.PendingChoices {
					if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
						pick = c
					}
				}
				if pick == nil {
					break
				}
				if err := g.ResolveManaChoice(pick.ID, me.ID, color); err != nil {
					t.Fatalf("ResolveManaChoice %s: %v", color, err)
				}
			}
			if got := effectstest.PoolColours(me); got[tc.a] != 1 || got[tc.b] != 1 || me.ManaPool.Total() != 2 {
				t.Errorf("pool %v, want one {%s} and one {%s}", got, tc.a, tc.b)
			}
			if got := effectstest.CounterCount(g, land, game.CounterStorage); got != 1 {
				t.Errorf("storage counters left %d, want 1", got)
			}
		})
	}
}

// --- creatures ---------------------------------------------------------------

func TestTspB2PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		keywords               []string
	}{
		{"Unyaro Bees", tspB2UnyaroBeesOracle, "Creature — Insect", []string{"flying"}},
		{"Spiketail Drakeling", tspB2SpiketailDrakelingOracle, "Creature — Drake", []string{"flying"}},
		{"Drudge Reavers", tspB2DrudgeReaversOracle, "Creature — Skeleton", []string{"flash"}},
		{"Drifter il-Dal", tspB2DrifterIlDalOracle, "Creature — Human Wizard", []string{"shadow"}},
		{"Bogardan Rager", tspB2BogardanRagerOracle, "Creature — Elemental", []string{"flash"}},
		{"Viashino Bladescout", tspB2ViashinoBladescoutOracle, "Creature — Lizard Scout", []string{"flash"}},
		{"Tivadar of Thorn", tspB2TivadarOfThornOracle, "Legendary Creature — Human Knight", []string{"first strike", "protection from red"}},
		{"Celestial Crusader", tspB2CelestialCrusaderOracle, "Creature — Spirit", []string{"flash", "split second", "flying"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			for _, kw := range tc.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s does not have %s", tc.name, kw)
				}
			}
		})
	}
}

func TestTspB2SageOfEpityrLooksAtTheTopFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Sage of Epityr", "Creature — Human Wizard", tspB2SageOfEpityrOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LookAtTopChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no look-at prompt")
	}
	if len(c.ScryCards) != 4 {
		t.Errorf("looking at %d cards, want 4", len(c.ScryCards))
	}
}

func TestTspB2AmrouSeekersBlockedOnlyByArtifactOrWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	seekers := tspB2NemB1Push(g, me.ID, "Amrou Seekers", "Creature — Kithkin Rebel", tspB2AmrouSeekersOracle, 2, 2, "W")
	green := tspB2NemB1Push(g, them.ID, "Green Bear", tspB2BearType, "", 2, 2, "G")
	white := tspB2NemB1Push(g, them.ID, "White Knight", "Creature — Human Knight", "", 2, 2, "W")
	golem := tspB2NemB1Push(g, them.ID, "Golem", "Artifact Creature — Golem", "", 2, 2)
	effectstest.AttackIntoBlocks(t, g, them.ID, seekers)
	if effectstest.BlockRefusal(t, g, green, seekers) == "" {
		t.Error("a green non-artifact creature blocked")
	}
	if got := effectstest.BlockRefusal(t, g, white, seekers); got != "" {
		t.Errorf("a white creature was refused: %q", got)
	}
	if got := effectstest.BlockRefusal(t, g, golem, seekers); got != "" {
		t.Errorf("an artifact creature was refused: %q", got)
	}
}

func TestTspB2AmrouScoutFindsASmallRebel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Small Rebel", TypeLine: "Creature — Human Rebel", ManaCost: "{1}{W}"},
		game.Card{Name: "Big Rebel", TypeLine: "Creature — Human Rebel", ManaCost: "{3}{W}"},
		game.Card{Name: "Small Soldier", TypeLine: "Creature — Human Soldier", ManaCost: "{W}"},
	)
	scout := tspB2NemB1Push(g, me.ID, "Amrou Scout", "Creature — Kithkin Rebel Scout", tspB2AmrouScoutOracle, 2, 1)
	tspB2NemB2Activate(t, g, me.ID, scout, "{C}{C}{C}{C}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Big Rebel") != uuid.Nil {
		t.Error("a mana value 4 Rebel was offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Small Soldier") != uuid.Nil {
		t.Error("a non-Rebel was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Small Rebel")
	if effectstest.FindBattlefieldByName(g, "Small Rebel") == uuid.Nil {
		t.Error("the Rebel did not enter the battlefield")
	}
}

func TestTspB2UnyaroBeesPumpAndSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bees := effectstest.B12Push(g, me.ID, "Unyaro Bees", "Creature — Insect", tspB2UnyaroBeesOracle, 0, 1)
	effectstest.DcMana(t, g, me.ID, "{G}")
	effectstest.ApaActivate(t, g, me, bees, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, bees); p != 1 || tough != 2 {
		t.Errorf("the Bees are %d/%d, want 1/2", p, tough)
	}
	life := opp.Life
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{G}")
	effectstest.ApaActivate(t, g, me, bees, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if g.Battlefield.Contains(bees) {
		t.Error("the Bees were not sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
}

func TestTspB2SpiketailDrakelingCountersUnlessTwoIsPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		drake := effectstest.B12Push(g, me.ID, "Spiketail Drakeling", "Creature — Drake", tspB2SpiketailDrakelingOracle, 2, 2)
		spell := effectstest.CastCatalogSpell(t, g, "A Bear", tspB2BearType, "", nil)
		if err := g.ActivateCatalogAbility(me.ID, drake, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(spell)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if g.Battlefield.Contains(drake) {
			t.Error("the Drakeling was not sacrificed as part of the cost")
		}
		effectstest.PassUntilTaxed(t, g, me.ID)
		if pay {
			effectstest.FillPool(me, 2)
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if g.Stack.Contains(spell) != pay {
			t.Errorf("paid=%v: the spell is still on the stack = %v", pay, g.Stack.Contains(spell))
		}
	}
}

func TestTspB2TrespasserIlVecDiscardsForShadow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rogue := effectstest.B12Push(g, me.ID, "Trespasser il-Vec", "Creature — Human Rogue", tspB2TrespasserIlVecOracle, 3, 1)
	fodder := effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, rogue, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, rogue, "shadow") {
		t.Error("the Trespasser did not gain shadow")
	}
	if me.Hand.Contains(fodder) {
		t.Error("the discard was not paid")
	}
}

func TestTspB2DrudgeReaversRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	reavers := effectstest.B12Push(g, me.ID, "Drudge Reavers", "Creature — Skeleton", tspB2DrudgeReaversOracle, 2, 2)
	effectstest.DcMana(t, g, me.ID, "{B}")
	effectstest.ApaActivate(t, g, me, reavers, 0, game.ActivateAbilityParams{})
	if got := effectstest.FindCardAnywhere(t, g, reavers).RegenerationShields; got != 1 {
		t.Errorf("regeneration shields %d, want 1", got)
	}
}

func TestTspB2ErrantDoomsayersTapsOnlyASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	doomsayers := effectstest.B12Push(g, me.ID, "Errant Doomsayers", "Creature — Human Rebel", tspB2ErrantDoomsayersOracle, 1, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big", tspB2BearType, 3, 3)
	small := effectstest.B12Creature(g, opp.ID, "Small", tspB2BearType, 3, 2)
	if err := g.ActivateCatalogAbility(me.ID, doomsayers, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(big)}); err == nil {
		t.Fatal("a creature with toughness 3 was targeted")
	}
	effectstest.ApaActivate(t, g, me, doomsayers, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(small)})
	if !effectstest.HmlTapped(t, g, small) {
		t.Error("the toughness-2 creature is untapped")
	}
}

func TestTspB2DrifterIlDalIsSacrificedUnlessUIsPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		drifter := effectstest.B12Push(g, me.ID, "Drifter il-Dal", "Creature — Human Wizard", tspB2DrifterIlDalOracle, 2, 1)
		effectstest.WalkToUpkeepOfSeat0(t, g)
		if pay {
			effectstest.DcMana(t, g, me.ID, "{U}")
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.UpkeepPromptFor(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if got := g.Battlefield.Contains(drifter); got != pay {
			t.Errorf("pay=%v: still on the battlefield = %v", pay, got)
		}
	}
}

func TestTspB2IcatianCrierMakesTwoCitizens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	crier := tspB2NemB1Push(g, me.ID, "Icatian Crier", "Creature — Human Spellshaper", tspB2IcatianCrierOracle, 1, 1, "W")
	tspB2NemB2Shape(t, g, crier, "{C}{W}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := tspB2CountOnBattlefield(g, "Citizen"); got != 2 {
		t.Errorf("%d Citizen tokens, want 2", got)
	}
}

func TestTspB2GreenseekerFindsABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Grizzly Bears", TypeLine: tspB2BearType, ManaCost: "{1}{G}"},
	)
	seeker := tspB2NemB1Push(g, me.ID, "Greenseeker", "Creature — Elf Spellshaper", tspB2GreenseekerOracle, 1, 1, "G")
	tspB2NemB2Shape(t, g, seeker, "{G}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil {
		t.Error("a creature card was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	if !tspB2HandHas(me, "Forest") {
		t.Error("the Forest is not in hand")
	}
}

func TestTspB2MagusOfTheDiskEntersTappedAndWipes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	cast := effectstest.CastHoldCreature(t, g, "Magus of the Disk", "Creature — Human Wizard", tspB2MagusOfTheDiskOracle, 2, 4)
	if !effectstest.IsTapped(g, cast) {
		t.Error("Magus of the Disk did not enter tapped")
	}

	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	magus := effectstest.B12Push(g, me.ID, "Magus of the Disk", "Creature — Human Wizard", tspB2MagusOfTheDiskOracle, 2, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", tspB2BearType, 2, 2)
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	aura := effectstest.B12Permanent(g, opp.ID, "Charm", "Enchantment")
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	effectstest.DcMana(t, g, me.ID, "{C}")
	effectstest.ApaActivate(t, g, me, magus, 0, game.ActivateAbilityParams{})
	for name, id := range map[string]uuid.UUID{"the Magus": magus, "the creature": bear, "the artifact": relic, "the enchantment": aura} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", name)
		}
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was destroyed")
	}
}

func TestTspB2ScarwoodTreefolkEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastHoldCreature(t, g, "Scarwood Treefolk", "Creature — Treefolk", tspB2ScarwoodTreefolkOracle, 3, 5)
	if !effectstest.IsTapped(g, id) {
		t.Error("Scarwood Treefolk did not enter tapped")
	}
}

func TestTspB2EntersTriggersPumpAndGrant(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		p, tough               int
		keyword                string
	}{
		{"Bogardan Rager", tspB2BogardanRagerOracle, "Creature — Elemental", 6, 2, ""},
		{"Viashino Bladescout", tspB2ViashinoBladescoutOracle, "Creature — Lizard Scout", 2, 2, "first strike"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.B12Creature(g, me.ID, "Bear", tspB2BearType, 2, 2)
			effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 1)
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.AnswerPickTarget(t, g, bear)
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, bear); p != tc.p || tough != tc.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
			if tc.keyword != "" && !effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
		})
	}
}

func TestTspB2TivadarDestroysAGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	goblin := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	spare := effectstest.B12Creature(g, opp.ID, "Other Goblin", "Creature — Goblin", 1, 1)
	effectstest.CastHoldCreature(t, g, "Tivadar of Thorn", "Legendary Creature — Human Knight", tspB2TivadarOfThornOracle, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, goblin)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(goblin) {
		t.Error("the Goblin survived")
	}
	if !g.Battlefield.Contains(spare) {
		t.Error("the Goblin that was not targeted was destroyed")
	}
}

func TestTspB2CelestialCrusaderPumpsOtherWhiteCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := tspB2NemB1Push(g, me.ID, "Celestial Crusader", "Creature — Spirit", tspB2CelestialCrusaderOracle, 2, 2, "W")
	theirs := tspB2NemB1Push(g, opp.ID, "Their Knight", "Creature — Human Knight", "", 2, 2, "W")
	green := tspB2NemB1Push(g, me.ID, "My Bear", tspB2BearType, "", 2, 2, "G")
	if p, tough := effectstest.PtOf(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("an opponent's white creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, green); p != 2 || tough != 2 {
		t.Errorf("a green creature is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, crusader); p != 2 || tough != 2 {
		t.Errorf("the Crusader is %d/%d, want 2/2", p, tough)
	}
}

func TestTspB2VisceridDeepwalkerPumpsAndSuspends(t *testing.T) {
	effectstest.S40SuspendAction(t, tspB2VisceridDeepwalkerOracle, 4, "{U}")
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.B12Push(g, me.ID, "Viscerid Deepwalker", "Creature — Homarid Warrior", tspB2VisceridDeepwalkerOracle, 2, 3)
	effectstest.DcMana(t, g, me.ID, "{U}")
	effectstest.ApaActivate(t, g, me, walker, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, walker); p != 3 || tough != 3 {
		t.Errorf("the Deepwalker is %d/%d, want 3/3", p, tough)
	}
}

// --- spells ------------------------------------------------------------------

func TestTspB2BewilderShrinksAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", tspB2BearType, 4, 4)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bewilder", "Instant", tspB2BewilderOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 1 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 1/4", p, tough)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestTspB2OrcishCannonadeHitsTheTargetAndYouAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine, theirs, hand := me.Life, opp.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Orcish Cannonade", "Instant", tspB2OrcishCannonadeOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, theirs-2)
	}
	if me.Life != mine-3 {
		t.Errorf("my life %d, want %d", me.Life, mine-3)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestTspB2SquallLineHitsFlyersAndPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 3, Keywords: []string{"flying"}})
	walker := effectstest.B12Creature(g, opp.ID, "Walker", tspB2BearType, 2, 3)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	effectstest.CastXSpell(t, g, "Squall Line", "Instant", tspB2SquallLineOracle, "{X}{G}{G}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, flyer).DamageMarked; got != 2 {
		t.Errorf("the flyer has %d damage, want 2", got)
	}
	if got := effectstest.FindCardAnywhere(t, g, walker).DamageMarked; got != 0 {
		t.Errorf("a creature without flying has %d damage, want 0", got)
	}
	for _, p := range g.Seats {
		if p.Life != before[p.ID]-2 {
			t.Errorf("%s life %d, want %d", p.Name, p.Life, before[p.ID]-2)
		}
	}
}

func TestTspB2HauntingHymnDiscardsFourInYourMainPhase(t *testing.T) {
	for _, main := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		tspB2Fill(opp, 6)
		want := 2
		if main {
			want = 4
			effectstest.CastCatalogSpell(t, g, "Haunting Hymn", "Instant", tspB2HauntingHymnOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
		} else {
			tspB2CastInEndStep(t, g, "Haunting Hymn", tspB2HauntingHymnOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
		}
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.DiscardOwed(g, opp.ID); n != want {
			t.Errorf("main=%v: the opponent owes %d discards, want %d", main, n, want)
		}
	}
}

func TestTspB2CarefulConsiderationDrawsFourAndDiscards(t *testing.T) {
	for _, main := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		want := 3
		var hand int
		if main {
			want = 2
			effectstest.AdvanceToMain(t, g)
			hand = me.Hand.Size()
			effectstest.CastCatalogSpell(t, g, "Careful Consideration", "Instant", tspB2CarefulConsiderationOracl, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
		} else {
			effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
			hand = me.Hand.Size()
			tspB2CastInEndStep(t, g, "Careful Consideration", tspB2CarefulConsiderationOracl, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != hand+4 {
			t.Errorf("main=%v: hand %d before the discard, want %d", main, got, hand+4)
		}
		if n := effectstest.DiscardOwed(g, me.ID); n != want {
			t.Errorf("main=%v: owes %d discards, want %d", main, n, want)
		}
	}
}

func TestTspB2FortifyEachMode(t *testing.T) {
	for _, tc := range []struct {
		mode     int
		p, tough int
	}{
		{0, 4, 2},
		{1, 2, 4},
	} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
		mine := effectstest.B12Creature(g, me.ID, "My Bear", tspB2BearType, 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", tspB2BearType, 2, 2)
		effectstest.CastModal(t, g, "Fortify", "Instant", tspB2FortifyOracle, []int{tc.mode}, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, mine); p != tc.p || tough != tc.tough {
			t.Errorf("mode %d: my Bear is %d/%d, want %d/%d", tc.mode, p, tough, tc.p, tc.tough)
		}
		if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
			t.Errorf("mode %d: an opponent's Bear is %d/%d, want 2/2", tc.mode, p, tough)
		}
	}
}

func TestTspB2FeeblenessShrinksTheEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", tspB2BearType, 3, 3)
	effectstest.CastCatalogSpell(t, g, "Feebleness", "Enchantment — Aura", tspB2FeeblenessOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 1 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 1/2", p, tough)
	}
}

func TestTspB2TrompTheDomainsCountsBasicLandTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	mine := effectstest.B12Creature(g, me.ID, "My Bear", tspB2BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", tspB2BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Tromp the Domains", "Sorcery", tspB2TrompTheDomainsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("my Bear is %d/%d, want 4/4 (two basic land types)", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("my Bear has no trample")
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's Bear is %d/%d, want 2/2", p, tough)
	}
}

func TestTspB2WurmcallingMakesAnXByXWurm(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastXSpell(t, g, "Wurmcalling", "Sorcery", tspB2WurmcallingOracle, "{X}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	wurm := effectstest.FindBattlefieldByName(g, "Wurm")
	if wurm == uuid.Nil {
		t.Fatal("no Wurm token")
	}
	if p, tough := effectstest.PtOf(t, g, wurm); p != 3 || tough != 3 {
		t.Errorf("the Wurm is %d/%d, want 3/3", p, tough)
	}
}

func TestTspB2SpellBurstCountersASpellOfManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", tspB2BearType, "", nil)
	effectstest.CastXSpell(t, g, "Spell Burst", "Instant", tspB2SpellBurstOracle, "{X}{U}", 0, effectstest.CardTarget(spell))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(spell) {
		t.Error("the spell is still on the stack")
	}
	if effectstest.FindBattlefieldByName(g, "A Bear") != uuid.Nil {
		t.Error("the countered spell resolved")
	}
}

func TestTspB2VolcanicAwakeningDestroysALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	effectstest.CastCatalogSpell(t, g, "Volcanic Awakening", "Sorcery", tspB2VolcanicAwakeningOracle, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
}

func TestTspB2GroundRiftStopsAGroundBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	flyer := effectstest.ApaPush(g, them.ID, them.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	ground := tspB2NemB1Push(g, them.ID, "Ground", tspB2BearType, "", 2, 2, "G")
	other := tspB2NemB1Push(g, them.ID, "Other", tspB2BearType, "", 2, 2, "G")
	attacker := tspB2NemB1Push(g, me.ID, "Attacker", tspB2BearType, "", 2, 2, "G")
	if err := effectstest.CastCatalogSpellErr(t, g, "Ground Rift", "Sorcery", tspB2GroundRiftOracle, effectstest.CardTarget(flyer)); err == nil {
		t.Fatal("a creature with flying was targeted")
	}
	effectstest.CastCatalogSpell(t, g, "Ground Rift", "Sorcery", tspB2GroundRiftOracle, effectstest.CardTarget(ground))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
	if effectstest.BlockRefusal(t, g, ground, attacker) == "" {
		t.Error("the target blocked")
	}
	if got := effectstest.BlockRefusal(t, g, other, attacker); got != "" {
		t.Errorf("another creature was refused: %q", got)
	}
}

func TestTspB2TemporalEddyTucksALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	effectstest.CastCatalogSpell(t, g, "Temporal Eddy", "Sorcery", tspB2TemporalEddyOracle, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Fatal("the land is still on the battlefield")
	}
	top, err := opp.Library.Top()
	if err != nil || top.Name != "Their Land" {
		t.Errorf("the top of the owner's library is %q (%v), want Their Land", top.Name, err)
	}
}

func TestTspB2MwonvuliAcidMossTradesALandForAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
	)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	effectstest.CastCatalogSpell(t, g, "Mwonvuli Acid-Moss", "Sorcery", tspB2MwonvuliAcidMossOracle, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the target land survived")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Plains") != uuid.Nil {
		t.Error("a Plains was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	forest := effectstest.FindBattlefieldByName(g, "Forest")
	if forest == uuid.Nil {
		t.Fatal("the Forest did not enter the battlefield")
	}
	if !effectstest.IsTapped(g, forest) {
		t.Error("the Forest entered untapped")
	}
}

func TestTspB2MolderDestroysAtXAndGainsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	relic := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Relic", TypeLine: "Artifact", ManaCost: "{2}"})
	life := me.Life
	effectstest.CastXSpell(t, g, "Molder", "Instant", tspB2MolderOracle, "{X}{G}", 2, effectstest.CardTarget(relic))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact survived")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

// Local copies of helpers that live in another test package.

// tspB2NemB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func tspB2NemB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}

// tspB2NemB2Activate activates ability 0 of `id` under the strict mana gate
// after adding `mana` to the pool.
func tspB2NemB2Activate(t *testing.T, g *game.Game, player, id uuid.UUID, mana string, p game.ActivateAbilityParams) {
	t.Helper()
	if mana != "" {
		effectstest.DcMana(t, g, player, mana)
	}
	p.Strict = true
	if err := g.ActivateCatalogAbility(player, id, 0, p); err != nil {
		t.Fatalf("activate: %v", err)
	}
}

// tspB2NemB2Shape activates a spellshaper's ability at `target`, paying with
// a fresh card from hand.
func tspB2NemB2Shape(t *testing.T, g *game.Game, shaper uuid.UUID, mana string, targets []game.TargetRef) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	tspB2NemB2Activate(t, g, me.ID, shaper, mana, game.ActivateAbilityParams{Targets: targets, DiscardIDs: []uuid.UUID{fodder}})
}
