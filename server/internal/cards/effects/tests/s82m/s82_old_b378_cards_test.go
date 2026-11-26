package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b378_cards_test.go — Born of the Gods, batch 378 (S82 older sets,
// #5559): twelve cards of damage, pump, library and activated abilities.
// Every test plays the real catalog entry through the engine.

const (
	oldB378Forgestoker = "53de1e6f-9471-4cdc-ac67-66d47da4d7eb"
	oldB378Volley      = "93b6083b-f49a-4641-b76f-8ccbec22f91d"
	oldB378Firedancer  = "b30edff0-3716-4dbb-b445-5329b08e36f4"
	oldB378Skyreaping  = "dcdc60ae-1858-4b51-8b64-58a4942bc3d7"
	oldB378Phalanx     = "0edd4894-0df0-4a4e-8a96-d20360ea81c7"
	oldB378BlackOak    = "5fc888d1-241f-45e1-9c20-3c5ae0aadc30"
	oldB378EyeGouge    = "89661ef9-cb42-401a-ae45-b57bb8dac033"
	oldB378Peregrinate = "2aff9be5-f6d4-4cdd-ab59-20202bc2f144"
	oldB378Plea        = "358cfcf3-8ad0-4877-a59a-76ee269169fb"
	oldB378Sanguimancy = "f0a85525-b174-41d9-8871-870af29d3480"
	oldB378Vortex      = "4acb586c-dd87-452c-b3d8-728a4396aae6"
	oldB378Spider      = "f0ad4a60-ce7f-4117-b8b0-266ffd0447c2"
	oldB378Bolt        = "e07f50fb-98b2-40b7-9610-b8e9438be394"

	oldB378Bear = "Creature — Bear"
)

var oldB378Cards = map[string]string{
	oldB378Forgestoker: "Forgestoker Dragon", oldB378Volley: "Lightning Volley", oldB378Firedancer: "Satyr Firedancer",
	oldB378Skyreaping: "Skyreaping", oldB378Phalanx: "Akroan Phalanx", oldB378BlackOak: "Black Oak of Odunos",
	oldB378EyeGouge: "Eye Gouge", oldB378Peregrinate: "Peregrination", oldB378Plea: "Plea for Guidance",
	oldB378Sanguimancy: "Sanguimancy", oldB378Vortex: "Vortex Elemental", oldB378Spider: "Graverobber Spider",
}

// oldB378Activate floats mana, activates ability idx of id with targets and
// settles the stack.
func oldB378Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, idx int, mana string, targets []game.TargetRef, tapIDs ...uuid.UUID) error {
	t.Helper()
	if mana != "" {
		effectstest.FloatMana(t, g, who, mana)
	}
	err := g.ActivateCatalogAbility(who.ID, id, idx, game.ActivateAbilityParams{Strict: true, Targets: targets, TapIDs: tapIDs})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func oldB378Library(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

func TestOldB378EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB378Cards {
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

func TestOldB378ForgestokerDragonPingsAndStopsABlockOnlyWhileAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Forgestoker Dragon", "Creature — Dragon", oldB378Forgestoker, 5, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", oldB378Bear, 1, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB378Bear, 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := oldB378Activate(t, g, me, dragon, 0, "{C}{R}", effectstest.CardTarget(small)); err == nil {
		t.Fatal("the ability was activated while the Dragon was not attacking")
	}
	effectstest.DeclareAttack(t, g, opp.ID, dragon)
	if err := oldB378Activate(t, g, me, dragon, 0, "{C}{R}", effectstest.CardTarget(small)); err != nil {
		t.Fatalf("activate on the 1/1: %v", err)
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 1/1 survived 1 damage")
	}
	if err := oldB378Activate(t, g, me, dragon, 0, "{C}{R}", effectstest.CardTarget(big)); err != nil {
		t.Fatalf("activate on the 3/3: %v", err)
	}
	effectstest.AssertRestrictions(t, g, big, game.CantBlock)
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("the Dragon has no flying")
	}
}

func TestOldB378LightningVolleyGivesCreaturesYouControlAPingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB378Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB378Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Volley", "Instant", oldB378Volley, nil)
	effectstest.PassPriorityAroundTable(t, g)
	life := effectstest.LifeOf(g, opp.ID)
	var err error
	for idx := 0; idx < 3; idx++ {
		if err = oldB378Activate(t, g, me, mine, idx, "", effectstest.PlayerRefs(opp.ID)); err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life-1 {
		t.Errorf("opponent life %d → %d, want -1", life, got)
	}
	for idx := 0; idx < 3; idx++ {
		if err := g.ActivateCatalogAbility(opp.ID, theirs, idx, game.ActivateAbilityParams{Strict: true, Targets: effectstest.PlayerRefs(me.ID)}); err == nil {
			t.Error("an opponent's creature gained the ping")
		}
	}
}

func TestOldB378SatyrFiredancerHitsACreatureForTheSpellsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB378Library(me, 3)
	effectstest.B12Push(g, me.ID, "Satyr Firedancer", "Enchantment Creature — Satyr", oldB378Firedancer, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB378Bear, 3, 3)
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Bolt of Keranos", "Sorcery", oldB378Bolt, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		effectstest.AnswerScryKeepAll(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("the Firedancer did not ask for a target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-3 {
		t.Errorf("opponent life %d → %d, want -3", life, got)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 3/3 survived 3 damage from the Firedancer")
	}
}

