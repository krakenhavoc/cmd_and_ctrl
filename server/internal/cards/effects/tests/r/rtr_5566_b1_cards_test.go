package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rtr_5566_b1_cards_test.go — the first batch of Return to Ravnica cards
// (#5566). Every test plays the real catalog entry through the engine.

var rtrB1Cards = map[string]string{
	"Axebane Stag":           "a45d7135-c042-40d9-818c-74fa1c7af8ad",
	"Catacomb Slug":          "1b0b5264-2810-4ac5-8b2e-8abbf4016287",
	"Cobblebrute":            "6d4ed1f2-0918-444c-95ef-d58beb5aa216",
	"Golgari Longlegs":       "f1534fbf-eab6-4f12-8037-e71d6861b636",
	"Archweaver":             "2ceb76e2-1692-48fc-ab20-9587d1374172",
	"Concordia Pegasus":      "93c35f9e-3346-4269-b1af-df9f216bfd7a",
	"Sunspire Griffin":       "0f51ab3d-aae9-4d8d-a881-e8835dcd7866",
	"Vassal Soul":            "ae2329a1-e37b-44aa-8766-077910fa1197",
	"Daggerdrome Imp":        "894979ba-e4c7-4914-b0d3-133cbaed38c2",
	"Hover Barrier":          "a09cce6f-c957-49d7-af99-a7b2bce786e8",
	"Hussar Patrol":          "bd73581c-c4bd-46cb-8004-75aeafa24288",
	"Minotaur Aggressor":     "7576bfba-eb4f-4d3e-aec5-52daee84d1d6",
	"Rakdos Ragemutt":        "8f4c7419-bbbe-4ddc-83ae-fd9ea3e86c3d",
	"Risen Sanctuary":        "a2198d36-4e8c-4d64-bda5-a95d85398155",
	"Rubbleback Rhino":       "33e87ab4-f680-4463-a48d-289fed7e196e",
	"Skyline Predator":       "9bc6d4ac-6fca-4574-9cd8-10a5528575b3",
	"Tenement Crasher":       "ee11cecc-4772-4bb2-9af8-a2a99eb037f9",
	"Towering Indrik":        "78bc8b7c-ca94-485c-b7bf-53c3724b7339",
	"Trained Caracal":        "a939f01b-afce-4d92-a897-1a6dc3d788c6",
	"Dead Reveler":           "5cee3505-0291-4f1b-9e5f-2909a31acd8e",
	"Gore-House Chainwalker": "dbe3e317-3c39-4b38-ae06-eae02a910c37",
	"Rakdos Cackler":         "763a2d27-52a8-43d9-b649-0f6c0cf87a07",
	"Spawn of Rix Maadi":     "39d4373b-35b4-4ceb-b4d5-9ca474365eb0",
	"Carnival Hellsteed":     "34ae5b49-f8bf-4aa8-bb62-92d6d3118b26",
	"Splatter Thug":          "3baee36b-d64f-448c-ae45-c1ec124c92ec",
	"Thrill-Kill Assassin":   "f13f5705-e3e0-4e1e-a5b3-61e04e23a374",
	"Bloodfray Giant":        "5f83fd2f-c4cb-4bcf-a4db-0238df251740",
	"Slitherhead":            "4b6966f6-052f-4d1d-a3bf-1d3cc91c71d4",
	"Sewer Shambler":         "5d111adc-a91a-437d-8aac-3b246e980bd8",
	"Terrus Wurm":            "0dfbee5a-0dc6-4f5b-abbd-9d8601de7107",
	"Korozda Monitor":        "55195474-c7ea-4c0a-b504-af15e64ec7e9",
	"Zanikev Locust":         "e29012bb-4a1d-409e-8db9-50920fdf68a2",
	"Sluiceway Scorpion":     "bb5ea7c7-741a-4edd-9696-b9881b70afe0",
	"Drudge Beetle":          "f5126c01-f765-44f4-9ced-0fc157f8b897",
	"Centaur Healer":         "ff435f84-73ab-4a16-8af6-6576af4e4edb",
	"Runewing":               "cc300649-ae50-47e7-ad0f-ffbdffddb989",
	"Stealer of Secrets":     "8f9535f2-f00c-40d3-9997-d1c0ece15f9f",
	"Slum Reaper":            "0cf742c1-feb5-4108-9976-37fa58528212",
	"Viashino Racketeer":     "18b09c2d-20cb-40e9-8209-8a1ec2edde97",
	"Keening Apparition":     "b9212596-3f65-41c9-9ee8-1f4cec7f0ecf",
	"Bellows Lizard":         "667dae0b-5006-4e38-be80-eae79c2a2a95",
	"Explosive Impact":       "5eae77d0-7653-4b07-a701-aabea13991bf",
	"Call of the Conclave":   "e1635acd-ed1e-4038-a11a-6518df285253",
	"Goblin Rally":           "9ccb6ce8-ccb6-4809-a1cc-cebd38744815",
	"Ultimate Price":         "d740453b-d8f8-48d0-8c50-72a7cdb791b5",
	"Isperia's Skywatch":     "5101fa05-3cca-499c-8c48-198ba0f8c672",
	"Lyev Skyknight":         "191ec087-7817-4c56-b2d5-66de5c663536",
}

