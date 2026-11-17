package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nem_b2_cards_test.go — Nemesis, batch 2 (#5618). Each test plays the
// real catalog entry and asserts what its text says.

const (
	nemB2AirBladderOracle      = "25c8a3b0-c393-4fc6-a373-808cd8e9850c"
	nemB2BattlefieldPercherOr  = "2836c144-7268-411a-b651-351cc4ef8fe9"
	nemB2BelbesPercherOracle   = "81b51bb1-89b9-46a0-ad7a-a43c0d214f11"
	nemB2BolaWarriorOracle     = "13f41253-364f-40a5-a7cd-84d8785b2460"
	nemB2CarrionWallOracle     = "ce8750d4-1b74-4b02-afd8-6b639c1a9a78"
	nemB2ChieftainOracle       = "8aae3a28-dcf3-460e-89f9-4b0c96d6f55e"
	nemB2DarkTriumphOracle     = "885630a8-c43b-4804-acd6-14a7c7321d73"
	nemB2DefiantFalconOracle   = "dc206e6a-e447-48d4-a91a-7c746320e2b4"
	nemB2DownhillChargeOracle  = "8018419c-ea19-499b-a33a-7c20b35bce7b"
	nemB2OverseerOracle        = "89a3a6a3-e5f9-4e28-8372-99219db3f976"
	nemB2SlideOracle           = "ad779760-1197-4248-8e21-7a048122640f"
	nemB2StrikeOracle          = "d67048eb-eaf1-4b5b-a8e3-a004bab438f3"
	nemB2SurgeOracle           = "fb1755a0-3334-419b-8cb5-5a3ac7fa5b13"
	nemB2ThopterOracle         = "a3169d1a-fd8d-474e-9e9f-1bd7db510e8b"
	nemB2InfiltrateOracle      = "081764b4-ef96-44aa-836c-05353efd215c"
	nemB2MindSwordsOracle      = "4af20f8c-2ba0-49d5-a64e-d3668cec30cd"
	nemB2MoggcatcherOracle     = "4845dc21-7534-4896-b298-4d44cd7a3b27"
	nemB2NetterOracle          = "8cd65dcd-e36d-47b1-b011-961c79004cb0"
	nemB2OffBalanceOracle      = "38d9401e-2840-455d-8478-8c976588d4ab"
	nemB2PlagueWitchOracle     = "532ab335-4873-405b-84b3-c447207172b5"
	nemB2RathiFiendOracle      = "b336d819-f299-4de3-9b23-51f83da59efb"
	nemB2IntimidatorOracle     = "a24f4198-2412-4b73-8e20-4ced3a826301"
	nemB2SeahunterOracle       = "59e1899f-a9e8-48aa-b3fb-0b7d9fd6859c"
	nemB2SealOfDoomOracle      = "71d54ba3-91fe-4a14-8b99-c284b4c75013"
	nemB2SealOfStrengthOracle  = "e41a68b3-e1cb-4f51-bd54-68882d2cc015"
	nemB2ShriekingMoggOracle   = "088a63d4-22d2-4061-8aef-2aa3fc348639"
	nemB2SilkenfistFighterOr   = "bf5ceba2-9def-4ff4-9291-740ca90c1c4d"
	nemB2SilkenfistOrderOracle = "5cca974b-1a6c-485f-adc0-81d2b7dca286"
	nemB2PoacherOracle         = "4f5921c1-b932-4d6b-bb8e-01992578abdc"
	nemB2HomunculusOracle      = "0b76bb3b-1ada-4c78-bd13-3464371d4fa0"
	nemB2ThugOracle            = "78bbd213-e8f4-40da-aef5-853244e95756"
	nemB2StampedeDriverOracle  = "3c652a31-3b56-4050-9613-d0481e54dfd7"
	nemB2BiologistOracle       = "1e2c4508-54ad-45c2-b35f-678f8c376b27"
	nemB2DisciplineOracle      = "90e46d04-9d53-4a24-aff0-f8c61e1db97d"
	nemB2MachinistOracle       = "50bbbb39-4709-478f-bb51-716c7e4f8323"
	nemB2ZeppelinOracle        = "dbd86a37-8ac1-4bdb-9ba1-7caee99ab5f7"
	nemB2TricksterMageOracle   = "d5a9fd97-535c-4aa4-9777-6cf8a413b6ec"
	nemB2ViciousHungerOracle   = "38ea22cd-2c5d-4f66-a111-207aca4c67c3"
	nemB2WanderingEyeOracle    = "59015068-6868-4235-981b-96137123c685"
)

func TestNemB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		nemB2AirBladderOracle:      "Air Bladder",
		nemB2BattlefieldPercherOr:  "Battlefield Percher",
		nemB2BelbesPercherOracle:   "Belbe's Percher",
		nemB2BolaWarriorOracle:     "Bola Warrior",
		nemB2CarrionWallOracle:     "Carrion Wall",
		nemB2ChieftainOracle:       "Chieftain en-Dal",
		nemB2DarkTriumphOracle:     "Dark Triumph",
		nemB2DefiantFalconOracle:   "Defiant Falcon",
		nemB2DownhillChargeOracle:  "Downhill Charge",
		nemB2OverseerOracle:        "Flowstone Overseer",
		nemB2SlideOracle:           "Flowstone Slide",
		nemB2StrikeOracle:          "Flowstone Strike",
		nemB2SurgeOracle:           "Flowstone Surge",
		nemB2ThopterOracle:         "Flowstone Thopter",
		nemB2InfiltrateOracle:      "Infiltrate",
		nemB2MindSwordsOracle:      "Mind Swords",
		nemB2MoggcatcherOracle:     "Moggcatcher",
		nemB2NetterOracle:          "Netter en-Dal",
		nemB2OffBalanceOracle:      "Off Balance",
		nemB2PlagueWitchOracle:     "Plague Witch",
		nemB2RathiFiendOracle:      "Rathi Fiend",
		nemB2IntimidatorOracle:     "Rathi Intimidator",
		nemB2SeahunterOracle:       "Seahunter",
		nemB2SealOfDoomOracle:      "Seal of Doom",
		nemB2SealOfStrengthOracle:  "Seal of Strength",
		nemB2ShriekingMoggOracle:   "Shrieking Mogg",
		nemB2SilkenfistFighterOr:   "Silkenfist Fighter",
		nemB2SilkenfistOrderOracle: "Silkenfist Order",
		nemB2PoacherOracle:         "Skyshroud Poacher",
		nemB2HomunculusOracle:      "Sneaky Homunculus",
		nemB2ThugOracle:            "Spineless Thug",
		nemB2StampedeDriverOracle:  "Stampede Driver",
		nemB2BiologistOracle:       "Stronghold Biologist",
		nemB2DisciplineOracle:      "Stronghold Discipline",
		nemB2MachinistOracle:       "Stronghold Machinist",
		nemB2ZeppelinOracle:        "Stronghold Zeppelin",
		nemB2TricksterMageOracle:   "Trickster Mage",
		nemB2ViciousHungerOracle:   "Vicious Hunger",
		nemB2WanderingEyeOracle:    "Wandering Eye",
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

// nemB2CastCreature casts a creature card with printed power and
// toughness from the active seat's hand at a main phase, so it survives
// the state-based check once it resolves.
func nemB2CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := handCardAtMain(t, g, name, typeLine, oracle)
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == id {
			me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = power, toughness
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// nemB2Activate activates ability 0 of `id` under the strict mana gate
// after adding `mana` to the pool.
func nemB2Activate(t *testing.T, g *game.Game, player, id uuid.UUID, mana string, p game.ActivateAbilityParams) {
	t.Helper()
	if mana != "" {
		dcMana(t, g, player, mana)
	}
	p.Strict = true
	if err := g.ActivateCatalogAbility(player, id, 0, p); err != nil {
		t.Fatalf("activate: %v", err)
	}
}

// --- the Perchers and Air Bladder ------------------------------------------------

func TestNemB2PerchersBlockOnlyFlyers(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Belbe's Percher", nemB2BelbesPercherOracle},
		{"Battlefield Percher", nemB2BattlefieldPercherOr},
		{"Stronghold Zeppelin", nemB2ZeppelinOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, them := seatsForCombat(g)
			percher := nemB1Push(g, them.ID, tc.name, "Creature — Bird", tc.oracle, 2, 2, "B")
			if !hasEffectiveKeyword(t, g, percher, "flying") {
				t.Fatalf("%s does not have flying", tc.name)
			}
			ground := nemB1Push(g, me.ID, "Ground attacker", "Creature — Bear", "", 2, 2, "G")
			flyer := nemB1Flyer(g, me.ID, "Flying attacker", "U")
			attackIntoBlocks(t, g, them.ID, ground, flyer)
			if blockRefusal(t, g, percher, ground) == "" {
				t.Errorf("%s blocked a creature without flying", tc.name)
			}
			if got := blockRefusal(t, g, percher, flyer); got != "" {
				t.Errorf("%s was refused a flyer: %q", tc.name, got)
			}
		})
	}
}

