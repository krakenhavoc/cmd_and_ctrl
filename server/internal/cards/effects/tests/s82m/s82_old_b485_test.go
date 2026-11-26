package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 485 (#5581): Rise of the Eldrazi, instants and sorceries,
// card draw and library, counterspells and bounce, removal, enchantments.

const (
	roe485GelatinousGenesis = "291e4b91-17c7-4b7b-8057-f0d361424016"
	roe485RealitySpasm      = "34a2fc4c-ab56-4ecd-95d9-04316046843c"
	roe485PerishTheThought  = "8955996a-76c9-496a-9238-5fd89bcdaae4"
	roe485AuraFinesse       = "445c46d1-b571-4fe9-92cc-2873242d748f"
	roe485ExplosiveRevel    = "49f01296-7b30-4d84-bfa3-6ed58cd23a33"
	roe485Nighthaze         = "c0044283-7034-4d1d-876a-a40ea15b7cc6"
	roe485LayBare           = "4d3ac460-79ac-4eef-9446-1b9fe0261556"
	roe485UnifiedWill       = "f90cc072-750c-4a87-bf6c-37192d15e691"
	roe485BalaGedScorpion   = "753b4258-bc1d-47f4-b994-356507b0816a"
	roe485GrotagSiegeRunner = "f76969e7-b075-417c-bc15-1b94a17fef54"
	roe485GravityWell       = "33f757ad-c636-4f17-ba9e-40856f83efd0"
	roe485BearType          = "Creature — Bear"
)

func TestRoeB485CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		roe485GelatinousGenesis: "Gelatinous Genesis", roe485RealitySpasm: "Reality Spasm",
		roe485PerishTheThought: "Perish the Thought", roe485AuraFinesse: "Aura Finesse",
		roe485ExplosiveRevel: "Explosive Revelation", roe485Nighthaze: "Nighthaze",
		roe485LayBare: "Lay Bare", roe485UnifiedWill: "Unified Will",
		roe485BalaGedScorpion: "Bala Ged Scorpion", roe485GrotagSiegeRunner: "Grotag Siege-Runner",
		roe485GravityWell: "Gravity Well",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// --- Gelatinous Genesis ------------------------------------------------------

func TestRoeB485GelatinousGenesisMakesXOozesOfSizeX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastFor(t, g, me, "Gelatinous Genesis", "Sorcery", roe485GelatinousGenesis, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Ooze"); n != 3 {
		t.Fatalf("%d Ooze tokens, want 3", n)
	}
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.Name == "Ooze" && (c.CurrentPower() != 3 || c.CurrentToughness() != 3) {
			t.Errorf("an Ooze is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
		}
	}
}

func TestRoeB485GelatinousGenesisWithXZeroMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastFor(t, g, me, "Gelatinous Genesis", "Sorcery", roe485GelatinousGenesis, game.CastSpellParams{XValue: 0})
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Ooze"); n != 0 {
		t.Fatalf("%d Ooze tokens, want 0", n)
	}
}

// --- Reality Spasm -----------------------------------------------------------

func TestRoeB485RealitySpasmTapsXTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", roe485BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", roe485BearType, 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", roe485BearType, 2, 2)
	effectstest.CastFor(t, g, me, "Reality Spasm", "Instant", roe485RealitySpasm, game.CastSpellParams{
		XValue: 2, Modes: []int{0}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B31Tapped(t, g, a) || !effectstest.B31Tapped(t, g, b) {
		t.Error("both targets should be tapped")
	}
	if effectstest.B31Tapped(t, g, c) {
		t.Error("an untargeted permanent was tapped")
	}
}

func TestRoeB485RealitySpasmUntapsXTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", roe485BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", roe485BearType, 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardByID(g, a).Tapped = true
		effectstest.FindBattlefieldCardByID(g, b).Tapped = true
	})
	effectstest.CastFor(t, g, me, "Reality Spasm", "Instant", roe485RealitySpasm, game.CastSpellParams{
		XValue: 2, Modes: []int{1}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B31Tapped(t, g, a) || effectstest.B31Tapped(t, g, b) {
		t.Error("both targets should be untapped")
	}
}

func TestRoeB485RealitySpasmRefusesTheWrongNumberOfTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", roe485BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, "Reality Spasm", "Instant", "", roe485RealitySpasm, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		XValue: 2, Modes: []int{0}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}},
	}); err == nil {
		t.Fatal("X = 2 with one target was accepted")
	}
}

// --- Perish the Thought ------------------------------------------------------

