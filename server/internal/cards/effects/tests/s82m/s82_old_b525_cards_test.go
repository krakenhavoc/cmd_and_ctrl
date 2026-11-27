package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b525_cards_test.go — Eventide, S82 older sets batch 525 (#5590):
// pump and combat tricks, Auras.

const (
	eveAssault   = "023f9e20-51fc-493d-a7aa-c9fcbdc7cd04"
	eveCaptain   = "3ac276ca-91d5-4c7d-a355-e56c67ac3ba9"
	eveThirst    = "724390aa-240d-4023-9dcb-6bdb2c2090f5"
	eveQuillspik = "b8c748ed-7ba2-4f92-a9de-5080382c2fc2"
	evePouncer   = "f695a073-7235-45c8-bc42-f1e29bb086ee"
	eveEdge      = "16ef24ea-4d84-43e5-bd4f-b6efadc87616"
	eveGift      = "3548489b-053e-4f81-a0be-07422cbe357d"
	eveScourge   = "4a1126a3-f8ce-4790-aa51-1e3f4b63eb18"
	eveFavor     = "56568fc4-7133-40e2-833c-29ecc1f953d6"
	eveClout     = "6e412163-a93a-4738-a8f0-a4d27c960360"
)

func TestS82OldB525EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eveAssault: "Unnerving Assault", eveCaptain: "Duergar Mine-Captain", eveThirst: "Cankerous Thirst",
		eveQuillspik: "Quillspike", evePouncer: "Groundling Pouncer", eveEdge: "Edge of the Divinity",
		eveGift: "Gift of the Deity", eveScourge: "Scourge of the Nobilis", eveFavor: "Favor of the Overbeing",
		eveClout: "Clout of the Dominus",
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

// eveCast puts the spell in my hand with its printed cost and colours, floats the
// pool and casts it for real (mana is read from the payment).
func eveCast(t *testing.T, g *game.Game, me *game.Player, name, oracle, cost string, colors []string, pool string, targets []game.TargetRef) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, ManaCost: cost,
		Colors: colors, Owner: me.ID, Controller: me.ID})
	effectstest.AddPool(t, g, me.ID, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func eveStats(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id)
}

func TestS82OldB525UnnervingAssaultReadsTheColoursSpent(t *testing.T) {
	for _, row := range []struct {
		pool         string
		mine, theirs int
	}{{"{C}{C}{U}", 2, 1}, {"{C}{C}{R}", 3, 2}, {"{C}{R}{U}", 3, 1}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		mine := effectstest.B16Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := effectstest.B16Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		eveCast(t, g, me, "Unnerving Assault", eveAssault, "{2}{U/R}", []string{"U", "R"}, row.pool, nil)
		if p, _ := eveStats(t, g, mine); p != row.mine {
			t.Errorf("%s: my creature has power %d, want %d", row.pool, p, row.mine)
		}
		if p, _ := eveStats(t, g, theirs); p != row.theirs {
			t.Errorf("%s: their creature has power %d, want %d", row.pool, p, row.theirs)
		}
	}
}

func TestS82OldB525CankerousThirstAsksOncePerColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	eveCast(t, g, me, "Cankerous Thirst", eveThirst, "{3}{B/G}", []string{"B", "G"}, "{C}{C}{B}{G}", effectstest.CardRefs(bear))
	effectstest.AnswerConfirmFor(t, g, me, true) // -3/-3: the black half
	effectstest.AnswerConfirmFor(t, g, me, true) // +3/+3: the green half
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := eveStats(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("both halves net out: %d/%d, want 3/3", p, tough)
	}
}

