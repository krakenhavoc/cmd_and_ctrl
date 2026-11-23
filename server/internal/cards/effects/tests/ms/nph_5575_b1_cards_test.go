package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nph_5575_b1_cards_test.go — New Phyrexia, batch 1 (#5575). Every test
// plays the real catalog entry through the engine.

const (
	nphFlamebornVironOracle        = "3fd2651d-89fb-4bd1-af4a-f7eae7ef9fd2"
	nphLoxodonConvertOracle        = "d0a80814-c69c-457c-ba19-4e1ba3fbd485"
	nphRottedHystrixOracle         = "e600afe3-8d5b-46ab-b98c-f0bacb58ab2a"
	nphDarksteelRelicOracle        = "2bda1fb1-528c-4e66-89f1-bb1a85e9181e"
	nphHovermyrOracle              = "b58722c5-84f2-43c2-be7e-3577c2f2a60a"
	nphPorcelainLegionnaireOracle  = "66442396-f804-47f3-8daa-0d43d8cfb482"
	nphShriekRaptorOracle          = "426bfd00-e727-49f4-be2d-df5289d2cc61"
	nphSlashPantherOracle          = "76558381-c4ae-4bc3-8c7b-fa0b717f0282"
	nphSpinedThopterOracle         = "45f2af87-0235-49f8-a63d-370faff3a097"
	nphSpireMonitorOracle          = "447b9412-2649-4548-8e9b-eed5e92ad92a"
	nphThunderingTanadonOracle     = "487382ea-0050-4a3f-bcf6-0ce41d6f4ebf"
	nphVaultSkirgeOracle           = "64cf5d59-7bcd-4b0b-a160-c8468d4c0f60"
	nphVulshokRefugeeOracle        = "7220769a-9d63-4211-9365-9feaf4d9a518"
	nphGlistenerElfOracle          = "9d95d173-5c7f-4e0c-bcdc-9b90fcd7339b"
	nphLostLeoninOracle            = "c21b5c5b-9191-4748-b111-a51631e0bc5a"
	nphRazorSwineOracle            = "4d5649a1-5716-4560-a956-84300bf6a2f4"
	nphOgreMenialOracle            = "e7c879fa-a34b-42d9-8edf-29254380c1f7"
	nphImmolatingSouleaterOracle   = "fb260f5a-b937-435b-ab18-dcc57d6845df"
	nphMoltensteelDragonOracle     = "dc2d67bb-29b3-43d7-8ebf-2100c5287bbf"
	nphToxicNimOracle              = "afc2a3a7-eb69-4fcb-b3a3-444dccb735d7"
	nphFallenFerromancerOracle     = "22446354-e5c3-4f19-9731-79a4562639d1"
	nphInsatiableSouleaterOracle   = "f81c0ddc-cd3c-42c0-a7d9-0057b69a5df1"
	nphPestilentSouleaterOracle    = "c4779671-2a01-4eef-847c-ede36091fc4a"
	nphDeathHoodCobraOracle        = "747dfffd-880c-48fe-9380-eb56071f5910"
	nphTrespassingSouleaterOracle  = "4d19b4d1-49d1-44e9-a453-dcb08b5fa6e4"
	nphBlindingSouleaterOracle     = "f5f330a5-14a0-4bfe-ba46-4616692b7bd3"
	nphDementiaBatOracle           = "95c76175-e3a6-4bf4-a8da-d1179851efe5"
	nphViralDrakeOracle            = "3b1647d1-9aec-454f-a215-e28633cb7a2d"
	nphSpinebiterOracle            = "9b844a0c-590f-46c7-b53e-c99398e0d8c0"
	nphKilnWalkerOracle            = "ca868ac4-6144-402e-ae23-629628d02883"
	nphPithDrillerOracle           = "62d293a6-4fad-4aaa-af87-62d705e243ca"
	nphPriestOfUrabraskOracle      = "e4bd8910-770b-4220-8a26-2673491f4a3e"
	nphInquisitorExarchOracle      = "7deae378-9291-487e-9f0b-1deaf37218a8"
	nphTormentorExarchOracle       = "ebc52f9a-ba0c-48ed-bb04-ed086efeb060"
	nphDeceiverExarchOracle        = "3c939ea6-68b7-4965-b1d3-af1d3dc79778"
	nphMarrowShardsOracle          = "be2ae8dc-afae-4bcf-8a97-08a2764d1148"
	nphVaporSnagOracle             = "a1408bae-ed5f-4837-a1a7-438dc30a2786"
	nphGlissasScornOracle          = "f30f17f0-bd2d-4f10-9739-111917acc100"
	nphVictoriousDestructionOracle = "fa5d6c02-53b2-4144-87cb-e223b529ba7d"
	nphPsychicBarrierOracle        = "2e712090-2284-4325-8ee8-38cd0e2bedd8"
	nphMutagenicGrowthOracle       = "c17b9e5b-58b6-4ccf-8283-945ba77429e7"
	nphWhipflareOracle             = "a6b599db-8d45-4a11-9765-ae7eb235b4b2"
	nphMindcullingOracle           = "955a7f4d-7100-4d8b-818c-be8651d0ebe6"
	nphGrimAfflictionOracle        = "411ba1a2-30bb-4618-b874-6be0d996f07d"
	nphCorrosiveGaleOracle         = "6143073f-4110-4a37-bc0c-89fbda2fc839"
	nphArtillerizeOracle           = "d14a8d90-13c0-4eb0-ab7b-86edb2c7fb96"
	nphSickleslicerOracle          = "7a004a4b-b53b-4b8e-86f5-528c1082e146"
	nphNecropouncerOracle          = "9e33b295-fab8-42c7-be43-baf9c17d063f"

	nphBearType = "Creature — Bear"
)

// nphB1ChooseMode answers the open mode_pick prompt with the printed
// bullet `mode`, then the bullet's target prompt when one is asked, and
// settles the stack.
func nphB1ChooseMode(t *testing.T, g *game.Game, chooser uuid.UUID, mode int, target *game.TargetRef) {
	t.Helper()
	c := effectstest.ModePickChoiceFor(g, chooser)
	if c == nil {
		t.Fatal("no mode_pick prompt is open")
	}
	if err := g.ResolveModePick(c.ID, chooser, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if target != nil {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, *target); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// nphB1CardRef is a card target reference.
func nphB1CardRef(id uuid.UUID) *game.TargetRef {
	return &game.TargetRef{Kind: game.TargetCard, ID: id}
}

// --- every card is in the catalog ----------------------------------------

func TestNph5575B1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		nphFlamebornVironOracle:        "Flameborn Viron",
		nphLoxodonConvertOracle:        "Loxodon Convert",
		nphRottedHystrixOracle:         "Rotted Hystrix",
		nphDarksteelRelicOracle:        "Darksteel Relic",
		nphHovermyrOracle:              "Hovermyr",
		nphPorcelainLegionnaireOracle:  "Porcelain Legionnaire",
		nphShriekRaptorOracle:          "Shriek Raptor",
		nphSlashPantherOracle:          "Slash Panther",
		nphSpinedThopterOracle:         "Spined Thopter",
		nphSpireMonitorOracle:          "Spire Monitor",
		nphThunderingTanadonOracle:     "Thundering Tanadon",
		nphVaultSkirgeOracle:           "Vault Skirge",
		nphVulshokRefugeeOracle:        "Vulshok Refugee",
		nphGlistenerElfOracle:          "Glistener Elf",
		nphLostLeoninOracle:            "Lost Leonin",
		nphRazorSwineOracle:            "Razor Swine",
		nphOgreMenialOracle:            "Ogre Menial",
		nphImmolatingSouleaterOracle:   "Immolating Souleater",
		nphMoltensteelDragonOracle:     "Moltensteel Dragon",
		nphToxicNimOracle:              "Toxic Nim",
		nphFallenFerromancerOracle:     "Fallen Ferromancer",
		nphInsatiableSouleaterOracle:   "Insatiable Souleater",
		nphPestilentSouleaterOracle:    "Pestilent Souleater",
		nphDeathHoodCobraOracle:        "Death-Hood Cobra",
		nphTrespassingSouleaterOracle:  "Trespassing Souleater",
		nphBlindingSouleaterOracle:     "Blinding Souleater",
		nphDementiaBatOracle:           "Dementia Bat",
		nphViralDrakeOracle:            "Viral Drake",
		nphSpinebiterOracle:            "Spinebiter",
		nphKilnWalkerOracle:            "Kiln Walker",
		nphPithDrillerOracle:           "Pith Driller",
		nphPriestOfUrabraskOracle:      "Priest of Urabrask",
		nphInquisitorExarchOracle:      "Inquisitor Exarch",
		nphTormentorExarchOracle:       "Tormentor Exarch",
		nphDeceiverExarchOracle:        "Deceiver Exarch",
		nphMarrowShardsOracle:          "Marrow Shards",
		nphVaporSnagOracle:             "Vapor Snag",
		nphGlissasScornOracle:          "Glissa's Scorn",
		nphVictoriousDestructionOracle: "Victorious Destruction",
		nphPsychicBarrierOracle:        "Psychic Barrier",
		nphMutagenicGrowthOracle:       "Mutagenic Growth",
		nphWhipflareOracle:             "Whipflare",
		nphMindcullingOracle:           "Mindculling",
		nphGrimAfflictionOracle:        "Grim Affliction",
		nphCorrosiveGaleOracle:         "Corrosive Gale",
		nphArtillerizeOracle:           "Artillerize",
		nphSickleslicerOracle:          "Sickleslicer",
		nphNecropouncerOracle:          "Necropouncer",
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

// --- printed keywords -----------------------------------------------------------

func TestNph5575B1PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Darksteel Relic", "Artifact", nphDarksteelRelicOracle, 0, 0, []string{"indestructible"}},
		{"Hovermyr", "Artifact Creature — Myr", nphHovermyrOracle, 1, 2, []string{"flying", "vigilance"}},
		{"Porcelain Legionnaire", "Artifact Creature — Phyrexian Soldier", nphPorcelainLegionnaireOracle, 3, 1, []string{"first strike"}},
		{"Shriek Raptor", "Creature — Phyrexian Bird", nphShriekRaptorOracle, 2, 3, []string{"flying", "infect"}},
		{"Slash Panther", "Artifact Creature — Phyrexian Cat", nphSlashPantherOracle, 4, 2, []string{"haste"}},
		{"Spined Thopter", "Artifact Creature — Phyrexian Thopter", nphSpinedThopterOracle, 2, 1, []string{"flying"}},
		{"Spire Monitor", "Creature — Phyrexian Drake", nphSpireMonitorOracle, 3, 3, []string{"flash", "flying"}},
		{"Thundering Tanadon", "Artifact Creature — Phyrexian Beast", nphThunderingTanadonOracle, 5, 4, []string{"trample"}},
		{"Vault Skirge", "Artifact Creature — Phyrexian Imp", nphVaultSkirgeOracle, 1, 1, []string{"flying", "lifelink"}},
		{"Vulshok Refugee", "Creature — Human Warrior", nphVulshokRefugeeOracle, 3, 2, []string{"protection from red"}},
		{"Glistener Elf", "Creature — Phyrexian Elf Warrior", nphGlistenerElfOracle, 1, 1, []string{"infect"}},
		{"Lost Leonin", "Creature — Phyrexian Cat Soldier", nphLostLeoninOracle, 2, 1, []string{"infect"}},
		{"Razor Swine", "Creature — Phyrexian Boar", nphRazorSwineOracle, 2, 1, []string{"first strike", "infect"}},
		{"Ogre Menial", "Creature — Phyrexian Ogre", nphOgreMenialOracle, 0, 4, []string{"infect"}},
		{"Moltensteel Dragon", "Artifact Creature — Phyrexian Dragon", nphMoltensteelDragonOracle, 4, 4, []string{"flying"}},
		{"Toxic Nim", "Creature — Phyrexian Zombie", nphToxicNimOracle, 4, 1, []string{"infect"}},
		{"Fallen Ferromancer", "Creature — Phyrexian Human Shaman", nphFallenFerromancerOracle, 1, 1, []string{"infect"}},
		{"Dementia Bat", "Creature — Phyrexian Bat", nphDementiaBatOracle, 2, 2, []string{"flying"}},
		{"Viral Drake", "Creature — Phyrexian Drake", nphViralDrakeOracle, 1, 4, []string{"flying", "infect"}},
		{"Spinebiter", "Creature — Phyrexian Beast", nphSpinebiterOracle, 3, 4, []string{"infect"}},
		{"Deceiver Exarch", "Creature — Phyrexian Cleric", nphDeceiverExarchOracle, 1, 4, []string{"flash"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			effectstest.AssertKeywords(t, g, id, tc.keywords...)
		})
	}
}

func TestNph5575B1DarksteelRelicSurvivesADestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	relic := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Darksteel Relic", "Artifact", nphDarksteelRelicOracle, false)
	effectstest.B27Kill(g, relic)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(relic) {
		t.Error("an indestructible artifact was destroyed")
	}
}

