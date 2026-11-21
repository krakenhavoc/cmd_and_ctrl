package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ravnica: Clue Edition bonus-sheet cards, batch 13 (#5929).

const (
	bnsB13Harpooner   = "332b6ed7-90de-46c0-85ad-1ad708b2c1e1"
	bnsB13Warrior     = "b05d000d-39b4-4db5-bd32-9c77a7434127"
	bnsB13Wrangler    = "635dda8c-3761-4c87-8723-8cd09878516b"
	bnsB13Enforcer    = "bd13392e-e72a-4dbb-ba9d-68f68e47e30e"
	bnsB13Leapfrog    = "6614dbe5-d7c7-4a51-91db-912764050269"
	bnsB13Lightning   = "1f4738db-be2e-4d02-ad0c-1b74dd6bbfdc"
	bnsB13Blackguard  = "8ccd20c7-b7fc-4ab1-a98b-3bbcc2079339"
	bnsB13Biomancer   = "afeba4d7-b659-4fb1-84dd-27704fbb2621"
	bnsB13Vampire     = "6031d854-10a8-43da-b5c4-b561194bcd4e"
	bnsB13Mentor      = "56c8b610-0753-45ce-b1cc-6a5e58538286"
	bnsB13Goblin      = "cbd4e160-9eb1-42de-aee8-6f1d1398f633"
	bnsB13Guildmage   = "877bfa3f-6e30-4cb3-831e-d56849b01334"
	bnsB13CreatureTyp = "Creature — Bear"
)

func TestBnsB13CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB13Harpooner: "Kraul Harpooner", bnsB13Warrior: "Kraul Warrior", bnsB13Wrangler: "Kronch Wrangler",
		bnsB13Enforcer: "Law-Rune Enforcer", bnsB13Leapfrog: "Leapfrog", bnsB13Lightning: "Living Lightning",
		bnsB13Blackguard: "Masked Blackguard", bnsB13Biomancer: "Master Biomancer", bnsB13Vampire: "Memory Vampire",
		bnsB13Mentor: "Ordruun Mentor", bnsB13Goblin: "Ornery Goblin", bnsB13Guildmage: "Orzhov Guildmage",
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

func bnsB13Activate(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID, idx int, params game.ActivateAbilityParams) error {
	t.Helper()
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, src, idx, params)
}

// bnsB13Hand casts a creature with real power and toughness from hand.
func bnsB13Hand(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID})
	effectstest.AdvanceToMain(t, g)
	p.LandDropsPerTurn++
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// bnsB13Settle answers whatever the table asks until nothing is pending and
// the stack is empty: it orders triggers, accepts optional triggers, and picks
// `target` for a single-target prompt (none when it is uuid.Nil).
func bnsB13Settle(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	idle := 0
	for i := 0; i < 32 && idle < 2; i++ {
		switch {
		case effectstest.TriggerOrderPrompt(g) != nil:
			ch := effectstest.TriggerOrderPrompt(g)
			if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, ch.TriggerOrderIDs); err != nil {
				t.Fatal(err)
			}
			idle = 0
		case effectstest.LatestTriggerPromptFor(g, me.ID) != nil:
			if err := g.ResolveTriggerPrompt(effectstest.LatestTriggerPromptFor(g, me.ID).ID, me.ID, true); err != nil {
				t.Fatal(err)
			}
			idle = 0
		case effectstest.LatestPickTarget(g, me.ID) != nil:
			pick := effectstest.LatestPickTarget(g, me.ID)
			var refs []game.TargetRef
			if target != uuid.Nil {
				refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: target})
			}
			if err := g.ResolvePickTargets(pick.ID, me.ID, refs); err != nil {
				t.Fatalf("ResolvePickTargets: %v", err)
			}
			idle = 0
		case len(g.PendingChoices) > 0:
			return // a prompt the caller answers itself
		default:
			effectstest.PassPriorityAroundTable(t, g)
			if len(g.PendingChoices) == 0 && g.Stack.Size() == 0 {
				idle++
			} else {
				idle = 0
			}
		}
	}
}

// --- Kraul Warrior / Masked Blackguard ---------------------------------------

