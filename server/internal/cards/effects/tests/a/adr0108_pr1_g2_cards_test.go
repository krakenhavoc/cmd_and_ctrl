package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The "if that creature would die this turn" spells: a creature the
// spell kills is exiled, and one whose damage was all prevented is
// marked anyway (the Disintegrate ruling), so it is exiled when it is
// destroyed later this turn.
func TestP1G2ThatCreatureSpellsExileWhatTheyKillAndMarkAShieldedTarget(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		setup                  func(g *game.Game, me *game.Player)
	}{
		{"Combustion Technique", "Instant — Lesson", effectstest.P1g2CombustionOracle, nil},
		{"Blitz of the Thunder-Raptor", "Instant", effectstest.P1g2BlitzOracle, func(_ *game.Game, me *game.Player) {
			effectstest.P1g2Graveyard(me, "Sorcery", 2)
		}},
		{"Narset's Rebuke", "Instant", effectstest.P1g2NarsetOracle, nil},
		{"Burn the Accursed", "Instant", effectstest.P1g2AccursedOracle, nil},
		{"Smite the Deathless", "Instant", effectstest.P1g2SmiteOracle, nil},
		{"Wilt in the Heat", "Instant", effectstest.P1g2WiltOracle, nil},
		{"Bouncer's Beatdown", "Instant", effectstest.P1g2BouncerOracle, func(g *game.Game, me *game.Player) {
			effectstest.P1Creature(g, me.ID, 2, 2)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.P1g2Seats(g)
			if tc.setup != nil {
				tc.setup(g, me)
			}
			bear := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.Pr6Card(bear))
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear it killed")

			shielded := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
			effectstest.P1ShieldCreature(g, shielded)
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.Pr6Card(shielded))
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.P1g2WantDamage(t, g, shielded, 0, "the shielded bear")
			effectstest.P1Destroy(t, g, shielded)
			effectstest.P1WantZone(t, g, shielded, game.ZoneExile, "the shielded bear destroyed later")
		})
	}
}

// Combustion Technique: 2 plus the Lesson cards in your graveyard, and
// nothing else in it.
func TestP1G2CombustionTechniqueCountsLessonCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	effectstest.P1g2Graveyard(me, "Sorcery — Lesson", 2)
	effectstest.P1g2Graveyard(me, "Instant", 3)
	effectstest.P1g2Graveyard(opp, "Sorcery — Lesson", 3)
	big := effectstest.P1Creature(g, opp.ID, 6, 6)
	effectstest.CastCatalogSpell(t, g, "Combustion Technique", "Instant — Lesson", effectstest.P1g2CombustionOracle, effectstest.Pr6Card(big))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1g2WantDamage(t, g, big, 4, "the 6/6")
}

// Blitz of the Thunder-Raptor: the instant and sorcery cards in your
// graveyard, not the creature cards; and a planeswalker it kills is
// exiled.
func TestP1G2BlitzCountsInstantsAndSorceriesAndExilesAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	effectstest.P1g2Graveyard(me, "Instant", 2)
	effectstest.P1g2Graveyard(me, "Sorcery", 1)
	effectstest.P1g2Graveyard(me, "Creature — Bear", 4)
	big := effectstest.P1Creature(g, opp.ID, 6, 6)
	effectstest.CastCatalogSpell(t, g, "Blitz of the Thunder-Raptor", "Instant", effectstest.P1g2BlitzOracle, effectstest.Pr6Card(big))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1g2WantDamage(t, g, big, 3, "the 6/6")

	walker := effectstest.P1g2Planeswalker(g, opp.ID, 3)
	effectstest.CastCatalogSpell(t, g, "Blitz of the Thunder-Raptor", "Instant", effectstest.P1g2BlitzOracle, effectstest.Pr6Card(walker))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, walker, game.ZoneExile, "the planeswalker Blitz killed")
}

// Narset's Rebuke adds {U}{R}{W}.
func TestP1G2NarsetsRebukeAddsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	bear := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.CastCatalogSpell(t, g, "Narset's Rebuke", "Instant", effectstest.P1g2NarsetOracle, effectstest.Pr6Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	got := map[string]int{}
	for _, c := range effectstest.Batch01PoolColors(me) {
		got[c]++
	}
	if len(me.ManaPool) != 3 || got["U"] != 1 || got["R"] != 1 || got["W"] != 1 {
		t.Errorf("pool %v, want U R W", effectstest.Batch01PoolColors(me))
	}
}