// --- activated abilities -------------------------------------------------------------

// Each firebreathing row pumps +1/+0; a Phyrexian symbol is paid with its
// colour or with 2 life.
func TestNph5575B1FirebreathingRows(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		mana                   []string
		lifeSymbols            int
	}{
		{"Ogre Menial", "Creature — Phyrexian Ogre", nphOgreMenialOracle, 0, 4, []string{"R"}, 0},
		{"Immolating Souleater", "Artifact Creature — Phyrexian Dog", nphImmolatingSouleaterOracle, 1, 1, []string{"R"}, 0},
		{"Moltensteel Dragon", "Artifact Creature — Phyrexian Dragon", nphMoltensteelDragonOracle, 4, 4, nil, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			effectstest.ApaMana(me, tc.mana...)
			life := me.Life
			effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{PhyrexianLife: tc.lifeSymbols})
			if p, tough := effectstest.PtOf(t, g, id); p != tc.power+1 || tough != tc.toughness {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.power+1, tc.toughness)
			}
			if got := life - me.Life; got != 2*tc.lifeSymbols {
				t.Errorf("paid %d life, want %d", got, 2*tc.lifeSymbols)
			}
		})
	}
}

func TestNph5575B1SelfKeywordGrants(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		row                    int
		mana                   []string
		lifeSymbols            int
		keyword                string
	}{
		{"Insatiable Souleater", "Artifact Creature — Phyrexian Beast", nphInsatiableSouleaterOracle, 0, []string{"G"}, 0, "trample"},
		{"Pestilent Souleater", "Artifact Creature — Phyrexian Insect", nphPestilentSouleaterOracle, 0, nil, 1, "infect"},
		{"Death-Hood Cobra reach", "Creature — Phyrexian Snake", nphDeathHoodCobraOracle, 0, []string{"C", "G"}, 0, "reach"},
		{"Death-Hood Cobra deathtouch", "Creature — Phyrexian Snake", nphDeathHoodCobraOracle, 1, []string{"C", "G"}, 0, "deathtouch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Fatalf("it has %s before the activation", tc.keyword)
			}
			effectstest.ApaMana(me, tc.mana...)
			effectstest.ApaActivate(t, g, me, id, tc.row, game.ActivateAbilityParams{PhyrexianLife: tc.lifeSymbols})
			if !effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("it has no %s after the activation", tc.keyword)
			}
		})
	}
}

