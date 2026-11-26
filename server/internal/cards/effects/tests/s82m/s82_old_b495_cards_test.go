package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b495_cards_test.go — Zendikar life gain and loss and removal, S82 OLD
// batch 495 (#5583), each driven through the engine.

const (
	zenGuulDraz   = "00d3fcdf-6959-4437-9bad-93a8625510a1"
	zenHagra      = "5e2c1e0e-0a10-416a-9b50-96ee0cbbc24e"
	zenLandbind   = "63500b81-8fac-4208-91d3-0fe642899d87"
	zenNeedlebite = "c8fedb0e-27d5-4fae-8058-3e9666a9661d"
	zenLacerator  = "dc261758-0982-4b7d-9986-042c03cc8366"
	zenPitfall    = "26e75f0b-3a02-4989-8429-b6276e98ac33"
	zenHaloHunter = "293560ec-accf-49c6-9bf4-6d37f56a4eb7"
	zenDesecrated = "3c9242fd-915b-4f6d-b4ee-830c9d80c9e8"
	zenLightcast  = "a6f14387-4dfd-4d6b-a6b1-2ae658a70de6"
	zenRavenous   = "e0bea449-7601-48ee-a522-f1bbe6905f6d"
)

func TestS82OldB495EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		zenGuulDraz: "Guul Draz Vampire", zenHagra: "Hagra Diabolist", zenLandbind: "Landbind Ritual",
		zenNeedlebite: "Needlebite Trap", zenLacerator: "Vampire Lacerator", zenPitfall: "Pitfall Trap",
		zenHaloHunter: "Halo Hunter", zenDesecrated: "Desecrated Earth", zenLightcast: "Devout Lightcaster",
		zenRavenous: "Ravenous Trap",
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

func zenSetLife(g *game.Game, p *game.Player, life int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, life-p.Life) })
}

func TestS82OldB495GuulDrazVampireGrowsAndGainsIntimidateAgainstALowOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	v := effectstest.PushCatalogPermanent(g, me.ID, "Guul Draz Vampire", "Creature — Vampire Rogue", zenGuulDraz, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == v {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 1, 1
			}
		}
	})
	zenSetLife(g, opp, 11)
	if p, tough := effectstest.PtOf(t, g, v); p != 1 || tough != 1 {
		t.Fatalf("at 11 life the Vampire is %d/%d, want 1/1", p, tough)
	}
	if effectstest.P1g2HasKeyword(g, v, "intimidate") {
		t.Error("intimidate at 11 life")
	}
	zenSetLife(g, opp, 10)
	if p, tough := effectstest.PtOf(t, g, v); p != 3 || tough != 2 {
		t.Errorf("at 10 life the Vampire is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.P1g2HasKeyword(g, v, "intimidate") {
		t.Error("no intimidate at 10 life")
	}
	// Its controller's own life does not matter.
	zenSetLife(g, opp, 40)
	zenSetLife(g, me, 3)
	if p, _ := effectstest.PtOf(t, g, v); p != 1 {
		t.Errorf("the controller's low life pumped it: power %d", p)
	}
}

func TestS82OldB495VampireLaceratorCostsLifeUnlessAnOpponentIsLow(t *testing.T) {
	run := func(t *testing.T, oppLife int) (before, after int) {
		g := effectstest.NewCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
		effectstest.PushCatalogPermanent(g, me.ID, "Vampire Lacerator", "Creature — Vampire Warrior", zenLacerator, false)
		zenSetLife(g, opp, oppLife)
		for _, p := range g.Seats {
			if p.ID != me.ID && p.ID != opp.ID {
				zenSetLife(g, p, 40)
			}
		}
		before = me.Life
		effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
		effectstest.PassPriorityAroundTable(t, g)
		return before, me.Life
	}
	if b, a := run(t, 20); a != b-1 {
		t.Errorf("no low opponent: life %d -> %d, want -1", b, a)
	}
	if b, a := run(t, 10); a != b {
		t.Errorf("an opponent at 10: life %d -> %d, want no loss", b, a)
	}
}

func TestS82OldB495LandbindRitualGainsTwoPerPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.PushPermanentForTest(g, me.ID, "Plains", "", "Basic Land — Plains")
	}
	effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	effectstest.PushPermanentForTest(g, opp.ID, "Plains", "", "Basic Land — Plains")
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Landbind Ritual", "Sorcery", zenLandbind, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+6 {
		t.Errorf("life %d -> %d, want +6 for three Plains", before, me.Life)
	}
}