func TestNemB2BattlefieldPercherPumps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	percher := nemB1Push(g, me.ID, "Battlefield Percher", "Creature — Bird", nemB2BattlefieldPercherOr, 2, 2, "B")
	nemB2Activate(t, g, me.ID, percher, "{C}{B}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, percher); p != 3 || tough != 3 {
		t.Errorf("Battlefield Percher is %d/%d, want 3/3", p, tough)
	}
}

func TestNemB2AirBladderGivesFlyingAndLimitsBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	bear := nemB1Push(g, them.ID, "Their Bear", "Creature — Bear", "", 2, 2, "G")
	castCatalogSpell(t, g, "Air Bladder", "Enchantment — Aura", nemB2AirBladderOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Fatal("the enchanted creature does not have flying")
	}
	ground := nemB1Push(g, me.ID, "Ground attacker", "Creature — Bear", "", 2, 2, "G")
	flyer := nemB1Flyer(g, me.ID, "Flying attacker", "U")
	attackIntoBlocks(t, g, them.ID, ground, flyer)
	if blockRefusal(t, g, bear, ground) == "" {
		t.Error("the enchanted creature blocked a creature without flying")
	}
	if got := blockRefusal(t, g, bear, flyer); got != "" {
		t.Errorf("the enchanted creature was refused a flyer: %q", got)
	}
}