func TestNph5575B1TrespassingSouleaterCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	eater := effectstest.B12Push(g, me.ID, "Trespassing Souleater", "Artifact Creature — Phyrexian Construct", nphTrespassingSouleaterOracle, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", nphBearType, 0, 4)
	effectstest.ApaMana(me, "U")
	effectstest.ApaActivate(t, g, me, eater, 0, game.ActivateAbilityParams{})
	effectstest.BrAttack(t, g, eater)
	if err := g.DeclareBlocker(blocker, eater); err == nil {
		t.Error("a creature blocked the Souleater after its ability resolved")
	}
}

func TestNph5575B1ToxicNimRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	nim := effectstest.B12Push(g, me.ID, "Toxic Nim", "Creature — Phyrexian Zombie", nphToxicNimOracle, 4, 1)
	effectstest.ApaMana(me, "B")
	effectstest.ApaActivate(t, g, me, nim, 0, game.ActivateAbilityParams{})
	effectstest.B27Kill(g, nim)
	if !g.Battlefield.Contains(nim) {
		t.Error("the shielded Nim was destroyed")
	}
}

// The ping comes from a source with infect, so a creature is dealt it as
// a -1/-1 counter.
func TestNph5575B1FallenFerromancerPingsWithInfect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	shaman := effectstest.B12Push(g, me.ID, "Fallen Ferromancer", "Creature — Phyrexian Human Shaman", nphFallenFerromancerOracle, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", nphBearType, 2, 2)
	effectstest.ApaMana(me, "C", "R")
	effectstest.ApaActivate(t, g, me, shaman, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if got := effectstest.CounterCount(g, bear, game.CounterMinusOne); got != 1 {
		t.Errorf("the Bear has %d -1/-1 counters, want 1", got)
	}
	if !effectstest.HmlTapped(t, g, shaman) {
		t.Error("the Ferromancer did not tap")
	}
}

func TestNph5575B1BlindingSouleaterTapsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	eater := effectstest.B12Push(g, me.ID, "Blinding Souleater", "Artifact Creature — Phyrexian Cleric", nphBlindingSouleaterOracle, 1, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", nphBearType, 2, 2)
	effectstest.ApaMana(me, "W")
	effectstest.ApaActivate(t, g, me, eater, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !effectstest.HmlTapped(t, g, bear) {
		t.Error("the target creature is untapped")
	}
	if !effectstest.HmlTapped(t, g, eater) {
		t.Error("the Souleater did not tap")
	}
}

func TestNph5575B1DementiaBatSacrificesForTwoDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bat := effectstest.B12Push(g, me.ID, "Dementia Bat", "Creature — Phyrexian Bat", nphDementiaBatOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "C", "C", "C", "C", "B")
	effectstest.ApaActivate(t, g, me, bat, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if !effectstest.InGraveyard(me, bat) {
		t.Error("the Bat was not sacrificed")
	}
	if n := effectstest.DiscardOwed(g, opp.ID); n != 2 {
		t.Errorf("the target player owes %d discards, want 2", n)
	}
}

func TestNph5575B1ViralDrakeProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	drake := effectstest.B12Push(g, me.ID, "Viral Drake", "Creature — Phyrexian Drake", nphViralDrakeOracle, 1, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", nphBearType, 2, 2)
	effectstest.B28AddCounter(t, g, bear, 1)
	effectstest.ApaMana(me, "C", "C", "C", "U")
	effectstest.ApaActivate(t, g, me, drake, 0, game.ActivateAbilityParams{})
	effectstest.AnswerProliferate(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 2 {
		t.Errorf("the Bear has %d +1/+1 counters, want 2", got)
	}
}

// --- static ---------------------------------------------------------------------------

func TestNph5575B1BlockedSpinebiterMayAssignAsThoughUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	biter := effectstest.ApaPush(g, g.Seats[0].ID, g.Seats[0].ID, game.Card{
		Name: "Spinebiter", TypeLine: "Creature — Phyrexian Beast", OracleID: nphSpinebiterOracle, Power: 3, Toughness: 4,
	})
	wall := effectstest.MscWlBlocker(g, "Wall", 0, 5)
	effectstest.MscWlBlockedCombat(t, g, []uuid.UUID{biter}, map[uuid.UUID]uuid.UUID{wall: biter})
	c := effectstest.MscWlDamagePrompt(t, g)
	if f := c.DamageAssignment; !f.AsThoughUnblocked || len(f.BlockerIDs) != 1 || f.AttackerPower != 3 {
		t.Fatalf("the prompt is %+v, want one blocker, power 3, as-though-unblocked", f)
	}
}

// --- triggered abilities ----------------------------------------------------------

func TestNph5575B1KilnWalkerPumpsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	walker := effectstest.B12Push(g, me.ID, "Kiln Walker", "Artifact Creature — Phyrexian Construct", nphKilnWalkerOracle, 0, 3)
	effectstest.DeclareAttack(t, g, them.ID, walker)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, walker); p != 3 || tough != 3 {
		t.Errorf("Kiln Walker is %d/%d while attacking, want 3/3", p, tough)
	}
}