func TestOldB378SatyrFiredancerIgnoresSpellDamageToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB378Library(me, 3)
	effectstest.B12Push(g, me.ID, "Satyr Firedancer", "Enchantment Creature — Satyr", oldB378Firedancer, 1, 1)
	hit := effectstest.B12Creature(g, opp.ID, "Hit", oldB378Bear, 3, 3)
	other := effectstest.B12Creature(g, opp.ID, "Other", oldB378Bear, 3, 3)
	effectstest.CastCatalogSpell(t, g, "Bolt of Keranos", "Sorcery", oldB378Bolt, effectstest.CardTarget(hit))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		effectstest.AnswerScryKeepAll(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the Firedancer triggered on damage to a creature")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("the other creature was dealt damage")
	}
}

func TestOldB378SkyreapingDamagesEveryFlyerByGreenDevotion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Green Thing", TypeLine: "Creature — Elf", ManaCost: "{1}{G}{G}", Colors: []string{"G"}, Power: 1, Toughness: 1})
	small := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Small Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 2, Keywords: []string{"flying"}})
	big := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Big Flyer", TypeLine: "Creature — Bird", Power: 3, Toughness: 3, Keywords: []string{"flying"}})
	ground := effectstest.B12Creature(g, opp.ID, "Ground", oldB378Bear, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Skyreaping", "Sorcery", oldB378Skyreaping, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 1/2 flyer survived 2 damage")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(ground) {
		t.Error("a 3/3 flyer or a ground creature died to 2 damage")
	}
}

func TestOldB378AkroanPhalanxPumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	phalanx := effectstest.B12Push(g, me.ID, "Akroan Phalanx", "Creature — Human Soldier", oldB378Phalanx, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB378Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB378Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := oldB378Activate(t, g, me, phalanx, 0, "{C}{C}{R}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 3 {
		t.Errorf("my bear has power %d, want 3", p)
	}
	if p, _ := effectstest.PtOf(t, g, phalanx); p != 4 {
		t.Errorf("the Phalanx has power %d, want 4", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("their bear has power %d, want 2", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, phalanx, "vigilance") {
		t.Error("the Phalanx has no vigilance")
	}
}

func TestOldB378BlackOakTapsAnotherCreatureToGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oak := effectstest.B12Push(g, me.ID, "Black Oak of Odunos", "Creature — Zombie Treefolk", oldB378BlackOak, 0, 5)
	effectstest.AdvanceToMain(t, g)
	if err := oldB378Activate(t, g, me, oak, 0, "{B}", nil); err == nil {
		t.Fatal("the ability was paid with no other creature to tap")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB378Bear, 2, 2)
	if err := oldB378Activate(t, g, me, oak, 0, "{B}", nil, bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, oak); p != 1 || tough != 6 {
		t.Errorf("the Oak is %d/%d, want 1/6", p, tough)
	}
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("the other creature is not tapped")
	}
}

