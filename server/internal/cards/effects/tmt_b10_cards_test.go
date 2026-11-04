package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b10_cards_test.go — Teenage Mutant Ninja Turtles, batch 10 (#4457).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tmtB10RaphLeoOracle     = "57f39ce2-ecbb-4262-8140-03aba5c4303f"
	tmtB10RaphMikeyOracle   = "82563a5f-cae4-4ece-876c-0df245d2dc58"
	tmtB10AttitudeOracle    = "eaf194a7-d95f-4738-9052-92ad00443a76"
	tmtB10DestroyerOracle   = "fdb8e862-c642-4b7b-b061-400de6061ffe"
	tmtB10RatKingOracle     = "a5656dab-5ce6-48fd-9eae-2c06969bb8df"
	tmtB10RenetOracle       = "6c754808-86a6-43d3-8aeb-96ec19f7396c"
	tmtB10ArmorOracle       = "bda5d0aa-3ed0-4f8c-80ae-f9cb8ed0ac82"
	tmtB10SplinterOracle    = "fb2cd508-f5cd-4d5b-b8e7-3cae1f39fd58"
	tmtB10CloningOracle     = "bd768e8f-8f04-487e-85e9-2828d6fd8745"
	tmtB10LastRoninOracle   = "e1865bec-1744-40c2-9031-7d6363be5333"
	tmtB10NeutrinosOracle   = "b1462c82-7702-451b-b612-990353a79ac4"
	tmtB10TokkaOracle       = "86ef86de-104f-4eae-a2db-893b50e95ad7"
	tmtB10TurtlesTimeOracle = "803971e6-67f3-4fda-aeb6-69ef8ea4b54e"
)

func TestTmtB10EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmtB10RaphLeoOracle:     "Raph & Leo, Sibling Rivals",
		tmtB10RaphMikeyOracle:   "Raph & Mikey, Troublemakers",
		tmtB10AttitudeOracle:    "Raphael, Most Attitude",
		tmtB10DestroyerOracle:   "Raphael, Ninja Destroyer",
		tmtB10RatKingOracle:     "Rat King, Verminister",
		tmtB10RenetOracle:       "Renet, Temporal Apprentice",
		tmtB10ArmorOracle:       "Shredder's Armor",
		tmtB10SplinterOracle:    "Splinter, Radical Rat",
		tmtB10CloningOracle:     "The Cloning of Shredder",
		tmtB10LastRoninOracle:   "The Last Ronin",
		tmtB10NeutrinosOracle:   "The Neutrinos",
		tmtB10TokkaOracle:       "Tokka & Rahzar, Terrible Twos",
		tmtB10TurtlesTimeOracle: "Turtles in Time",
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

// tmtB10Zone names the zone a card is in, or "" when it is in none.
func tmtB10Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var kind game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

// tmtB10Library pushes a card on top of a player's library.
func tmtB10Library(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// tmtB10Graveyard pushes a card into a player's graveyard.
func tmtB10Graveyard(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// tmtB10AnswerAttackTarget answers "what does it attack?" with the first
// option when the put asked it.
func tmtB10AnswerAttackTarget(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	if latestOptionPickFor(g, chooser) != nil {
		answerOptionPick(t, g, chooser, 0)
	}
}

// tmtB10LastRonin pushes The Last Ronin with `lore` lore counters already
// on it, so the next counter crosses one chapter only.
func tmtB10LastRonin(g *game.Game, owner uuid.UUID, lore int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Last Ronin", TypeLine: "Enchantment — Saga", OracleID: tmtB10LastRoninOracle,
		ManaCost: "{4}{B}{G}", Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLore: lore},
	})
}

// --- Raph & Leo, Sibling Rivals ----------------------------------------------