func TestNph5575B1PithDrillerPutsACounterOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", nphBearType, 2, 2)
	effectstest.CastHoldCreature(t, g, "Pith Driller", "Artifact Creature — Phyrexian Horror", nphPithDrillerOracle, 2, 4)
	effectstest.EoeABPickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterMinusOne); got != 1 {
		t.Errorf("the Bear has %d -1/-1 counters, want 1", got)
	}
}

func TestNph5575B1PriestOfUrabraskAddsThreeRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.CastHoldCreature(t, g, "Priest of Urabrask", "Creature — Phyrexian Human Cleric", nphPriestOfUrabraskOracle, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.McPoolCount(me, "R"); got != 3 {
		t.Errorf("red mana in the pool = %d, want 3", got)
	}
}

func TestNph5575B1InquisitorExarchEachMode(t *testing.T) {
	t.Run("gain", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := effectstest.Top100ActiveSeat(g)
		life := me.Life
		effectstest.CastHoldCreature(t, g, "Inquisitor Exarch", "Creature — Phyrexian Cleric", nphInquisitorExarchOracle, 2, 2)
		nphB1ChooseMode(t, g, me.ID, 0, nil)
		if me.Life != life+2 {
			t.Errorf("life %d, want %d", me.Life, life+2)
		}
	})
	t.Run("drain", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
		life, theirs := me.Life, opp.Life
		effectstest.CastHoldCreature(t, g, "Inquisitor Exarch", "Creature — Phyrexian Cleric", nphInquisitorExarchOracle, 2, 2)
		ref := effectstest.PlayerTarget(opp.ID)
		nphB1ChooseMode(t, g, me.ID, 1, &ref)
		if opp.Life != theirs-2 {
			t.Errorf("the opponent is at %d, want %d", opp.Life, theirs-2)
		}
		if me.Life != life {
			t.Errorf("your life moved to %d", me.Life)
		}
	})
}