func TestS82OldB495HagraDiabolistDrainsByTheNumberOfAllies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Other Ally", "Creature — Human Ally", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Hagra Diabolist", TypeLine: "Creature — Ogre Shaman Ally", OracleID: zenHagra, Power: 3, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d -> %d, want -2 for two Allies", before, opp.Life)
	}
	// Another Ally entering triggers it again, for three; declining costs nothing.
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Third Ally", TypeLine: "Creature — Kor Ally", Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("declined, life is %d, want %d", opp.Life, before-2)
	}
}

func TestS82OldB495NeedlebiteTrapIsCheapOnlyAfterAnOpponentGainedLife(t *testing.T) {
	cast := func(t *testing.T, gained bool) (*game.Game, *game.Player, *game.Player, error) {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[0], g.Seats[1]
		if gained {
			g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 2) })
		}
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Needlebite Trap", TypeLine: "Instant — Trap", ManaCost: "{5}{B}{B}", OracleID: zenNeedlebite, Owner: me.ID, Controller: me.ID})
		effectstest.FloatForTest(g, me, "B")
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "trap", Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
		return g, me, opp, err
	}
	if _, _, _, err := cast(t, false); err == nil {
		t.Error("the trap cost was offered with no life gained")
	}
	g, me, opp, err := cast(t, true)
	if err != nil {
		t.Fatalf("the trap cost was refused after an opponent gained life: %v", err)
	}
	mine, theirs := me.Life, opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-5 || me.Life != mine+5 {
		t.Errorf("life: opponent %d -> %d, you %d -> %d; want -5 and +5", theirs, opp.Life, mine, me.Life)
	}
}

func TestS82OldB495PitfallTrapNeedsExactlyOneAttackerWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	atk, def := effectstest.SeatsForCombat(g)
	walker := effectstest.B12Creature(g, atk.ID, "Walker", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, atk.ID, "Second", "Creature — Bear", 2, 2)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: atk.ID, Controller: atk.ID})
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	def.Hand.PushTop(game.Card{InstanceID: id, Name: "Pitfall Trap", TypeLine: "Instant — Trap", ManaCost: "{2}{W}", OracleID: zenPitfall, Owner: def.ID, Controller: def.ID})
	cast := func(target uuid.UUID) error {
		effectstest.FloatForTest(g, def, "W")
		return g.CastSpell(def.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "trap", Targets: effectstest.CardTarget(target)})
	}
	if err := cast(walker); err == nil {
		t.Fatal("the trap cost was offered with nobody attacking")
	}
	g.WithWriteLock(func() { def.ManaPool = game.ManaPool{} })
	effectstest.DeclareAttack(t, g, def.ID, walker, flyer)
	g.WithWriteLock(func() { def.ManaPool = game.ManaPool{} })
	if err := cast(walker); err == nil {
		t.Fatal("the trap cost was offered with two attackers")
	}
	g.WithWriteLock(func() { def.ManaPool = game.ManaPool{} })
	_ = second
	// Two attackers: the full cost is not paid either, so only the target rule is under test.
	effectstest.FloatForTest(g, def, "CCW")
	if err := g.CastSpell(def.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(flyer)}); err == nil {
		t.Error("an attacking creature with flying was a legal target")
	}
	if err := g.CastSpell(def.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(second)}); err == nil {
		t.Error("a creature that is not attacking was a legal target")
	}
	if err := g.CastSpell(def.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(walker)}); err != nil {
		t.Fatalf("cast at the attacking non-flyer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the attacking creature without flying was not destroyed")
	}
}

func TestS82OldB495PitfallTrapPaysWWhenExactlyOneCreatureAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	atk, def := effectstest.SeatsForCombat(g)
	walker := effectstest.B12Creature(g, atk.ID, "Walker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, def.ID, walker)
	id := uuid.New()
	def.Hand.PushTop(game.Card{InstanceID: id, Name: "Pitfall Trap", TypeLine: "Instant — Trap", ManaCost: "{2}{W}", OracleID: zenPitfall, Owner: def.ID, Controller: def.ID})
	effectstest.FloatForTest(g, def, "W")
	if err := g.CastSpell(def.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "trap", Targets: effectstest.CardTarget(walker)}); err != nil {
		t.Fatalf("the trap cost was refused with one attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the lone attacker survived")
	}
}

