package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b220_cards_test.go — S82 OLD batch 220 (#5531, Rivals of Ixalan:
// graveyard recursion, token makers, equipment, instants and sorceries,
// artifacts). Each test plays the real catalog entry.

const (
	oldB220JungleCreeper = "acc091c0-3891-41b9-b5f6-7537fe9493fe"
	oldB220Recover       = "f4361602-63da-430b-adbb-36d74e987aa6"
	oldB220BafflingEnd   = "03c35670-84ac-4825-b395-cefb76ceb89a"
	oldB220CaptainsHook  = "25143290-2dd5-4def-a857-a1db6d3ed2eb"
	oldB220SphinxsDecree = "7d615b37-8dc6-4142-bf9a-9815dd99025a"
	oldB220Amalgam       = "909d7e77-ad7c-4af2-88db-f22442cfa531"
	oldB220Bear          = "Creature — Bear"
)

func TestOldB220EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB220JungleCreeper: "Jungle Creeper",
		oldB220Recover:       "Recover",
		oldB220BafflingEnd:   "Baffling End",
		oldB220CaptainsHook:  "Captain's Hook",
		oldB220SphinxsDecree: "Sphinx's Decree",
		oldB220Amalgam:       "Awakened Amalgam",
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

func oldB220Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB220HasSubtype(t *testing.T, g *game.Game, id uuid.UUID, subtype string) bool {
	t.Helper()
	c := effectstest.LayeredCard(t, g, id)
	return c.HasSubtypeWord(subtype)
}

func oldB220Graveyard(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Jungle Creeper and Recover -------------------------------------------------

func TestOldB220JungleCreeperReturnsItselfFromTheGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB220Seats(g)
	id := oldB220Graveyard(me, "Jungle Creeper", "Creature — Elemental", oldB220JungleCreeper)
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{B}{G}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("paid {3}{B}{G} with four mana")
	}
	effectstest.AddPool(t, g, me.ID, "{C}")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("the Creeper did not return to its owner's hand")
	}
	// From the battlefield the ability does nothing: it only works in a graveyard.
	bf := effectstest.B12Push(g, me.ID, "Jungle Creeper", "Creature — Elemental", oldB220JungleCreeper, 3, 3)
	effectstest.AddPool(t, g, me.ID, "{B}{G}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bf, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the graveyard ability was activated from the battlefield")
	}
}

func TestOldB220RecoverReturnsACreatureCardAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB220Seats(g)
	creature := oldB220Graveyard(me, "Dead Bear", oldB220Bear, "")
	land := oldB220Graveyard(me, "Forest", "Basic Land — Forest", "")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Recover", "Sorcery", oldB220Recover, effectstest.CardRefs(land)); err == nil {
		t.Fatal("a land card was a legal target")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Recover", "Sorcery", oldB220Recover, effectstest.CardRefs(creature))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(creature) {
		t.Error("the creature card did not return to hand")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d (the creature and a drawn card)", got, hand+2)
	}
}

func TestOldB220RecoverFizzlesWithoutItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB220Seats(g)
	creature := oldB220Graveyard(me, "Dead Bear", oldB220Bear, "")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Recover", "Sorcery", oldB220Recover, effectstest.CardRefs(creature))
	g.WithWriteLock(func() { me.Graveyard.Remove(creature) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d: a spell whose only target is gone does nothing (CR 608.2b)", got, hand)
	}
}

// --- Baffling End ---------------------------------------------------------------

func TestOldB220BafflingEndExilesACheapCreatureAndGivesADinosaurWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB220Seats(g)
	cheap := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cheap", TypeLine: oldB220Bear, ManaCost: "{2}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	dear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dear", TypeLine: oldB220Bear, ManaCost: "{2}{G}{G}",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mine", TypeLine: oldB220Bear, ManaCost: "{1}",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	end := effectstest.CastCatalogSpell(t, g, "Baffling End", "Enchantment", oldB220BafflingEnd, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); c != nil {
		for _, bad := range []uuid.UUID{dear, mine} {
			for _, offered := range c.PickTargetCards {
				if offered == bad {
					t.Errorf("a creature that is not an opposing creature with mana value 3 or less was offered")
				}
			}
		}
	}
	effectstest.AnswerPickTargetIfOpen(t, g, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, opp, cheap); z != "exile" {
		t.Errorf("the cheap creature is in %s, want exile", z)
	}
	if !g.Battlefield.Contains(dear) || !g.Battlefield.Contains(mine) {
		t.Error("a creature that was not the target was exiled")
	}
	// Destroying the enchantment: the targeted opponent creates a 3/3 Dinosaur and
	// the exiled creature stays in exile.
	effectstest.P1Destroy(t, g, end)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	dinos := effectstest.S46Tokens(g, opp.ID, "Dinosaur")
	if len(dinos) != 1 || dinos[0].Power != 3 || dinos[0].Toughness != 3 || !game.HasKeyword(&dinos[0], "trample") {
		t.Errorf("tokens %+v, want one 3/3 green Dinosaur with trample for the opponent", dinos)
	}
	if len(effectstest.S46Tokens(g, me.ID, "Dinosaur")) != 0 {
		t.Error("the controller got a Dinosaur")
	}
	if z, _ := effectstest.CardWhere(g, opp, cheap); z != "exile" {
		t.Errorf("the exiled creature came back (%s)", z)
	}
}

// --- Captain's Hook ---------------------------------------------------------------

