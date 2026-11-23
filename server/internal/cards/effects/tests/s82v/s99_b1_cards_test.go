package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s99_b1_cards_test.go — Starter 1999, batch 1 (#5620). Each test plays
// the real catalog entry through the engine and asserts what its text
// says.

const (
	s99B1AngelOfLightOracle     = "d12cf640-ff60-4935-a2c7-28ab788de42c"
	s99B1ChampionLancerOracle   = "2681bfb5-127b-4046-85ac-72be584999cb"
	s99B1CinderStormOracle      = "c7971968-c4ee-4437-92ec-1238bce1935f"
	s99B1DakmorGhoulOracle      = "6b25b7d7-ebb1-474c-9898-e027dbf0bbdd"
	s99B1DakmorLancerOracle     = "00efb113-ba5a-4749-b42b-5693b980f26f"
	s99B1DevoutMonkOracle       = "8a0f4e8e-b541-4327-80fe-4fecc23e0df3"
	s99B1EagerCadetOracle       = "39780477-01ba-40c6-8fe6-2178efb10b92"
	s99B1GoblinChariotOracle    = "6ef0dd51-31df-4fd3-8153-49a427a7fdcf"
	s99B1GoblinCommandoOracle   = "76c02534-35e3-4950-b4b3-90c679cdf6a7"
	s99B1GoblinSettlerOracle    = "04dd449f-ff3a-4ed7-8241-244d58b371dc"
	s99B1LoyalSentryOracle      = "54b1a496-1f8e-4581-8260-bad57b25c6e3"
	s99B1PrideOfLionsOracle     = "5b5e0df8-f8f0-40a1-9814-89d73a026ad0"
	s99B1RoyalFalconOracle      = "a247a7e7-83bb-46f8-9a31-85b0d8ede919"
	s99B1RoyalTrooperOracle     = "e31510d5-b74e-4a8d-94e8-8bd098cac3c7"
	s99B1SeaEagleOracle         = "acb57162-7093-4a3c-9818-d3b61ce757c6"
	s99B1ShriekingSpecterOracle = "f6adfb5f-a21c-4e86-bd16-a9c1e355af34"
	s99B1SilverbackApeOracle    = "a72f9437-d652-4a64-99df-a307a6dd9a0d"
	s99B1SquallOracle           = "16fe0abe-04fa-4023-8ee0-043a076e78bb"
	s99B1StreamOfAcidOracle     = "1b9c9083-62c3-46fd-a2a0-6de7de88a3c3"
	s99B1ThunderDragonOracle    = "70d9579f-c8a6-4406-afe1-9291fe5ee736"
	s99B1TidingsOracle          = "72897780-094d-4a21-8b1c-419a9defd2fb"
	s99B1TrainedOrggOracle      = "da67de0f-6b0d-4100-9260-334f8c235c59"
	s99B1VeteranCavalierOracle  = "2d02cfce-a142-4b3b-a01f-4e156e75164e"
	s99B1VizzerdrixOracle       = "a6649a99-c9cd-474e-91f2-cc8e75496864"
	s99B1WillowElfOracle        = "f88db161-3d00-4f45-b2a4-cbdc4f8e5312"
)

func TestS99B1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s99B1AngelOfLightOracle:     "Angel of Light",
		s99B1ChampionLancerOracle:   "Champion Lancer",
		s99B1CinderStormOracle:      "Cinder Storm",
		s99B1DakmorGhoulOracle:      "Dakmor Ghoul",
		s99B1DakmorLancerOracle:     "Dakmor Lancer",
		s99B1DevoutMonkOracle:       "Devout Monk",
		s99B1EagerCadetOracle:       "Eager Cadet",
		s99B1GoblinChariotOracle:    "Goblin Chariot",
		s99B1GoblinCommandoOracle:   "Goblin Commando",
		s99B1GoblinSettlerOracle:    "Goblin Settler",
		s99B1LoyalSentryOracle:      "Loyal Sentry",
		s99B1PrideOfLionsOracle:     "Pride of Lions",
		s99B1RoyalFalconOracle:      "Royal Falcon",
		s99B1RoyalTrooperOracle:     "Royal Trooper",
		s99B1SeaEagleOracle:         "Sea Eagle",
		s99B1ShriekingSpecterOracle: "Shrieking Specter",
		s99B1SilverbackApeOracle:    "Silverback Ape",
		s99B1SquallOracle:           "Squall",
		s99B1StreamOfAcidOracle:     "Stream of Acid",
		s99B1ThunderDragonOracle:    "Thunder Dragon",
		s99B1TidingsOracle:          "Tidings",
		s99B1TrainedOrggOracle:      "Trained Orgg",
		s99B1VeteranCavalierOracle:  "Veteran Cavalier",
		s99B1VizzerdrixOracle:       "Vizzerdrix",
		s99B1WillowElfOracle:        "Willow Elf",
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

// The five vanilla bodies print no rules text and register none.
func TestS99B1VanillaCreaturesHaveNoAbilities(t *testing.T) {
	for _, oracle := range []string{
		s99B1EagerCadetOracle, s99B1SilverbackApeOracle, s99B1TrainedOrggOracle,
		s99B1VizzerdrixOracle, s99B1WillowElfOracle,
	} {
		spec, _ := effects.Lookup(oracle)
		if len(spec.PrintedKeywords) != 0 || len(spec.Triggered) != 0 || len(spec.Activated) != 0 ||
			len(spec.Static) != 0 || len(spec.Replacements) != 0 || spec.OnResolve != nil {
			t.Errorf("%s registers an ability it does not print", spec.Name)
		}
	}
}

// --- keyword bodies ------------------------------------------------------------

func TestS99B1KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Angel of Light", "Creature — Angel", s99B1AngelOfLightOracle, []string{"flying", "vigilance"}},
		{"Royal Falcon", "Creature — Bird", s99B1RoyalFalconOracle, []string{"flying"}},
		{"Sea Eagle", "Creature — Bird", s99B1SeaEagleOracle, []string{"flying"}},
		{"Shrieking Specter", "Creature — Specter", s99B1ShriekingSpecterOracle, []string{"flying"}},
		{"Thunder Dragon", "Creature — Dragon", s99B1ThunderDragonOracle, []string{"flying"}},
		{"Goblin Chariot", "Creature — Goblin Warrior", s99B1GoblinChariotOracle, []string{"haste"}},
		{"Veteran Cavalier", "Creature — Human Knight", s99B1VeteranCavalierOracle, []string{"vigilance"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Flying: a ground creature can't block Sea Eagle.
func TestS99B1SeaEagleIsNotBlockedByAGroundCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eagle := effectstest.B12Push(g, me.ID, "Sea Eagle", "Creature — Bird", s99B1SeaEagleOracle, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.PorBlock(t, g, eagle, opp.ID, bear); err == nil {
		t.Error("a ground creature blocked Sea Eagle")
	}
}

// Haste: a Goblin Chariot cast this turn attacks.
func TestS99B1GoblinChariotAttacksTheTurnItArrives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	chariot := effectstest.CastHoldCreature(t, g, "Goblin Chariot", "Creature — Goblin Warrior", s99B1GoblinChariotOracle, 2, 2)
	cadet := effectstest.CastHoldCreature(t, g, "Eager Cadet", "Creature — Human Soldier", s99B1EagerCadetOracle, 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(cadet, opp.ID); err == nil {
		t.Error("a summoning-sick creature without haste attacked")
	}
	if err := g.DeclareAttacker(chariot, opp.ID); err != nil {
		t.Errorf("Goblin Chariot couldn't attack the turn it arrived: %v", err)
	}
}

func TestS99B1VeteranCavalierAttacksWithoutTapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cavalier := effectstest.B12Push(g, me.ID, "Veteran Cavalier", "Creature — Human Knight", s99B1VeteranCavalierOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, cavalier)
	if effectstest.AlliTapped(t, g, cavalier) {
		t.Error("Veteran Cavalier tapped to attack")
	}
}

// --- spells --------------------------------------------------------------------

func TestS99B1CinderStormDealsSevenToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Cinder Storm", "Sorcery", s99B1CinderStormOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-7 {
		t.Errorf("life %d -> %d, want -7", life, opp.Life)
	}

	g = effectstest.NewCatalogGame(t)
	big := effectstest.B12Creature(g, g.Seats[1].ID, "Big", "Creature — Giant", 7, 7)
	effectstest.CastCatalogSpell(t, g, "Cinder Storm", "Sorcery", s99B1CinderStormOracle, effectstest.CardTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a 7-toughness creature survived Cinder Storm")
	}
}