func TestTmtB10RaphAndLeoUntapsTargetsAndAddsOneCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rl := b27Push(g, me.ID, "Raph & Leo, Sibling Rivals", "Legendary Creature — Mutant Ninja Turtle", tmtB10RaphLeoOracle, "{1}{R/W}{R/W}", 2, 4, "R", "W")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	start := len(g.Events)
	const label = "Raph & Leo, Sibling Rivals — untap one or two target attacking creatures; an additional combat phase"

	declareAttack(t, g, opp.ID, rl, bear, other)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(rl, bear)...)
	passPriorityAroundTable(t, g)
	if countResolved(g, start, label) != 1 {
		t.Fatal("Raph & Leo's trigger did not resolve once")
	}
	if isTapped(g, rl) || isTapped(g, bear) {
		t.Error("the two targets should be untapped")
	}
	if !isTapped(g, other) {
		t.Error("an attacker that was not a target stays tapped")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat || next.PhaseOrdinal != 2 {
		t.Fatalf("after the first combat: %+v, want a second combat", next)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(rl, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker in the second combat: %v", err)
	}
	lockInAttacks(t, g)
	if latestPickTarget(g, me.ID) != nil || triggerOnStack(g, rl) != nil {
		t.Error("the second combat is not the first: no trigger")
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after the second combat: %+v, want no third combat", next)
	}
}

// --- Raph & Mikey, Troublemakers ---------------------------------------------

func TestTmtB10RaphAndMikeyPutsTheFirstCreatureCardAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rm := b27Push(g, me.ID, "Raph & Mikey, Troublemakers", "Legendary Creature — Mutant Ninja Turtle", tmtB10RaphMikeyOracle, "{5}{R/G}{R/G}", 7, 7, "R", "G")
	creature := tmtB10Library(me, "Hill Giant", "Creature — Giant", "{3}{R}")
	land := tmtB10Library(me, "Mountain", "Basic Land — Mountain", "")
	spell := tmtB10Library(me, "Shock", "Instant", "{R}")

	declareAttack(t, g, opp.ID, rm)
	passPriorityAroundTable(t, g)
	tmtB10AnswerAttackTarget(t, g, me.ID)
	c, ok := battlefieldCard(g, creature)
	if !ok {
		t.Fatalf("the creature card is in %q, want the battlefield", tmtB10Zone(g, creature))
	}
	if !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("Hill Giant: tapped %v, attacking %v; want tapped and attacking", c.Tapped, c.AttackingTarget)
	}
	n := len(me.Library.Cards)
	bottom := map[uuid.UUID]bool{me.Library.Cards[0].InstanceID: true, me.Library.Cards[1].InstanceID: true}
	if n < 2 || !bottom[land] || !bottom[spell] {
		t.Error("the revealed land and instant should be the bottom two cards of the library")
	}
}

// --- Raphael, Most Attitude ---------------------------------------------------

func TestTmtB10RaphaelMostAttitudeExilesThenPlaysOneCardOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	raph := b27Push(g, me.ID, "Raphael, Most Attitude", "Legendary Creature — Mutant Ninja Turtle", tmtB10AttitudeOracle, "{3}{R}", 4, 3, "R")
	assertKeywords(t, g, raph, "menace")
	first := tmtB10Library(me, "Gitaxian Probe", "Instant", "{0}")
	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 black Rat"), 1) })
		passPriorityAroundTable(t, g)
		answerLatestTriggerPrompt(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
		if i == 0 {
			tmtB10Library(me, "Mental Note", "Instant", "{0}")
		}
	}
	if tmtB10Zone(g, first) != game.ZoneExile {
		t.Fatalf("the top card is in %q, want exile", tmtB10Zone(g, first))
	}
	var exiled []uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.ExiledWith.ID == raph {
			exiled = append(exiled, c.InstanceID)
		}
	}
	if len(exiled) != 2 {
		t.Fatalf("%d cards linked to Raphael, want 2", len(exiled))
	}

	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, exiled[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("a card exiled with Raphael was cast before he attacked")
	}
	declareAttack(t, g, opp.ID, raph)
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, exiled[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting a card exiled with Raphael: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, exiled[1], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("a second card exiled with Raphael was cast: the permission is for one card")
	}
}

func TestTmtB10RaphaelMostAttitudeDeclinedExileLeavesTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b27Push(g, me.ID, "Raphael, Most Attitude", "Legendary Creature — Mutant Ninja Turtle", tmtB10AttitudeOracle, "{3}{R}", 4, 3, "R")
	top := tmtB10Library(me, "Ornithopter", "Artifact Creature — Thopter", "{0}")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 black Rat"), 1) })
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if tmtB10Zone(g, top) != game.ZoneLibrary {
		t.Errorf("declined: the top card is in %q", tmtB10Zone(g, top))
	}
}