const rtrB1BearType = "Creature — Bear"

// rtrB1TokensOf is every token creature p controls.
func rtrB1TokensOf(g *game.Game, p uuid.UUID) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.IsToken() && c.Controller == p {
				out = append(out, c)
			}
		}
	})
	return out
}

func TestRtr5566B1EveryCardIsRegisteredFull(t *testing.T) {
	if len(rtrB1Cards) != 47 {
		t.Fatalf("%d cards in the batch table, want 47", len(rtrB1Cards))
	}
	for name, oracle := range rtrB1Cards {
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

// --- printed keywords ------------------------------------------------------

func TestRtr5566B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine string
		keywords       []string
	}{
		{"Archweaver", "Creature — Spider", []string{"reach", "trample"}},
		{"Concordia Pegasus", "Creature — Pegasus", []string{"flying"}},
		{"Sunspire Griffin", "Creature — Griffin", []string{"flying"}},
		{"Vassal Soul", "Creature — Spirit", []string{"flying"}},
		{"Daggerdrome Imp", "Creature — Imp", []string{"flying", "lifelink"}},
		{"Hover Barrier", "Creature — Illusion Wall", []string{"defender", "flying"}},
		{"Hussar Patrol", "Creature — Human Knight", []string{"flash", "vigilance"}},
		{"Minotaur Aggressor", "Creature — Minotaur Berserker", []string{"first strike", "haste"}},
		{"Rakdos Ragemutt", "Creature — Elemental Dog", []string{"lifelink", "haste"}},
		{"Risen Sanctuary", "Creature — Elemental", []string{"vigilance"}},
		{"Rubbleback Rhino", "Creature — Rhino", []string{"hexproof"}},
		{"Skyline Predator", "Creature — Drake", []string{"flash", "flying"}},
		{"Tenement Crasher", "Creature — Beast", []string{"haste"}},
		{"Towering Indrik", "Creature — Beast", []string{"reach"}},
		{"Trained Caracal", "Creature — Cat", []string{"lifelink"}},
		{"Carnival Hellsteed", "Creature — Nightmare Horse", []string{"first strike", "haste"}},
		{"Splatter Thug", "Creature — Human Warrior", []string{"first strike"}},
		{"Thrill-Kill Assassin", "Creature — Human Assassin", []string{"deathtouch"}},
		{"Bloodfray Giant", "Creature — Giant", []string{"trample"}},
		{"Sewer Shambler", "Creature — Zombie", []string{"swampwalk"}},
		{"Korozda Monitor", "Creature — Lizard", []string{"trample"}},
		{"Zanikev Locust", "Creature — Insect", []string{"flying"}},
		{"Sluiceway Scorpion", "Creature — Scorpion", []string{"deathtouch"}},
		{"Runewing", "Creature — Bird", []string{"flying"}},
		{"Isperia's Skywatch", "Creature — Vedalken Knight", []string{"flying"}},
		{"Lyev Skyknight", "Creature — Human Knight", []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, rtrB1Cards[tc.name], 2, 2)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Flash: Skyline Predator and Hussar Patrol are cast outside a main phase.
func TestRtr5566B1FlashCreaturesCastAtBeginningOfCombat(t *testing.T) {
	for _, tc := range []struct{ Name, TypeLine string }{
		{"Skyline Predator", "Creature — Drake"},
		{"Hussar Patrol", "Creature — Human Knight"},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.B13HandCard(me, tc.Name, tc.TypeLine, rtrB1Cards[tc.Name])
			effectstest.AdvanceTo(t, g, game.StepBeginCombat)
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("%s at the beginning of combat: %v", tc.Name, err)
			}
		})
	}
}