func TestBnsB13KraulWarriorPumpsByThreeEachActivationAndNeedsSixMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Kraul Warrior", "Creature — Insect Warrior", bnsB13Warrior, 2, 2)
	effectstest.FillPoolColored(me, "G", 5)
	if err := bnsB13Activate(t, g, me, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with five mana")
	}
	effectstest.FillPoolColored(me, "G", 7)
	for i := 0; i < 2; i++ {
		if err := bnsB13Activate(t, g, me, id, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 8 || tough != 8 {
		t.Fatalf("after two activations %d/%d, want 8/8", p, tough)
	}
	effectstest.NextTurnOfMine(t, g)
	if p := effectstest.EffectivePower(t, g, id); p != 2 {
		t.Errorf("the pump outlived the turn: power %d", p)
	}
}

func TestBnsB13MaskedBlackguardHasFlashAndPumpsByOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Masked Blackguard", "Creature — Human Rogue", bnsB13Blackguard, 2, 1)
	effectstest.AssertKeywords(t, g, id, "flash")
	effectstest.FillPoolColored(me, "B", 2)
	if err := bnsB13Activate(t, g, me, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with two mana")
	}
	effectstest.FillPoolColored(me, "B", 3)
	if err := bnsB13Activate(t, g, me, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 3 || tough != 2 {
		t.Errorf("pumped to %d/%d, want 3/2", p, tough)
	}
}

// --- Law-Rune Enforcer --------------------------------------------------------

func TestBnsB13LawRuneEnforcerTapsOnlyManaValueTwoOrGreater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	enf := effectstest.B12Push(g, me.ID, "Law-Rune Enforcer", "Creature — Human Soldier", bnsB13Enforcer, 1, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == enf {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	cheap := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cheap", TypeLine: bnsB13CreatureTyp,
		ManaCost: "{1}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	dear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dear", TypeLine: bnsB13CreatureTyp,
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.FillPoolColored(me, "C", 1)
	if err := bnsB13Activate(t, g, me, enf, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.CardRef(cheap)}}); err == nil {
		t.Fatal("targeted a mana value 1 creature")
	}
	effectstest.FillPoolColored(me, "C", 1)
	if err := bnsB13Activate(t, g, me, enf, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.CardRef(dear)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, dear); !c.Tapped {
		t.Error("the mana value 2 creature was not tapped")
	}
	if c := effectstest.B12Card(t, g, cheap); c.Tapped {
		t.Error("the mana value 1 creature was tapped")
	}
	if c := effectstest.B12Card(t, g, enf); !c.Tapped {
		t.Error("the Enforcer did not tap for its cost")
	}
}

// --- Ornery Goblin -----------------------------------------------------------

func TestBnsB13OrneryGoblinPokesEachCreatureItBlocksOrIsBlockedBy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goblin := effectstest.B12Push(g, opp.ID, "Ornery Goblin", "Creature — Goblin Warrior", bnsB13Goblin, 2, 1)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", bnsB13CreatureTyp, 2, 4)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", bnsB13CreatureTyp, 2, 4)
	effectstest.WoeS2Block(g, opp.ID, goblin, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, attacker); got != 1 {
		t.Errorf("the blocked creature has %d damage, want 1", got)
	}
	if got := effectstest.Pr6Marked(g, bystander); got != 0 {
		t.Errorf("a creature outside the block has %d damage, want 0", got)
	}
	first := effectstest.B12Creature(g, opp.ID, "First Blocker", bnsB13CreatureTyp, 2, 4)
	second := effectstest.B12Creature(g, opp.ID, "Second Blocker", bnsB13CreatureTyp, 2, 4)
	mine := effectstest.B12Push(g, me.ID, "Ornery Goblin", "Creature — Goblin Warrior", bnsB13Goblin, 2, 1)
	effectstest.WoeS2Block(g, opp.ID, first, mine)
	effectstest.WoeS2Block(g, opp.ID, second, mine)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{first, second} {
		if got := effectstest.Pr6Marked(g, id); got != 1 {
			t.Errorf("a blocker has %d damage, want 1", got)
		}
	}
}

// --- Orzhov Guildmage ----------------------------------------------------------

func TestBnsB13OrzhovGuildmageGainsOneForATargetAndEveryoneLosesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Orzhov Guildmage", "Creature — Human Wizard", bnsB13Guildmage, 2, 2)
	oppLife := opp.Life
	effectstest.FillPoolColored(me, "W", 3)
	if err := bnsB13Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("gain activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife+1 {
		t.Errorf("target player life %d -> %d, want +1", oppLife, opp.Life)
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.FillPoolColored(me, "B", 3)
	if err := bnsB13Activate(t, g, me, id, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("loss activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-1 {
			t.Errorf("%s life %d -> %d, want -1", p.Name, lives[p.ID], p.Life)
		}
	}
	effectstest.FillPoolColored(me, "W", 2)
	if err := bnsB13Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("activated the gain ability with two mana")
	}
}

// --- Leapfrog ------------------------------------------------------------------