func TestS99B1TidingsDrawsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Tidings", "Sorcery", s99B1TidingsOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+4 {
		t.Errorf("hand %d -> %d, want +4", hand, me.Hand.Size())
	}
}

func TestS99B1SquallHitsOnlyCreaturesWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirFlyer := effectstest.PorFlyer(g, opp.ID, "Their Bird")
	myFlyer := effectstest.PorFlyer(g, me.ID, "My Bird")
	ground := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Squall", "Sorcery", s99B1SquallOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirFlyer) || g.Battlefield.Contains(myFlyer) {
		t.Error("a 1/1 flyer survived Squall")
	}
	if !g.Battlefield.Contains(ground) || effectstest.Pr6Marked(g, ground) != 0 {
		t.Error("Squall dealt damage to a creature without flying")
	}
}

func TestS99B1StreamOfAcidDestroysALandOrANonblackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	black := effectstest.B12Creature(g, g.Seats[1].ID, "Zombie", "Creature — Zombie", 2, 2)
	effectstest.AlliSetColors(g, black, "B")
	if err := effectstest.CastCatalogSpellErr(t, g, "Stream of Acid", "Sorcery", s99B1StreamOfAcidOracle, effectstest.CardTarget(black)); err == nil {
		t.Error("Stream of Acid targeted a black creature")
	}

	g = effectstest.NewCatalogGame(t)
	green := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AlliSetColors(g, green, "G")
	effectstest.CastCatalogSpell(t, g, "Stream of Acid", "Sorcery", s99B1StreamOfAcidOracle, effectstest.CardTarget(green))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the green creature survived Stream of Acid")
	}

	g = effectstest.NewCatalogGame(t)
	land := effectstest.B12Permanent(g, g.Seats[1].ID, "Swamp", "Basic Land — Swamp")
	effectstest.CastCatalogSpell(t, g, "Stream of Acid", "Sorcery", s99B1StreamOfAcidOracle, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived Stream of Acid")
	}
}