// --- unleash ---------------------------------------------------------------

func TestRtr5566B1UnleashEntersWithACounterAndCantBlock(t *testing.T) {
	cases := []struct {
		name, typeLine string
		power, tough   int
	}{
		{"Dead Reveler", "Creature — Zombie", 2, 3},
		{"Gore-House Chainwalker", "Creature — Human Warrior", 2, 1},
		{"Rakdos Cackler", "Creature — Devil", 1, 1},
		{"Spawn of Rix Maadi", "Creature — Horror", 5, 3},
		{"Carnival Hellsteed", "Creature — Nightmare Horse", 5, 4},
		{"Splatter Thug", "Creature — Human Warrior", 2, 2},
		{"Thrill-Kill Assassin", "Creature — Human Assassin", 1, 2},
		{"Bloodfray Giant", "Creature — Giant", 4, 3},
	}
	for _, tc := range cases {
		for _, apply := range []bool{true, false} {
			t.Run(tc.name, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, rtrB1Cards[tc.name], tc.power, tc.tough)
				answerUnleash(t, g, apply)
				effectstest.PassPriorityAroundTable(t, g)
				if !effectstest.BattlefieldHas(g, id) {
					t.Fatalf("%s did not enter", tc.name)
				}
				want := 0
				if apply {
					want = 1
				}
				if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != want {
					t.Errorf("unleash %v: %d +1/+1 counters, want %d", apply, got, want)
				}
				if got := game.Restricted(effectstest.ApaLive(g, id), game.CantBlock); got != apply {
					t.Errorf("unleash %v: can't block = %v", apply, got)
				}
			})
		}
	}
}

// --- scavenge --------------------------------------------------------------

func TestRtr5566B1ScavengeForItsPower(t *testing.T) {
	cases := []struct {
		name, typeLine, mana string
		power, tough         int
	}{
		{"Slitherhead", "Creature — Plant Zombie", "", 1, 1},
		{"Sewer Shambler", "Creature — Zombie", "{C}{C}{B}", 2, 1},
		{"Terrus Wurm", "Creature — Zombie Wurm", "{C}{C}{C}{C}{C}{C}{B}", 5, 5},
		{"Korozda Monitor", "Creature — Lizard", "{C}{C}{C}{C}{C}{G}{G}", 3, 3},
		{"Zanikev Locust", "Creature — Insect", "{C}{C}{B}{B}", 3, 3},
		{"Sluiceway Scorpion", "Creature — Scorpion", "{C}{B}{G}", 2, 2},
		{"Drudge Beetle", "Creature — Insect", "{C}{C}{C}{C}{C}{G}", 2, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.B12Creature(g, me.ID, "Bear", rtrB1BearType, 2, 2)
			id, _ := effectstest.ActivateFromGraveyard(t, g, tc.name, tc.typeLine, rtrB1Cards[tc.name], tc.power, tc.tough,
				tc.mana, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != tc.power {
				t.Errorf("the target got %d counters, want %d", got, tc.power)
			}
			if me.Graveyard.Contains(id) {
				t.Errorf("%s is still in the graveyard", tc.name)
			}
		})
	}
}

// --- triggered abilities ---------------------------------------------------