// --- Raphael, Ninja Destroyer ------------------------------------------------

func TestTmtB10RaphaelNinjaDestroyerMustBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	raph := reqCreature(g, me.ID, "Raphael, Ninja Destroyer", tmtB10DestroyerOracle, 4, 4, "haste")
	reqCreature(g, opp.ID, "Spiked Wall", "", 3, 5)
	reqToBlockers(t, g, raph)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Raphael, Ninja Destroyer must be blocked if able." {
		t.Errorf("sentence = %q", got)
	}
}

func TestTmtB10RaphaelNinjaDestroyerAddsThatMuchRedThatStays(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	raph := b27Push(g, me.ID, "Raphael, Ninja Destroyer", "Legendary Creature — Mutant Ninja Turtle", tmtB10DestroyerOracle, "{2}{R}{R}", 4, 4, "R")
	src := b12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Pinger", "Creature — Elf", 1, 1)
	mshA4Damage(g, src, raph, 3)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 3})
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"R": 3})
}

// --- Rat King, Verminister ---------------------------------------------------

func TestTmtB10RatKingDisappearMakesARatAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	king := b27Push(g, me.ID, "Rat King, Verminister", "Legendary Creature — Rat Avatar", tmtB10RatKingOracle, "{1}{B}", 1, 1, "B")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Rat"); got != 1 {
		t.Errorf("%d Rat tokens, want 1", got)
	}
	if plusOnes(g, king) != 1 {
		t.Errorf("Rat King has %d +1/+1 counters, want 1", plusOnes(g, king))
	}
}

func TestTmtB10RatKingDisappearNeedsAPermanentToHaveLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	king := b27Push(g, me.ID, "Rat King, Verminister", "Legendary Creature — Rat Avatar", tmtB10RatKingOracle, "{1}{B}", 1, 1, "B")
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Rat"); got != 0 || plusOnes(g, king) != 0 {
		t.Errorf("nothing left: %d Rats, %d counters; want none", got, plusOnes(g, king))
	}
}

func TestTmtB10RatKingReturnsTheTargetAndEveryCardWithItsName(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	king := b27Push(g, me.ID, "Rat King, Verminister", "Legendary Creature — Rat Avatar", tmtB10RatKingOracle, "{1}{B}", 1, 1, "B")
	r1 := b12Creature(g, me.ID, "Rat", "Creature — Rat", 1, 1)
	r2 := b12Creature(g, me.ID, "Rat", "Creature — Rat", 1, 1)
	target := tmtB10Graveyard(me, "Relentless Rats", "Creature — Rat", "{1}{B}{B}")
	twin := tmtB10Graveyard(me, "Relentless Rats", "Creature — Rat", "{1}{B}{B}")
	other := tmtB10Graveyard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	if err := g.ActivateCatalogAbility(me.ID, king, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{king, r1, r2},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{target, twin} {
		c, ok := battlefieldCard(g, id)
		if !ok || !c.Tapped {
			t.Errorf("%s: on battlefield %v, tapped %v; want tapped on the battlefield", id, ok, c.Tapped)
		}
	}
	if tmtB10Zone(g, other) != game.ZoneGraveyard {
		t.Error("a card with another name stays in the graveyard")
	}
}

// --- Renet, Temporal Apprentice ----------------------------------------------

func TestTmtB10RenetBouncesOtherNonlandPermanentsThatEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	old := b12Creature(g, opp.ID, "Old Bear", "Creature — Bear", 2, 2)
	var fresh uuid.UUID
	g.WithWriteLock(func() {
		ids, _ := g.CreateTokensForEffect(opp.ID, TokenCard("1/1 black Rat"), 1, game.TokenEntryOptions{})
		if len(ids) > 0 {
			fresh = ids[0]
		}
	})
	passPriorityAroundTable(t, g)
	land := pushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	g.WithWriteLock(func() { g.TurnTally.Entered[land]++ })
	renet := castCatalogSpell(t, g, "Renet, Temporal Apprentice", "Legendary Creature — Human Wizard", tmtB10RenetOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fresh) {
		t.Error("the token that entered this turn should have been returned (and ceased to exist)")
	}
	if !g.Battlefield.Contains(old) {
		t.Error("a permanent that was already on the battlefield stays")
	}
	if !g.Battlefield.Contains(renet) {
		t.Error("Renet returns only OTHER permanents")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land stays")
	}
}

