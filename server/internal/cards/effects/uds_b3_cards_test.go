package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// uds_b3_cards_test.go — Urza's Destiny, batch 3 (#5621). Each test plays
// the real catalog entry through the engine and asserts what its text
// says.

const (
	udsB3AttritionOracle         = "a85ee0cf-39ce-49f2-874a-f6f54ca1775f"
	udsB3BloodshotCyclopsOracle  = "84d47913-53ac-4d4d-bac3-bc0306d43e12"
	udsB3BraidwoodSextantOracle  = "b44816b4-44ef-446d-a254-b4b7bc015a42"
	udsB3MentalDisciplineOracle  = "b22080d6-a9ed-4bdd-a604-058e0e3e9463"
	udsB3FodderCannonOracle      = "aaf171bd-a4bb-4ce4-836a-da193c94f42e"
	udsB3BlizzardElementalOracle = "dd489f30-d1cf-4236-8390-9c21cd8659e0"
	udsB3MantisEngineOracle      = "57bea272-5b38-444e-acc2-6104af2ece22"
	udsB3TemporalAdeptOracle     = "fd9b5462-6578-4bf6-8026-23eaf5af3eee"
	udsB3MasterHealerOracle      = "7695b61a-8d44-44b1-9783-ea3114111c3e"
	udsB3SerraAdvocateOracle     = "4c67292a-7fcc-4646-8a06-99a7b4b3e005"
	udsB3OppositionOracle        = "bf0b252d-1295-44e4-bac3-113b5732a2a9"
	udsB3FieldSurgeonOracle      = "730809d7-cb8c-495c-8972-7e85b967358e"
	udsB3RofellosOracle          = "3015882b-4897-4d2c-8e33-6731c85a0d03"
	udsB3RescueOracle            = "f0ae687a-7222-4521-a514-ba10373fed7e"
	udsB3SoulFeastOracle         = "8186fd80-015f-470c-9e1c-cbf45764a057"
	udsB3SolidarityOracle        = "7b094a2b-2b69-4a9b-9829-c1aabc53f75e"
	udsB3FlameJetOracle          = "8a9d6a30-e2c2-4da5-ad3b-d64fb2b5543c"
	udsB3RecklessAbandonOracle   = "f24d5187-0f80-4528-bf44-03ae146f3551"
	udsB3MultanisDecreeOracle    = "b233473d-d22b-45e4-9fe7-55d1170a788b"
	udsB3GoblinMasonsOracle      = "cfec9d17-2d11-4b89-8b30-84bd25a38466"
	udsB3ThievingMagpieOracle    = "2f074636-6a36-46a3-b872-add79e19aa8b"
	udsB3TelepathicSpiesOracle   = "8162730e-2573-4bc0-827f-5b7b3dbf3bff"
	udsB3CapashenStandardOracle  = "75510429-41bb-409e-b6fe-04a8bb174c6b"
	udsB3IlluminatedWingsOracle  = "7a22389a-34e7-4726-a551-f6fbc225cefe"
)

func TestUdsB3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		udsB3AttritionOracle:         "Attrition",
		udsB3BloodshotCyclopsOracle:  "Bloodshot Cyclops",
		udsB3BraidwoodSextantOracle:  "Braidwood Sextant",
		udsB3MentalDisciplineOracle:  "Mental Discipline",
		udsB3FodderCannonOracle:      "Fodder Cannon",
		udsB3BlizzardElementalOracle: "Blizzard Elemental",
		udsB3MantisEngineOracle:      "Mantis Engine",
		udsB3TemporalAdeptOracle:     "Temporal Adept",
		udsB3MasterHealerOracle:      "Master Healer",
		udsB3SerraAdvocateOracle:     "Serra Advocate",
		udsB3OppositionOracle:        "Opposition",
		udsB3FieldSurgeonOracle:      "Field Surgeon",
		udsB3RofellosOracle:          "Rofellos, Llanowar Emissary",
		udsB3RescueOracle:            "Rescue",
		udsB3SoulFeastOracle:         "Soul Feast",
		udsB3SolidarityOracle:        "Solidarity",
		udsB3FlameJetOracle:          "Flame Jet",
		udsB3RecklessAbandonOracle:   "Reckless Abandon",
		udsB3MultanisDecreeOracle:    "Multani's Decree",
		udsB3GoblinMasonsOracle:      "Goblin Masons",
		udsB3ThievingMagpieOracle:    "Thieving Magpie",
		udsB3TelepathicSpiesOracle:   "Telepathic Spies",
		udsB3CapashenStandardOracle:  "Capashen Standard",
		udsB3IlluminatedWingsOracle:  "Illuminated Wings",
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