func TestS82OldB495RavenousTrapNeedsThreeCardsInOneOpponentsGraveyard(t *testing.T) {
	cast := func(t *testing.T, discards []int) (*game.Game, *game.Player, error) {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[0]
		for seat, n := range discards {
			p := g.Seats[seat+1]
			for i := 0; i < n; i++ {
				id := uuid.New()
				p.Hand.PushTop(game.Card{InstanceID: id, Name: "Fodder", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
				g.WithWriteLock(func() { _ = g.DiscardCardsKeyedForEffect(p.ID, []uuid.UUID{id}, uuid.Nil, game.KeyedTail{}) })
			}
		}
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ravenous Trap", TypeLine: "Instant — Trap", ManaCost: "{2}{B}{B}", OracleID: zenRavenous, Owner: me.ID, Controller: me.ID})
		return g, me, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "trap", Targets: []game.TargetRef{effectstest.PlayerTarget(g.Seats[1].ID)}})
	}
	if _, _, err := cast(t, []int{2}); err == nil {
		t.Error("the {0} cost was offered with two cards")
	}
	if _, _, err := cast(t, []int{2, 2}); err == nil {
		t.Error("the {0} cost was offered with two cards each in two graveyards")
	}
	g, _, err := cast(t, []int{3})
	if err != nil {
		t.Fatalf("the {0} cost was refused with three cards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(g.Seats[1].Graveyard.Cards); n != 0 {
		t.Errorf("the target's graveyard still holds %d cards", n)
	}
}

func TestS82OldB495DesecratedEarthDestroysALandAndItsControllerDiscards(t *testing.T) {
	for _, indestructible := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: opp.ID, Controller: opp.ID})
		if indestructible {
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == land {
						g.Battlefield.Cards[i].Keywords = []string{"indestructible"}
					}
				}
			})
		}
		g.WithWriteLock(func() { opp.Hand.Cards = nil })
		card := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: card, Name: "Only Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		mineBefore := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Desecrated Earth", "Sorcery", zenDesecrated, effectstest.CardTarget(land))
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
			t.Fatalf("indestructible=%v: its controller was not asked to discard", indestructible)
		}
		effectstest.AnswerDiscard(t, g, opp.ID, card)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(land) == !indestructible {
			t.Errorf("indestructible=%v: the land's survival is wrong", indestructible)
		}
		if opp.Hand.Contains(card) || !opp.Graveyard.Contains(card) {
			t.Errorf("indestructible=%v: its controller did not discard", indestructible)
		}
		if me.Hand.Size() != mineBefore {
			t.Errorf("the caster's hand changed")
		}
	}
}

func TestS82OldB495HaloHunterDestroysAnAngel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	angel := effectstest.B12Creature(g, opp.ID, "Angel", "Creature — Angel", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	hunter := effectstest.TdCastFromHand(t, g, game.Card{Name: "Halo Hunter", TypeLine: "Creature — Demon", OracleID: zenHaloHunter, Power: 6, Toughness: 3, Keywords: []string{"intimidate"}}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, angel)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(angel) {
		t.Error("the Angel survived")
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(hunter) {
		t.Error("something other than the Angel was destroyed")
	}
}

func TestS82OldB495DevoutLightcasterExilesABlackPermanentAndHasProtection(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	black := effectstest.SeedColoredCreature(g, opp.ID, "Black Bear", []string{"B"}, 0)
	green := effectstest.SeedColoredCreature(g, opp.ID, "Green Bear", []string{"G"}, 0)
	caster := effectstest.TdCastFromHand(t, g, game.Card{Name: "Devout Lightcaster", TypeLine: "Creature — Kor Cleric", OracleID: zenLightcast, Power: 2, Toughness: 2,
		Colors: []string{"W"}, Keywords: []string{"protection from black"}}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, black)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(black) || !g.Exile.Contains(black) {
		t.Error("the black permanent was not exiled")
	}
	if !g.Battlefield.Contains(green) || !g.Battlefield.Contains(caster) {
		t.Error("something other than the black permanent left")
	}
	if !effectstest.P1g2HasKeyword(g, caster, "protection from black") {
		t.Error("the Lightcaster lacks protection from black")
	}
}