// Burn the Accursed: 2 damage to the creature's controller, read as the
// spell resolves.
func TestP1G2BurnTheAccursedHitsTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.P1g2Seats(g)
	bear := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Burn the Accursed", "Instant", effectstest.P1g2AccursedOracle, effectstest.Pr6Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("controller's life %d → %d, want -2", before, opp.Life)
	}
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear")
}

// Smite the Deathless: an indestructible creature loses indestructible,
// so the 3 damage kills it and it is exiled; one the damage cannot kill
// loses it until end of turn.
func TestP1G2SmiteTheDeathlessStripsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.P1g2Seats(g)
	small := effectstest.P1g2Indestructible(g, opp.ID, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Smite the Deathless", "Instant", effectstest.P1g2SmiteOracle, effectstest.Pr6Card(small))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, small, game.ZoneExile, "the indestructible 2/2")

	big := effectstest.P1g2Indestructible(g, opp.ID, 5, 5)
	effectstest.CastCatalogSpell(t, g, "Smite the Deathless", "Instant", effectstest.P1g2SmiteOracle, effectstest.Pr6Card(big))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.P1g2HasKeyword(g, big, "indestructible") {
		t.Fatal("the 5/5 still has indestructible")
	}
	effectstest.P1Destroy(t, g, big)
	effectstest.P1WantZone(t, g, big, game.ZoneExile, "the 5/5 destroyed later")
}

// Burn from Within: both riders need the damage. An indestructible
// creature dealt lethal damage loses indestructible and is exiled; one
// whose damage was prevented keeps it; a player is just dealt X.
func TestP1G2BurnFromWithinNeedsTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.P1g2Seats(g)
	hit := effectstest.P1g2Indestructible(g, opp.ID, 3, 3)
	effectstest.CastXSpell(t, g, "Burn from Within", "Sorcery", effectstest.P1g2WithinOracle, "{X}{R}", 3, effectstest.Pr6Card(hit))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, hit, game.ZoneExile, "the indestructible 3/3 dealt 3")

	shielded := effectstest.P1g2Indestructible(g, opp.ID, 3, 3)
	effectstest.P1ShieldCreature(g, shielded)
	effectstest.CastXSpell(t, g, "Burn from Within", "Sorcery", effectstest.P1g2WithinOracle, "{X}{R}", 3, effectstest.Pr6Card(shielded))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, shielded, "indestructible") {
		t.Error("a creature dealt no damage lost indestructible")
	}
	effectstest.P1Destroy(t, g, shielded)
	effectstest.P1WantZone(t, g, shielded, game.ZoneBattlefield, "the shielded indestructible creature")

	before := opp.Life
	effectstest.CastXSpell(t, g, "Burn from Within", "Sorcery", effectstest.P1g2WithinOracle, "{X}{R}", 2, effectstest.Pr6Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("player's life %d → %d, want -2", before, opp.Life)
	}
}

// Betrayer's Bargain: the sacrifice branch is paid, and the target is
// exiled.
func TestP1G2BetrayersBargainPaysItsCostAndExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	fodder := effectstest.P1Creature(g, me.ID, 1, 1)
	bear := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	branch := 0
	effectstest.P1g2Cast(t, g, "Betrayer's Bargain", "Instant", effectstest.P1g2BargainOracle, "", game.CastSpellParams{
		Targets: effectstest.Pr6Card(bear), CostBranch: &branch, SacrificeIDs: []uuid.UUID{fodder},
	})
	effectstest.P1WantZone(t, g, fodder, game.ZoneGraveyard, "the sacrificed creature")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear")
}