// --- enters triggers -------------------------------------------------------------

func TestS99B1DakmorGhoulDrainsTheTargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	mine, theirs, bystander := me.Life, opp.Life, other.Life
	effectstest.CastHoldCreature(t, g, "Dakmor Ghoul", "Creature — Zombie", s99B1DakmorGhoulOracle, 2, 2)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine+2 || opp.Life != theirs-2 || other.Life != bystander {
		t.Errorf("life: me %d -> %d (want +2), target %d -> %d (want -2), other %d -> %d (want 0)",
			mine, me.Life, theirs, opp.Life, bystander, other.Life)
	}
}

func TestS99B1DakmorLancerDestroysANonblackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	black := effectstest.UsgB3Colored(g, opp.ID, "Black Beast", "B")
	green := effectstest.UsgB3Colored(g, opp.ID, "Green Beast", "G")
	effectstest.CastHoldCreature(t, g, "Dakmor Lancer", "Creature — Human Knight", s99B1DakmorLancerOracle, 3, 3)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: black}); err == nil {
		t.Fatal("the trigger targeted a black creature")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, green)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the nonblack creature was not destroyed")
	}
	if !g.Battlefield.Contains(black) {
		t.Error("the black creature was destroyed")
	}
}

func TestS99B1DevoutMonkGainsOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Devout Monk", "Creature — Human Monk Cleric", s99B1DevoutMonkOracle, 1, 1)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
}

func TestS99B1GoblinCommandoDealsTwoToTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Goblin Commando", "Creature — Goblin", s99B1GoblinCommandoOracle, 2, 2)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/2 survived Goblin Commando's 2 damage")
	}
}

func TestS99B1GoblinSettlerDestroysTargetLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Goblin Settler", "Creature — Goblin", s99B1GoblinSettlerOracle, 1, 1)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("the trigger targeted a creature")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
}

// Every creature without flying is dealt 3, the controller's own
// included; the flyers and the Dragon are not.
func TestS99B1ThunderDragonHitsEachCreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	tough := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	flyer := effectstest.PorFlyer(g, opp.ID, "Bird")
	dragon := effectstest.CastHoldCreature(t, g, "Thunder Dragon", "Creature — Dragon", s99B1ThunderDragonOracle, 5, 5)
	if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(mine) {
		t.Error("a 3-toughness creature without flying survived")
	}
	if got := effectstest.Pr6Marked(g, tough); got != 3 {
		t.Errorf("the 0/4 has %d damage, want 3", got)
	}
	if got := effectstest.Pr6Marked(g, flyer); got != 0 {
		t.Errorf("the flyer has %d damage, want 0", got)
	}
	if got := effectstest.Pr6Marked(g, dragon); got != 0 {
		t.Errorf("Thunder Dragon has %d damage, want 0", got)
	}
}

// --- Champion Lancer -------------------------------------------------------------

// A creature's ability deals it nothing; a spell's damage is not
// prevented.
func TestS99B1ChampionLancerPreventsDamageFromCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lancer := effectstest.B12Push(g, opp.ID, "Champion Lancer", "Creature — Human Knight", s99B1ChampionLancerOracle, 3, 3)
	effectstest.CastHoldCreature(t, g, "Goblin Commando", "Creature — Goblin", s99B1GoblinCommandoOracle, 2, 2)
	effectstest.PickTriggerTarget(t, g, me.ID, lancer)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, lancer); got != 0 {
		t.Fatalf("Champion Lancer was dealt %d damage by a creature, want 0", got)
	}

	effectstest.CastCatalogSpell(t, g, "Cinder Storm", "Sorcery", s99B1CinderStormOracle, effectstest.CardTarget(lancer))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lancer) {
		t.Error("Champion Lancer survived a spell's 7 damage")
	}
}

// Combat damage from the creature it blocks is prevented; the Lancer's
// own damage is dealt as usual.
func TestS99B1ChampionLancerTakesNoCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Brute", TypeLine: "Creature — Giant", Power: 5, Toughness: 5})
	lancer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{
		Name: "Champion Lancer", TypeLine: "Creature — Human Knight", OracleID: s99B1ChampionLancerOracle, Power: 3, Toughness: 3,
	})
	effectstest.Pr7bAttack(t, g, opp.ID, atk)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{lancer: atk})
	effectstest.Pr7bDamageStep(t, g)
	if got := effectstest.Pr6Marked(g, lancer); got != 0 {
		t.Errorf("Champion Lancer was dealt %d combat damage, want 0", got)
	}
	if got := effectstest.Pr6Marked(g, atk); got != 3 {
		t.Errorf("the attacker was dealt %d, want 3", got)
	}
}

// --- Loyal Sentry ----------------------------------------------------------------

func TestS99B1LoyalSentryDestroysWhatItBlocksAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Brute", "Creature — Giant", 5, 5)
	sentry := effectstest.B12Push(g, opp.ID, "Loyal Sentry", "Creature — Human Soldier", s99B1LoyalSentryOracle, 1, 1)
	if err := effectstest.PorBlock(t, g, atk, opp.ID, sentry); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	if n := effectstest.TriggersOnStackFrom(g, sentry); n != 1 {
		t.Fatalf("%d triggers from Loyal Sentry, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(atk) {
		t.Error("the creature Loyal Sentry blocked was not destroyed")
	}
	if g.Battlefield.Contains(sentry) {
		t.Error("Loyal Sentry was not destroyed")
	}
}

