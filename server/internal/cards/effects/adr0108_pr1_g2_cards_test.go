package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr1_g2_cards_test.go — ADR 0108 Delivery PR 1 (#1886, #1887),
// card group 2: each card against the board that tells its wording
// apart.

const (
	p1g2CombustionOracle = "522a04ff-cbfe-47b0-bd29-ef6fc27a6905"
	p1g2BlitzOracle      = "d2db3f9c-26fe-487b-bbb7-8bc6f33456f1"
	p1g2NarsetOracle     = "247075c5-62f8-41a1-91b3-562ff0aabb30"
	p1g2AccursedOracle   = "b5aae42b-3fde-4f10-b85e-882c528badef"
	p1g2SmiteOracle      = "37994591-3494-4314-a7da-49c112b0866f"
	p1g2WithinOracle     = "8a1ccbdb-3d89-42fb-a731-6db4241acf24"
	p1g2BargainOracle    = "bfb3d862-d92d-4a4f-9a22-671b55d954fd"
	p1g2WiltOracle       = "3bed0b60-5944-44bb-9ddd-c82f323e6d20"
	p1g2BouncerOracle    = "36ac8fc6-98ad-499b-9adf-046433e2c122"
	p1g2ZenithOracle     = "82e61db8-4625-488f-8a5f-66ace9bbf34a"
	p1g2LavaOracle       = "e42fb51e-254a-43ac-ad02-9f0fad0f4c8a"
	p1g2WoundOracle      = "a30159ae-f6a6-4e29-bca2-769d3657d310"
	p1g2TorchOracle      = "10d33e95-3e5d-447e-ba4a-acd3c33b4045"
	p1g2SpikefieldOracle = "81036c9f-fe0a-45a7-bcd5-0d344f31055a"
)

// p1g2Cast casts a catalog spell from the active player's hand with the
// whole announcement in `params` (targets, optional costs, cost branch,
// sacrifices, X).
func p1g2Cast(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// p1g2Seats is the active player and the next one.
func p1g2Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func p1g2Indestructible(g *game.Game, owner uuid.UUID, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Darksteel Bear", TypeLine: "Artifact Creature — Bear",
		Power: power, Toughness: toughness, Keywords: []string{"indestructible"}, Owner: owner, Controller: owner,
	})
}

func p1g2Planeswalker(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: loyalty}, Owner: owner, Controller: owner,
	})
}

func p1g2Graveyard(p *game.Player, typeLine string, n int) {
	for i := 0; i < n; i++ {
		p.Graveyard.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Old Card", TypeLine: typeLine, Power: 1, Toughness: 1,
			Owner: p.ID, Controller: p.ID,
		})
	}
}

func p1g2WantDamage(t *testing.T, g *game.Game, id uuid.UUID, want int, what string) {
	t.Helper()
	if got := damageMarkedOn(g, id); got != want {
		t.Errorf("%s has %d damage, want %d", what, got, want)
	}
}

func p1g2HasKeyword(g *game.Game, id uuid.UUID, kw string) bool {
	var has bool
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			has = game.HasKeyword(&c, kw)
		}
	})
	return has
}

// The "if that creature would die this turn" spells: a creature the
// spell kills is exiled, and one whose damage was all prevented is
// marked anyway (the Disintegrate ruling), so it is exiled when it is
// destroyed later this turn.
func TestP1G2ThatCreatureSpellsExileWhatTheyKillAndMarkAShieldedTarget(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		setup                  func(g *game.Game, me *game.Player)
	}{
		{"Combustion Technique", "Instant — Lesson", p1g2CombustionOracle, nil},
		{"Blitz of the Thunder-Raptor", "Instant", p1g2BlitzOracle, func(_ *game.Game, me *game.Player) {
			p1g2Graveyard(me, "Sorcery", 2)
		}},
		{"Narset's Rebuke", "Instant", p1g2NarsetOracle, nil},
		{"Burn the Accursed", "Instant", p1g2AccursedOracle, nil},
		{"Smite the Deathless", "Instant", p1g2SmiteOracle, nil},
		{"Wilt in the Heat", "Instant", p1g2WiltOracle, nil},
		{"Bouncer's Beatdown", "Instant", p1g2BouncerOracle, func(g *game.Game, me *game.Player) {
			p1Creature(g, me.ID, 2, 2)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := p1g2Seats(g)
			if tc.setup != nil {
				tc.setup(g, me)
			}
			bear := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, pr6Card(bear))
			passPriorityAroundTable(t, g)
			p1WantZone(t, g, bear, game.ZoneExile, "the bear it killed")

			shielded := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
			p1ShieldCreature(g, shielded)
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, pr6Card(shielded))
			passPriorityAroundTable(t, g)
			p1g2WantDamage(t, g, shielded, 0, "the shielded bear")
			p1Destroy(t, g, shielded)
			p1WantZone(t, g, shielded, game.ZoneExile, "the shielded bear destroyed later")
		})
	}
}