func TestBnsB13LeapfrogFliesOnlyAfterYouCastAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	frog := effectstest.B12Push(g, me.ID, "Leapfrog", "Creature — Frog", bnsB13Leapfrog, 3, 1)
	if effectstest.HasEffectiveKeyword(t, g, frog, "flying") {
		t.Fatal("flying before any spell was cast")
	}
	creature := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: creature, Name: "Bear", TypeLine: bnsB13CreatureTyp, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.LandDropsPerTurn++
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, frog, "flying") {
		t.Fatal("a creature spell gave flying")
	}
	effectstest.WoeS2CastInstant(t, g, me, "")
	if !effectstest.HasEffectiveKeyword(t, g, frog, "flying") {
		t.Fatal("no flying after casting an instant")
	}
	effectstest.NextTurnOfMine(t, g)
	if effectstest.HasEffectiveKeyword(t, g, frog, "flying") {
		t.Error("still flying on the next turn")
	}
}

func TestBnsB13LeapfrogIgnoresOpponentsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	frog := effectstest.B12Push(g, me.ID, "Leapfrog", "Creature — Frog", bnsB13Leapfrog, 3, 1)
	effectstest.WoeS2CastInstant(t, g, opp, "")
	if effectstest.HasEffectiveKeyword(t, g, frog, "flying") {
		t.Error("an opponent's instant gave flying")
	}
}

// --- Kronch Wrangler -----------------------------------------------------------

func TestBnsB13KronchWranglerGrowsOnlyForCreaturesOfPowerFourOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	w := effectstest.B12Push(g, me.ID, "Kronch Wrangler", "Creature — Human Warrior", bnsB13Wrangler, 2, 1)
	effectstest.AssertKeywords(t, g, w, "trample")
	for _, c := range []struct {
		owner *game.Player
		power int
	}{{me, 3}, {opp, 5}} {
		id := effectstest.B12Creature(g, c.owner.ID, "Visitor", bnsB13CreatureTyp, c.power, c.power)
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id}) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.B12Card(t, g, w).Counters[game.CounterPlusOne]; n != 0 {
		t.Fatalf("%d counters after a power 3 creature of mine and a power 5 creature of theirs", n)
	}
	big := effectstest.B12Creature(g, me.ID, "Big", bnsB13CreatureTyp, 4, 4)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: big}) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Card(t, g, w).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("%d counters after a power 4 creature entered, want 1", n)
	}
}

// --- Ordruun Mentor ------------------------------------------------------------

func TestBnsB13OrdruunMentorMentorsAndGivesFirstStrikeToAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mentor := effectstest.B12Push(g, me.ID, "Ordruun Mentor", "Creature — Minotaur Soldier", bnsB13Mentor, 3, 2)
	small := effectstest.B12Creature(g, me.ID, "Bear", bnsB13CreatureTyp, 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, mentor, small)
	bnsB13Settle(t, g, me, small)
	if n := effectstest.B12Card(t, g, small).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("mentor counters on the lesser attacker: %d, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, small, "first strike") {
		t.Error("the attacker did not gain first strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, mentor, "first strike") {
		t.Error("the Mentor itself gained first strike")
	}
}

// --- Living Lightning ----------------------------------------------------------

func TestBnsB13LivingLightningReturnsAnInstantOrSorceryWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bolt := effectstest.SeedGraveyardCard(t, g, "Shock", "Instant", "")
	bear := effectstest.SeedGraveyardCard(t, g, "Dead Bear", bnsB13CreatureTyp, "")
	id := effectstest.B12Push(g, me.ID, "Living Lightning", "Creature — Elemental Shaman", bnsB13Lightning, 3, 2)
	g.DestroyPermanentsForEffect([]uuid.UUID{id})
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bolt}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := bnsB13HandCard(me, bolt); !ok {
		t.Error("the instant did not return to hand")
	}
	if _, ok := bnsB13HandCard(me, bear); ok {
		t.Error("a creature card returned to hand")
	}
}

func TestBnsB13LivingLightningDoesNothingWithNoInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedGraveyardCard(t, g, "Dead Bear", bnsB13CreatureTyp, "")
	id := effectstest.B12Push(g, me.ID, "Living Lightning", "Creature — Elemental Shaman", bnsB13Lightning, 3, 2)
	handBefore := len(me.Hand.Cards)
	g.DestroyPermanentsForEffect([]uuid.UUID{id})
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != handBefore {
		t.Errorf("hand grew from %d to %d with nothing to return", handBefore, len(me.Hand.Cards))
	}
}