// --- Shredder's Armor ----------------------------------------------------------

func TestTmtB10ShreddersArmorAttachesOnEntryAndEquipsBySacrificeOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	fodder := b12Permanent(g, me.ID, "Fodder", "Artifact")
	fodder2 := b12Permanent(g, me.ID, "Fodder Two", "Artifact")
	land := pushCatalogPermanent(g, me.ID, "Swamp", "Basic Land — Swamp", "", false)

	armor := castCatalogSpell(t, g, "Shredder's Armor", "Artifact — Equipment", tmtB10ArmorOracle, nil)
	passPriorityAroundTable(t, g)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, bear) != 4 {
		t.Fatalf("equipped Bear has power %d, want 4 (+2/+1)", effectivePower(t, g, bear))
	}

	if err := g.ActivateCatalogAbility(me.ID, armor, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{land},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: other}},
	}); err == nil {
		t.Fatal("a land paid the equip cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, armor, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{armor},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: other}},
	}); err == nil {
		t.Fatal("the Equipment paid for its own equip")
	}
	if err := g.ActivateCatalogAbility(me.ID, armor, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: other}},
	}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed artifact is still on the battlefield")
	}
	if effectivePower(t, g, other) != 4 || effectivePower(t, g, bear) != 2 {
		t.Errorf("after equip: Other Bear %d, Bear %d; want 4 and 2", effectivePower(t, g, other), effectivePower(t, g, bear))
	}
	if err := g.ActivateCatalogAbility(me.ID, armor, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder2},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("equip was activated twice in one turn")
	}
}

// --- Splinter, Radical Rat -----------------------------------------------------

func TestTmtB10SplinterDoublesNinjaTriggersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b27Push(g, me.ID, "Splinter, Radical Rat", "Legendary Creature — Mutant Ninja Rat", tmtB10SplinterOracle, "{1}{W/B}{W/B}", 2, 4, "W", "B")
	// The Neutrinos' alliance is a triggered ability of a creature; as a
	// Ninja it triggers twice, as an Elf once.
	ninja := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Neutrinos", TypeLine: "Legendary Creature — Elf Ninja", OracleID: tmtB10NeutrinosOracle,
		Power: 2, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 black Rat"), 1) })
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, ninja); got != 4 {
		t.Errorf("Ninja Neutrinos power %d, want 4 (alliance twice)", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	b27Push(g2, me2.ID, "Splinter, Radical Rat", "Legendary Creature — Mutant Ninja Rat", tmtB10SplinterOracle, "{1}{W/B}{W/B}", 2, 4, "W", "B")
	elf := b27Push(g2, me2.ID, "The Neutrinos", "Legendary Creature — Elf Rebel", tmtB10NeutrinosOracle, "{2}{R}{W}", 2, 4, "R", "W")
	g2.WithWriteLock(func() { _ = g2.CreateTokenForEffect(me2.ID, TokenCard("1/1 black Rat"), 1) })
	passPriorityAroundTable(t, g2)
	if got := effectivePower(t, g2, elf); got != 3 {
		t.Errorf("non-Ninja Neutrinos power %d, want 3 (alliance once)", got)
	}
}

func TestTmtB10SplinterMakesATargetNinjaUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	splinter := b27Push(g, me.ID, "Splinter, Radical Rat", "Legendary Creature — Mutant Ninja Rat", tmtB10SplinterOracle, "{1}{W/B}{W/B}", 2, 4, "W", "B")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	fillPool(me, 1)
	fillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, splinter, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a non-Ninja was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, splinter, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: splinter}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, splinter, game.CantBeBlocked)
}

// --- The Cloning of Shredder ----------------------------------------------------

