package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b406_cards_test.go — Dragon's Maze, batch 406 (S82 older sets,
// #5564): twelve cards of name locks, hand attack, draw replacement, graveyard
// grants and spell damage. Every test plays the real catalog entry through the
// engine.

const (
	oldB406Council   = "464d5db8-5da1-4ced-aebf-33b3b6cb5bb2"
	oldB406Collector = "6bef2904-1ebe-4e7e-879a-79f4968174fb"
	oldB406Gatekeep  = "7c6d213b-448a-4231-9f6f-65c8ad49a5f1"
	oldB406Scrivener = "af66da4b-2f36-4dd2-a5c6-18e6023c5aba"
	oldB406Blast     = "cfe82e0f-7ac0-4de3-b1b3-a2d519b00bbd"
	oldB406Torturer  = "b1ad2fd3-9f39-49e7-92c5-57ab0e1da5b7"
	oldB406Mastiff   = "bcbaef07-1d5b-457b-b373-77d3a2884b39"
	oldB406Varolz    = "b6b83b88-0660-4227-895e-78ac43d01b66"
	oldB406Morgue    = "ee3b72ac-35b6-45d0-bb50-5a5899b291b1"
	oldB406Render    = "3966bd03-a98e-43b6-a0c0-10d4d81c173c"
	oldB406Commando  = "92e2f982-f8b9-4c3b-ab50-14c025370e5c"
	oldB406Animist   = "beebed34-3a9b-4b51-aacb-1e85d76cf326"

	oldB406Bear = "Creature — Bear"
)

var oldB406Cards = map[string]string{
	oldB406Council: "Council of the Absolute", oldB406Collector: "Sin Collector", oldB406Gatekeep: "Smelt-Ward Gatekeepers",
	oldB406Scrivener: "Blood Scrivener", oldB406Blast: "Blast of Genius", oldB406Torturer: "Hired Torturer",
	oldB406Mastiff: "Boros Mastiff", oldB406Varolz: "Varolz, the Scar-Striped", oldB406Morgue: "Morgue Burst",
	oldB406Render: "Render Silent", oldB406Commando: "Blaze Commando", oldB406Animist: "Feral Animist",
}

func TestOldB406EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB406Cards {
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

func oldB406Hand(p *game.Player, name, typeLine, cost string) uuid.UUID {
	return effectstest.HandCardFull(p, name, typeLine, cost, "", nil)
}

func TestOldB406CouncilOfTheAbsoluteLocksOpponentsAndDiscountsYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.CastCatalogSpell(t, g, "Council of the Absolute", "Creature — Human Advisor", oldB406Council, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "Divination")
	effectstest.ToMainForCost(t, g)
	theirs := oldB406Hand(opp, "Divination", "Sorcery", "{2}{U}")
	if err := g.CastSpell(opp.ID, theirs, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("an opponent cast a spell with the chosen name")
	}
	mine := oldB406Hand(me, "Divination", "Sorcery", "{2}{U}")
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting the named spell for {U}: %v", err)
	}
	plain := oldB406Hand(me, "Plain Sorcery", "Sorcery", "{2}{U}")
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a spell with another name cost {2} less")
	}
}

func TestOldB406SinCollectorExilesAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(opp,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	effectstest.CastCatalogSpell(t, g, "Sin Collector", "Creature — Human Cleric", oldB406Collector, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[0]}); err == nil {
		t.Fatal("a creature card was accepted as an instant or sorcery card")
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[1]) {
		t.Error("the instant was not exiled")
	}
	if !opp.Hand.Contains(ids[0]) {
		t.Error("the creature card left the hand")
	}
}

func TestOldB406SmeltWardGatekeepersThreatensOnlyWithTwoGates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: oldB406Bear,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Tapped: true})
	effectstest.B12Push(g, me.ID, "Gate One", "Land — Gate", "", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Smelt-Ward Gatekeepers", "Creature — Human Warrior", oldB406Gatekeep, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the trigger went on the stack with only one Gate")
	}
	effectstest.B12Push(g, me.ID, "Gate Two", "Land — Gate", "", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Smelt-Ward Gatekeepers", "Creature — Human Warrior", oldB406Gatekeep, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MscM2PickTargets(t, g, theirs)
	c, ok := g.LookupCardForEffect(theirs)
	if !ok || c.Controller != me.ID {
		t.Fatal("the target did not change control")
	}
	if c.Tapped {
		t.Error("the stolen creature was not untapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("the stolen creature has no haste")
	}
}

func TestOldB406BloodScrivenerDoublesAnEmptyHandDrawAndCostsALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Blood Scrivener", "Creature — Zombie Wizard", oldB406Scrivener, 2, 1)
	for me.Hand.Size() > 0 {
		me.Hand.Cards = me.Hand.Cards[:0]
	}
	life := effectstest.LifeOf(g, me.ID)
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand after the replaced draw = %d, want 2", got)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-1 {
		t.Errorf("life %d → %d, want -1", life, got)
	}
	life = effectstest.LifeOf(g, me.ID)
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand after a draw with cards in hand = %d, want 3", got)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life {
		t.Errorf("life changed to %d on an ordinary draw", got)
	}
}

func TestOldB406HiredTorturerDrainsAndRevealsARandomCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	torturer := effectstest.B12Push(g, me.ID, "Hired Torturer", "Creature — Human Rogue", oldB406Torturer, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, torturer, "defender") {
		t.Error("the Torturer has no defender")
	}
	oppHand := opp.Hand.Size()
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, torturer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.PlayerRefs(opp.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent life %d → %d, want -2", life, got)
	}
	if opp.Hand.Size() != oppHand {
		t.Error("the revealed card left the hand")
	}
	if !effectstest.TappedNow(g, torturer) {
		t.Error("the Torturer was not tapped as a cost")
	}
}