func bnsB13HandCard(p *game.Player, id uuid.UUID) (game.Card, bool) {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// --- Master Biomancer ----------------------------------------------------------

func TestBnsB13MasterBiomancerGivesEveryOtherCreatureCountersAndTheMutantType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bio := effectstest.B12Push(g, me.ID, "Master Biomancer", "Creature — Elf Wizard", bnsB13Biomancer, 2, 4)
	id := bnsB13Hand(t, g, me, "Plain Bear", bnsB13CreatureTyp, "", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.B12Card(t, g, id)
	if n := bear.Counters[game.CounterPlusOne]; n != 2 {
		t.Errorf("entered with %d counters, want 2 (the Biomancer's power)", n)
	}
	if !bnsB13HasSubtype(bear.Effective().Subtypes, "Mutant") {
		t.Errorf("not a Mutant: %v", bear.Effective().Subtypes)
	}
	if !bnsB13HasSubtype(bear.Effective().Subtypes, "Bear") {
		t.Errorf("lost its other types: %v", bear.Effective().Subtypes)
	}
	if n := effectstest.B12Card(t, g, bio).Counters[game.CounterPlusOne]; n != 0 {
		t.Errorf("the Biomancer put %d counters on itself", n)
	}
	_ = opp
}

func TestBnsB13MasterBiomancerIgnoresNoncreaturesAndOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Master Biomancer", "Creature — Elf Wizard", bnsB13Biomancer, 2, 4)
	art := bnsB13Hand(t, g, me, "Trinket", "Artifact", "", 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Card(t, g, art).Counters[game.CounterPlusOne]; n != 0 {
		t.Errorf("a noncreature got %d counters", n)
	}
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Their Bear", TypeLine: bnsB13CreatureTyp, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	if n := effectstest.B12Card(t, g, id).Counters[game.CounterPlusOne]; n != 0 {
		t.Errorf("an opposing creature got %d counters", n)
	}
}

func bnsB13HasSubtype(subs []string, want string) bool {
	for _, s := range subs {
		if s == want {
			return true
		}
	}
	return false
}

// --- Kraul Harpooner -----------------------------------------------------------

func bnsB13Fliers(g *game.Game, opp *game.Player) (flier, ground uuid.UUID) {
	flier = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bat", TypeLine: "Creature — Bat",
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	ground = effectstest.B12Creature(g, opp.ID, "Ground Bear", bnsB13CreatureTyp, 2, 2)
	return
}

func TestBnsB13KraulHarpoonerPumpsByGraveyardCreaturesAndFightsAFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	flier, _ := bnsB13Fliers(g, opp)
	effectstest.SeedGraveyardCreature(me, "Dead One", "{1}")
	effectstest.SeedGraveyardCreature(me, "Dead Two", "{2}")
	effectstest.SeedGraveyardCard(t, g, "Old Shock", "Instant", "")
	id := bnsB13Hand(t, g, me, "Kraul Harpooner", "Creature — Insect Warrior", bnsB13Harpooner, 3, 5)
	for i := 0; i < 8 && g.Stack.Size() > 0 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: flier}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, id); p != 5 {
		t.Errorf("power %d, want 3 + 2 creature cards = 5", p)
	}
	if bnsB13OnBattlefield(g, flier) {
		t.Errorf("the 2/2 flier survived a fight with a 5-power Harpooner")
	}
	if got := effectstest.Pr6Marked(g, id); got != 2 {
		t.Errorf("the Harpooner took %d from the fight, want 2", got)
	}
}

func TestBnsB13KraulHarpoonerCanDeclineTheFightAndStillGetsThePump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	flier, _ := bnsB13Fliers(g, opp)
	effectstest.SeedGraveyardCreature(me, "Dead One", "{1}")
	id := bnsB13Hand(t, g, me, "Kraul Harpooner", "Creature — Insect Warrior", bnsB13Harpooner, 3, 5)
	for i := 0; i < 8 && g.Stack.Size() > 0 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: flier}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, id); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if got := effectstest.Pr6Marked(g, flier); got != 0 {
		t.Errorf("the flier took %d with the fight declined", got)
	}
}

func TestBnsB13KraulHarpoonerWithNoFlierStillPumpsAndFightsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	_, ground := bnsB13Fliers(g, opp)
	effectstest.SeedGraveyardCreature(me, "Dead One", "{1}")
	id := bnsB13Hand(t, g, me, "Kraul Harpooner", "Creature — Insect Warrior", bnsB13Harpooner, 3, 5)
	bnsB13Settle(t, g, me, uuid.Nil)
	if p := effectstest.EffectivePower(t, g, id); p != 4 {
		t.Errorf("power %d, want 4: the pump needs no target", p)
	}
	if got := effectstest.Pr6Marked(g, ground); got != 0 {
		t.Errorf("a creature without flying took %d", got)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm {
			t.Error("asked about a fight with no target")
		}
	}
}