// --- the Spellshapers ------------------------------------------------------------

func TestNemB2SpellshapersNeedADiscard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	witch := nemB1Push(g, me.ID, "Plague Witch", "Creature — Elf Spellshaper", nemB2PlagueWitchOracle, 1, 1, "B")
	victim := nemB1Push(g, them.ID, "Victim", "Creature — Test", "", 1, 1, "G")
	dcMana(t, g, me.ID, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(victim)}); err == nil {
		t.Fatal("Plague Witch was activated without discarding a card")
	}
	fodder := handCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardTarget(victim), DiscardIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("Plague Witch: %v", err)
	}
	if !me.Graveyard.Contains(fodder) || !b16Tapped(t, g, witch) {
		t.Fatal("the cost did not discard the card and tap the Witch")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 1/1 survived -1/-1")
	}
}

// nemB2Shape activates a spellshaper's ability at `target`, paying with
// a fresh card from hand.
func nemB2Shape(t *testing.T, g *game.Game, shaper uuid.UUID, mana string, targets []game.TargetRef) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := handCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	nemB2Activate(t, g, me.ID, shaper, mana, game.ActivateAbilityParams{Targets: targets, DiscardIDs: []uuid.UUID{fodder}})
}

func TestNemB2BolaWarriorStopsABlock(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	warrior := nemB1Push(g, me.ID, "Bola Warrior", "Creature — Human Spellshaper Warrior", nemB2BolaWarriorOracle, 1, 1, "R")
	bear := nemB1Push(g, them.ID, "Their Bear", "Creature — Bear", "", 2, 2, "G")
	other := nemB1Push(g, them.ID, "Their Other", "Creature — Bear", "", 2, 2, "G")
	attacker := nemB1Push(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2, "G")
	nemB2Shape(t, g, warrior, "{R}", cardTarget(bear))
	passPriorityAroundTable(t, g)
	attackIntoBlocks(t, g, them.ID, attacker)
	if blockRefusal(t, g, bear, attacker) == "" {
		t.Error("the target blocked")
	}
	if got := blockRefusal(t, g, other, attacker); got != "" {
		t.Errorf("another creature was refused: %q", got)
	}
}

func TestNemB2NetterEnDalStopsAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	netter := nemB1Push(g, me.ID, "Netter en-Dal", "Creature — Human Spellshaper", nemB2NetterOracle, 1, 1, "W")
	held := nemB1Push(g, me.ID, "Held", "Creature — Bear", "", 2, 2, "G")
	free := nemB1Push(g, me.ID, "Free", "Creature — Bear", "", 2, 2, "G")
	nemB2Shape(t, g, netter, "{W}", cardTarget(held))
	passPriorityAroundTable(t, g)
	advanceToStepInTurn(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(held, them.ID); err == nil {
		t.Error("the target attacked")
	}
	if err := g.DeclareAttacker(free, them.ID); err != nil {
		t.Errorf("another creature was refused: %v", err)
	}
}