func TestNph5575B1TormentorExarchEachMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mode         int
		wantP, wantT int
	}{
		{"pump", 0, 4, 3},
		{"shrink", 1, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
			bear := effectstest.B12Creature(g, opp.ID, "Bear", nphBearType, 2, 3)
			effectstest.CastHoldCreature(t, g, "Tormentor Exarch", "Creature — Phyrexian Cleric", nphTormentorExarchOracle, 2, 2)
			nphB1ChooseMode(t, g, me.ID, tc.mode, nphB1CardRef(bear))
			if p, tough := effectstest.PtOf(t, g, bear); p != tc.wantP || tough != tc.wantT {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestNph5575B1DeceiverExarchEachMode(t *testing.T) {
	t.Run("untap", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := effectstest.Top100ActiveSeat(g)
		mine := effectstest.B12Creature(g, me.ID, "Mine", nphBearType, 2, 2)
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, mine).Tapped = true })
		effectstest.CastHoldCreature(t, g, "Deceiver Exarch", "Creature — Phyrexian Cleric", nphDeceiverExarchOracle, 1, 4)
		nphB1ChooseMode(t, g, me.ID, 0, nphB1CardRef(mine))
		if effectstest.HmlTapped(t, g, mine) {
			t.Error("your permanent is still tapped")
		}
	})
	t.Run("tap", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", nphBearType, 2, 2)
		effectstest.CastHoldCreature(t, g, "Deceiver Exarch", "Creature — Phyrexian Cleric", nphDeceiverExarchOracle, 1, 4)
		nphB1ChooseMode(t, g, me.ID, 1, nphB1CardRef(theirs))
		if !effectstest.HmlTapped(t, g, theirs) {
			t.Error("the opponent's permanent is untapped")
		}
	})
}

