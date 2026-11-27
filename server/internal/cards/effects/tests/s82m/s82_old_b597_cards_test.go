package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b597_cards_test.go — Betrayers of Kamigawa, damage, batch 597
// (#5602, S82 OLD). Every test plays the real catalog entry.

const (
	oldB597AuraBarbs   = "1aa8fdc3-efeb-450d-9540-2ef0efaf270d"
	oldB597IshiIshi    = "30265fce-28e6-4994-83f7-0e4c4b22df35"
	oldB597Frostling   = "6f6ac768-4c97-44db-933a-f2d442e7f665"
	oldB597Gatekeeper  = "7792d3af-5c64-4d0a-a2b3-b903d787b0ab"
	oldB597Clash       = "7ab18e48-f935-405b-8658-b49052d20529"
	oldB597MatsuSniper = "c20728ac-c01c-470f-a43b-de6104836645"
)

func TestOldB597CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB597AuraBarbs: "Aura Barbs", oldB597IshiIshi: "Ishi-Ishi, Akki Crackshot", oldB597Frostling: "Frostling",
		oldB597Gatekeeper: "Shinka Gatekeeper", oldB597Clash: "Clash of Realities", oldB597MatsuSniper: "Matsu-Tribe Sniper",
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

func oldB597Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestOldB597FrostlingSacrificesToPingACreature(t *testing.T) {
	g, me, opp := oldB597Table(t)
	frost := effectstest.PushCatalogPermanent(g, me.ID, "Frostling", "Creature — Spirit", oldB597Frostling, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, frost, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(bear)})
	if g.Battlefield.Contains(frost) {
		t.Error("the Frostling was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if n := effectstest.DamageMarkedOn(g, bear); n != 1 {
		t.Errorf("the target has %d damage, want 1", n)
	}
}

func TestOldB597ShinkaGatekeeperReflectsTheDamageToItsController(t *testing.T) {
	g, me, opp := oldB597Table(t)
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Shinka Gatekeeper", "Creature — Ogre Warrior", oldB597Gatekeeper, false)
	source := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.DealDamageTo(t, g, source, keeper, 2)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if me.Life != life-2 {
		t.Errorf("life %d, want %d (2 damage reflected)", me.Life, life-2)
	}
}

func TestOldB597IshiIshiPunishesOnlyAnOpponentsSpiritOrArcaneSpell(t *testing.T) {
	g, me, opp := oldB597Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Ishi-Ishi, Akki Crackshot", "Legendary Creature — Goblin Warrior", oldB597IshiIshi, false)
	life := opp.Life
	// My own Arcane spell does nothing.
	effectstest.CastCatalogSpell(t, g, "My Arcane Spell", "Instant — Arcane", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Fatalf("my own spell dealt damage: opponent %d (was %d)", opp.Life, life)
	}
	// An opponent's Bear does nothing; their Spirit and Arcane spells do.
	oldB597OpponentCasts(t, g, opp, "Their Trick", "Instant — Trap")
	if opp.Life != life {
		t.Fatalf("a non-Spirit, non-Arcane spell dealt damage: %d, was %d", opp.Life, life)
	}
	oldB597OpponentCasts(t, g, opp, "Their Spirit", "Instant — Spirit")
	if opp.Life != life-2 {
		t.Fatalf("a Spirit spell left the opponent at %d, want %d", opp.Life, life-2)
	}
	oldB597OpponentCasts(t, g, opp, "Their Arcane", "Instant — Arcane")
	if opp.Life != life-4 {
		t.Errorf("an Arcane spell left the opponent at %d, want %d", opp.Life, life-4)
	}
}

// oldB597OpponentCasts has the player cast a card with priority, then settles
// the trigger and the spell.
func oldB597OpponentCasts(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
}

func TestOldB597MatsuTribeSniperPingsAFlyerAndFreezesIt(t *testing.T) {
	g, me, opp := oldB597Table(t)
	sniper := effectstest.PushCatalogPermanent(g, me.ID, "Matsu-Tribe Sniper", "Creature — Snake Warrior Archer", oldB597MatsuSniper, false)
	ground := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bird", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 3, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"flying"}})
	if err := effectstest.ActivateOn(t, g, me.ID, sniper, 0, ground); err == nil {
		t.Fatal("a creature without flying was a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, sniper, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(bird)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if n := effectstest.DamageMarkedOn(g, bird); n != 1 {
		t.Errorf("the flyer has %d damage, want 1", n)
	}
	if !effectstest.IsTapped(g, bird) {
		t.Error("the damaged creature was not tapped")
	}
	if len(effectstest.HmlUntapSkips(t, g, bird)) == 0 {
		t.Error("the damaged creature does not skip its next untap step")
	}
}

func TestOldB597AuraBarbsHitsEnchantmentsThenEnchantedCreatures(t *testing.T) {
	g, me, opp := oldB597Table(t)
	effectstest.B12Permanent(g, me.ID, "My Shrine", "Enchantment")
	effectstest.B12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	effectstest.B12Permanent(g, opp.ID, "Their Other Shrine", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Pacifism", effectstest.AuraTypeLine, "", false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	myLife, theirLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Aura Barbs", "Instant — Arcane", oldB597AuraBarbs, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	// I control a Shrine and the Aura; the opponent controls two Shrines.
	if me.Life != myLife-4 {
		t.Errorf("my life %d, want %d (two enchantments)", me.Life, myLife-4)
	}
	if opp.Life != theirLife-4 {
		t.Errorf("their life %d, want %d (two enchantments)", opp.Life, theirLife-4)
	}
	if n := effectstest.DamageMarkedOn(g, bear); n != 2 {
		t.Errorf("the enchanted creature has %d damage, want 2", n)
	}
}

func TestOldB597ClashOfRealitiesSpiritsHitNonSpiritsAndBackAgain(t *testing.T) {
	g, me, opp := oldB597Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Clash of Realities", "Enchantment", oldB597Clash, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	ghost := effectstest.B12Creature(g, opp.ID, "Their Ghost", "Creature — Spirit", 5, 5)

	// A Spirit entering may deal 3 to a non-Spirit creature only.
	effectstest.CastCatalogSpell(t, g, "Newcomer Spirit", "Creature — Spirit", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("an entering Spirit got no target prompt")
	}
	if !effectstest.HasID(p.PickTargetCards, bear) || effectstest.HasID(p.PickTargetCards, ghost) {
		t.Fatalf("the Spirit's trigger offers %v, want only the non-Spirit %v", p.PickTargetCards, bear)
	}
	effectstest.PickIfAsked(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if n := effectstest.DamageMarkedOn(g, bear); n != 3 {
		t.Errorf("the non-Spirit has %d damage, want 3", n)
	}

	// A non-Spirit creature entering may deal 3 to a Spirit creature only.
	effectstest.CastCatalogSpell(t, g, "Newcomer Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	p = effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("an entering non-Spirit got no target prompt")
	}
	if effectstest.HasID(p.PickTargetCards, bear) || !effectstest.HasID(p.PickTargetCards, ghost) {
		t.Fatalf("the non-Spirit's trigger offers %v, want only the Spirit %v", p.PickTargetCards, ghost)
	}
	effectstest.PickIfAsked(t, g, me.ID, ghost)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if n := effectstest.DamageMarkedOn(g, ghost); n != 0 {
		t.Errorf("declining still dealt %d damage", n)
	}
}