func TestNemB2StrongholdCountersSplitByCreature(t *testing.T) {
	for _, tc := range []struct {
		name, oracle                      string
		counters, refuses                 string
		countersTypeLine, refusesTypeLine string
	}{
		{"Stronghold Biologist", nemB2BiologistOracle, "Grizzly Bears", "Divination", "Creature — Bear", "Sorcery"},
		{"Stronghold Machinist", nemB2MachinistOracle, "Divination", "Grizzly Bears", "Sorcery", "Creature — Bear"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			shaper := nemB1Push(g, me.ID, tc.name, "Creature — Human Spellshaper", tc.oracle, 1, 1, "U")
			wrong := handCardFull(me, tc.refuses, tc.refusesTypeLine, "{1}", "", nil)
			if err := g.CastSpell(me.ID, wrong, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast %s: %v", tc.refuses, err)
			}
			fodder := handCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
			dcMana(t, g, me.ID, "{U}{U}")
			if err := g.ActivateCatalogAbility(me.ID, shaper, 0, game.ActivateAbilityParams{
				Strict: true, Targets: cardTarget(wrong), DiscardIDs: []uuid.UUID{fodder},
			}); err == nil {
				t.Fatalf("%s targeted a %s", tc.name, tc.refusesTypeLine)
			}
			passPriorityAroundTable(t, g)
			right := handCardFull(me, tc.counters, tc.countersTypeLine, "{1}", "", nil)
			if err := g.CastSpell(me.ID, right, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast %s: %v", tc.counters, err)
			}
			nemB2Shape(t, g, shaper, "{U}{U}", cardTarget(right))
			passPriorityAroundTable(t, g)
			if !me.Graveyard.Contains(right) || g.Battlefield.Contains(right) {
				t.Errorf("%s was not countered", tc.counters)
			}
		})
	}
}

func TestNemB2StampedeDriverPumpsYourTeam(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	driver := nemB1Push(g, me.ID, "Stampede Driver", "Creature — Human Spellshaper", nemB2StampedeDriverOracle, 1, 1, "G")
	bear := nemB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2, "G")
	theirs := nemB1Push(g, them.ID, "Their Bear", "Creature — Bear", "", 2, 2, "G")
	nemB2Shape(t, g, driver, "{C}{G}", nil)
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{driver: 2, bear: 3} {
		if p, tough := ptOf(t, g, id); p != want || tough != want {
			t.Errorf("your creature is %d/%d, want %d/%d", p, tough, want, want)
		}
		if !hasEffectiveKeyword(t, g, id, "trample") {
			t.Error("your creature did not gain trample")
		}
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 || hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Errorf("the opponent's creature changed: %d/%d", p, tough)
	}
}

func TestNemB2TricksterMageTapsOrUntaps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	mage := nemB1Push(g, me.ID, "Trickster Mage", "Creature — Human Spellshaper", nemB2TricksterMageOracle, 1, 1, "U")
	land := seedLand(g, them.ID, "Forest", "Basic Land — Forest", "")
	nemB2Shape(t, g, mage, "{U}", cardTarget(land))
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	if !b16Tapped(t, g, land) {
		t.Fatal("the land was not tapped")
	}
	findBattlefieldCardForTest(g, mage).Tapped = false
	nemB2Shape(t, g, mage, "{U}", cardTarget(land))
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 1)
	if b16Tapped(t, g, land) {
		t.Error("the land was not untapped")
	}
}

// --- the Seals -------------------------------------------------------------------

func TestNemB2SealOfDoomDestroysANonblackCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	seal := nemB1Push(g, me.ID, "Seal of Doom", "Enchantment", nemB2SealOfDoomOracle, 0, 0, "B")
	black := nemB1Push(g, them.ID, "Black", "Creature — Zombie", "", 2, 2, "B")
	green := nemB1Push(g, them.ID, "Green", "Creature — Bear", "", 2, 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, seal, 0, game.ActivateAbilityParams{Targets: cardTarget(black)}); err == nil {
		t.Fatal("Seal of Doom targeted a black creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, seal, 0, game.ActivateAbilityParams{Targets: cardTarget(green)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(seal) {
		t.Error("the Seal was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the nonblack creature survived")
	}
}

func TestNemB2SealOfStrengthPumps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seal := nemB1Push(g, me.ID, "Seal of Strength", "Enchantment", nemB2SealOfStrengthOracle, 0, 0, "G")
	bear := nemB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, seal, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 || !me.Graveyard.Contains(seal) {
		t.Errorf("the creature is %d/%d, want 5/5 with the Seal sacrificed", p, tough)
	}
}

// --- Flowstone -------------------------------------------------------------------

func TestNemB2FlowstoneOverseerPumpsATarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	overseer := nemB1Push(g, me.ID, "Flowstone Overseer", "Creature — Beast", nemB2OverseerOracle, 4, 4, "R")
	frail := nemB1Push(g, them.ID, "Frail", "Creature — Test", "", 3, 1, "G")
	bear := nemB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	nemB2Activate(t, g, me.ID, overseer, "{R}{R}", game.ActivateAbilityParams{Targets: cardTarget(frail)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(frail) {
		t.Error("an X/1 survived +1/-1")
	}
	nemB2Activate(t, g, me.ID, overseer, "{R}{R}", game.ActivateAbilityParams{Targets: cardTarget(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 1 {
		t.Errorf("the 2/2 is %d/%d, want 3/1", p, tough)
	}
}

func TestNemB2FlowstoneSlideGivesEveryCreaturePlusXMinusX(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	mine := nemB1Push(g, me.ID, "My Giant", "Creature — Giant", "", 3, 4, "R")
	bear := nemB1Push(g, them.ID, "Their Bear", "Creature — Bear", "", 2, 2, "G")
	slide := handCardAtMain(t, g, "Flowstone Slide", "Sorcery", nemB2SlideOracle)
	if err := g.CastSpell(me.ID, slide, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 survived +2/-2")
	}
	if p, tough := ptOf(t, g, mine); p != 5 || tough != 2 {
		t.Errorf("the caster's 3/4 is %d/%d, want 5/2", p, tough)
	}
}

func TestNemB2FlowstoneStrikeGivesPlusOneMinusOneAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := nemB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	castCatalogSpell(t, g, "Flowstone Strike", "Instant", nemB2StrikeOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 1 {
		t.Errorf("the 2/2 is %d/%d, want 3/1", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the target did not gain haste")
	}
}

func TestNemB2FlowstoneSurgeIsYourAnthem(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	nemB1Push(g, me.ID, "Flowstone Surge", "Enchantment", nemB2SurgeOracle, 0, 0, "R")
	mine := nemB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2, "G")
	theirs := nemB1Push(g, them.ID, "Their Bear", "Creature — Bear", "", 2, 2, "G")
	if p, tough := ptOf(t, g, mine); p != 3 || tough != 1 {
		t.Errorf("your 2/2 is %d/%d, want 3/1", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the opponent's 2/2 is %d/%d, want 2/2", p, tough)
	}
}

func TestNemB2FlowstoneThopterPumpsAndFlies(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	thopter := nemB1Push(g, me.ID, "Flowstone Thopter", "Artifact Creature — Thopter", nemB2ThopterOracle, 4, 4)
	if hasEffectiveKeyword(t, g, thopter, "flying") {
		t.Fatal("the Thopter flies before the ability")
	}
	nemB2Activate(t, g, me.ID, thopter, "{C}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, thopter); p != 5 || tough != 3 || !hasEffectiveKeyword(t, g, thopter, "flying") {
		t.Errorf("the Thopter is %d/%d, flying %v; want a 5/3 flyer", p, tough, hasEffectiveKeyword(t, g, thopter, "flying"))
	}
}

// --- alternative costs -----------------------------------------------------------

func TestNemB2SwampSpellsSacrificeACreatureInstead(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Dark Triumph", "Instant", nemB2DarkTriumphOracle},
		{"Mind Swords", "Sorcery", nemB2MindSwordsOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := handCardAtMain(t, g, tc.name, tc.typeLine, tc.oracle)
			fodder := nemB1Push(g, me.ID, "Fodder", "Creature — Bear", "", 2, 2, "G")
			if err := castForAltCost(g, id, "", "sacrifice", []uuid.UUID{fodder}, nil); err == nil {
				t.Fatalf("%s was cast for a creature with no Swamp", tc.name)
			}
			if !g.Battlefield.Contains(fodder) || !me.Hand.Contains(id) {
				t.Fatal("the refused cast paid something anyway")
			}
			seedLand(g, me.ID, "Swamp", "Basic Land — Swamp", "")
			if err := castForAltCost(g, id, "", "sacrifice", []uuid.UUID{fodder}, nil); err != nil {
				t.Fatalf("cast with a Swamp: %v", err)
			}
			if g.Battlefield.Contains(fodder) {
				t.Error("the creature was not sacrificed")
			}
		})
	}
}

func TestNemB2DarkTriumphPumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	mine := nemB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2, "G")
	theirs := nemB1Push(g, them.ID, "Their Bear", "Creature — Bear", "", 2, 2, "G")
	castCatalogSpell(t, g, "Dark Triumph", "Instant", nemB2DarkTriumphOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 4 || tough != 2 {
		t.Errorf("your 2/2 is %d/%d, want 4/2", p, tough)
	}
	if p, _ := ptOf(t, g, theirs); p != 2 {
		t.Errorf("the opponent's creature has power %d, want 2", p)
	}
}

func TestNemB2MindSwordsExilesTwoFromEachHand(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Mind Swords", "Sorcery", nemB2MindSwordsOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Hand.Size() < 2 {
			t.Fatalf("seat %s has %d cards; the test needs two", p.Name, p.Hand.Size())
		}
		picks := []uuid.UUID{p.Hand.Cards[0].InstanceID, p.Hand.Cards[1].InstanceID}
		before := p.Hand.Size()
		answerChooseCards(t, g, p.ID, picks...)
		if p.Hand.Size() != before-2 {
			t.Errorf("seat %s: hand %d -> %d, want two fewer", p.Name, before, p.Hand.Size())
		}
		for _, id := range picks {
			if !g.Exile.Contains(id) {
				t.Errorf("seat %s: a chosen card is not in exile", p.Name)
			}
		}
	}
}