func TestTmtB10CloningOfShredderCopiesTheExiledCardEachChapter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := tmtB10Graveyard(me, "Oroku Saki", "Legendary Creature — Human Ninja", "{2}{B}")
	advanceToMain(t, g)
	saga := castCatalogSpell(t, g, "The Cloning of Shredder", "Enchantment — Saga", tmtB10CloningOracle, nil)
	passPriorityAroundTable(t, g)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(victim)...)
	passPriorityAroundTable(t, g)
	if tmtB10Zone(g, victim) != game.ZoneExile {
		t.Fatalf("the creature card is in %q, want exile", tmtB10Zone(g, victim))
	}
	copies := tmtB1Named(g, "Oroku Saki")
	if len(copies) != 1 {
		t.Fatalf("%d copies after chapter I, want 1", len(copies))
	}
	c := tmtB1Card(t, g, copies[0])
	if c.IsLegendary() || !c.HasSubtypeWord("Mutant") || !c.HasSubtypeWord("Ninja") || !c.IsToken() {
		t.Errorf("copy type line %q: want a nonlegendary Human Ninja Mutant token", c.TypeLine)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
	passPriorityAroundTable(t, g)
	if n := len(tmtB1Named(g, "Oroku Saki")); n != 2 {
		t.Errorf("%d copies after chapter II, want 2", n)
	}
}

// --- The Last Ronin ----------------------------------------------------------------

func TestTmtB10LastRoninChapterOneDestroysAllCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "The Last Ronin", "Enchantment — Saga", tmtB10LastRoninOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("chapter I should destroy every creature")
	}
}

func TestTmtB10LastRoninChapterTwoMillsThenReturnsACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	saga := tmtB10LastRonin(g, me.ID, 1)
	milled := tmtB10Library(me, "Hill Giant", "Creature — Giant", "{3}{R}")
	for i := 0; i < 3; i++ {
		tmtB10Library(me, "Mountain", "Basic Land — Mountain", "")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
	passPriorityAroundTable(t, g)
	if tmtB10Zone(g, milled) != game.ZoneGraveyard {
		t.Fatalf("the fourth card is in %q, want the graveyard", tmtB10Zone(g, milled))
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(milled)...)
	passPriorityAroundTable(t, g)
	if tmtB10Zone(g, milled) != game.ZoneHand {
		t.Errorf("the milled creature card is in %q, want the hand", tmtB10Zone(g, milled))
	}
}

func TestTmtB10LastRoninChapterThreePumpsALoneAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	saga := tmtB10LastRonin(g, me.ID, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 3 {
		t.Errorf("the lone attacker has %d +1/+1 counters, want 3", plusOnes(g, bear))
	}
	assertKeywords(t, g, bear, "trample", "lifelink", "indestructible")
}

func TestTmtB10LastRoninChapterThreeIgnoresTwoAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	saga := tmtB10LastRonin(g, me.ID, 2)
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if plusOnes(g, a) != 0 || plusOnes(g, b) != 0 {
		t.Error("two attackers: neither attacks alone")
	}
}

// --- The Neutrinos ----------------------------------------------------------------------

func TestTmtB10NeutrinosAllianceAndBlinkAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	neut := b27Push(g, me.ID, "The Neutrinos", "Legendary Creature — Elf Rebel", tmtB10NeutrinosOracle, "{2}{R}{W}", 2, 4, "R", "W")
	assertKeywords(t, g, neut, "flying")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 black Rat"), 1) })
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, neut); got != 3 {
		t.Fatalf("alliance: power %d, want 3", got)
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, neut)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	tmtB10AnswerAttackTarget(t, g, me.ID)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Bear" {
			back = c
		}
	}
	if back == nil {
		t.Fatal("the blinked Bear did not come back")
	}
	if !back.Tapped || back.AttackingTarget == uuid.Nil || back.Controller != me.ID {
		t.Errorf("Bear: tapped %v, attacking %v, controller mine %v", back.Tapped, back.AttackingTarget, back.Controller == me.ID)
	}
}

// --- Tokka & Rahzar, Terrible Twos -----------------------------------------------------