func TestRtr5566B1CentaurHealerGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Centaur Healer", "Creature — Centaur Cleric", rtrB1Cards["Centaur Healer"], 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestRtr5566B1RunewingDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wing := effectstest.B12Push(g, me.ID, "Runewing", "Creature — Bird", rtrB1Cards["Runewing"], 2, 2)
	hand := me.Hand.Size()
	effectstest.B27Kill(g, wing)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestRtr5566B1StealerOfSecretsDrawsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stealer := effectstest.B12Push(g, me.ID, "Stealer of Secrets", "Creature — Human Rogue", rtrB1Cards["Stealer of Secrets"], 2, 2)
	hand := me.Hand.Size()
	effectstest.DealCombatDamage(g, me.ID, stealer, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

// Each player has exactly one candidate, so no one is asked: every
// opponent's creature and the Reaper itself are sacrificed.
func TestRtr5566B1SlumReaperEachPlayerSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var victims []uuid.UUID
	for _, s := range g.Seats {
		if s.ID != me.ID {
			victims = append(victims, effectstest.B12Creature(g, s.ID, "Victim", rtrB1BearType, 2, 2))
		}
	}
	reaper := effectstest.CastHoldCreature(t, g, "Slum Reaper", "Creature — Horror", rtrB1Cards["Slum Reaper"], 4, 2)
	effectstest.PassPriorityAroundTable(t, g)
	for i, v := range victims {
		if effectstest.BattlefieldHas(g, v) {
			t.Errorf("opponent %d kept its creature", i)
		}
	}
	if effectstest.BattlefieldHas(g, reaper) {
		t.Error("the Reaper's controller sacrificed nothing")
	}
}

func TestRtr5566B1ViashinoRacketeerLootsOnlyForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Drawn")
	discard := effectstest.PushHandCard(g, me)
	effectstest.CastHoldCreature(t, g, "Viashino Racketeer", "Creature — Lizard Rogue", rtrB1Cards["Viashino Racketeer"], 2, 1)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, discard)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand after discard one, draw one: %d, want %d", got, hand)
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the chosen card was not discarded")
	}
}

func TestRtr5566B1DetainOnEntry(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		power, tough   int
	}{
		{"Isperia's Skywatch", "Creature — Vedalken Knight", 3, 3},
		{"Lyev Skyknight", "Creature — Human Knight", 3, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := effectstest.B12Creature(g, me.ID, "My Bear", rtrB1BearType, 2, 2)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", rtrB1BearType, 2, 2)
			land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
			effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, rtrB1Cards[tc.name], tc.power, tc.tough)
			pick := effectstest.LatestPickTarget(g, me.ID)
			if pick == nil {
				t.Fatal("no target asked")
			}
			if effectstest.HasID(pick.PickTargetCards, mine) {
				t.Error("a creature I control was offered")
			}
			if effectstest.HasID(pick.PickTargetCards, land) {
				t.Error("a land was offered")
			}
			effectstest.PickCard(t, g, me.ID, theirs)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AssertRestrictions(t, g, theirs, effectstest.BnsB10Detained)
			effectstest.AssertRestrictions(t, g, mine, 0)
		})
	}
}

// Lyev Skyknight's detain reaches a noncreature permanent.
func TestRtr5566B1LyevSkyknightDetainsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.CastHoldCreature(t, g, "Lyev Skyknight", "Creature — Human Knight", rtrB1Cards["Lyev Skyknight"], 3, 1)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, rock, effectstest.BnsB10Detained)
}

// --- activated abilities ---------------------------------------------------

func TestRtr5566B1KeeningApparitionDestroysAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	spirit := effectstest.B12Push(g, me.ID, "Keening Apparition", "Creature — Spirit", rtrB1Cards["Keening Apparition"], 2, 2)
	creature := effectstest.B12Creature(g, opp.ID, "Bear", rtrB1BearType, 2, 2)
	aura := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	if err := g.ActivateCatalogAbility(me.ID, spirit, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(creature)}); err == nil {
		t.Fatal("the Apparition targeted a creature")
	}
	effectstest.ApaActivate(t, g, me, spirit, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(aura)})
	if effectstest.BattlefieldHas(g, aura) {
		t.Error("the enchantment survived")
	}
	if effectstest.BattlefieldHas(g, spirit) {
		t.Error("the Apparition was not sacrificed")
	}
}