// udsB3Activate floats `mana` and activates row `row` of `id` under the
// strict mana gate.
func udsB3Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, row int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	floatForTest(g, who, mana)
	params.Strict = true
	return g.ActivateCatalogAbility(who.ID, id, row, params)
}

func TestUdsB3KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Blizzard Elemental", "Creature — Elemental", udsB3BlizzardElementalOracle, []string{"flying"}},
		{"Serra Advocate", "Creature — Angel", udsB3SerraAdvocateOracle, []string{"flying"}},
		{"Thieving Magpie", "Creature — Bird", udsB3ThievingMagpieOracle, []string{"flying"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !hasKeywordOnBattlefield(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

func TestUdsB3AttritionSacrificesToDestroyANonblackCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attrition := pushCatalogPermanent(g, me.ID, "Attrition", "Enchantment", udsB3AttritionOracle, true)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	other := b12Creature(g, me.ID, "Other Fodder", "Creature — Rat", 1, 1)
	black := tmpB1Push(g, opp.ID, "Black Knight", "Creature — Human Knight", "", 2, 2, "B")
	green := tmpB1Push(g, opp.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	if err := udsB3Activate(t, g, me, attrition, 0, "B", game.ActivateAbilityParams{
		Targets: cardRefs(black), SacrificeIDs: []uuid.UUID{fodder},
	}); err == nil {
		t.Fatal("a black creature was accepted as the target")
	}
	if err := udsB3Activate(t, g, me, attrition, 0, "", game.ActivateAbilityParams{
		Targets: cardRefs(green), SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the nonblack creature was not destroyed")
	}
	// No {T}: a second activation the same turn.
	green2 := tmpB1Push(g, opp.ID, "Green Bear Two", "Creature — Bear", "", 2, 2, "G")
	if err := udsB3Activate(t, g, me, attrition, 0, "B", game.ActivateAbilityParams{
		Targets: cardRefs(green2), SacrificeIDs: []uuid.UUID{other},
	}); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green2) {
		t.Error("the second activation did not destroy its target")
	}
}

func TestUdsB3BloodshotCyclopsFlingsTheSacrificedPower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sick := pushCatalogPermanent(g, me.ID, "Bloodshot Cyclops", "Creature — Cyclops Giant", udsB3BloodshotCyclopsOracle, true)
	ogre := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := udsB3Activate(t, g, me, sick, 0, "", game.ActivateAbilityParams{
		Targets: playerRefs(opp.ID), SacrificeIDs: []uuid.UUID{ogre},
	}); err == nil {
		t.Fatal("a summoning-sick Cyclops paid {T} (CR 302.6)")
	}
	cyclops := b12Push(g, me.ID, "Bloodshot Cyclops", "Creature — Cyclops Giant", udsB3BloodshotCyclopsOracle, 4, 4)
	life := opp.Life
	if err := udsB3Activate(t, g, me, cyclops, 0, "", game.ActivateAbilityParams{
		Targets: playerRefs(opp.ID), SacrificeIDs: []uuid.UUID{ogre},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3 (the Ogre's power)", life, opp.Life)
	}
}

func TestUdsB3FodderCannonShootsATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cannon := pushCatalogPermanent(g, me.ID, "Fodder Cannon", "Artifact", udsB3FodderCannonOracle, true)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Giant", 4, 4)
	if err := udsB3Activate(t, g, me, cannon, 0, "CCC", game.ActivateAbilityParams{
		Targets: cardRefs(victim), SacrificeIDs: []uuid.UUID{fodder},
	}); err == nil {
		t.Fatal("activated for {3}; it costs {4}")
	}
	if err := udsB3Activate(t, g, me, cannon, 0, "C", game.ActivateAbilityParams{
		Targets: cardRefs(victim), SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || g.Battlefield.Contains(victim) {
		t.Error("the fodder is sacrificed and the 4/4 takes 4")
	}
	if !g.Battlefield.Contains(cannon) || !b13Tapped(t, g, cannon) {
		t.Error("the Cannon stays, tapped")
	}
}

func TestUdsB3BraidwoodSextantFetchesABasicLandToHand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sextant := pushCatalogPermanent(g, me.ID, "Braidwood Sextant", "Artifact", udsB3BraidwoodSextantOracle, true)
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	urborg := stapleLibraryCard(me, "Urborg, Tomb of Yawgmoth", "Legendary Land")
	if err := udsB3Activate(t, g, me, sextant, 0, "CC", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(sextant) {
		t.Error("the Sextant is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := func() error {
		beginSearchIfAsked(t, g, c)
		return g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{urborg})
	}(); err == nil {
		t.Error("a nonbasic land was accepted")
	}
	answerSearchByID(t, g, me.ID, forest)
	if !me.Hand.Contains(forest) {
		t.Error("the basic land did not reach the hand")
	}
}

func TestUdsB3MentalDisciplineDiscardsToDraw(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	md := pushCatalogPermanent(g, me.ID, "Mental Discipline", "Enchantment", udsB3MentalDisciplineOracle, true)
	junk := handCardFull(me, "Junk", "Sorcery", "{1}", "", nil)
	if err := udsB3Activate(t, g, me, md, 0, "U", game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{junk}}); err == nil {
		t.Fatal("activated for {U}; it costs {1}{U}")
	}
	if err := udsB3Activate(t, g, me, md, 0, "C", game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{junk}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(junk) {
		t.Error("the card was not discarded as the cost")
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestUdsB3BlizzardElementalUntapsItself(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	el := b12Push(g, me.ID, "Blizzard Elemental", "Creature — Elemental", udsB3BlizzardElementalOracle, 5, 5)
	b16Tap(g, el)
	if err := udsB3Activate(t, g, me, el, 0, "CC", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated for {2}; it costs {3}{U}")
	}
	if err := udsB3Activate(t, g, me, el, 0, "CU", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b13Tapped(t, g, el) {
		t.Error("the Elemental is still tapped")
	}
}

func TestUdsB3MantisEngineGainsFlyingAndFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	engine := pushCatalogPermanent(g, me.ID, "Mantis Engine", "Artifact Creature — Insect", udsB3MantisEngineOracle, true)
	for row, kw := range []string{"flying", "first strike"} {
		if hasKeywordOnBattlefield(t, g, engine, kw) {
			t.Fatalf("the Engine has %s before the activation", kw)
		}
		if err := udsB3Activate(t, g, me, engine, row, "CC", game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate row %d: %v", row, err)
		}
		passPriorityAroundTable(t, g)
		if !hasKeywordOnBattlefield(t, g, engine, kw) {
			t.Errorf("the Engine did not gain %s", kw)
		}
	}
}

func TestUdsB3TemporalAdeptBouncesAnyPermanent(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	adept := b12Push(g, me.ID, "Temporal Adept", "Creature — Human Wizard", udsB3TemporalAdeptOracle, 1, 1)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := udsB3Activate(t, g, me, adept, 0, "CUU", game.ActivateAbilityParams{Targets: cardRefs(rock)}); err == nil {
		t.Fatal("activated without {U}{U}{U}")
	}
	if err := udsB3Activate(t, g, me, adept, 0, "U", game.ActivateAbilityParams{Targets: cardRefs(rock)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
}

func TestUdsB3RescueReturnsOnlyYourOwnPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Rescue", "Instant", udsB3RescueOracle, cardTarget(theirs)); err == nil {
		t.Fatal("Rescue targeted a permanent you don't control")
	}
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Rescue", "Instant", udsB3RescueOracle, cardTarget(mine))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("your creature did not return to your hand")
	}
}

func TestUdsB3MasterHealerPreventsFour(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	healer := b12Push(g, me.ID, "Master Healer", "Creature — Human Cleric", udsB3MasterHealerOracle, 1, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 6)
	if err := udsB3Activate(t, g, me, healer, 0, "", game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 5) })
	if got := damageMarkedOn(g, bear); got != 1 {
		t.Errorf("5 damage into a 4-shield: marked %d, want 1", got)
	}
}

func TestUdsB3FieldSurgeonTapsACreatureToPreventOne(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	// Summoning sick: tapping a creature is not the {T} symbol (CR 302.6).
	surgeon := pushCatalogPermanent(g, me.ID, "Field Surgeon", "Creature — Human Cleric", udsB3FieldSurgeonOracle, true)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 4)
	if err := udsB3Activate(t, g, me, surgeon, 0, "", game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{surgeon}, Targets: cardRefs(bear),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !b13Tapped(t, g, surgeon) {
		t.Error("the Surgeon did not tap itself to pay")
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := damageMarkedOn(g, bear); got != 2 {
		t.Errorf("3 damage into a 1-shield: marked %d, want 2", got)
	}
}

func TestUdsB3OppositionTapsACreatureToTapATarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opposition := pushCatalogPermanent(g, me.ID, "Opposition", "Enchantment", udsB3OppositionOracle, true)
	mine := pushCatalogPermanent(g, me.ID, "Fresh Bear", "Creature — Bear", "", true)
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	ench := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	if err := udsB3Activate(t, g, me, opposition, 0, "", game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{mine}, Targets: cardRefs(ench),
	}); err == nil {
		t.Fatal("an enchantment was accepted as the target")
	}
	if err := udsB3Activate(t, g, me, opposition, 0, "", game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{mine}, Targets: cardRefs(land),
	}); err != nil {
		t.Fatalf("activate with a summoning-sick creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, land) || !b13Tapped(t, g, mine) {
		t.Error("the land and the paying creature are both tapped")
	}
	if err := udsB3Activate(t, g, me, opposition, 0, "", game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{mine}, Targets: cardRefs(land),
	}); err == nil {
		t.Error("a tapped creature paid the cost again")
	}
}

func TestUdsB3SerraAdvocatePumpsOnlyACombatant(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advocate := b12Push(g, me.ID, "Serra Advocate", "Creature — Angel", udsB3SerraAdvocateOracle, 2, 2)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	raider := b12Creature(g, me.ID, "Raider", "Creature — Bear", 2, 2)
	if err := udsB3Activate(t, g, me, advocate, 0, "", game.ActivateAbilityParams{Targets: cardRefs(idle)}); err == nil {
		t.Fatal("a creature out of combat was accepted")
	}
	arnSetCombat(g, raider, opp.ID, uuid.Nil)
	if err := udsB3Activate(t, g, me, advocate, 0, "", game.ActivateAbilityParams{Targets: cardRefs(raider)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, raider); p != 4 || tough != 4 {
		t.Errorf("the attacker is %d/%d, want 4/4", p, tough)
	}
}

func TestUdsB3RofellosAddsGreenPerForest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rofellos := pushCatalogPermanent(g, me.ID, "Rofellos, Llanowar Emissary", "Legendary Creature — Elf Druid", udsB3RofellosOracle, false)
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	if err := g.ActivateManaAbility(me.ID, rofellos, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	got := poolColors(me)
	if len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("pool %v, want G G (two Forests you control)", got)
	}
}

func TestUdsB3SoulFeastDrainsFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	myLife, oppLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Soul Feast", "Sorcery", udsB3SoulFeastOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-4 || me.Life != myLife+4 {
		t.Errorf("opponent %d -> %d, me %d -> %d: want -4 and +4", oppLife, opp.Life, myLife, me.Life)
	}
}

