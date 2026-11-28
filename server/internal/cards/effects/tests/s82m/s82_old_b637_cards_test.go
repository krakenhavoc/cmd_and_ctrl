package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b637_cards_test.go — Planeshift, batch 637 (#5615). Each test plays
// the real catalog entry and asserts what its text says.

const (
	b637PygmyKavuOracle   = "17407b73-0dbe-4a71-84ec-3aaca9651f5b"
	b637AuraBlastOracle   = "4e3c3bdc-667d-42ec-b960-159a39c53cb3"
	b637WeatherlightOrcl  = "57889a8e-0039-45c2-a608-ab315d51aa3a"
	b637LashknifeOracle   = "bfde73b2-46bd-4fdb-9468-35819af609f7"
	b637UrzasGuiltOracle  = "b0aeeeb6-490c-4638-a4ad-498e4d85413e"
	b637ImplodeOracle     = "7e529372-cec7-40d7-a1ac-1dce1764f8b2"
	b637ConfoundOracle    = "ca8a0e81-f885-47ee-82bb-fb25ec61bb1c"
	b637SkyshroudOracle   = "ffcedb73-538a-45ea-8564-7b92624537e0"
	b637CreatureTypeBear  = "Creature — Bear"
	b637SearchCreatureTyp = "Creature — Bear"
)

func TestOldB637EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b637PygmyKavuOracle:  "Pygmy Kavu",
		b637AuraBlastOracle:  "Aura Blast",
		b637WeatherlightOrcl: "Skyship Weatherlight",
		b637LashknifeOracle:  "Lashknife Barrier",
		b637UrzasGuiltOracle: "Urza's Guilt",
		b637ImplodeOracle:    "Implode",
		b637ConfoundOracle:   "Confound",
		b637SkyshroudOracle:  "Skyshroud Blessing",
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

func TestOldB637PygmyKavuDrawsForEachBlackCreatureYourOpponentsControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for _, c := range []game.Card{
		{Name: "Black One", TypeLine: b637CreatureTypeBear, Colors: []string{"B"}, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID},
		{Name: "Black Two", TypeLine: b637CreatureTypeBear, Colors: []string{"B", "G"}, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID},
		{Name: "Green One", TypeLine: b637CreatureTypeBear, Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID},
		{Name: "My Black", TypeLine: b637CreatureTypeBear, Colors: []string{"B"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID},
	} {
		c.InstanceID = uuid.New()
		effectstest.PushBattlefieldCardWithTimestamp(g, c)
	}
	// The hand starts without the Kavu: the cast helper adds it and casts it.
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Pygmy Kavu", "Creature — Kavu", b637PygmyKavuOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d: two black creatures belong to an opponent", got, hand+2)
	}
}

func TestOldB637AuraBlastDestroysAnEnchantmentAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ench := effectstest.B12Permanent(g, opp.ID, "Odd Ward", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b637CreatureTypeBear, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Aura Blast", "Instant", b637AuraBlastOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("a creature was accepted as the target of Aura Blast")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Aura Blast", "Instant", b637AuraBlastOracle, effectstest.CardTarget(ench))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment was not destroyed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (the spell cast from hand is gone, one card drawn)", got, hand+1)
	}
}

func TestOldB637ImplodeDestroysALandAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b637CreatureTypeBear, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Implode", "Sorcery", b637ImplodeOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("a creature was accepted as the target of Implode")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Implode", "Sorcery", b637ImplodeOracle, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestOldB637ConfoundCountersASpellThatTargetsACreatureAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b637CreatureTypeBear, 2, 2)
	// A spell aimed at a player is not a legal target for Confound.
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if err := g.CastSpell(me.ID, mustPutConfound(me), game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err == nil {
		t.Fatal("Confound accepted a spell that targets only a player")
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { opp.Life = 40 })
	bolt2 := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTarget(bear))
	hand := me.Hand.Size()
	id := mustPutConfound(me)
	hand++
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt2}}}); err != nil {
		t.Fatalf("cast Confound: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, bear); got != 0 {
		t.Errorf("the countered Bolt still dealt %d damage", got)
	}
	// Confound left the hand (-1) and drew one (+1).
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d", got, hand)
	}
}

