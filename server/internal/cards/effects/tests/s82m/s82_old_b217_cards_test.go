package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b217_cards_test.go — S82 OLD batch 217 (#5531, Rivals of Ixalan):
// Charging Tuskodon, Dinosaur Hunter, Mutiny, Form of the Dinosaur, Reckless
// Rage, Sea Legs, Squire's Devotion, Journey to Eternity // Atzal, Cave of
// Eternity, See Red and Tilonalli's Crown. Each test plays the real catalog
// entry through the engine.

const (
	oldB217Tuskodon = "1198d47a-4235-4678-bd2e-20f2a97b5924"
	oldB217Hunter   = "5c58f2eb-2617-4d10-a393-6e4a9eddd1dd"
	oldB217Mutiny   = "647b4cb4-569d-4938-b56c-31a70e68ee41"
	oldB217Form     = "68af06fd-a715-47c6-b439-c9fcb1b01a1d"
	oldB217Rage     = "d7e2b2ba-cb5e-41f1-ad33-baf56c0b72ed"
	oldB217SeaLegs  = "2e66dc2f-44f3-44f1-8ed7-7f991b3718be"
	oldB217Squire   = "4770ad4a-a346-4108-8f61-a077155d4bd1"
	oldB217Journey  = "7d6ccd0b-df16-40b2-930b-bcde0b6ef73f"
	oldB217SeeRed   = "dd33af3e-ab8c-486c-baae-a37cc857f180"
	oldB217Crown    = "fbc2f469-09a5-4641-b990-b483d5d15c0d"
)

func TestOldB217EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB217Tuskodon:       "Charging Tuskodon",
		oldB217Hunter:         "Dinosaur Hunter",
		oldB217Mutiny:         "Mutiny",
		oldB217Form:           "Form of the Dinosaur",
		oldB217Rage:           "Reckless Rage",
		oldB217SeaLegs:        "Sea Legs",
		oldB217Squire:         "Squire's Devotion",
		oldB217Journey:        "Journey to Eternity",
		oldB217Journey + "#1": "Atzal, Cave of Eternity",
		oldB217SeeRed:         "See Red",
		oldB217Crown:          "Tilonalli's Crown",
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

func oldB217Seats(g *game.Game) (me, opp, third *game.Player) {
	n := len(g.Seats)
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n]
}

// --- Charging Tuskodon -------------------------------------------------------

func TestOldB217TuskodonDoublesCombatDamageToAPlayerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB217Seats(g)
	tusk := effectstest.B12Push(g, me.ID, "Charging Tuskodon", "Creature — Dinosaur", oldB217Tuskodon, 4, 4)
	before := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, tusk)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if got := before - opp.Life; got != 8 {
		t.Errorf("the defending player lost %d, want 8 (double 4)", got)
	}
}

func TestOldB217TuskodonDoesNotDoubleNoncombatDamageOrDamageToCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB217Seats(g)
	tusk := effectstest.B12Push(g, me.ID, "Charging Tuskodon", "Creature — Dinosaur", oldB217Tuskodon, 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 9)
	before := opp.Life
	effectstest.P2Ping(t, g, tusk, opp.ID, 3)
	if got := before - opp.Life; got != 3 {
		t.Errorf("noncombat damage to a player was %d, want 3", got)
	}
	effectstest.Damage(t, g, tusk, bear, 4)
	if got := effectstest.DamageOn(t, g, bear); got != 4 {
		t.Errorf("damage to a creature was %d, want 4", got)
	}
}

// --- Dinosaur Hunter ---------------------------------------------------------

func TestOldB217HunterDestroysADinosaurItDamages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB217Seats(g)
	hunter := effectstest.B12Push(g, me.ID, "Dinosaur Hunter", "Creature — Human Pirate", oldB217Hunter, 2, 2)
	dino := effectstest.B12Creature(g, opp.ID, "Big Dino", "Creature — Dinosaur", 6, 6)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 6, 6)
	effectstest.Damage(t, g, hunter, bear, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a non-Dinosaur was destroyed")
	}
	effectstest.Damage(t, g, hunter, dino, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dino) {
		t.Error("the Dinosaur survived the Hunter's damage")
	}
}

// --- Mutiny ------------------------------------------------------------------