func TestUdsB3SolidarityGivesYourCreaturesPlusZeroFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Solidarity", "Instant", udsB3SolidarityOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 2 || tough != 7 {
		t.Errorf("your creature is %d/%d, want 2/7", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestUdsB3FlameJetHitsAPlayerAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Flame Jet", "Sorcery", udsB3FlameJetOracle, cardTarget(bear)); err == nil {
		t.Fatal("Flame Jet targeted a creature")
	}
	g = newCatalogGame(t)
	opp = g.Seats[1]
	life := opp.Life
	castCatalogSpell(t, g, "Flame Jet", "Sorcery", udsB3FlameJetOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3", life, opp.Life)
	}
	g = newCatalogGame(t)
	seedLibrary(g.Seats[g.Turn.ActiveSeat], "A", "B")
	ons1Cycle(t, g, "Flame Jet", "Sorcery", udsB3FlameJetOracle, "{C}{C}")
}

func TestUdsB3RecklessAbandonSacrificesACreatureToDealFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	toMainForCost(t, g)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	life := opp.Life
	cast := func(sac uuid.UUID) error {
		id := handCardFull(me, "Reckless Abandon", "Sorcery", "{R}", udsB3RecklessAbandonOracle, []string{"R"})
		floatForTest(g, me, "R")
		return g.CastSpell(me.ID, id, game.CastSpellParams{
			Targets: []game.TargetRef{playerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{sac}, Strict: true,
		})
	}
	if err := cast(rock); err == nil {
		t.Fatal("an artifact paid the creature sacrifice")
	}
	me.ManaPool.EmptyPool()
	if err := cast(fodder); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want -4", life, opp.Life)
	}
}