// --- the spells --------------------------------------------------------------------

func TestNph5575B1MarrowShardsHitsOnlyAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", nphBearType, 1, 3)
	idle := effectstest.B12Creature(g, me.ID, "Idle", nphBearType, 1, 3)
	effectstest.DeclareAttack(t, g, them.ID, attacker)
	id := effectstest.B13HandCard(me, "Marrow Shards", "Instant", nphMarrowShardsOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, attacker).DamageMarked; got != 1 {
		t.Errorf("the attacker has %d damage, want 1", got)
	}
	if got := effectstest.FindCardAnywhere(t, g, idle).DamageMarked; got != 0 {
		t.Errorf("a creature that isn't attacking has %d damage", got)
	}
}

func TestNph5575B1VaporSnagBouncesAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", nphBearType, 2, 2)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Vapor Snag", "Instant", nphVaporSnagOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
	if opp.Life != life-1 {
		t.Errorf("its controller is at %d, want %d", opp.Life, life-1)
	}
}

func TestNph5575B1DestroyAndItsControllerLosesOne(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		good, bad              string
	}{
		{"Glissa's Scorn", "Instant", nphGlissasScornOracle, "Artifact", nphBearType},
		{"Victorious Destruction", "Sorcery", nphVictoriousDestructionOracle, "Basic Land — Forest", "Enchantment"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			opp := effectstest.Top100Opponent(g)
			bad := effectstest.B12Permanent(g, opp.ID, "Bad", tc.bad)
			if err := effectstest.CastCatalogSpellErr(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(bad)); err == nil {
				t.Fatalf("%s targeted a %s", tc.name, tc.bad)
			}
			good := effectstest.B12Permanent(g, opp.ID, "Good", tc.good)
			life := opp.Life
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(good))
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(good) {
				t.Error("the target survived")
			}
			if opp.Life != life-1 {
				t.Errorf("its controller is at %d, want %d", opp.Life, life-1)
			}
		})
	}
}