func TestS82OldB525CankerousThirstBlackOnlyShrinksAndMayBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	eveCast(t, g, me, "Cankerous Thirst", eveThirst, "{3}{B/G}", []string{"B", "G"}, "{C}{C}{C}{B}", effectstest.CardRefs(bear))
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("a 3/3 under -3/-3 should have died")
	}
	// Green only: the caster declines and the creature is unchanged.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[g2.Turn.ActiveSeat], g2.Seats[1]
	b2 := effectstest.B16Creature(g2, opp2.ID, "Bear", "Creature — Bear", 3, 3)
	eveCast(t, g2, me2, "Cankerous Thirst", eveThirst, "{3}{B/G}", []string{"B", "G"}, "{C}{C}{C}{G}", effectstest.CardRefs(b2))
	effectstest.AnswerConfirmFor(t, g2, me2, false)
	if p, _ := eveStats(t, g2, b2); p != 3 {
		t.Errorf("declined: power %d, want 3", p)
	}
}

func TestS82OldB525DuergarMineCaptainNeedsAnUntapAndPumpsAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	cap := effectstest.B12Push(g, me.ID, "Duergar Mine-Captain", "Creature — Dwarf Soldier", eveCaptain, 2, 1)
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	effectstest.AddPool(t, g, me.ID, "{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, cap, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {Q} with the Captain untapped")
	}
	effectstest.TapForTest(t, g, cap)
	effectstest.B16Activate(t, g, me.ID, cap, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := eveStats(t, g, atk); p != 3 {
		t.Errorf("the attacker has power %d, want 3", p)
	}
	if p, _ := eveStats(t, g, idle); p != 2 {
		t.Errorf("a creature that is not attacking has power %d, want 2", p)
	}
	if effectstest.B16Tapped(t, g, cap) {
		t.Error("{Q} did not untap the Captain")
	}
}

func TestS82OldB525QuillspikeRemovesAMinusCounterFromAnyOfYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spike := effectstest.B12Push(g, me.ID, "Quillspike", "Creature — Beast", eveQuillspik, 1, 1)
	other := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Marked", TypeLine: "Creature — Bear",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterMinusOne: 1}})
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{B}")
	effectstest.B16Activate(t, g, me.ID, spike, 0, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{other}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := eveStats(t, g, spike); p != 4 || tough != 4 {
		t.Errorf("Quillspike is %d/%d, want 4/4", p, tough)
	}
	if c, ok := effectstest.BattlefieldCard(g, other); !ok || c.Counters[game.CounterMinusOne] != 0 {
		t.Errorf("the other creature kept its -1/-1 counter: %+v", c.Counters)
	}
	// No counter left to remove: a strict activation is refused.
	effectstest.AddPool(t, g, me.ID, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, spike, 0, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{other}}); err == nil {
		t.Error("activated with no -1/-1 counter to remove")
	}
}

func TestS82OldB525GroundlingPouncerNeedsAnOpposingFlyerAndOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pouncer := effectstest.B12Push(g, me.ID, "Groundling Pouncer", "Creature — Faerie", evePouncer, 2, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, pouncer, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no opposing creature with flying")
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"flying"}})
	effectstest.B16Activate(t, g, me.ID, pouncer, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := eveStats(t, g, pouncer); p != 3 || tough != 4 {
		t.Errorf("Pouncer is %d/%d, want 3/4", p, tough)
	}
	effectstest.AssertKeywords(t, g, pouncer, "flying")
	if err := g.ActivateCatalogAbility(me.ID, pouncer, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second activation the same turn was allowed")
	}
}

// eveAura enchants a creature of the given colours with the Aura and returns it.
func eveAura(t *testing.T, g *game.Game, name, oracle string, colors ...string) uuid.UUID {
	t.Helper()
	host := effectstest.B43Creature(g, g.Seats[0].ID, "Host", "Creature — Bear", 2, 2, colors...)
	effectstest.CastCatalogSpell(t, g, name, "Enchantment — Aura", oracle, effectstest.CardRefs(host))
	effectstest.PassPriorityAroundTable(t, g)
	return host
}