func TestUdsB3MultanisDecreeDestroysEnchantmentsForTwoLifeEach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := b12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	theirs := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	life := me.Life
	castCatalogSpell(t, g, "Multani's Decree", "Sorcery", udsB3MultanisDecreeOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("an enchantment survived")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("an artifact was destroyed")
	}
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 (two enchantments)", life, me.Life)
	}
}

func TestUdsB3GoblinMasonsDestroyOnlyAWall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	masons := b12Push(g, me.ID, "Goblin Masons", "Creature — Goblin", udsB3GoblinMasonsOracle, 2, 1)
	wall := b12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	killCreature(t, g, me.ID, masons)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the dies trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Error("a non-Wall was accepted as the target")
	}
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: wall}, true)
	if g.Battlefield.Contains(wall) {
		t.Error("the Wall was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear was destroyed")
	}
}

func TestUdsB3ThievingMagpieDrawsOnDamageToAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B")
	magpie := b12Push(g, me.ID, "Thieving Magpie", "Creature — Bird", udsB3ThievingMagpieOracle, 1, 3)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(magpie, opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(magpie, bear, 1) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("damage to a creature drew a card: %d drawn in all", got)
	}
}

func TestUdsB3TelepathicSpiesLookAtAnOpponentsHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	handCardFull(opp, "Secret", "Instant", "{U}", "", nil)
	castCatalogSpell(t, g, "Telepathic Spies", "Creature — Human Wizard", udsB3TelepathicSpiesOracle, nil)
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	for i := range opp.Hand.Cards {
		if !opp.Hand.Cards[i].IsKnownTo(me.ID) {
			t.Error("the Spies' controller did not see the opponent's hand")
		}
		if opp.Hand.Cards[i].IsKnownTo(other.ID) {
			t.Error("\"look at\" showed the hand to another player")
		}
	}
}

