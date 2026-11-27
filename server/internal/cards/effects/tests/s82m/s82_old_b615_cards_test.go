package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o615IvoryMask     = "028bc726-0e86-4c53-9648-625153ae4b01"
	o615Choke         = "057fa60b-10b0-4612-be0d-157076c82241"
	o615WarpedDevo    = "22c49415-e8d2-4b4c-b021-57ee54670f86"
	o615VernalBloom   = "238b270c-152a-48e5-a955-65f5cb80c174"
	o615Cowardice     = "2692dd66-c200-4d0b-9427-30deaa2342ab"
	o615RollingStones = "7ce533d0-8b6e-46d8-b21e-a2932bdffb67"
	o615DeathPits     = "aead53ef-ffcf-4ae2-ab81-40aa9627adc9"
	o615SacredGround  = "b18e773c-611b-4c18-8b2d-8d3a7e5ddc93"

	o615StoneRain = "6e880df6-fc00-43d2-a9c8-f575f40b78c6"
	o615Unsummon  = "837182db-1bf3-4a2c-bd01-1af9d9873561"
	o615Vendetta  = "50a16634-ebf7-477f-ad6f-23332599f838"
)

func o615Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func TestS82OldB615EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		o615IvoryMask: "Ivory Mask", o615Choke: "Choke", o615WarpedDevo: "Warped Devotion",
		o615VernalBloom: "Vernal Bloom", o615Cowardice: "Cowardice", o615RollingStones: "Rolling Stones",
		o615DeathPits: "Death Pits of Rath", o615SacredGround: "Sacred Ground",
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

func TestS82OldB615IvoryMaskGivesItsControllerShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushCatalogPermanent(g, opp.ID, "Ivory Mask", "Enchantment", o615IvoryMask, false)
	bolt := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := effectstest.CastCatalogSpellErr(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070", bolt); err == nil {
		t.Error("a player with shroud was targeted")
	}
	own := []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}
	if err := effectstest.CastCatalogSpellErr(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070", own); err != nil {
		t.Errorf("a player without shroud could not be targeted: %v", err)
	}
}

func TestS82OldB615RollingStonesLetsWallsAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall", Keywords: []string{"defender"},
		Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wall, opp.ID); err == nil {
		t.Fatal("a Wall attacked without Rolling Stones")
	}
	g2 := effectstest.NewCatalogGame(t)
	me, opp = g2.Seats[0], g2.Seats[1]
	effectstest.PushCatalogPermanent(g2, opp.ID, "Rolling Stones", "Enchantment", o615RollingStones, false)
	wall = effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall", Keywords: []string{"defender"},
		Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	other := effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Sentry", TypeLine: "Creature — Soldier", Keywords: []string{"defender"},
		Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(wall, opp.ID); err != nil {
		t.Errorf("a Wall with defender could not attack under Rolling Stones: %v", err)
	}
	if err := g2.DeclareAttacker(other, opp.ID); err == nil {
		t.Error("a non-Wall with defender attacked")
	}
}

func TestS82OldB615ChokeKeepsIslandsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0]
	effectstest.PushPermanentForTest(g, g.Seats[1].ID, "Choke", o615Choke, "Enchantment")
	island := effectstest.PushPermanentForTest(g, owner.ID, "Island", "", "Basic Land — Island")
	dual := effectstest.PushPermanentForTest(g, owner.ID, "Steam Vents", "", "Land — Island Mountain")
	forest := effectstest.PushPermanentForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	for _, id := range []uuid.UUID{island, dual, forest} {
		if err := g.TapCard(id, true); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.UntapStaticTapped(t, g, island) || !effectstest.UntapStaticTapped(t, g, dual) {
		t.Error("an Island untapped under Choke")
	}
	if effectstest.UntapStaticTapped(t, g, forest) {
		t.Error("a Forest was kept tapped")
	}
}

func TestS82OldB615VernalBloomAddsGToTheForestsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Vernal Bloom", "Enchantment", o615VernalBloom, false)
	mine := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Permanent(g, them.ID, "Forest", "Basic Land — Forest")
	island := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.TapForMana(t, g, me.ID, mine)
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("my pool = %v, want two green", got)
	}
	effectstest.TapForMana(t, g, them.ID, theirs)
	if got := effectstest.SortedPool(them); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("their pool = %v, want two green: 'its controller', not 'you'", got)
	}
	me.ManaPool = nil
	effectstest.TapForMana(t, g, me.ID, island)
	if got := effectstest.SortedPool(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("an Island's pool = %v, want only the one blue", got)
	}
}

func TestS82OldB615CowardiceReturnsATargetedCreatureAndTheSpellFizzles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushCatalogPermanent(g, g.Seats[g.Turn.ActiveSeat].ID, "Cowardice", "Enchantment", o615Cowardice, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Vendetta", "Instant", o615Vendetta, o615Target(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the targeted creature was not returned to its owner's hand")
	}
	if opp.Graveyard.Contains(bear) {
		t.Error("the spell still destroyed the creature")
	}
}

func TestS82OldB615DeathPitsDestroysAnyCreatureDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Death Pits of Rath", "Enchantment", o615DeathPits, false)
	src := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 1)
	big := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 6)
	effectstest.DealDamageTo(t, g, src, big, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a creature dealt 1 damage survived Death Pits")
	}
	if !opp.Graveyard.Contains(big) {
		t.Error("the creature is not in its owner's graveyard")
	}
	// A noncreature permanent is not touched, and damage to a player is not damage to a creature.
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.DealCombatDamageToPlayer(g, src, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) || !g.Battlefield.Contains(src) {
		t.Error("damage to a player destroyed something")
	}
}

func TestS82OldB615SacredGroundReturnsALandAnOpponentDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, owner := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushCatalogPermanent(g, owner.ID, "Sacred Ground", "Enchantment", o615SacredGround, false)
	land := effectstest.B12Permanent(g, owner.ID, "Forest", "Basic Land — Forest")
	_ = caster
	effectstest.CastCatalogSpell(t, g, "Stone Rain", "Sorcery", o615StoneRain, o615Target(land))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Fatalf("the land did not return to the battlefield (graveyard %v)", owner.Graveyard.Contains(land))
	}
}

func TestS82OldB615SacredGroundIgnoresYourOwnEffectsAndOtherCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushCatalogPermanent(g, me.ID, "Sacred Ground", "Enchantment", o615SacredGround, false)
	// Your own spell destroys your own land: not an opponent's.
	mine := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Stone Rain", "Sorcery", o615StoneRain, o615Target(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("Sacred Ground returned a land of mine that my own spell destroyed")
	}
	// An opponent's land goes to the opponent's graveyard, not mine.
	theirs := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Stone Rain", "Sorcery", o615StoneRain, o615Target(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("Sacred Ground returned a land that was not put into my graveyard")
	}
}

func TestS82OldB615WarpedDevotionMakesTheHandOwnerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushCatalogPermanent(g, g.Seats[g.Turn.ActiveSeat].ID, "Warped Devotion", "Enchantment", o615WarpedDevo, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	before := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Unsummon", "Instant", o615Unsummon, o615Target(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
		t.Fatal("the creature's owner was not asked to discard")
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// The Bear came back (+1) and one card was discarded (-1).
	if got := opp.Hand.Size(); got != before {
		t.Errorf("hand size %d, want %d (one returned, one discarded)", got, before)
	}
	if !opp.Hand.Contains(bear) && !opp.Graveyard.Contains(bear) {
		t.Error("the Bear is neither in hand nor in the graveyard")
	}
}