// --- Memory Vampire ------------------------------------------------------------

func bnsB13VampireHits(t *testing.T, g *game.Game, me, opp *game.Player, damage int, victims ...*game.Player) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	v := effectstest.PushCatalogPermanent(g, me.ID, "Memory Vampire", "Creature — Vampire Detective", bnsB13Vampire, false)
	effectstest.AssertKeywords(t, g, v, "flying")
	effectstest.DealCombatDamageToPlayer(g, v, opp.ID, damage)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		var refs []game.TargetRef
		for _, p := range victims {
			refs = append(refs, game.TargetRef{Kind: game.TargetPlayer, ID: p.ID})
		}
		if err := g.ResolvePickTargets(pick.ID, me.ID, refs); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	return v
}

func TestBnsB13MemoryVampireMillsEachTargetedPlayerByTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	meBefore, oppBefore, thirdBefore := len(me.Library.Cards), len(opp.Library.Cards), len(third.Library.Cards)
	bnsB13VampireHits(t, g, me, opp, 4, opp, me)
	if got := oppBefore - len(opp.Library.Cards); got != 4 {
		t.Errorf("the defender milled %d, want 4", got)
	}
	if got := meBefore - len(me.Library.Cards); got != 4 {
		t.Errorf("I milled %d, want 4 (I was a target)", got)
	}
	if got := thirdBefore - len(third.Library.Cards); got != 0 {
		t.Errorf("an untargeted player milled %d", got)
	}
}

func TestBnsB13MemoryVampireMayTargetNobody(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oppBefore := len(opp.Library.Cards)
	bnsB13VampireHits(t, g, me, opp, 3)
	if got := oppBefore - len(opp.Library.Cards); got != 0 {
		t.Errorf("milled %d with no targets", got)
	}
}

func TestBnsB13MemoryVampireCollectingEvidenceCastsANonlandCardFreeFromTheDefendersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	// Evidence worth 9 in my graveyard.
	e1 := effectstest.SeedGraveyardCreature(me, "Evidence A", "{4}{R}")
	e2 := effectstest.SeedGraveyardCreature(me, "Evidence B", "{3}{G}")
	prize := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: prize, Name: "Prize Bear", TypeLine: bnsB13CreatureTyp, ManaCost: "{5}{G}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	land := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Old Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	v := effectstest.PushCatalogPermanent(g, me.ID, "Memory Vampire", "Creature — Vampire Detective", bnsB13Vampire, false)
	effectstest.DealCombatDamageToPlayer(g, v, opp.ID, 1)
	bnsB13Settle(t, g, me, uuid.Nil)
	effectstest.AnswerChooseCards(t, g, me.ID, e1, e2)
	bnsB13Settle(t, g, me, prize)
	offer := bnsB13CastOffer(g, me.ID)
	if offer == nil {
		t.Fatal("no free-cast offer for the targeted card")
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, prize, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting the offered card: %v", err)
	}
	bnsB13Settle(t, g, me, uuid.Nil)
	if !bnsB13OnBattlefield(g, prize) {
		t.Error("the targeted card was not cast and resolved")
	}
	if bnsB13OnBattlefield(g, land) {
		t.Error("a land entered")
	}
	if _, ok := effectstest.GraveyardCard(opp, land); !ok {
		t.Error("the land left the graveyard")
	}
	if _, ok := effectstest.GraveyardCard(me, e1); ok {
		t.Error("evidence was not exiled")
	}
}

func TestBnsB13MemoryVampireCantCollectEvidenceWithTooLittleInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedGraveyardCreature(me, "Evidence A", "{4}")
	v := effectstest.PushCatalogPermanent(g, me.ID, "Memory Vampire", "Creature — Vampire Detective", bnsB13Vampire, false)
	effectstest.DealCombatDamageToPlayer(g, v, opp.ID, 1)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.EvidenceEvents(g)); n != 0 {
		t.Errorf("%d collections from a graveyard worth 4", n)
	}
}

func bnsB13OnBattlefield(g *game.Game, id uuid.UUID) bool {
	return effectstest.FindBattlefieldCardForTest(g, id) != nil
}

func bnsB13CastOffer(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast && c.Chooser == chooser {
			return c
		}
	}
	return nil
}