// Combustion Technique: 2 plus the Lesson cards in your graveyard, and
// nothing else in it.
func TestP1G2CombustionTechniqueCountsLessonCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	p1g2Graveyard(me, "Sorcery — Lesson", 2)
	p1g2Graveyard(me, "Instant", 3)
	p1g2Graveyard(opp, "Sorcery — Lesson", 3)
	big := p1Creature(g, opp.ID, 6, 6)
	castCatalogSpell(t, g, "Combustion Technique", "Instant — Lesson", p1g2CombustionOracle, pr6Card(big))
	passPriorityAroundTable(t, g)
	p1g2WantDamage(t, g, big, 4, "the 6/6")
}

// Blitz of the Thunder-Raptor: the instant and sorcery cards in your
// graveyard, not the creature cards; and a planeswalker it kills is
// exiled.
func TestP1G2BlitzCountsInstantsAndSorceriesAndExilesAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	p1g2Graveyard(me, "Instant", 2)
	p1g2Graveyard(me, "Sorcery", 1)
	p1g2Graveyard(me, "Creature — Bear", 4)
	big := p1Creature(g, opp.ID, 6, 6)
	castCatalogSpell(t, g, "Blitz of the Thunder-Raptor", "Instant", p1g2BlitzOracle, pr6Card(big))
	passPriorityAroundTable(t, g)
	p1g2WantDamage(t, g, big, 3, "the 6/6")

	walker := p1g2Planeswalker(g, opp.ID, 3)
	castCatalogSpell(t, g, "Blitz of the Thunder-Raptor", "Instant", p1g2BlitzOracle, pr6Card(walker))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, walker, game.ZoneExile, "the planeswalker Blitz killed")
}

// Narset's Rebuke adds {U}{R}{W}.
func TestP1G2NarsetsRebukeAddsMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	bear := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
	castCatalogSpell(t, g, "Narset's Rebuke", "Instant", p1g2NarsetOracle, pr6Card(bear))
	passPriorityAroundTable(t, g)
	got := map[string]int{}
	for _, c := range batch01PoolColors(me) {
		got[c]++
	}
	if len(me.ManaPool) != 3 || got["U"] != 1 || got["R"] != 1 || got["W"] != 1 {
		t.Errorf("pool %v, want U R W", batch01PoolColors(me))
	}
}

// Burn the Accursed: 2 damage to the creature's controller, read as the
// spell resolves.
func TestP1G2BurnTheAccursedHitsTheController(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := p1g2Seats(g)
	bear := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
	before := opp.Life
	castCatalogSpell(t, g, "Burn the Accursed", "Instant", p1g2AccursedOracle, pr6Card(bear))
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("controller's life %d → %d, want -2", before, opp.Life)
	}
	p1WantZone(t, g, bear, game.ZoneExile, "the bear")
}

// Smite the Deathless: an indestructible creature loses indestructible,
// so the 3 damage kills it and it is exiled; one the damage cannot kill
// loses it until end of turn.
func TestP1G2SmiteTheDeathlessStripsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := p1g2Seats(g)
	small := p1g2Indestructible(g, opp.ID, 2, 2)
	castCatalogSpell(t, g, "Smite the Deathless", "Instant", p1g2SmiteOracle, pr6Card(small))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, small, game.ZoneExile, "the indestructible 2/2")

	big := p1g2Indestructible(g, opp.ID, 5, 5)
	castCatalogSpell(t, g, "Smite the Deathless", "Instant", p1g2SmiteOracle, pr6Card(big))
	passPriorityAroundTable(t, g)
	if p1g2HasKeyword(g, big, "indestructible") {
		t.Fatal("the 5/5 still has indestructible")
	}
	p1Destroy(t, g, big)
	p1WantZone(t, g, big, game.ZoneExile, "the 5/5 destroyed later")
}

// Burn from Within: both riders need the damage. An indestructible
// creature dealt lethal damage loses indestructible and is exiled; one
// whose damage was prevented keeps it; a player is just dealt X.
func TestP1G2BurnFromWithinNeedsTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := p1g2Seats(g)
	hit := p1g2Indestructible(g, opp.ID, 3, 3)
	castXSpell(t, g, "Burn from Within", "Sorcery", p1g2WithinOracle, "{X}{R}", 3, pr6Card(hit))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, hit, game.ZoneExile, "the indestructible 3/3 dealt 3")

	shielded := p1g2Indestructible(g, opp.ID, 3, 3)
	p1ShieldCreature(g, shielded)
	castXSpell(t, g, "Burn from Within", "Sorcery", p1g2WithinOracle, "{X}{R}", 3, pr6Card(shielded))
	passPriorityAroundTable(t, g)
	if !p1g2HasKeyword(g, shielded, "indestructible") {
		t.Error("a creature dealt no damage lost indestructible")
	}
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneBattlefield, "the shielded indestructible creature")

	before := opp.Life
	castXSpell(t, g, "Burn from Within", "Sorcery", p1g2WithinOracle, "{X}{R}", 2, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("player's life %d → %d, want -2", before, opp.Life)
	}
}