func TestNemB2DownhillChargeCountsMountainsAfterTheCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := handCardAtMain(t, g, "Downhill Charge", "Instant", nemB2DownhillChargeOracle)
	bear := nemB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	paid := seedLand(g, me.ID, "Mountain A", "Basic Land — Mountain", "")
	seedLand(g, me.ID, "Mountain B", "Basic Land — Mountain", "")
	seedLand(g, me.ID, "Mountain C", "Basic Land — Mountain", "")
	if err := castForAltCost(g, id, "", "sacrifice", []uuid.UUID{paid}, cardTarget(bear)); err != nil {
		t.Fatalf("cast for a Mountain: %v", err)
	}
	if g.Battlefield.Contains(paid) {
		t.Fatal("the Mountain was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the 2/2 is %d/%d, want 4/2 for the two Mountains left", p, tough)
	}
}

// --- small spells ----------------------------------------------------------------

func TestNemB2InfiltrateMakesItUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	attacker := nemB1Push(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2, "U")
	blocker := nemB1Push(g, them.ID, "Blocker", "Creature — Bear", "", 2, 2, "G")
	castCatalogSpell(t, g, "Infiltrate", "Instant", nemB2InfiltrateOracle, cardTarget(attacker))
	passPriorityAroundTable(t, g)
	attackIntoBlocks(t, g, them.ID, attacker)
	if blockRefusal(t, g, blocker, attacker) == "" {
		t.Error("the target was blocked")
	}
}

func TestNemB2OffBalanceStopsAttackingAndBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	mine := nemB1Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2, "W")
	attacker := nemB1Push(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2, "W")
	theirs := nemB1Push(g, them.ID, "Theirs", "Creature — Bear", "", 2, 2, "G")
	castCatalogSpell(t, g, "Off Balance", "Instant", nemB2OffBalanceOracle, cardTarget(mine))
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Off Balance", "Instant", nemB2OffBalanceOracle, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	advanceToStepInTurn(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mine, them.ID); err == nil {
		t.Error("the target attacked")
	}
	attackIntoBlocks(t, g, them.ID, attacker)
	if blockRefusal(t, g, theirs, attacker) == "" {
		t.Error("the target blocked")
	}
}

func TestNemB2ViciousHungerDealsTwoAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	bear := nemB1Push(g, them.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	life := me.Life
	castCatalogSpell(t, g, "Vicious Hunger", "Sorcery", nemB2ViciousHungerOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestNemB2StrongholdDisciplineCostsLifePerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	nemB1Push(g, me.ID, "Mine A", "Creature — Bear", "", 2, 2, "G")
	nemB1Push(g, me.ID, "Mine B", "Creature — Bear", "", 2, 2, "G")
	nemB1Push(g, them.ID, "Theirs", "Creature — Bear", "", 2, 2, "G")
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	castCatalogSpell(t, g, "Stronghold Discipline", "Sorcery", nemB2DisciplineOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := 0
		switch p.ID {
		case me.ID:
			want = 2
		case them.ID:
			want = 1
		}
		if got := before[p.ID] - p.Life; got != want {
			t.Errorf("seat %s lost %d, want %d", p.Name, got, want)
		}
	}
}

// --- the recruiters --------------------------------------------------------------

func TestNemB2RecruitersFindTheirTribe(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, mana, subtype string
		maxMV                                 int
	}{
		{"Moggcatcher", nemB2MoggcatcherOracle, "Creature — Human Mercenary", "{C}{C}{C}", "Goblin", -1},
		{"Seahunter", nemB2SeahunterOracle, "Creature — Human Mercenary", "{C}{C}{C}", "Merfolk", -1},
		{"Skyshroud Poacher", nemB2PoacherOracle, "Creature — Human Rebel", "{C}{C}{C}", "Elf", -1},
		{"Defiant Falcon", nemB2DefiantFalconOracle, "Creature — Rebel Bird", "{C}{C}{C}{C}", "Rebel", 3},
		{"Rathi Fiend", nemB2RathiFiendOracle, "Creature — Phyrexian Horror Mercenary", "{C}{C}{C}", "Mercenary", 3},
		{"Rathi Intimidator", nemB2IntimidatorOracle, "Creature — Phyrexian Horror Mercenary", "{C}{C}", "Mercenary", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			seedSearchLibrary(me,
				game.Card{Name: "Small " + tc.subtype, TypeLine: "Creature — " + tc.subtype, ManaCost: "{1}"},
				game.Card{Name: "Big " + tc.subtype, TypeLine: "Creature — " + tc.subtype, ManaCost: "{5}{B}"},
				game.Card{Name: "Edge " + tc.subtype, TypeLine: "Creature — " + tc.subtype, ManaCost: "{1}{B}{B}"},
				game.Card{Name: tc.subtype + " Sorcery", TypeLine: "Kindred Sorcery — " + tc.subtype, ManaCost: "{B}"},
				game.Card{Name: "Small Soldier", TypeLine: "Creature — Human Soldier", ManaCost: "{W}"},
			)
			recruiter := nemB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			nemB2Activate(t, g, me.ID, recruiter, tc.mana, game.ActivateAbilityParams{})
			passPriorityAroundTable(t, g)
			c := searchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("the search names a quality, so it asks")
			}
			if searchOptionNamed(g, c, tc.subtype+" Sorcery") != uuid.Nil {
				t.Error("a Kindred sorcery is not a permanent card")
			}
			if searchOptionNamed(g, c, "Small Soldier") != uuid.Nil {
				t.Errorf("a non-%s was offered", tc.subtype)
			}
			big := searchOptionNamed(g, c, "Big "+tc.subtype) != uuid.Nil
			edge := searchOptionNamed(g, c, "Edge "+tc.subtype) != uuid.Nil
			if big != (tc.maxMV < 0) {
				t.Errorf("a mana value 6 card offered = %v, bound %d", big, tc.maxMV)
			}
			if edge != (tc.maxMV < 0 || tc.maxMV >= 3) {
				t.Errorf("a mana value 3 card offered = %v, bound %d", edge, tc.maxMV)
			}
			answerSearchNamed(t, g, me.ID, "Small "+tc.subtype)
			if findBattlefieldByName(g, "Small "+tc.subtype) == uuid.Nil {
				t.Error("the found card did not enter the battlefield")
			}
		})
	}
}

func TestNemB2RathiFiendMakesEachPlayerLoseThree(t *testing.T) {
	g := newCatalogGame(t)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	nemB2CastCreature(t, g, "Rathi Fiend", "Creature — Phyrexian Horror Mercenary", nemB2RathiFiendOracle, 2, 2)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if got := before[p.ID] - p.Life; got != 3 {
			t.Errorf("seat %s lost %d, want 3", p.Name, got)
		}
	}
}

func TestNemB2KeywordRecruiters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	falcon := nemB1Push(g, me.ID, "Defiant Falcon", "Creature — Rebel Bird", nemB2DefiantFalconOracle, 1, 1, "W")
	intimidator := nemB1Push(g, me.ID, "Rathi Intimidator", "Creature — Phyrexian Horror Mercenary", nemB2IntimidatorOracle, 2, 1, "B")
	if !hasEffectiveKeyword(t, g, falcon, "flying") {
		t.Error("Defiant Falcon does not have flying")
	}
	if !hasEffectiveKeyword(t, g, intimidator, "fear") {
		t.Error("Rathi Intimidator does not have fear")
	}
}

// --- creatures -------------------------------------------------------------------

func TestNemB2CarrionWallRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	wall := nemB1Push(g, me.ID, "Carrion Wall", "Creature — Wall", nemB2CarrionWallOracle, 3, 2, "B")
	if !hasEffectiveKeyword(t, g, wall, "defender") {
		t.Fatal("Carrion Wall does not have defender")
	}
	nemB2Activate(t, g, me.ID, wall, "{C}{B}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, wall)
	if !g.Battlefield.Contains(wall) {
		t.Error("the regeneration shield did not save the Wall")
	}
}

func TestNemB2SpinelessThugCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	thug := nemB1Push(g, them.ID, "Spineless Thug", "Creature — Phyrexian Zombie Mercenary", nemB2ThugOracle, 2, 2, "B")
	attacker := nemB1Push(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2, "G")
	attackIntoBlocks(t, g, them.ID, attacker)
	if blockRefusal(t, g, thug, attacker) == "" {
		t.Error("Spineless Thug blocked")
	}
}