// Wilt in the Heat: {2} less once a card has left your graveyard this
// turn — not an opponent's graveyard — and when it is cast from yours.
func TestP1G2WiltInTheHeatCostsLessAfterACardLeftYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	price := func(from game.ZoneKind) int {
		return effectstest.SelfPriced(t, g, me, effectstest.P1g2WiltOracle, "Instant", "{2}{R}{W}", from, nil).ManaValue()
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	effectstest.P1Creature(g, me.ID, 3, 3)
	effectstest.P1Creature(g, me.ID, 1, 1)
	effectstest.P1Creature(g, opp.ID, 9, 9)
	big := effectstest.P1Creature(g, opp.ID, 6, 6)
	effectstest.CastCatalogSpell(t, g, "Bouncer's Beatdown", "Instant", effectstest.P1g2BouncerOracle, effectstest.Pr6Card(big))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1g2WantDamage(t, g, big, 3, "the 6/6")

	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black Bear", TypeLine: "Creature — Bear", Colors: []string{"B"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if got := effectstest.SelfPriced(t, g, me, effectstest.P1g2BouncerOracle, "Instant", "{2}{G}", game.ZoneHand, effectstest.Pr6Card(black)).ManaValue(); got != 1 {
		t.Errorf("Bouncer's Beatdown at a black creature: %d, want 1", got)
	}
	if got := effectstest.SelfPriced(t, g, me, effectstest.P1g2BouncerOracle, "Instant", "{2}{G}", game.ZoneHand, effectstest.Pr6Card(big)).ManaValue(); got != 3 {
		t.Errorf("Bouncer's Beatdown at a non-black creature: %d, want 3", got)
	}
}

// Red Sun's Zenith: "a creature dealt damage this way", and the card
// ends up in its owner's library — even at X=0.
func TestP1G2RedSunsZenithExilesWhatItDamagedAndShufflesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.P1g2Seats(g)
	bear := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	z := effectstest.CastXSpell(t, g, "Red Sun's Zenith", "Sorcery", effectstest.P1g2ZenithOracle, "{X}{R}", 2, effectstest.Pr6Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear the Zenith killed")
	effectstest.P1WantZone(t, g, z, game.ZoneLibrary, "Red Sun's Zenith")

	shielded := effectstest.P1Creature(g, opp.ID, 5, 5)
	effectstest.P1ShieldCreature(g, shielded)
	effectstest.CastXSpell(t, g, "Red Sun's Zenith", "Sorcery", effectstest.P1g2ZenithOracle, "{X}{R}", 3, effectstest.Pr6Card(shielded))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, shielded)
	effectstest.P1WantZone(t, g, shielded, game.ZoneGraveyard, "a creature the Zenith dealt no damage")

	zero := effectstest.CastXSpell(t, g, "Red Sun's Zenith", "Sorcery", effectstest.P1g2ZenithOracle, "{X}{R}", 0, effectstest.Pr6Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, zero, game.ZoneLibrary, "a Red Sun's Zenith cast for 0")
}

// Scorching Lava: unkicked, a regenerating creature regenerates; kicked,
// it can't, it is exiled, and a shielded target is marked anyway.
func TestP1G2ScorchingLavaKickedStopsRegenerationAndExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.P1g2Seats(g)
	unkicked := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1Regenerate(t, g, unkicked)
	effectstest.CastCatalogSpell(t, g, "Scorching Lava", "Instant", effectstest.P1g2LavaOracle, effectstest.Pr6Card(unkicked))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, unkicked, game.ZoneBattlefield, "a regenerating bear hit by an unkicked Lava")

	kicked := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1Regenerate(t, g, kicked)
	effectstest.P1g2Cast(t, g, "Scorching Lava", "Instant", effectstest.P1g2LavaOracle, "", game.CastSpellParams{
		Targets: effectstest.Pr6Card(kicked), OptionalCosts: []int{0},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, kicked, game.ZoneExile, "a regenerating bear hit by a kicked Lava")

	shielded := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1ShieldCreature(g, shielded)
	effectstest.P1g2Cast(t, g, "Scorching Lava", "Instant", effectstest.P1g2LavaOracle, "", game.CastSpellParams{
		Targets: effectstest.Pr6Card(shielded), OptionalCosts: []int{0},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, shielded)
	effectstest.P1WantZone(t, g, shielded, game.ZoneExile, "a shielded bear after a kicked Lava")
}

// Necrotic Wound: -X/-X for the creature cards in your graveyard; the
// creature it shrinks to death is exiled, and one it doesn't is marked.
func TestP1G2NecroticWoundShrinksByCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	effectstest.P1g2Graveyard(me, "Creature — Zombie", 2)
	effectstest.P1g2Graveyard(me, "Sorcery", 3)
	bear := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.CastCatalogSpell(t, g, "Necrotic Wound", "Instant", effectstest.P1g2WoundOracle, effectstest.Pr6Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear shrunk to 0/0")

	big := effectstest.P1Creature(g, opp.ID, 3, 3)
	effectstest.CastCatalogSpell(t, g, "Necrotic Wound", "Instant", effectstest.P1g2WoundOracle, effectstest.Pr6Card(big))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, big, game.ZoneBattlefield, "the 3/3 at -2/-2")
	effectstest.P1Destroy(t, g, big)
	effectstest.P1WantZone(t, g, big, game.ZoneExile, "the shrunk 3/3 destroyed later")
}