func TestOldB406BorosMastiffGainsLifelinkOnlyInABattalion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mastiff := effectstest.B12Push(g, me.ID, "Boros Mastiff", "Creature — Dog", oldB406Mastiff, 2, 2)
	a := effectstest.B12Creature(g, me.ID, "A", oldB406Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", oldB406Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, mastiff, a)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, mastiff, "lifelink") {
		t.Error("lifelink without two other attackers")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	mastiff2 := effectstest.B12Push(g2, me2.ID, "Boros Mastiff", "Creature — Dog", oldB406Mastiff, 2, 2)
	a2 := effectstest.B12Creature(g2, me2.ID, "A", oldB406Bear, 2, 2)
	b2 := effectstest.B12Creature(g2, me2.ID, "B", oldB406Bear, 2, 2)
	_ = b
	effectstest.DeclareAttack(t, g2, opp2.ID, mastiff2, a2, b2)
	effectstest.PassPriorityAroundTable(t, g2)
	if !effectstest.HasEffectiveKeyword(t, g2, mastiff2, "lifelink") {
		t.Error("no lifelink with two other attackers")
	}
}

func TestOldB406VarolzGivesScavengeAtItsManaCostAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	varolz := effectstest.B12Push(g, me.ID, "Varolz, the Scar-Striped", "Legendary Creature — Troll Warrior", oldB406Varolz, 2, 2)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: oldB406Bear, ManaCost: "{1}{G}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	target := effectstest.B12Creature(g, opp.ID, "Target", oldB406Bear, 1, 1)
	effectstest.FloatMana(t, g, me, "{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{
		Strict: true, Ref: "grant:varolz-the-scar-striped/scavenge:0:0", Targets: effectstest.CardTarget(target),
	}); err != nil {
		t.Fatalf("scavenge from the graveyard for its mana cost: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, target, game.CounterPlusOne); got != 3 {
		t.Errorf("target has %d +1/+1 counters, want 3 (the exiled card's power)", got)
	}
	if !g.Exile.Contains(dead) {
		t.Error("the scavenged card was not exiled")
	}
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", oldB406Bear, 1, 1)
	effectstest.ApaActivate(t, g, me, varolz, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	effectstest.P1Destroy(t, g, varolz)
	if !g.Battlefield.Contains(varolz) {
		t.Error("Varolz did not regenerate")
	}
}

func TestOldB406MorgueBurstReturnsACreatureAndDealsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Ogre", TypeLine: "Creature — Ogre", ManaCost: "{3}{R}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Morgue Burst", "Sorcery", oldB406Morgue,
		append(effectstest.CardTarget(dead), effectstest.PlayerRefs(opp.ID)...))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card was not returned to hand")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent life %d → %d, want -4", life, got)
	}
}

func TestOldB406RenderSilentCountersAndSilencesItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastCatalogSpell(t, g, "Render Silent", "Instant", oldB406Render, effectstest.CardTarget(bolt))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	if me.Life != 40 {
		t.Errorf("the countered Bolt dealt damage: life %d", me.Life)
	}
	next := oldB406Hand(opp, "Second Bolt", "Instant", "")
	if err := g.CastSpell(opp.ID, next, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("the countered spell's controller cast another spell this turn")
	}
	if err := g.CastSpell(me.ID, oldB406Hand(me, "Mine", "Instant", ""), game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the caster of Render Silent was silenced: %v", err)
	}
}

func TestOldB406BlazeCommandoMakesTwoSoldiersWhenYourSpellDealsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Blaze Commando", "Creature — Minotaur Soldier", oldB406Commando, 5, 3)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.TrekB37Tokens(g, me.ID, "Soldier")); got != 2 {
		t.Errorf("%d Soldier tokens after a spell dealt damage, want 2 (one trigger for one damage instance)", got)
	}
}

func TestOldB406FeralAnimistDoublesItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	animist := effectstest.B12Push(g, me.ID, "Feral Animist", "Creature — Goblin Shaman", oldB406Animist, 2, 1)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, animist, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, animist); got != 4 {
		t.Errorf("power = %d, want 4", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, animist, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate again: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, animist); got != 8 {
		t.Errorf("power after a second activation = %d, want 8", got)
	}
}

func TestOldB406BlazeCommandoIgnoresDamageFromAnOpponentsSpellAndFromCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Blaze Commando", "Creature — Minotaur Soldier", oldB406Commando, 5, 3)
	effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.TrekB37Tokens(g, me.ID, "Soldier")); got != 0 {
		t.Errorf("%d Soldiers after an opponent's spell dealt damage, want 0", got)
	}
}

func TestOldB406BlastOfGeniusDrawsThenDiscardsAndDealsTheDiscardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	keep := effectstest.HandCardFull(me, "Cheap", "Instant", "{U}", "", nil)
	big := effectstest.HandCardFull(me, "Big", "Sorcery", "{4}{R}", "", nil)
	life := effectstest.LifeOf(g, opp.ID)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Blast of Genius", "Sorcery", oldB406Blast, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt after drawing three cards")
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand = %d after the draws, want %d", got, hand+3)
	}
	effectstest.AnswerDiscard(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-5 {
		t.Errorf("opponent life %d → %d, want -5 (the discarded card's mana value)", life, got)
	}
	if !me.Hand.Contains(keep) {
		t.Error("the other card left the hand")
	}
}