// Betrayer's Bargain: the sacrifice branch is paid, and the target is
// exiled.
func TestP1G2BetrayersBargainPaysItsCostAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	fodder := p1Creature(g, me.ID, 1, 1)
	bear := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
	branch := 0
	p1g2Cast(t, g, "Betrayer's Bargain", "Instant", p1g2BargainOracle, "", game.CastSpellParams{
		Targets: pr6Card(bear), CostBranch: &branch, SacrificeIDs: []uuid.UUID{fodder},
	})
	p1WantZone(t, g, fodder, game.ZoneGraveyard, "the sacrificed creature")
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneExile, "the bear")
}

// Wilt in the Heat: {2} less once a card has left your graveyard this
// turn — not an opponent's graveyard — and when it is cast from yours.
func TestP1G2WiltInTheHeatCostsLessAfterACardLeftYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	price := func(from game.ZoneKind) int {
		return selfPriced(t, g, me, p1g2WiltOracle, "Instant", "{2}{R}{W}", from, nil).ManaValue()
	}
	if got := price(game.ZoneHand); got != 4 {
		t.Fatalf("Wilt with nothing gone: %d, want 4", got)
	}
	if got := price(game.ZoneGraveyard); got != 2 {
		t.Errorf("Wilt cast from your own graveyard: %d, want 2", got)
	}
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Theirs", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(theirs); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	if got := price(game.ZoneHand); got != 4 {
		t.Errorf("Wilt after an opponent's card left their graveyard: %d, want 4", got)
	}
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "Mine", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(mine); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	if got := price(game.ZoneHand); got != 2 {
		t.Errorf("Wilt after your card left your graveyard: %d, want 2", got)
	}
}

// Bouncer's Beatdown: X is the greatest power among your creatures, and
// it costs {2} less when it targets a black permanent.
func TestP1G2BouncersBeatdownGreatestPowerAndBlackDiscount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	p1Creature(g, me.ID, 3, 3)
	p1Creature(g, me.ID, 1, 1)
	p1Creature(g, opp.ID, 9, 9)
	big := p1Creature(g, opp.ID, 6, 6)
	castCatalogSpell(t, g, "Bouncer's Beatdown", "Instant", p1g2BouncerOracle, pr6Card(big))
	passPriorityAroundTable(t, g)
	p1g2WantDamage(t, g, big, 3, "the 6/6")

	black := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black Bear", TypeLine: "Creature — Bear", Colors: []string{"B"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if got := selfPriced(t, g, me, p1g2BouncerOracle, "Instant", "{2}{G}", game.ZoneHand, pr6Card(black)).ManaValue(); got != 1 {
		t.Errorf("Bouncer's Beatdown at a black creature: %d, want 1", got)
	}
	if got := selfPriced(t, g, me, p1g2BouncerOracle, "Instant", "{2}{G}", game.ZoneHand, pr6Card(big)).ManaValue(); got != 3 {
		t.Errorf("Bouncer's Beatdown at a non-black creature: %d, want 3", got)
	}
}

// Red Sun's Zenith: "a creature dealt damage this way", and the card
// ends up in its owner's library — even at X=0.
func TestP1G2RedSunsZenithExilesWhatItDamagedAndShufflesItself(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := p1g2Seats(g)
	bear := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
	z := castXSpell(t, g, "Red Sun's Zenith", "Sorcery", p1g2ZenithOracle, "{X}{R}", 2, pr6Card(bear))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneExile, "the bear the Zenith killed")
	p1WantZone(t, g, z, game.ZoneLibrary, "Red Sun's Zenith")

	shielded := p1Creature(g, opp.ID, 5, 5)
	p1ShieldCreature(g, shielded)
	castXSpell(t, g, "Red Sun's Zenith", "Sorcery", p1g2ZenithOracle, "{X}{R}", 3, pr6Card(shielded))
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneGraveyard, "a creature the Zenith dealt no damage")

	zero := castXSpell(t, g, "Red Sun's Zenith", "Sorcery", p1g2ZenithOracle, "{X}{R}", 0, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, zero, game.ZoneLibrary, "a Red Sun's Zenith cast for 0")
}