func TestTmtB10TokkaAndRahzarPunishesAFreeCastOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b27Push(g, opp.ID, "Tokka & Rahzar, Terrible Twos", "Legendary Creature — Turtle Wolf Mutant", tmtB10TokkaOracle, "{B/R}{B/R}", 3, 2, "B", "R")
	life := me.Life

	// Paid in full: no damage.
	paid := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: paid, Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	fillPool(me, 1)
	fillPoolColored(me, "G", 1)
	if err := g.CastSpell(me.ID, paid, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("a spell paid in full cost %d life", life-me.Life)
	}

	// A free cast: 3 damage to the caster.
	free := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: free, Name: "Hill Giant", TypeLine: "Creature — Giant", ManaCost: "{3}{R}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, free, game.CastSpellParams{PaperPayment: &game.PaperPayment{Free: true}}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := life - me.Life; got != 3 {
		t.Errorf("a free cast cost %d life, want 3", got)
	}
}

func TestTmtB10TokkaAndRahzarCantBeCountered(t *testing.T) {
	spec, _ := Lookup(tmtB10TokkaOracle)
	if !spec.CantBeCountered {
		t.Error("Tokka & Rahzar can't be countered")
	}
}

// --- Turtles in Time --------------------------------------------------------------------

func TestTmtB10TurtlesInTimeAsksEveryoneBeforeAnyoneShufflesOrDraws(t *testing.T) {
	g := newCatalogGame(t)
	order := apnapPlayers(g)
	if len(order) < 4 {
		t.Fatalf("want a four-player table, got %d seats", len(order))
	}
	me, second := g.PlayerByIDForEffect(order[0]), g.PlayerByIDForEffect(order[1])
	third := g.PlayerByIDForEffect(order[2])
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, third.ID, "Their Bear", "Creature — Bear", 2, 2)
	myGrave := tmtB10Graveyard(me, "Old Card", "Sorcery", "{1}")
	thirdGrave := tmtB10Graveyard(third, "Their Old Card", "Sorcery", "{1}")

	spell := castCatalogSpell(t, g, "Turtles in Time", "Sorcery", tmtB10TurtlesTimeOracle, nil)
	passPriorityAroundTable(t, g)
	if tmtB10Zone(g, mine) != game.ZoneHand || tmtB10Zone(g, theirs) != game.ZoneHand {
		t.Fatal("every creature returns to its owner's hand")
	}
	hands := map[uuid.UUID]int{}
	for _, id := range order {
		hands[id] = g.PlayerByIDForEffect(id).Hand.Size()
	}
	// APNAP: the caster is asked first, and only the caster.
	if latestConfirmFor(g, second.ID) != nil {
		t.Fatal("the second player was asked before the first answered")
	}
	answerMayChoice(t, g, me.ID, true)
	answerMayChoice(t, g, second.ID, true)
	// Two players have said yes; nobody has shuffled or drawn yet.
	for _, p := range []*game.Player{me, second} {
		if p.Hand.Size() != hands[p.ID] {
			t.Errorf("%s's hand changed from %d to %d before every player answered", p.Name, hands[p.ID], p.Hand.Size())
		}
	}
	if tmtB10Zone(g, mine) != game.ZoneHand || tmtB10Zone(g, myGrave) != game.ZoneGraveyard {
		t.Error("the caster shuffled before every player answered")
	}
	answerMayChoice(t, g, third.ID, false)
	answerMayChoice(t, g, order[3], false)
	passPriorityAroundTable(t, g)

	for _, p := range []*game.Player{me, second} {
		if p.Hand.Size() != 7 {
			t.Errorf("%s shuffled and has %d cards, want 7", p.Name, p.Hand.Size())
		}
	}
	if z := tmtB10Zone(g, myGrave); z != game.ZoneLibrary && z != game.ZoneHand {
		t.Errorf("the caster's graveyard card is in %q, want shuffled in", z)
	}
	if third.Hand.Size() != hands[third.ID] || tmtB10Zone(g, thirdGrave) != game.ZoneGraveyard || tmtB10Zone(g, theirs) != game.ZoneHand {
		t.Error("a player who declined keeps their hand and graveyard")
	}
	if tmtB10Zone(g, spell) != game.ZoneExile {
		t.Errorf("Turtles in Time is in %q, want exile", tmtB10Zone(g, spell))
	}
}