// Torch the Tower: 2 damage, or 3 and a scry when bargained; a permanent
// it dealt damage is marked, one whose damage was prevented is not.
func TestP1G2TorchTheTowerBargainAndDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	big := effectstest.P1Creature(g, opp.ID, 3, 3)
	effectstest.CastCatalogSpell(t, g, "Torch the Tower", "Instant", effectstest.P1g2TorchOracle, effectstest.Pr6Card(big))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1g2WantDamage(t, g, big, 2, "the 3/3 unbargained")
	if c := effectstest.ScryChoiceFor(g, me.ID); c != nil {
		t.Error("an unbargained Torch scried")
	}
	effectstest.P1Destroy(t, g, big)
	effectstest.P1WantZone(t, g, big, game.ZoneExile, "the 3/3 Torch dealt damage, destroyed later")

	shielded := effectstest.P1Creature(g, opp.ID, 3, 3)
	effectstest.P1ShieldCreature(g, shielded)
	effectstest.CastCatalogSpell(t, g, "Torch the Tower", "Instant", effectstest.P1g2TorchOracle, effectstest.Pr6Card(shielded))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, shielded)
	effectstest.P1WantZone(t, g, shielded, game.ZoneGraveyard, "a creature Torch dealt no damage")

	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure",
		Owner: me.ID, Controller: me.ID,
	})
	target := effectstest.P1Creature(g, opp.ID, 3, 3)
	effectstest.P1g2Cast(t, g, "Torch the Tower", "Instant", effectstest.P1g2TorchOracle, "", game.CastSpellParams{
		Targets: effectstest.Pr6Card(target), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("a bargained Torch did not scry")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, target, game.ZoneExile, "the 3/3 a bargained Torch killed")

	walker := effectstest.P1g2Planeswalker(g, opp.ID, 2)
	effectstest.CastCatalogSpell(t, g, "Torch the Tower", "Instant", effectstest.P1g2TorchOracle, effectstest.Pr6Card(walker))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, walker, game.ZoneExile, "the planeswalker Torch killed")
}

// Spikefield Hazard: "a permanent dealt damage this way" — a planeswalker
// too — and a player is just dealt 1.
func TestP1G2SpikefieldHazardMarksAnyPermanentItDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.P1g2Seats(g)
	walker := effectstest.P1g2Planeswalker(g, opp.ID, 1)
	effectstest.CastCatalogSpell(t, g, "Spikefield Hazard", "Instant", effectstest.P1g2SpikefieldOracle, effectstest.Pr6Card(walker))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, walker, game.ZoneExile, "the planeswalker Spikefield Hazard killed")

	shielded := effectstest.P1Creature(g, opp.ID, 1, 1)
	effectstest.P1ShieldCreature(g, shielded)
	effectstest.CastCatalogSpell(t, g, "Spikefield Hazard", "Instant", effectstest.P1g2SpikefieldOracle, effectstest.Pr6Card(shielded))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, shielded)
	effectstest.P1WantZone(t, g, shielded, game.ZoneGraveyard, "a creature Spikefield Hazard dealt no damage")

	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Spikefield Hazard", "Instant", effectstest.P1g2SpikefieldOracle, effectstest.Pr6Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("player's life %d → %d, want -1", before, opp.Life)
	}
}