func TestOldB217MutinyMakesAnOpponentsCreatureHitAnotherOfTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, third := oldB217Seats(g)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 3)
	elsewhere := effectstest.B12Creature(g, third.ID, "Elsewhere", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Mutiny", "Sorcery", oldB217Mutiny, effectstest.CardRefs(big, elsewhere)); err == nil {
		t.Error("Mutiny was cast with a second target another player controls")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Mutiny", "Sorcery", oldB217Mutiny, effectstest.CardRefs(mine, small)); err == nil {
		t.Error("Mutiny was cast with a first target you control")
	}
	effectstest.CastCatalogSpell(t, g, "Mutiny", "Sorcery", oldB217Mutiny, effectstest.CardRefs(big, small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the second creature survived 5 damage")
	}
	if got := effectstest.DamageOn(t, g, big); got != 0 {
		t.Errorf("the first creature took %d, want 0", got)
	}
}

// --- Reckless Rage -----------------------------------------------------------

func TestOldB217RecklessRageHitsAnOpponentsCreatureAndYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB217Seats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 5)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 5)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Reckless Rage", "Instant", oldB217Rage, effectstest.CardRefs(mine, other)); err == nil {
		t.Error("Reckless Rage was cast with two creatures you control")
	}
	effectstest.CastCatalogSpell(t, g, "Reckless Rage", "Instant", oldB217Rage, effectstest.CardRefs(theirs, mine))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 4-toughness creature survived 4 damage")
	}
	if got := effectstest.DamageOn(t, g, mine); got != 2 {
		t.Errorf("your creature took %d, want 2", got)
	}
	if got := effectstest.DamageOn(t, g, other); got != 0 {
		t.Errorf("an untargeted creature took %d", got)
	}
}

// --- Form of the Dinosaur ----------------------------------------------------

func TestOldB217FormSetsYourLifeTotalTo15(t *testing.T) {
	for _, start := range []int{40, 5} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, _, _ := oldB217Seats(g)
		g.WithWriteLock(func() { me.Life = start })
		effectstest.CastCatalogSpell(t, g, "Form of the Dinosaur", "Enchantment", oldB217Form, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != 15 {
			t.Errorf("from %d life: life is %d, want 15", start, me.Life)
		}
	}
}

func TestOldB217FormFiresAtYourUpkeepAndTheCreatureHitsYouBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB217Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Form of the Dinosaur", "Enchantment", oldB217Form, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 3, 12)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 5, 5)
	_ = mine
	life := me.Life
	// Walk around the table to my next upkeep.
	mySeat := oldB217SeatIndex(g, me.ID)
	for i := 0; i < 2*len(g.Seats) && (i == 0 || g.Turn.ActiveSeat != mySeat); i++ {
		effectstest.EndTurn(t, g)
	}
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		effectstest.AdvanceToUpkeepOfSeat(t, g, mySeat)
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the targeted creature survived 15 damage")
	}
	if got := life - me.Life; got != 3 {
		t.Errorf("you took %d from the dead creature, want its power 3", got)
	}
}

func oldB217SeatIndex(g *game.Game, id uuid.UUID) int {
	for i, s := range g.Seats {
		if s.ID == id {
			return i
		}
	}
	return -1
}

// --- Sea Legs ----------------------------------------------------------------

func TestOldB217SeaLegsPumpsAPirateAndShrinksEverythingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB217Seats(g)
	pirate := effectstest.B12Creature(g, me.ID, "Pirate", "Creature — Human Pirate", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.Enchant(t, g, "Sea Legs", oldB217SeaLegs, pirate)
	effectstest.Enchant(t, g, "Sea Legs", oldB217SeaLegs, bear)
	if p, to := effectstest.PtOf(t, g, pirate); p != 2 || to != 4 {
		t.Errorf("the Pirate is %d/%d, want 2/4", p, to)
	}
	if p, to := effectstest.PtOf(t, g, bear); p != 1 || to != 3 {
		t.Errorf("the Bear is %d/%d, want 1/3", p, to)
	}
}

// --- Squire's Devotion -------------------------------------------------------

func TestOldB217SquiresDevotionPumpsGrantsLifelinkAndMakesAVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := oldB217Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "Squire's Devotion", oldB217Squire, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.PtOf(t, g, bear); p != 3 || to != 3 {
		t.Errorf("the enchanted creature is %d/%d, want 3/3", p, to)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the enchanted creature has no lifelink")
	}
	var found bool
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Vampire" && c.Power == 1 && c.Toughness == 1 {
			found = true
			if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "lifelink") {
				t.Error("the Vampire token has no lifelink")
			}
		}
	}
	if !found {
		t.Error("no 1/1 Vampire token was created")
	}
}