// Scorching Lava: unkicked, a regenerating creature regenerates; kicked,
// it can't, it is exiled, and a shielded target is marked anyway.
func TestP1G2ScorchingLavaKickedStopsRegenerationAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := p1g2Seats(g)
	unkicked := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
	p1Regenerate(t, g, unkicked)
	castCatalogSpell(t, g, "Scorching Lava", "Instant", p1g2LavaOracle, pr6Card(unkicked))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, unkicked, game.ZoneBattlefield, "a regenerating bear hit by an unkicked Lava")

	kicked := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
	p1Regenerate(t, g, kicked)
	p1g2Cast(t, g, "Scorching Lava", "Instant", p1g2LavaOracle, "", game.CastSpellParams{
		Targets: pr6Card(kicked), OptionalCosts: []int{0},
	})
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, kicked, game.ZoneExile, "a regenerating bear hit by a kicked Lava")

	shielded := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
	p1ShieldCreature(g, shielded)
	p1g2Cast(t, g, "Scorching Lava", "Instant", p1g2LavaOracle, "", game.CastSpellParams{
		Targets: pr6Card(shielded), OptionalCosts: []int{0},
	})
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneExile, "a shielded bear after a kicked Lava")
}

// Necrotic Wound: -X/-X for the creature cards in your graveyard; the
// creature it shrinks to death is exiled, and one it doesn't is marked.
func TestP1G2NecroticWoundShrinksByCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	p1g2Graveyard(me, "Creature — Zombie", 2)
	p1g2Graveyard(me, "Sorcery", 3)
	bear := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
	castCatalogSpell(t, g, "Necrotic Wound", "Instant", p1g2WoundOracle, pr6Card(bear))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneExile, "the bear shrunk to 0/0")

	big := p1Creature(g, opp.ID, 3, 3)
	castCatalogSpell(t, g, "Necrotic Wound", "Instant", p1g2WoundOracle, pr6Card(big))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, big, game.ZoneBattlefield, "the 3/3 at -2/-2")
	p1Destroy(t, g, big)
	p1WantZone(t, g, big, game.ZoneExile, "the shrunk 3/3 destroyed later")
}

// Torch the Tower: 2 damage, or 3 and a scry when bargained; a permanent
// it dealt damage is marked, one whose damage was prevented is not.
func TestP1G2TorchTheTowerBargainAndDealtDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	big := p1Creature(g, opp.ID, 3, 3)
	castCatalogSpell(t, g, "Torch the Tower", "Instant", p1g2TorchOracle, pr6Card(big))
	passPriorityAroundTable(t, g)
	p1g2WantDamage(t, g, big, 2, "the 3/3 unbargained")
	if c := scryChoiceFor(g, me.ID); c != nil {
		t.Error("an unbargained Torch scried")
	}
	p1Destroy(t, g, big)
	p1WantZone(t, g, big, game.ZoneExile, "the 3/3 Torch dealt damage, destroyed later")

	shielded := p1Creature(g, opp.ID, 3, 3)
	p1ShieldCreature(g, shielded)
	castCatalogSpell(t, g, "Torch the Tower", "Instant", p1g2TorchOracle, pr6Card(shielded))
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneGraveyard, "a creature Torch dealt no damage")

	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure",
		Owner: me.ID, Controller: me.ID,
	})
	target := p1Creature(g, opp.ID, 3, 3)
	p1g2Cast(t, g, "Torch the Tower", "Instant", p1g2TorchOracle, "", game.CastSpellParams{
		Targets: pr6Card(target), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("a bargained Torch did not scry")
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, target, game.ZoneExile, "the 3/3 a bargained Torch killed")

	walker := p1g2Planeswalker(g, opp.ID, 2)
	castCatalogSpell(t, g, "Torch the Tower", "Instant", p1g2TorchOracle, pr6Card(walker))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, walker, game.ZoneExile, "the planeswalker Torch killed")
}

// Spikefield Hazard: "a permanent dealt damage this way" — a planeswalker
// too — and a player is just dealt 1.
func TestP1G2SpikefieldHazardMarksAnyPermanentItDamaged(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := p1g2Seats(g)
	walker := p1g2Planeswalker(g, opp.ID, 1)
	castCatalogSpell(t, g, "Spikefield Hazard", "Instant", p1g2SpikefieldOracle, pr6Card(walker))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, walker, game.ZoneExile, "the planeswalker Spikefield Hazard killed")

	shielded := p1Creature(g, opp.ID, 1, 1)
	p1ShieldCreature(g, shielded)
	castCatalogSpell(t, g, "Spikefield Hazard", "Instant", p1g2SpikefieldOracle, pr6Card(shielded))
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneGraveyard, "a creature Spikefield Hazard dealt no damage")

	before := opp.Life
	castCatalogSpell(t, g, "Spikefield Hazard", "Instant", p1g2SpikefieldOracle, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("player's life %d → %d, want -1", before, opp.Life)
	}
}