func TestRtr5566B1BellowsLizardFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	lizard := effectstest.B12Push(g, me.ID, "Bellows Lizard", "Creature — Lizard", rtrB1Cards["Bellows Lizard"], 1, 1)
	for range 2 {
		effectstest.ApaMana(me, "R", "R")
		effectstest.ApaActivate(t, g, me, lizard, 0, game.ActivateAbilityParams{})
	}
	if p, tt := effectstest.EffectivePTOf(t, g, lizard); p != 3 || tt != 1 {
		t.Errorf("%d/%d, want 3/1", p, tt)
	}
	if err := g.ActivateCatalogAbility(me.ID, lizard, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no mana")
	}
}

// --- the spells ------------------------------------------------------------

func TestRtr5566B1ExplosiveImpactDealsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Explosive Impact", "Instant", rtrB1Cards["Explosive Impact"], []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-5)
	}
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 6, 6)
	effectstest.CastCatalogSpell(t, g, "Explosive Impact", "Instant", rtrB1Cards["Explosive Impact"], effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 5 {
		t.Errorf("the creature has %d damage, want 5", got)
	}
}

func TestRtr5566B1TokenSorceries(t *testing.T) {
	for _, tc := range []struct {
		name         string
		n, p, toughs int
		color        string
	}{
		{"Call of the Conclave", 1, 3, 3, "G"},
		{"Goblin Rally", 4, 1, 1, "R"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			effectstest.CastCatalogSpell(t, g, tc.name, "Sorcery", rtrB1Cards[tc.name], nil)
			effectstest.PassPriorityAroundTable(t, g)
			tokens := rtrB1TokensOf(g, me.ID)
			if len(tokens) != tc.n {
				t.Fatalf("%d tokens, want %d", len(tokens), tc.n)
			}
			for _, tok := range tokens {
				if tok.Power != tc.p || tok.Toughness != tc.toughs || !tok.HasColor(tc.color) {
					t.Errorf("token %s is %d/%d colors %v, want %d/%d %s", tok.Name, tok.Power, tok.Toughness, tok.Colors, tc.p, tc.toughs, tc.color)
				}
			}
		})
	}
}

func TestRtr5566B1UltimatePriceDestroysOnlyAMonocoloredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	gold := effectstest.B12Creature(g, opp.ID, "Gold", rtrB1BearType, 2, 2)
	effectstest.HmlSetColors(g, gold, "G", "W")
	grey := effectstest.B12Creature(g, opp.ID, "Grey", "Artifact Creature — Golem", 2, 2)
	green := effectstest.B12Creature(g, opp.ID, "Green", rtrB1BearType, 2, 2)
	effectstest.HmlSetColors(g, green, "G")
	for _, id := range []uuid.UUID{gold, grey} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Ultimate Price", "Instant", rtrB1Cards["Ultimate Price"], effectstest.CardTarget(id)); err == nil {
			t.Fatal("Ultimate Price targeted a creature that isn't monocolored")
		}
	}
	effectstest.CastCatalogSpell(t, g, "Ultimate Price", "Instant", rtrB1Cards["Ultimate Price"], effectstest.CardTarget(green))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, green) {
		t.Error("the monocolored creature survived")
	}
}

// Rubbleback Rhino's hexproof: an opponent's Ultimate Price can't target it.
func TestRtr5566B1RubblebackRhinoHasHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	rhino := effectstest.B12Push(g, opp.ID, "Rubbleback Rhino", "Creature — Rhino", rtrB1Cards["Rubbleback Rhino"], 3, 4)
	effectstest.HmlSetColors(g, rhino, "G")
	if err := effectstest.CastCatalogSpellErr(t, g, "Ultimate Price", "Instant", rtrB1Cards["Ultimate Price"], effectstest.CardTarget(rhino)); err == nil {
		t.Error("an opponent's spell targeted a creature with hexproof")
	}
}