func TestUdsB3AurasSacrificeThemselvesToDraw(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		keyword      string
	}{
		{"Capashen Standard", udsB3CapashenStandardOracle, 3, 3, ""},
		{"Illuminated Wings", udsB3IlluminatedWingsOracle, 2, 2, "flying"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			seedLibrary(me, "A", "B")
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			aura := auraCast(t, g, tc.name, tc.oracle, bear)
			if p, tough := ptOf(t, g, bear); p != tc.power || tough != tc.tough {
				t.Errorf("the enchanted Bear is %d/%d, want %d/%d", p, tough, tc.power, tc.tough)
			}
			if tc.keyword != "" && !hasKeywordOnBattlefield(t, g, bear, tc.keyword) {
				t.Errorf("the enchanted Bear has no %s", tc.keyword)
			}
			hand := me.Hand.Size()
			if err := udsB3Activate(t, g, me, aura, 0, "CC", game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if g.Battlefield.Contains(aura) {
				t.Error("the Aura was not sacrificed as the cost")
			}
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != 1 {
				t.Errorf("drew %d, want 1", got)
			}
			if p, tough := ptOf(t, g, bear); p != 2 || tough != 2 {
				t.Errorf("the Bear is %d/%d after the Aura left, want 2/2", p, tough)
			}
		})
	}
}