// You cast a creature, then counter it yourself: its controller, you,
// loses the life.
func TestNph5575B1PsychicBarrierCountersACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.CastCatalogSpell(t, g, "Bear", nphBearType, "", nil)
	life := me.Life
	id := effectstest.B13HandCard(me, "Psychic Barrier", "Instant", nphPsychicBarrierOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bear) || g.Battlefield.Contains(bear) {
		t.Error("the creature spell was not countered")
	}
	if me.Life != life-1 {
		t.Errorf("its controller is at %d, want %d", me.Life, life-1)
	}
}

func TestNph5575B1MutagenicGrowthPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", nphBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Mutagenic Growth", "Instant", nphMutagenicGrowthOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
}

func TestNph5575B1WhipflareSparesArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", nphBearType, 2, 3)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 3)
	effectstest.CastCatalogSpell(t, g, "Whipflare", "Sorcery", nphWhipflareOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, bear).DamageMarked; got != 2 {
		t.Errorf("the nonartifact creature has %d damage, want 2", got)
	}
	if got := effectstest.FindCardAnywhere(t, g, golem).DamageMarked; got != 0 {
		t.Errorf("the artifact creature has %d damage, want 0", got)
	}
}

func TestNph5575B1MindcullingDrawsTwoAndTheOpponentDiscardsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Mindculling", "Sorcery", nphMindcullingOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	// The cast card left the hand, then two were drawn.
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand is %d, want %d", got, hand+2)
	}
	if n := effectstest.DiscardOwed(g, opp.ID); n != 2 {
		t.Errorf("the opponent owes %d discards, want 2", n)
	}
}

func TestNph5575B1GrimAfflictionCountersThenProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", nphBearType, 3, 3)
	effectstest.CastCatalogSpell(t, g, "Grim Affliction", "Instant", nphGrimAfflictionOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProliferate(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterMinusOne); got != 2 {
		t.Errorf("the Bear has %d -1/-1 counters, want 2", got)
	}
}

func TestNph5575B1CorrosiveGaleHitsOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 3, Toughness: 3, Keywords: []string{"flying"}})
	walker := effectstest.B12Creature(g, opp.ID, "Walker", nphBearType, 3, 3)
	id := effectstest.B13HandCard(me, "Corrosive Gale", "Sorcery", nphCorrosiveGaleOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, flyer).DamageMarked; got != 2 {
		t.Errorf("the flyer has %d damage, want 2", got)
	}
	if got := effectstest.FindCardAnywhere(t, g, walker).DamageMarked; got != 0 {
		t.Errorf("the creature without flying has %d damage, want 0", got)
	}
}

func TestNph5575B1ArtillerizeSacrificesForFiveDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	id := effectstest.B13HandCard(me, "Artillerize", "Instant", nphArtillerizeOracle)
	target := []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: target}); err == nil {
		t.Fatal("cast with no sacrifice")
	}
	life := opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: target, SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if opp.Life != life-5 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-5)
	}
}

// --- living weapons ---------------------------------------------------------------

func TestNph5575B1LivingWeapons(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		wantP, wantT int
		haste        bool
	}{
		{"Sickleslicer", nphSickleslicerOracle, 2, 2, false},
		{"Necropouncer", nphNecropouncerOracle, 3, 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			_, germ := effectstest.CastLivingWeapon(t, g, tc.name, tc.oracle)
			if p, tough := effectstest.EffectivePTOf(t, g, germ); p != tc.wantP || tough != tc.wantT {
				t.Errorf("the Germ is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
			if got := effectstest.HasEffectiveKeyword(t, g, germ, "haste"); got != tc.haste {
				t.Errorf("the Germ has haste = %v, want %v", got, tc.haste)
			}
		})
	}
}