func TestNemB2SilkenfistsUntapWhenBlocked(t *testing.T) {
	for _, tc := range []struct {
		name, oracle     string
		power, toughness int
	}{
		{"Silkenfist Fighter", nemB2SilkenfistFighterOr, 1, 3},
		{"Silkenfist Order", nemB2SilkenfistOrderOracle, 3, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me, them := seatsForCombat(g)
			fist := nemB1Push(g, me.ID, tc.name, "Creature — Kor Soldier", tc.oracle, tc.power, tc.toughness, "W")
			wall := nemB1Push(g, them.ID, "Wall", "Creature — Wall", "", 0, 6, "W")
			mrd3BlockedBy(t, g, them.ID, fist, wall)
			if !b16Tapped(t, g, fist) {
				t.Fatal("the attacker did not tap to attack")
			}
			passPriorityAroundTable(t, g)
			if b16Tapped(t, g, fist) {
				t.Error("the blocked creature was not untapped")
			}
		})
	}
}

func TestNemB2ShriekingMoggTapsEveryOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	mine := nemB1Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2, "G")
	theirs := nemB1Push(g, them.ID, "Theirs", "Creature — Bear", "", 2, 2, "G")
	mogg := nemB2CastCreature(t, g, "Shrieking Mogg", "Creature — Goblin", nemB2ShriekingMoggOracle, 1, 1)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, mine) || !b16Tapped(t, g, theirs) {
		t.Error("another creature was left untapped")
	}
	if b16Tapped(t, g, mogg) {
		t.Error("the Mogg tapped itself")
	}
	if !hasEffectiveKeyword(t, g, mogg, "haste") {
		t.Error("Shrieking Mogg does not have haste")
	}
}

func TestNemB2WanderingEyeRevealsEveryHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eye := nemB1Push(g, me.ID, "Wandering Eye", "Creature — Illusion", nemB2WanderingEyeOracle, 1, 3, "U")
	if !hasEffectiveKeyword(t, g, eye, "flying") {
		t.Error("Wandering Eye does not have flying")
	}
	for _, p := range g.Seats {
		var got []string
		g.WithWriteLock(func() { got = g.HandRevealedLocked(p.ID) })
		if !slices.Contains(got, "Wandering Eye") {
			t.Errorf("seat %s's hand is not revealed: %v", p.Name, got)
		}
	}
}

func TestNemB2SneakyHomunculusAndPowerTwo(t *testing.T) {
	t.Run("attacking", func(t *testing.T) {
		g := newCatalogGame(t)
		me, them := seatsForCombat(g)
		homunculus := nemB1Push(g, me.ID, "Sneaky Homunculus", "Creature — Homunculus Illusion", nemB2HomunculusOracle, 1, 1, "U")
		big := nemB1Push(g, them.ID, "Big", "Creature — Bear", "", 2, 2, "G")
		small := nemB1Push(g, them.ID, "Small", "Creature — Bear", "", 1, 3, "G")
		attackIntoBlocks(t, g, them.ID, homunculus)
		if blockRefusal(t, g, big, homunculus) == "" {
			t.Error("a power 2 creature blocked it")
		}
		if got := blockRefusal(t, g, small, homunculus); got != "" {
			t.Errorf("a power 1 creature was refused: %q", got)
		}
	})
	t.Run("blocking", func(t *testing.T) {
		g := newCatalogGame(t)
		me, them := seatsForCombat(g)
		homunculus := nemB1Push(g, them.ID, "Sneaky Homunculus", "Creature — Homunculus Illusion", nemB2HomunculusOracle, 1, 1, "U")
		big := nemB1Push(g, me.ID, "Big", "Creature — Bear", "", 2, 2, "G")
		small := nemB1Push(g, me.ID, "Small", "Creature — Bear", "", 1, 3, "G")
		attackIntoBlocks(t, g, them.ID, big, small)
		if blockRefusal(t, g, homunculus, big) == "" {
			t.Error("it blocked a power 2 creature")
		}
		if got := blockRefusal(t, g, homunculus, small); got != "" {
			t.Errorf("it was refused a power 1 creature: %q", got)
		}
	})
}

func TestNemB2ChieftainEnDalGivesAttackersFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	chieftain := nemB1Push(g, me.ID, "Chieftain en-Dal", "Creature — Human Knight", nemB2ChieftainOracle, 2, 2, "W")
	other := nemB1Push(g, me.ID, "Other", "Creature — Bear", "", 2, 2, "G")
	idle := nemB1Push(g, me.ID, "Idle", "Creature — Bear", "", 2, 2, "G")
	declareAttack(t, g, them.ID, chieftain, other)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{chieftain, other} {
		if !hasEffectiveKeyword(t, g, id, "first strike") {
			t.Error("an attacking creature did not gain first strike")
		}
	}
	if hasEffectiveKeyword(t, g, idle, "first strike") {
		t.Error("a creature that did not attack gained first strike")
	}
}