func TestRoeB485PerishTheThoughtShufflesTheChosenCardIntoTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(opp,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: roe485BearType, ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	lib := opp.Library.Size()
	effectstest.CastAt(t, g, "Perish the Thought", "Sorcery", roe485PerishTheThought, effectstest.PbPlayer(opp.ID))
	p := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(p.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(ids[0]) || !opp.Hand.Contains(ids[1]) {
		t.Error("only the chosen card should have left the hand (a land is a legal pick)")
	}
	if !libraryHas(opp, ids[0]) || opp.Library.Size() != lib+1 {
		t.Errorf("the chosen card is not in their library (size %d, want %d)", opp.Library.Size(), lib+1)
	}
	if opp.Graveyard.Contains(ids[0]) {
		t.Error("the card was discarded instead of shuffled in")
	}
}

func libraryHas(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Library.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestRoeB485PerishTheThoughtRefusesAYouTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Perish the Thought", "Sorcery", roe485PerishTheThought,
		[]game.TargetRef{effectstest.PbPlayer(me.ID)}); err == nil {
		t.Fatal("targeting yourself was accepted")
	}
}

// --- Aura Finesse ------------------------------------------------------------

func TestRoeB485AuraFinesseMovesYourAuraAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", roe485BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", roe485BearType, 2, 2)
	aura := effectstest.B15Aura(g, me.ID, a)
	hand := me.Hand.Size()
	effectstest.CastFor(t, g, me, "Aura Finesse", "Instant", roe485AuraFinesse, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: aura}, {Kind: game.TargetCard, ID: b}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedToNow(t, g, aura, b) {
		t.Error("the Aura is not on the new creature")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the Finesse was cast from it, then one drawn)", me.Hand.Size(), hand+1)
	}
}

func TestRoeB485AuraFinesseRefusesAnOpponentsAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", roe485BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", roe485BearType, 2, 2)
	aura := effectstest.B15Aura(g, opp.ID, a)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Aura Finesse", "Instant", roe485AuraFinesse,
		[]game.TargetRef{{Kind: game.TargetCard, ID: aura}, {Kind: game.TargetCard, ID: b}}); err == nil {
		t.Fatal("an Aura you do not control was accepted")
	}
}

// --- Explosive Revelation ----------------------------------------------------

func TestRoeB485ExplosiveRevelationDealsTheNonlandCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Library.Cards = nil
	hit := effectstest.SeedLibraryTop(me, "Hit", "Creature — Bear", "{2}{R}")
	land := effectstest.SeedLibraryTop(me, "Mountain", "Basic Land — Mountain", "")
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastAt(t, g, "Explosive Revelation", "Sorcery", roe485ExplosiveRevel, effectstest.PbPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-3 {
		t.Errorf("life %d, want %d", got, life-3)
	}
	if !me.Hand.Contains(hit) {
		t.Error("the nonland card should be in hand")
	}
	if len(me.Library.Cards) != 1 || me.Library.Cards[0].InstanceID != land {
		t.Error("the revealed land should be on the bottom of the library")
	}
}

func TestRoeB485ExplosiveRevelationWithNoNonlandDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Library.Cards = nil
	land := effectstest.SeedLibraryTop(me, "Mountain", "Basic Land — Mountain", "")
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastAt(t, g, "Explosive Revelation", "Sorcery", roe485ExplosiveRevel, effectstest.PbPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life {
		t.Errorf("life %d, want %d", got, life)
	}
	if len(me.Library.Cards) != 1 || me.Library.Cards[0].InstanceID != land {
		t.Error("the land should still be in the library")
	}
}

func TestRoeB485ExplosiveRevelationCanTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Library.Cards = nil
	effectstest.SeedLibraryTop(me, "Hit", "Creature — Bear", "{3}{R}")
	victim := effectstest.B12Creature(g, opp.ID, "Victim", roe485BearType, 3, 3)
	effectstest.CastAt(t, g, "Explosive Revelation", "Sorcery", roe485ExplosiveRevel, game.TargetRef{Kind: game.TargetCard, ID: victim})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("4 damage should have killed the 3/3")
	}
}

// --- Nighthaze ---------------------------------------------------------------

func TestRoeB485NighthazeGrantsSwampwalkAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", roe485BearType, 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Nighthaze", "Sorcery", roe485Nighthaze, effectstest.RtCardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "swampwalk") {
		t.Error("the creature should have swampwalk")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

// --- Lay Bare and Unified Will -----------------------------------------------

// roe485Respond has seat 0 cast a Lightning Bolt at seat 1 and seat 1 answer
// with `name`, a counterspell targeting the Bolt. It returns the table.
func roe485Respond(t *testing.T, name, oracle string, mine, theirs int) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	caster, me := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	for i := 0; i < mine; i++ {
		effectstest.B12Creature(g, me.ID, "Mine", roe485BearType, 1, 1)
	}
	for i := 0; i < theirs; i++ {
		effectstest.B12Creature(g, caster.ID, "Theirs", roe485BearType, 1, 1)
	}
	bolt := effectstest.CastFor(t, g, caster, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtPlayerRef(me.ID)})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	id := effectstest.HandCardFull(me, name, "Instant", "", oracle, []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.RtCardTarget(bolt)}); err != nil {
		t.Fatalf("casting %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, caster, me
}

func TestRoeB485LayBareCountersAndLooksAtTheControllersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, me := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	secret := effectstest.HandCardFull(caster, "Secret", "Sorcery", "", "", nil)
	if caster.Hand.Cards[0].IsKnownTo(me.ID) {
		t.Fatal("the hand was known before the look")
	}
	bolt := effectstest.CastFor(t, g, caster, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtPlayerRef(me.ID)})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	life := effectstest.LifeOf(g, me.ID)
	id := effectstest.HandCardFull(me, "Lay Bare", "Instant", "", roe485LayBare, []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.RtCardTarget(bolt)}); err != nil {
		t.Fatalf("casting Lay Bare: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life {
		t.Errorf("life %d, want %d: the Bolt should be countered", got, life)
	}
	known := false
	for i := range caster.Hand.Cards {
		if caster.Hand.Cards[i].InstanceID == secret {
			known = caster.Hand.Cards[i].IsKnownTo(me.ID)
		}
	}
	if !known {
		t.Error("the caster of Lay Bare did not learn the spell controller's hand")
	}
}

func TestRoeB485UnifiedWillCountersWhenYouControlMoreCreatures(t *testing.T) {
	g, _, me := roe485Respond(t, "Unified Will", roe485UnifiedWill, 2, 1)
	if got := effectstest.LifeOf(g, me.ID); got != 40 {
		t.Errorf("life %d, want 40: the Bolt should be countered", got)
	}
}

func TestRoeB485UnifiedWillDoesNothingWithEqualCreatures(t *testing.T) {
	g, _, me := roe485Respond(t, "Unified Will", roe485UnifiedWill, 2, 2)
	if got := effectstest.LifeOf(g, me.ID); got != 37 {
		t.Errorf("life %d, want 37: equal counts counter nothing", got)
	}
}

func TestRoeB485UnifiedWillDoesNothingWithFewerCreatures(t *testing.T) {
	g, _, me := roe485Respond(t, "Unified Will", roe485UnifiedWill, 0, 3)
	if got := effectstest.LifeOf(g, me.ID); got != 37 {
		t.Errorf("life %d, want 37", got)
	}
}

// --- Bala Ged Scorpion -------------------------------------------------------

func TestRoeB485BalaGedScorpionDestroysAPowerOneOrLessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := effectstest.B12Creature(g, opp.ID, "Small", roe485BearType, 1, 4)
	big := effectstest.B12Creature(g, opp.ID, "Big", roe485BearType, 2, 2)
	effectstest.CastHoldCreature(t, g, "Bala Ged Scorpion", "Creature — Scorpion", roe485BalaGedScorpion, 2, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, big) {
		t.Error("a creature with power 2 is offered")
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the power 1 creature survived")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the power 2 creature was destroyed")
	}
}

func TestRoeB485BalaGedScorpionMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := effectstest.B12Creature(g, opp.ID, "Small", roe485BearType, 1, 1)
	effectstest.CastHoldCreature(t, g, "Bala Ged Scorpion", "Creature — Scorpion", roe485BalaGedScorpion, 2, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("declining the trigger still destroyed the creature")
	}
}

// --- Grotag Siege-Runner -----------------------------------------------------

func roe485Grotag(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grotag := effectstest.B12Push(g, me.ID, "Grotag Siege-Runner", "Creature — Goblin Rogue", roe485GrotagSiegeRunner, 2, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R")
	return g, me, opp, grotag
}

func TestRoeB485GrotagDestroysADefenderAndHitsItsController(t *testing.T) {
	g, me, opp, grotag := roe485Grotag(t)
	wall := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 8, Keywords: []string{"defender"}})
	life := effectstest.LifeOf(g, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, grotag, 0, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(wall)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) || g.Battlefield.Contains(grotag) {
		t.Error("the Wall and the sacrificed Grotag should both be gone")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("life %d, want %d", got, life-2)
	}
}

func TestRoeB485GrotagRefusesACreatureWithoutDefender(t *testing.T) {
	g, me, opp, grotag := roe485Grotag(t)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", roe485BearType, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, grotag, 0, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(bear)}); err == nil {
		t.Fatal("a creature without defender was accepted")
	}
}

// --- Gravity Well ------------------------------------------------------------

func TestRoeB485GravityWellTakesFlyingFromAnAttackerUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Gravity Well", TypeLine: "Enchantment", OracleID: roe485GravityWell})
	flier := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Flier", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	home := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Home Flier", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	ground := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Ground", "", 2, 2))
	effectstest.DeclareAttack(t, g, opp.ID, flier, ground)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, flier, "flying") {
		t.Error("the attacking flier kept flying")
	}
	if !effectstest.HasEffectiveKeyword(t, g, home, "flying") {
		t.Error("a flier that did not attack lost flying")
	}
}