func TestOldB220CaptainsHookGrantsPlusTwoMenaceAndPirate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB220Seats(g)
	effectstest.AdvanceToMain(t, g)
	hook := effectstest.SeedEquipment(g, me.ID, "Captain's Hook", oldB220CaptainsHook)
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.AddPool(t, g, me.ID, "{C}")
	effectstest.EquipTo(t, g, me.ID, hook, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("equipped Bear = %d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "menace") {
		t.Error("the equipped creature lacks menace")
	}
	if !oldB220HasSubtype(t, g, bear, "Pirate") {
		t.Error("the equipped creature is not a Pirate")
	}
	if !oldB220HasSubtype(t, g, bear, "Bear") {
		t.Error("the equipped creature lost its other creature type")
	}
}

func TestOldB220CaptainsHookDestroysThePermanentItMovesOffOfAndItsLeavingHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB220Seats(g)
	effectstest.AdvanceToMain(t, g)
	hook := effectstest.SeedEquipment(g, me.ID, "Captain's Hook", oldB220CaptainsHook)
	first, second := effectstest.SeedBear(g, me.ID), effectstest.SeedBear(g, me.ID)
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, hook, first)
	effectstest.EquipTo(t, g, me.ID, hook, second)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, first) {
		t.Error("the creature the Equipment moved off of was not destroyed")
	}
	if !effectstest.OnBattlefieldByID(g, second) {
		t.Error("the new host was destroyed")
	}
	// The Equipment leaving while attached also unattaches it.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, hook); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, second) {
		t.Error("the host survived the Equipment leaving")
	}
}

// --- Sphinx's Decree ----------------------------------------------------------------

func oldB220ToTurnOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == seat && g.Turn.Step == game.StepPrecombatMain); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != seat || g.Turn.Step != game.StepPrecombatMain {
		t.Fatalf("never reached seat %d's main phase", seat)
	}
}

func oldB220TryCast(g *game.Game, p *game.Player, name, typeLine, cost string) error {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true})
}

func TestOldB220SphinxsDecreeBansInstantsAndSorceriesOnlyInEachOpponentsNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB220Seats(g)
	meSeat := g.Turn.ActiveSeat
	effectstest.CastCatalogSpell(t, g, "Sphinx's Decree", "Sorcery", oldB220SphinxsDecree, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Still my turn: the opponent may cast an instant.
	effectstest.AddPool(t, g, opp.ID, "{R}")
	if err := oldB220TryCast(g, opp, "Opt", "Instant", "{R}"); err != nil {
		t.Errorf("an opponent could not cast an instant before their next turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, off := range []int{1, 2, 3} {
		seat := (meSeat + off) % len(g.Seats)
		p := g.Seats[seat]
		oldB220ToTurnOf(t, g, seat)
		effectstest.AddPool(t, g, p.ID, "{R}{R}{R}")
		if err := oldB220TryCast(g, p, "Shock", "Instant", "{R}"); err == nil {
			t.Errorf("seat %d cast an instant during their next turn", seat)
		}
		if err := oldB220TryCast(g, p, "Divination", "Sorcery", "{R}"); err == nil {
			t.Errorf("seat %d cast a sorcery during their next turn", seat)
		}
		if err := oldB220TryCast(g, p, "Bear", "Creature — Bear", "{R}"); err != nil {
			t.Errorf("seat %d: a creature spell was barred: %v", seat, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	// My own next turn: no ban on me.
	oldB220ToTurnOf(t, g, meSeat)
	effectstest.AddPool(t, g, g.Seats[meSeat].ID, "{R}")
	if err := oldB220TryCast(g, g.Seats[meSeat], "Shock", "Instant", "{R}"); err != nil {
		t.Errorf("the caster was banned from instants: %v", err)
	}
}

func TestOldB220SphinxsDecreeEndsWithThatTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB220Seats(g)
	meSeat := g.Turn.ActiveSeat
	effectstest.CastCatalogSpell(t, g, "Sphinx's Decree", "Sorcery", oldB220SphinxsDecree, nil)
	effectstest.PassPriorityAroundTable(t, g)
	oppSeat := (meSeat + 1) % len(g.Seats)
	for i := 0; i < 5; i++ {
		oldB220ToTurnOf(t, g, (oppSeat+i)%len(g.Seats))
	}
	oldB220ToTurnOf(t, g, oppSeat)
	effectstest.AddPool(t, g, opp.ID, "{R}")
	if err := oldB220TryCast(g, opp, "Shock", "Instant", "{R}"); err != nil {
		t.Errorf("the ban outlived their next turn: %v", err)
	}
}

// --- Awakened Amalgam ---------------------------------------------------------------

func TestOldB220AwakenedAmalgamCountsDifferentlyNamedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB220Seats(g)
	amalgam := effectstest.B12Push(g, me.ID, "Awakened Amalgam", "Artifact Creature — Golem", oldB220Amalgam, 0, 0)
	if p, tough := effectstest.EffectivePTOf(t, g, amalgam); p != 0 || tough != 0 {
		t.Fatalf("with no lands it is %d/%d, want 0/0", p, tough)
	}

	for _, name := range []string{"Forest", "Forest", "Island", "Swamp"} {
		effectstest.PushCatalogPermanent(g, me.ID, name, "Basic Land — "+name, "", false)
	}
	effectstest.PushCatalogPermanent(g, opp.ID, "Mountain", "Basic Land — Mountain", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Not a land", "Artifact", "", false)
	if p, tough := effectstest.EffectivePTOf(t, g, amalgam); p != 3 || tough != 3 {
		t.Errorf("with Forest, Forest, Island, Swamp it is %d/%d, want 3/3 (three names, mine only)", p, tough)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	if p, tough := effectstest.EffectivePTOf(t, g, amalgam); p != 4 || tough != 4 {
		t.Errorf("after a Plains it is %d/%d, want 4/4", p, tough)
	}
}
