package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_boundary_cards_test.go — #1009, the card side.
//
// Every "this turn" reader in the catalog now asks Game.TurnTally.
// The four that were still walking the event log back to
// EventBeginUpkeep are proved here, each with the case the walk got
// wrong: something that happened this turn BEFORE the upkeep event.
//
// An upkeep event that is not the first thing in the turn is
// reachable two ways — a permanent entering, an attack being
// declared or a Faerie connecting during the UNTAP step (#70,
// ADR 0070 makes untap-step choices and triggers real), and an EXTRA
// upkeep step (CR 500.8, ADR 0059 Decision 4). The tests spell it the
// second way because it needs no untap-step machinery: they do the
// thing, then put an upkeep event into the log behind it, which is
// precisely the shape the old walk answered "no" to.

// emitUpkeep puts an upkeep-began event into the log — the event the
// migrated walks used to stop at.
func emitUpkeep(g *game.Game, active uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: active})
	})
}

// Oran-Rief, the Vastwood: "{T}: Put a +1/+1 counter on each green
// creature that entered this turn." The creature entered, an upkeep
// event landed behind it, and it still entered this turn.
func TestB06OranRiefGrowsACreatureThatEnteredBeforeTheUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := seedPermanentWithOracle(g, me.ID, "Oran-Rief, the Vastwood", "Land", b06OranRiefOracle)
	old := seedCreature(g, "Old Bear", me.ID) // no ETB event at all
	fresh := b06GreenCreature(t, g, me.ID, "Fresh Elf")

	emitUpkeep(g, me.ID)

	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if counterCount(g, fresh, "+1/+1") != 1 {
		t.Error("the upkeep is not the turn boundary: a creature that entered before it entered this turn")
	}
	if counterCount(g, old, "+1/+1") != 0 {
		t.Error("a creature that never entered this turn must not grow")
	}
}

// Éowyn, Shieldmaiden's "another Human entered this turn" reads the
// subtype tally for the count and the per-object cell to decide
// whether Éowyn herself is one of the entries it counted. With an
// upkeep event behind her own entry, the old walk said she was not,
// and one Human — her — satisfied "another".
func TestB43EowynCountsHerOwnEntryFromBeforeTheUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eowyn := b43Catalog(g, me.ID, "Éowyn, Shieldmaiden", "Legendary Creature — Human Knight", b43EowynOracle, 5, 4)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: eowyn})
	})
	emitUpkeep(g, me.ID)

	source, ok := g.LookupCardForEffect(eowyn)
	if !ok {
		t.Fatal("setup: Éowyn is nowhere")
	}
	if b43AnotherHumanEnteredThisTurn(g, &source) {
		t.Error("Éowyn is the only Human that entered this turn, so there is no ANOTHER")
	}

	other := b12Creature(g, me.ID, "Human Soldier", "Creature — Human Soldier", 1, 1)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: other})
	})
	if !b43AnotherHumanEnteredThisTurn(g, &source) {
		t.Error("a second Human entered, so there is an ANOTHER")
	}
}

// Chart a Course's "unless you attacked this turn" is the
// AttacksDeclared cell. An attack declared before an upkeep event is
// still an attack this turn.
func TestB18AttackedThisTurnSeesAnAttackBeforeTheUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if b18AttackedThisTurn(g, me.ID) {
		t.Fatal("setup: nobody has attacked yet")
	}
	attackWith(t, g, opp.ID, bear)
	emitUpkeep(g, me.ID)

	if !b18AttackedThisTurn(g, me.ID) {
		t.Error("the upkeep is not the turn boundary: the attack was declared this turn")
	}
	if b18AttackedThisTurn(g, opp.ID) {
		t.Error("the defender did not attack")
	}
}

// Alela, Cunning Conqueror — the #596 half. A Faerie ROGUE TOKEN
// connects and then ceases to exist (CR 704.5d takes a dead token out
// of the graveyard), and the goad's target predicate is asked
// afterwards, because a target predicate is not handed the trigger's
// event. The set is recorded at the damage, so the player it hit is
// still in it.
func TestB33AlelaGoadReachesThePlayerAVanishedFaerieTokenHit(t *testing.T) {
	g := newCatalogGame(t)
	me, victim, other := g.Seats[0], g.Seats[1], g.Seats[2]
	var tok uuid.UUID
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(me.ID, game.Card{
			Name: "Faerie Rogue", TypeLine: "Token Creature — Faerie Rogue", Power: 1, Toughness: 1,
		}, 1, game.TokenEntryOptions{})
		if err != nil {
			t.Fatalf("create token: %v", err)
		}
		tok = ids[0]
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: me.ID, Source: tok,
			Target: victim.ID, Amount: 1, Combat: true,
		})
	})
	theirs := b12Creature(g, victim.ID, "Their Bear", "Creature — Bear", 2, 2)
	otherBear := b12Creature(g, other.ID, "Other Bear", "Creature — Bear", 2, 2)

	// The Faerie trades and is gone from every zone.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tok {
				g.Battlefield.Cards = append(g.Battlefield.Cards[:i], g.Battlefield.Cards[i+1:]...)
				break
			}
		}
	})
	if _, ok := g.LookupCardForEffect(tok); ok {
		t.Fatal("setup: the token should be nowhere")
	}

	hit, ok := g.LookupCardForEffect(theirs)
	if !ok {
		t.Fatal("setup: the victim's creature is nowhere")
	}
	if !b33CreatureOfPlayerHitByYourFaeries(g, me.ID, hit) {
		t.Error("a Faerie token that traded still hit that player: the goad must reach their creatures")
	}
	safe, ok := g.LookupCardForEffect(otherBear)
	if !ok {
		t.Fatal("setup: the third player's creature is nowhere")
	}
	if b33CreatureOfPlayerHitByYourFaeries(g, me.ID, safe) {
		t.Error("no Faerie of mine hit the third player")
	}
}

// Trygon Predator reads the same record by NAME. A Predator that died
// in the combat it connected in still names the player it hit.
func TestB32TrygonPredatorNamesThePlayerItHitAfterItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	pred := b12Push(g, me.ID, "Trygon Predator", "Creature — Beast", b32TrygonPredatorOracleForTest, 2, 3)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: me.ID, Source: pred,
			Target: victim.ID, Amount: 2, Combat: true,
		})
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == pred {
				g.Battlefield.Cards = append(g.Battlefield.Cards[:i], g.Battlefield.Cards[i+1:]...)
				break
			}
		}
	})
	rock := b12Permanent(g, victim.ID, "Their Signet", "Artifact")
	c, ok := g.LookupCardForEffect(rock)
	if !ok {
		t.Fatal("setup: the artifact is nowhere")
	}
	if !b32ArtifactOrEnchantmentOfPlayerHitByYourTrygonPredator(g, me.ID, c) {
		t.Error("the Predator hit that player this turn, wherever the Predator is now")
	}
}

const b32TrygonPredatorOracleForTest = "c744b5f4-fbcf-48b8-9d60-5e9c6ac297e0"