// The attacker left and came back before the trigger resolved: it is a
// new object (CR 400.7) and is spared, and the Sentry is still destroyed.
func TestS99B1LoyalSentrySparesANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Brute", "Creature — Giant", 5, 5)
	sentry := effectstest.B12Push(g, opp.ID, "Loyal Sentry", "Creature — Human Soldier", s99B1LoyalSentryOracle, 1, 1)
	if err := effectstest.PorBlock(t, g, atk, opp.ID, sentry); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.BounceAndReturnForTest(t, g, atk, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(atk) {
		t.Error("the returned creature was destroyed")
	}
	if g.Battlefield.Contains(sentry) {
		t.Error("Loyal Sentry was not destroyed")
	}
}

// --- Royal Trooper ---------------------------------------------------------------

func TestS99B1RoyalTrooperGetsTwoTwoWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	trooper := effectstest.B12Push(g, opp.ID, "Royal Trooper", "Creature — Human Soldier", s99B1RoyalTrooperOracle, 2, 2)
	if err := effectstest.PorBlock(t, g, atk, opp.ID, trooper); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, trooper); p != 4 || tt != 4 {
		t.Errorf("Royal Trooper is %d/%d, want 4/4", p, tt)
	}
}

// --- Shrieking Specter -----------------------------------------------------------

func TestS99B1ShriekingSpecterMakesTheDefendingPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.UsgB3Hand(opp, 2)
	effectstest.UsgB3Hand(other, 2)
	specter := effectstest.B12Push(g, me.ID, "Shrieking Specter", "Creature — Specter", s99B1ShriekingSpecterOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, specter)
	if n := effectstest.TriggersOnStackFrom(g, specter); n != 1 {
		t.Fatalf("%d triggers from Shrieking Specter, want 1", n)
	}
	yard := opp.Graveyard.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, opp.ID)
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 1 || opp.Graveyard.Size() != yard+1 {
		t.Errorf("defender hand %d, graveyard %d -> %d; want 1 card left and one discarded", opp.Hand.Size(), yard, opp.Graveyard.Size())
	}
	if other.Hand.Size() != 2 {
		t.Errorf("a player who was not attacked has %d cards, want 2", other.Hand.Size())
	}
}

// --- Pride of Lions --------------------------------------------------------------

func TestS99B1BlockedPrideOfLionsMayAssignAsThoughUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	pride := effectstest.ApaPush(g, g.Seats[0].ID, g.Seats[0].ID, game.Card{
		Name: "Pride of Lions", TypeLine: "Creature — Cat", OracleID: s99B1PrideOfLionsOracle, Power: 4, Toughness: 4,
	})
	wall := effectstest.MscWlBlocker(g, "Wall", 0, 5)
	effectstest.MscWlBlockedCombat(t, g, []uuid.UUID{pride}, map[uuid.UUID]uuid.UUID{wall: pride})

	c := effectstest.MscWlDamagePrompt(t, g)
	if f := c.DamageAssignment; !f.AsThoughUnblocked || len(f.BlockerIDs) != 1 || f.AttackerPower != 4 {
		t.Fatalf("the prompt is %+v, want one blocker, power 4, as-though-unblocked", f)
	}
	life := opp.Life
	if err := effectstest.MscWlAnswer(t, g, c, 4, game.DamageAssignmentEntry{BlockerID: wall, Amount: 0}); err != nil {
		t.Fatalf("all of it to the player: %v", err)
	}
	if opp.Life != life-4 {
		t.Errorf("the defending player went %d -> %d, want -4", life, opp.Life)
	}
	if got := effectstest.Pr6Marked(g, wall); got != 0 {
		t.Errorf("the blocker was dealt %d damage, want none", got)
	}
}