func TestS82OldB525EdgeOfTheDivinityPumpsByColour(t *testing.T) {
	for _, row := range []struct {
		colors     []string
		power, tou int
	}{{[]string{"W"}, 3, 4}, {[]string{"B"}, 4, 3}, {[]string{"W", "B"}, 5, 5}, {[]string{"G"}, 2, 2}, {nil, 2, 2}} {
		g := effectstest.NewCatalogGame(t)
		host := eveAura(t, g, "Edge of the Divinity", eveEdge, row.colors...)
		if p, tough := eveStats(t, g, host); p != row.power || tough != row.tou {
			t.Errorf("%v: %d/%d, want %d/%d", row.colors, p, tough, row.power, row.tou)
		}
	}
}

func TestS82OldB525GiftOfTheDeity(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	black := eveAura(t, g, "Gift of the Deity", eveGift, "B")
	if p, _ := eveStats(t, g, black); p != 3 {
		t.Errorf("black host power %d, want 3", p)
	}
	effectstest.AssertKeywords(t, g, black, "deathtouch")
	g2 := effectstest.NewCatalogGame(t)
	green := eveAura(t, g2, "Gift of the Deity", eveGift, "G")
	if p, _ := eveStats(t, g2, green); p != 3 {
		t.Errorf("green host power %d, want 3", p)
	}
	if got := effectstest.EffectiveAbilities(t, g2, green); effectstest.ContainsString(got, "deathtouch") {
		t.Errorf("a green host gets no deathtouch: %v", got)
	}
	if len(effectstest.EffectiveOf(t, g2, green).BlockRequirements) == 0 {
		t.Error("a green host should carry the Lure block requirement")
	}
	g3 := effectstest.NewCatalogGame(t)
	white := eveAura(t, g3, "Gift of the Deity", eveGift, "W")
	if p, _ := eveStats(t, g3, white); p != 2 {
		t.Errorf("white host power %d, want 2", p)
	}
}

func TestS82OldB525ScourgeOfTheNobilis(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	red := eveAura(t, g, "Scourge of the Nobilis", eveScourge, "R")
	if p, _ := eveStats(t, g, red); p != 3 {
		t.Errorf("red host power %d, want 3", p)
	}
	idx, _ := effectstest.GrantedActivatedIndex(t, g, red)
	if idx < 0 {
		t.Fatal("a red host has no granted firebreathing")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{W}")
	effectstest.B16Activate(t, g, me.ID, red, idx, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := eveStats(t, g, red); p != 4 {
		t.Errorf("after firebreathing power %d, want 4", p)
	}
	g2 := effectstest.NewCatalogGame(t)
	white := eveAura(t, g2, "Scourge of the Nobilis", eveScourge, "W")
	effectstest.AssertKeywords(t, g2, white, "lifelink")
	if i, _ := effectstest.GrantedActivatedIndex(t, g2, white); i >= 0 {
		t.Error("a white host must not get the firebreathing")
	}
}

func TestS82OldB525FavorAndCloutGrantByColour(t *testing.T) {
	for _, row := range []struct {
		name, oracle, color, kw string
	}{
		{"Favor of the Overbeing", eveFavor, "G", "vigilance"}, {"Favor of the Overbeing", eveFavor, "U", "flying"},
		{"Clout of the Dominus", eveClout, "U", "shroud"}, {"Clout of the Dominus", eveClout, "R", "haste"},
	} {
		g := effectstest.NewCatalogGame(t)
		host := eveAura(t, g, row.name, row.oracle, row.color)
		if p, _ := eveStats(t, g, host); p != 3 {
			t.Errorf("%s on %s: power %d, want 3", row.name, row.color, p)
		}
		effectstest.AssertKeywords(t, g, host, row.kw)
		g2 := effectstest.NewCatalogGame(t)
		off := eveAura(t, g2, row.name, row.oracle, "W")
		if p, _ := eveStats(t, g2, off); p != 2 {
			t.Errorf("%s on a white creature: power %d, want 2", row.name, p)
		}
	}
}