func mustPutConfound(p *game.Player) uuid.UUID {
	return effectstest.PutInHand(p, game.Card{Name: "Confound", TypeLine: "Instant", OracleID: b637ConfoundOracle})
}

func TestOldB637SkyshroudBlessingGivesEveryLandShroudAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b637CreatureTypeBear, 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Skyshroud Blessing", "Instant", b637SkyshroudOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "shroud") {
			t.Error("a land did not gain shroud")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "shroud") {
		t.Error("a creature gained shroud")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestOldB637LashknifeBarrierDrawsOnEntryAndReducesDamageToYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Lashknife Barrier", "Enchantment", b637LashknifeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: the trigger drew one", got, hand+1)
	}
	mine := effectstest.B12Creature(g, me.ID, "Mine", b637CreatureTypeBear, 2, 9)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b637CreatureTypeBear, 2, 9)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, mine); got != 2 {
		t.Errorf("your creature was dealt %d damage, want 2", got)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, theirs); got != 3 {
		t.Errorf("an opponent's creature was dealt %d damage, want 3", got)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - me.Life; got != 3 {
		t.Errorf("you lost %d life, want 3: the barrier guards creatures only", got)
	}
}

func TestOldB637UrzasGuiltDrawsDiscardsThenCostsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	before := map[uuid.UUID]int{}
	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID], life[p.ID] = p.Hand.Size(), p.Life
	}
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Urza's Guilt", "Sorcery", b637UrzasGuiltOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Nobody has lost life while a discard is still owed.
	asked := 0
	for round := 0; round < 8; round++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseCards {
				open = c
				break
			}
		}
		if open == nil {
			break
		}
		for _, p := range g.Seats {
			if p.Life != life[p.ID] {
				t.Fatalf("a player lost life before the last discard was chosen")
			}
		}
		if len(open.ChooseCards) < 3 {
			t.Fatalf("a discard prompt offers %d cards, want a hand's worth", len(open.ChooseCards))
		}
		if err := g.ResolveChooseCards(open.ID, open.Chooser, open.ChooseCards[:3]); err != nil {
			t.Fatalf("answer the discard: %v", err)
		}
		asked++
		effectstest.PassPriorityAroundTable(t, g)
	}
	if asked != 4 {
		t.Fatalf("%d discard prompts, want one per player (4)", asked)
	}
	for _, p := range g.Seats {
		want := before[p.ID] + 2 - 3
		if p.ID == me.ID {
			want-- // the spell itself left the hand
			want++ // CastCatalogSpell added it first
		}
		if got := p.Hand.Size(); got != want {
			t.Errorf("seat %v hand %d, want %d", p.ID, got, want)
		}
		if got := life[p.ID] - p.Life; got != 4 {
			t.Errorf("seat %v lost %d life, want 4", p.ID, got)
		}
	}
}

func TestOldB637SkyshipWeatherlightSearchesExilesAndReturnsOneAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Spare Bear", TypeLine: b637SearchCreatureTyp},
		game.Card{Name: "Spare Rod", TypeLine: "Artifact"},
		game.Card{Name: "Spare Spell", TypeLine: "Sorcery"},
	)
	effectstest.ToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Skyship Weatherlight", "Legendary Artifact", b637WeatherlightOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.SearchChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.HasID(pick.SearchCards, ids[2]) {
		t.Error("a sorcery was offered to a search for artifacts and creatures")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0], ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:2] {
		if effectstest.B12ZoneOf(g, id) != game.ZoneExile {
			t.Fatalf("a found card is in %v, want exile", effectstest.B12ZoneOf(g, id))
		}
	}
	ship := effectstest.BattlefieldIDsNamed(g, "Skyship Weatherlight")[0]
	if err := effectstest.RavB2Activate(t, g, me.ID, ship, 0, game.ActivateAbilityParams{}, "C", "C", "C", "C"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	inHand := 0
	for _, id := range ids[:2] {
		if effectstest.B12ZoneOf(g, id) == game.ZoneHand {
			inHand++
		}
	}
	if inHand != 1 {
		t.Errorf("%d of the exiled cards went to hand, want exactly 1", inHand)
	}
}