func TestOldB378EyeGougeShrinksAndDestroysACyclops(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB378Bear, 3, 3)
	cyclops := effectstest.B12Creature(g, opp.ID, "Cyclops", "Creature — Cyclops", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Eye Gouge", "Instant", oldB378EyeGouge, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 2/2", p, tough)
	}
	effectstest.CastCatalogSpell(t, g, "Eye Gouge", "Instant", oldB378EyeGouge, effectstest.CardRefs(cyclops))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cyclops) {
		t.Error("the Cyclops survived Eye Gouge")
	}
}

func TestOldB378PeregrinationFindsTwoBasicsAndScrys(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	oldB378Library(me, 3)
	a := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	b := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Peregrination", "Sorcery", oldB378Peregrinate, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !effectstest.B16Tapped(t, g, a) {
		t.Error("the first Forest is not on the battlefield tapped")
	}
	if !me.Hand.Contains(b) {
		t.Error("the second Forest is not in hand")
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("Peregrination did not scry after the search")
	}
}

func TestOldB378PleaForGuidanceTakesTwoEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	a := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Ench A", TypeLine: "Enchantment"})
	b := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Ench B", TypeLine: "Enchantment Creature — Spirit", Power: 1, Toughness: 1})
	c := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Ench C", TypeLine: "Enchantment"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: oldB378Bear})
	effectstest.CastCatalogSpell(t, g, "Plea for Guidance", "Sorcery", oldB378Plea, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) || me.Hand.Contains(c) {
		t.Errorf("hand has A %v B %v C %v, want true true false", me.Hand.Contains(a), me.Hand.Contains(b), me.Hand.Contains(c))
	}
}

func TestOldB378SanguimancyDrawsAndLosesByBlackDevotion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB378Library(me, 5)
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Black Thing", TypeLine: "Creature — Zombie", ManaCost: "{2}{B}{B}", Colors: []string{"B"}, Power: 1, Toughness: 1})
	hand, life := me.Hand.Size(), effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Sanguimancy", "Sorcery", oldB378Sanguimancy, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want +2", hand, me.Hand.Size())
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-2 {
		t.Errorf("life %d → %d, want -2", life, got)
	}
}

func TestOldB378VortexElementalTucksItselfAndItsCombatPartners(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB378Bear, 0, 3)
	vortex := effectstest.B12Push(g, opp.ID, "Vortex Elemental", "Creature — Elemental", oldB378Vortex, 0, 1)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", oldB378Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(vortex, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	if err := oldB378Activate(t, g, opp, vortex, 0, "{U}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(vortex) || g.Battlefield.Contains(attacker) {
		t.Error("the Vortex or the creature it blocked is still on the battlefield")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("an uninvolved creature was tucked")
	}
	if !opp.Library.Contains(vortex) || !me.Library.Contains(attacker) {
		t.Error("each creature is not in its owner's library")
	}
}

func TestOldB378VortexElementalMakesATargetBlockIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vortex := effectstest.B12Push(g, me.ID, "Vortex Elemental", "Creature — Elemental", oldB378Vortex, 0, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB378Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := oldB378Activate(t, g, me, vortex, 1, "{C}{C}{C}{U}{U}", effectstest.CardTarget(theirs)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.DeclareAttack(t, g, opp.ID, vortex)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers(nil); err == nil {
		t.Error("the opponent declared no blockers although the target must block if able")
	}
}

func TestOldB378GraverobberSpiderGrowsOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spider := effectstest.B12Push(g, me.ID, "Graverobber Spider", "Creature — Spider", oldB378Spider, 2, 4)
	for i := 0; i < 2; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead", TypeLine: oldB378Bear, Owner: me.ID, Controller: me.ID})
	}
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := oldB378Activate(t, g, me, spider, 0, "{C}{C}{C}{B}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, spider); p != 4 || tough != 6 {
		t.Errorf("the Spider is %d/%d, want 4/6", p, tough)
	}
	if err := oldB378Activate(t, g, me, spider, 0, "{C}{C}{C}{B}", nil); err == nil {
		t.Error("the ability was activated twice in a turn")
	}
}