// --- See Red -----------------------------------------------------------------

func TestOldB217SeeRedPumpsAndGivesFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := oldB217Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "See Red", oldB217SeeRed, bear)
	if p, to := effectstest.PtOf(t, g, bear); p != 4 || to != 3 {
		t.Errorf("the enchanted creature is %d/%d, want 4/3", p, to)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the enchanted creature has no first strike")
	}
}

func TestOldB217SeeRedIsSacrificedAtYourEndStepUnlessYouAttacked(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := oldB217Seats(g)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		aura := effectstest.PushCatalogPermanent(g, me.ID, "See Red", "Enchantment — Aura", oldB217SeeRed, false)
		g.WithWriteLock(func() { _ = g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: bear}) })
		if attacked {
			effectstest.DeclareAttack(t, g, opp.ID, bear)
		}
		effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(aura); got != attacked {
			t.Errorf("attacked=%v: the Aura is on the battlefield = %v, want %v", attacked, got, attacked)
		}
	}
}

// --- Tilonalli's Crown -------------------------------------------------------

func TestOldB217CrownPumpsGivesTrampleAndPingsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := oldB217Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
	effectstest.Enchant(t, g, "Tilonalli's Crown", oldB217Crown, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.PtOf(t, g, bear); p != 5 || to != 3 {
		t.Errorf("the enchanted creature is %d/%d, want 5/3", p, to)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the enchanted creature has no trample")
	}
	if got := effectstest.DamageOn(t, g, bear); got != 1 {
		t.Errorf("the creature took %d from the Aura, want 1", got)
	}
}

func TestOldB217CrownKillsAOneToughnessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := oldB217Seats(g)
	mouse := effectstest.B12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	effectstest.Enchant(t, g, "Tilonalli's Crown", oldB217Crown, mouse)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mouse) {
		t.Error("a 1-toughness creature survived the Crown's 1 damage")
	}
}

// --- Journey to Eternity // Atzal, Cave of Eternity --------------------------

func oldB217JourneyRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Journey to Eternity // Atzal, Cave of Eternity",
		Layout:   "transform",
		TypeLine: "Legendary Enchantment — Aura // Legendary Land",
		OracleID: uuid.MustParse(oldB217Journey),
		CardFaces: []cards.CardFace{
			{Name: "Journey to Eternity", TypeLine: "Legendary Enchantment — Aura", ManaCost: "{1}{B}{G}", Colors: []string{"B", "G"}},
			{Name: "Atzal, Cave of Eternity", TypeLine: "Legendary Land"},
		},
	}
}

func TestOldB217JourneyReturnsTheCreatureAndItselfTransformed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := oldB217Seats(g)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	journey := effectstest.ImportToBattlefield(t, g, oldB217JourneyRow(), me)
	if c := effectstest.AdCard(t, g, journey); !c.IsAttachedTo(bear) {
		t.Fatalf("setup: Journey to Eternity is attached to %+v", c.AttachedTo)
	}
	effectstest.RtaDestroy(t, g, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Bear") {
		t.Error("the enchanted creature did not come back as a new object")
	}
	var land *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Atzal, Cave of Eternity" {
			land = &g.Battlefield.Cards[i]
		}
	}
	if land == nil {
		t.Fatalf("Journey to Eternity did not return transformed (it is in %v)", effectstest.AdZone(g, journey))
	}
	if !land.IsLand() || land.Controller != me.ID {
		t.Errorf("Atzal is a land = %v controlled by %v", land.IsLand(), land.Controller)
	}
}

func TestOldB217AtzalReturnsACreatureFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _, _ := oldB217Seats(g)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	journey := effectstest.ImportToBattlefield(t, g, oldB217JourneyRow(), me)
	effectstest.RtaDestroy(t, g, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	// The creature came back; send it to the graveyard again for the land to return.
	effectstest.RtaDestroy(t, g, effectstest.RoonaOnly(t, g, "Bear").InstanceID)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	var back uuid.UUID
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Bear" {
			back = c.InstanceID
		}
	}
	land := effectstest.RoonaOnly(t, g, "Atzal, Cave of Eternity").InstanceID
	_ = journey
	effectstest.ApaMana(me, "B", "G", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(back)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Bear") {
		t.Error("Atzal did not return the creature card from the graveyard")
	}
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("Atzal did not tap for its ability")
	}
}
