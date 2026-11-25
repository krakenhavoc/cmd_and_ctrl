package testsso

import (
	"go/ast"
	"go/token"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// source_object_primitives_test.go — #1432, CR 400.7. The primitives
// that act on "this permanent" do nothing when the permanent under the
// source's instance ID is a NEW object: it left and came back while the
// ability waited. See source_object_guard.go and ADR 0018's 2026-09-24
// (#1432) amendment. The engine half is game/source_new_object_test.go.

// --- the primitives, one by one --------------------------------------

// primitiveCase is one primitive aimed at its own source. `setup` runs
// after the (optional) flicker, on whatever object is there by then;
// `acted` reports whether the primitive changed that object.
type primitiveCase struct {
	name     string
	typeLine string
	setup    func(t *testing.T, g *game.Game, src uuid.UUID)
	apply    func(ctx *effects.Context, src uuid.UUID) error
	acted    func(t *testing.T, g *game.Game, src uuid.UUID) bool
}

func onTheBattlefield(g *game.Game, id uuid.UUID) bool { return g.Battlefield.Contains(id) }

// flickerEpochBefore is the Flicker case's before-reading, keyed by the
// per-subtest source ID so parallel-safe runs cannot collide.
var flickerEpochBefore = map[uuid.UUID]int{}

// objectEpochOf is the CR 400.7 epoch of the card on the battlefield.
func objectEpochOf(g *game.Game, id uuid.UUID) int {
	epoch := -1
	g.ReadSnapshot(func() {
		if c, ok := effectstest.BattlefieldCard(g, id); ok {
			epoch = c.ObjectEpoch
		}
	})
	return epoch
}

func primitiveCases(opp uuid.UUID) []primitiveCase {
	gone := func(_ *testing.T, g *game.Game, src uuid.UUID) bool { return !onTheBattlefield(g, src) }
	return []primitiveCase{
		{name: "AddCounter", apply: func(ctx *effects.Context, src uuid.UUID) error {
			return effects.AddCounter{Target: src, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
		}, acted: func(_ *testing.T, g *game.Game, src uuid.UUID) bool {
			return effectstest.CounterCount(g, src, game.CounterPlusOne) != 0
		}},
		{name: "AddCounter removing", setup: func(t *testing.T, g *game.Game, src uuid.UUID) {
			if err := g.AddCounter(src, "charge", 2); err != nil {
				t.Fatal(err)
			}
		}, apply: func(ctx *effects.Context, src uuid.UUID) error {
			return effects.AddCounter{Target: src, Kind: "charge", N: -1}.Apply(ctx)
		}, acted: func(_ *testing.T, g *game.Game, src uuid.UUID) bool {
			return effectstest.CounterCount(g, src, "charge") != 2
		}},
		{name: "SacrificePermanent", apply: func(ctx *effects.Context, src uuid.UUID) error {
			return effects.SacrificePermanent{Target: src}.Apply(ctx)
		}, acted: gone},
		{name: "TapTarget", apply: func(ctx *effects.Context, src uuid.UUID) error {
			return effects.TapTarget{Target: src}.Apply(ctx)
		}, acted: func(t *testing.T, g *game.Game, src uuid.UUID) bool { return effectstest.B16Tapped(t, g, src) }},
		{name: "UntapTarget", setup: func(_ *testing.T, g *game.Game, src uuid.UUID) { effectstest.B16Tap(g, src) },
			apply: func(ctx *effects.Context, src uuid.UUID) error { return effects.UntapTarget{Target: src}.Apply(ctx) },
			acted: func(t *testing.T, g *game.Game, src uuid.UUID) bool { return !effectstest.B16Tapped(t, g, src) }},
		{name: "BounceToHand", apply: func(ctx *effects.Context, src uuid.UUID) error {
			return effects.BounceToHand{Target: src}.Apply(ctx)
		}, acted: gone},
		{name: "ExileTarget", apply: func(ctx *effects.Context, src uuid.UUID) error {
			return effects.ExileTarget{Target: src}.Apply(ctx)
		}, acted: gone},
		{name: "Flicker", setup: func(_ *testing.T, g *game.Game, src uuid.UUID) {
			flickerEpochBefore[src] = objectEpochOf(g, src)
		}, apply: func(ctx *effects.Context, src uuid.UUID) error {
			return effects.Flicker{Target: src}.Apply(ctx)
		}, acted: func(_ *testing.T, g *game.Game, src uuid.UUID) bool {
			// A flicker comes back, so it is judged by the epoch.
			return objectEpochOf(g, src) != flickerEpochBefore[src]
		}},
		{name: "DestroyTarget", apply: func(ctx *effects.Context, src uuid.UUID) error {
			return effects.DestroyTarget{Target: src}.Apply(ctx)
		}, acted: gone},
		{name: "PutIntoLibrary", apply: func(ctx *effects.Context, src uuid.UUID) error {
			return effects.PutIntoLibrary{Card: src}.Apply(ctx)
		}, acted: gone},
		{name: "BoostUntilEOT", apply: func(ctx *effects.Context, src uuid.UUID) error {
			return effects.BoostUntilEOT{Target: src, Power: 3}.Apply(ctx)
		}, acted: func(t *testing.T, g *game.Game, src uuid.UUID) bool {
			return effectstest.EffectivePower(t, g, src) != 2
		}},
		{name: "GrantKeywordUntilEOT", apply: func(ctx *effects.Context, src uuid.UUID) error {
			return effects.GrantKeywordUntilEOT{Target: src, Keywords: []string{"flying"}}.Apply(ctx)
		}, acted: func(t *testing.T, g *game.Game, src uuid.UUID) bool {
			return slices.Contains(effectstest.EffectiveAbilities(t, g, src), "flying")
		}},
		{name: "BecomeCreatureUntilEOT (zero target = the source)", typeLine: "Artifact — Vehicle",
			apply: func(ctx *effects.Context, _ uuid.UUID) error { return effects.BecomeCreatureUntilEOT{}.Apply(ctx) },
			acted: func(t *testing.T, g *game.Game, src uuid.UUID) bool {
				return slices.Contains(effectstest.EffectiveTypes(t, g, src), "Creature")
			}},
		{name: "GainControl", apply: func(ctx *effects.Context, src uuid.UUID) error {
			return effects.GainControl{Target: src, Controller: opp}.Apply(ctx)
		}, acted: func(_ *testing.T, g *game.Game, src uuid.UUID) bool {
			var ctl uuid.UUID
			g.ReadSnapshot(func() {
				c, _ := effectstest.BattlefieldCard(g, src)
				ctl = c.Controller
			})
			return ctl == opp
		}},
	}
}

// Every primitive, aimed at "this": after a flicker in response it does
// nothing to the new object, and with no flicker it does what it says.
func TestPrimitivesDoNotActOnASourceThatCameBackAsANewObject(t *testing.T) {
	for _, flicker := range []bool{false, true} {
		for _, tc := range primitiveCases(uuid.Nil) {
			name := tc.name + "/live"
			if flicker {
				name = tc.name + "/flickered"
			}
			t.Run(name, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me, opp := g.Seats[0], g.Seats[1]
				tc := tc
				for _, c := range primitiveCases(opp.ID) {
					if c.name == tc.name {
						tc = c
					}
				}
				typeLine := tc.typeLine
				if typeLine == "" {
					typeLine = "Creature — Bear"
				}
				src := effectstest.PushDiesCreatureForTest(g, me.ID, "Relic Bear", "", typeLine, 2, 2)
				item := &game.StackItem{ID: uuid.New(), Kind: game.StackItemTriggered,
					Controller: me.ID, Owner: me.ID, SourceCardID: src}
				g.WithWriteLock(func() {
					ref, _ := g.PermanentRefForEffect(src)
					item.SourceObject = ref
				})
				if flicker {
					effectstest.FlickerInResponse(t, g, src)
				}
				if tc.setup != nil {
					tc.setup(t, g, src)
				}
				g.WithWriteLock(func() {
					if err := tc.apply(effects.NewContext(g, item), src); err != nil {
						t.Fatalf("Apply: %v", err)
					}
				})
				if got := tc.acted(t, g, src); got == flicker {
					if flicker {
						t.Errorf("%s acted on the new object its source became (CR 400.7)", tc.name)
					} else {
						t.Errorf("%s did nothing to a live source", tc.name)
					}
				}
			})
		}
	}
}

// "If you do" is told no: a sacrifice / exile / bounce of a new object
// did not happen, so the clause hanging off it hears false.
func TestTheThenClauseOfASkippedSelfActHearsNo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := effectstest.PushDiesCreatureForTest(g, me.ID, "Relic Bear", "", "Creature — Bear", 2, 2)
	item := &game.StackItem{ID: uuid.New(), Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID, SourceCardID: src}
	g.WithWriteLock(func() { item.SourceObject, _ = g.PermanentRefForEffect(src) })
	effectstest.FlickerInResponse(t, g, src)
	heard := map[string]bool{}
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, item)
		testHeardExile = heard
		_ = effects.SacrificePermanent{Target: src, ThenTail: effects.WithWord(ctx.Tail(testHeardThenTail), "sacrifice")}.Apply(ctx)
		_ = effects.ExileTarget{Target: src, ThenTail: ctx.Tail(testHeardExileTail)}.Apply(ctx)
		_ = effects.BounceToHand{Target: src, ThenTail: effects.WithWord(ctx.Tail(testHeardThenTail), "bounce")}.Apply(ctx)
	})
	for _, k := range []string{"sacrifice", "exile", "bounce"} {
		ok, ran := heard[k]
		if !ran || ok {
			t.Errorf("%s's tail: ran=%v told=%v, want it told false", k, ran, ok)
		}
	}
	if !onTheBattlefield(g, src) {
		t.Error("the new object left the battlefield")
	}
}

// withoutNewSourceObject drops only the source, and only when it is a
// new object — the list primitives (TurnFaceDown, PhaseOut, AirbendAll).
func TestWithoutNewSourceObjectDropsOnlyTheNewSource(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := effectstest.PushDiesCreatureForTest(g, me.ID, "Relic Bear", "", "Creature — Bear", 2, 2)
	other := effectstest.PushDiesCreatureForTest(g, me.ID, "Other Bear", "", "Creature — Bear", 2, 2)
	item := &game.StackItem{ID: uuid.New(), Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID, SourceCardID: src}
	g.WithWriteLock(func() { item.SourceObject, _ = g.PermanentRefForEffect(src) })
	g.WithWriteLock(func() {
		if got := effects.NewContext(g, item).WithoutNewSourceObject([]uuid.UUID{src, other}); len(got) != 2 {
			t.Errorf("live source dropped: %v", got)
		}
	})
	effectstest.FlickerInResponse(t, g, src)
	g.WithWriteLock(func() {
		got := effects.NewContext(g, item).WithoutNewSourceObject([]uuid.UUID{src, other})
		if len(got) != 1 || got[0] != other {
			t.Errorf("got %v, want only the other bear", got)
		}
	})
}

// b09SourceStillOnBattlefield — the guard ~20 card bodies put in front
// of an act on "this" (and of the payoff that follows it, as Coalition
// Relic's mana does) — means the same OBJECT, not the same card.
func TestSourceStillOnBattlefieldMeansTheSameObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := effectstest.PushDiesCreatureForTest(g, me.ID, "Relic Bear", "", "Creature — Bear", 2, 2)
	item := &game.StackItem{ID: uuid.New(), Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID, SourceCardID: src}
	g.WithWriteLock(func() { item.SourceObject, _ = g.PermanentRefForEffect(src) })
	g.WithWriteLock(func() {
		if !effects.B09SourceStillOnBattlefield(g, item) {
			t.Error("a live source reads as gone")
		}
	})
	effectstest.FlickerInResponse(t, g, src)
	g.WithWriteLock(func() {
		if effects.B09SourceStillOnBattlefield(g, item) {
			t.Error("a source that left and came back reads as still here (CR 400.7)")
		}
	})
}

// --- proof cards ------------------------------------------------------

// Bartolomé del Presidio — counter on self, an activated ability. Fed a
// Bear, then flickered in response: the new Bartolomé gets no counter.
func TestBartolomeFlickeredInResponseGetsNoCounter(t *testing.T) {
	for _, flicker := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bart := effectstest.B12Push(g, me.ID, "Bartolomé del Presidio", "Legendary Creature — Vampire Knight", effectstest.B16BartolomeDelPresidioOracle, 2, 1)
		bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
		effectstest.AdvanceToMain(t, g)
		if err := g.ActivateCatalogAbility(me.ID, bart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err != nil {
			t.Fatal(err)
		}
		if flicker {
			effectstest.FlickerInResponse(t, g, bart)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := 1
		if flicker {
			want = 0
		}
		if got := effectstest.CounterCount(g, bart, game.CounterPlusOne); got != want {
			t.Errorf("flicker=%v: +1/+1 counters = %d, want %d", flicker, got, want)
		}
	}
}

// Underworld Breach — "sacrifice it" at the beginning of the end step.
// A Breach flickered in response is a new object and stays.
func TestUnderworldBreachFlickeredInResponseIsNotSacrificed(t *testing.T) {
	for _, flicker := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		breach := effectstest.PushPermanentForTest(g, me.ID, "Underworld Breach", "27e0948b-9916-473b-8d8c-a51bdfbc7457", "Enchantment")
		effectstest.AdvanceToEndStep(t, g)
		if effectstest.TriggerOnStack(g, breach) == nil {
			t.Fatal("Underworld Breach did not trigger at the end step")
		}
		if flicker {
			effectstest.FlickerInResponse(t, g, breach)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := onTheBattlefield(g, breach); got != flicker {
			t.Errorf("flicker=%v: Breach on the battlefield = %v", flicker, got)
		}
	}
}

// Devoted Druid — "untap this creature". The Druid is flickered in
// response and the new one is tapped by something else; the ability
// does not untap it.
func TestDevotedDruidDoesNotUntapANewObject(t *testing.T) {
	for _, flicker := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		druid := effectstest.B12Push(g, me.ID, "Devoted Druid", "Creature — Elf Druid", effectstest.DevotedDruidOracle, 0, 2)
		effectstest.AdvanceToMain(t, g)
		effectstest.B16Tap(g, druid)
		if err := g.ActivateCatalogAbility(me.ID, druid, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatal(err)
		}
		if flicker {
			effectstest.FlickerInResponse(t, g, druid)
			effectstest.B16Tap(g, druid)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.B12Card(t, g, druid).Tapped; got != flicker {
			t.Errorf("flicker=%v: tapped = %v", flicker, got)
		}
	}
}

// Mulldrifter, evoked — the engine's sacrifice-on-entry trigger
// (alternative_cost.go). Flickered in response it is a new object that
// was never evoked, and stays; its draw still happens.
func TestEvokedMulldrifterFlickeredInResponseIsNotSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastWithAltCost(t, g, "Mulldrifter", "Creature — Elemental", effectstest.MulldrifterOracle, "evoke")
	effectstest.PassUntilOnBattlefield(t, g, id)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, caster.ID)
	effectstest.FlickerInResponse(t, g, id)
	hand := len(caster.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if !onTheBattlefield(g, id) {
		t.Error("the new Mulldrifter was sacrificed by the evoke trigger of the one that left (CR 400.7)")
	}
	if got := len(caster.Hand.Cards); got != hand+2 {
		t.Errorf("hand %d → %d, want the two cards the draw does not need the Mulldrifter for", hand, got)
	}
}

// Krenko, Tin Street Kingpin — the counter and the power. Flickered in
// response, the new Krenko gets no counter and the Goblins are the
// departed Krenko's last-known power (1). Killed in response: the same
// one Goblin, and no counter on the card in the graveyard.
func TestKrenkoThatIsNotTheAttackerAnyMoreMakesLastKnownPowerGoblins(t *testing.T) {
	for _, how := range []string{"flicker", "kill"} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		krenko := effectstest.PushDiesCreatureForTest(g, me.ID, "Krenko, Tin Street Kingpin", effectstest.KrenkoKingpinOrcl,
			"Legendary Creature — Goblin", 1, 2)
		effectstest.DeclareAttack(t, g, opp.ID, krenko)
		if effectstest.TriggersOnStackFrom(g, krenko) != 1 {
			t.Fatal("Krenko did not trigger")
		}
		if how == "flicker" {
			effectstest.FlickerInResponse(t, g, krenko)
		} else {
			g.WithWriteLock(func() {
				if err := g.DestroyPermanentForEffect(krenko); err != nil {
					t.Fatal(err)
				}
			})
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CountBattlefieldNamed(g, me.ID, "Goblin"); got != 1 {
			t.Errorf("%s: Goblins = %d, want 1 (the last-known power)", how, got)
		}
		var counters int
		g.ReadSnapshot(func() {
			if c, ok := g.LookupCardForEffect(krenko); ok {
				counters = c.Counters[game.CounterPlusOne]
			}
		})
		if counters != 0 {
			t.Errorf("%s: the Krenko that is not the attacker got %d +1/+1 counters", how, counters)
		}
	}
	spec, _ := effects.Lookup(effectstest.KrenkoKingpinOrcl)
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Krenko is complete now: %v %v", spec.Completeness, spec.Caveats)
	}
}

// --- the exceptions: text that follows the card ----------------------

const (
	selfReturnProbeOracle = "test-1432-self-return-probe"
	diesTuckProbeOracle   = "test-1432-dies-tuck-probe"
)

// "{B}: Return this card from your graveyard to the battlefield. Put a
// +1/+1 counter on it." The ability names the graveyard card; the effect
// moves it onto the battlefield itself, and CR 400.7 lets the rest of
// the effect find the permanent it made: the counter lands. (A flicker
// primitive would not show this — it returns the card under a fresh
// instance ID — so the move here is a reanimation, which keeps it, as
// unearth's "it gains haste" does.)
func TestAnEffectThatMovedItsOwnSourceStillFindsIt(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: selfReturnProbeOracle,
		Name:     "Self-Return Probe",
		Activated: []effects.ActivatedAbility{{
			Label: "{B}: Return this card from your graveyard to the battlefield. Put a +1/+1 counter on it.",
			Cost:  effects.ManaCost("{B}"),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
					return err
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			},
		}},
	})
	g := effectstest.NewCatalogGame(t)
	probe, _ := effectstest.ActivateFromGraveyard(t, g, "Self-Return Probe", "Creature — Construct",
		selfReturnProbeOracle, 1, 1, "{B}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !onTheBattlefield(g, probe) {
		t.Fatal("the probe did not return")
	}
	if got := effectstest.CounterCount(g, probe, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the returned probe = %d, want 1 — the effect lost the permanent it made", got)
	}
}

// "When this dies, put it on the bottom of its owner's library." The
// trigger names the permanent that died; the card it acts on is in the
// graveyard. Off the battlefield, the primitive follows the card.
func TestADiesTriggerStillActsOnTheCardInTheZoneItWentTo(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: diesTuckProbeOracle,
		Name:     "Dies-Tuck Probe",
		Triggered: []game.TriggeredAbility{effects.WhenThisDies("Dies-Tuck Probe — put it on the bottom", func(g *game.Game, item *game.StackItem) error {
			return effects.PutIntoLibrary{Card: item.SourceCardID, ToBottom: true}.Apply(effects.NewContext(g, item))
		})},
	})
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	probe := effectstest.PushDiesCreatureForTest(g, me.ID, "Dies-Tuck Probe", diesTuckProbeOracle, "Creature — Construct", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(probe); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Library.Contains(probe) {
		t.Error("the dies trigger did not put the card that died into its library")
	}
}

// --- the regression: no primitive may forget to ask ------------------

// Every Apply in this package whose primitive names a card to act on —
// a `Target`, `Card` or `Targets` field of uuid type — either asks the
// #1432 question or delegates to one that does, or is listed below with
// the reason it must not. A new primitive fails here until its author
// decides which.
func TestEveryPrimitiveThatActsOnACardAsksAboutItsSource(t *testing.T) {
	exempt := map[string]string{
		"DealDamage":                          "damage FROM the source reads last-known information (CR 608.2h); damage TO it is dealt by a mass-effect loop as often as by \"this\", and is not an act the source's text names as \"this\"",
		"ReturnFromGraveyard":                 "follows the card: the card is in a graveyard, never a permanent that left and came back",
		"ReturnFromGraveyardTogether":         "follows the cards: each is in a graveyard, never a permanent that left and came back",
		"ReturnFromExile":                     "follows the card: the card is in exile",
		"ReturnFromExileTogether":             "follows the cards: each is in exile, never a permanent that left and came back",
		"GrantFlashbackToCard":                "a card in a graveyard",
		"GrantHarmonizeToCard":                "a card in a graveyard",
		"OfferGraveyardCast":                  "a card in a graveyard, refused unless it is still there",
		"PlotExiled":                          "a card in exile",
		"CastTargetFromYourGraveyardThisTurn": "a card in a graveyard, a permission over the card itself",
		"CreateTokenCopy":                     "reads copiable values; last-known information is the answer for a source that left (CR 707.4)",
	}
	ps := scanPrimitives(t)
	structs, applies, guarded := ps.structs, ps.applies, ps.guarded
	var missing []string
	for name, body := range applies {
		_ = body
		st, ok := structs[name]
		if !ok || !actsOnACard(st) || guarded[name] {
			continue
		}
		if _, ok := exempt[name]; ok {
			continue
		}
		missing = append(missing, name)
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("these primitives act on a card by ID and never ask whether it is their source as a NEW object (#1432, CR 400.7): %v\n"+
			"Guard the act with ctx.IsNewSourceObject(target) (source_object_guard.go), or add the type to the exemption table with the reason it follows the card.", missing)
	}
	for name := range exempt {
		if _, ok := applies[name]; !ok {
			t.Errorf("exempt primitive %s no longer exists — drop it from the table", name)
		}
		if guarded[name] {
			t.Errorf("exempt primitive %s asks the question now — drop it from the table", name)
		}
	}
}

// primitiveScan is the package parsed once for the #1432 / #1463
// regressions: every struct type, every Apply body, and the set of
// primitives that ask the new-object question — directly, or by
// delegating to one that does.
type primitiveScan struct {
	structs map[string]*ast.StructType
	applies map[string]*ast.BlockStmt
	guarded map[string]bool
}

func scanPrimitives(t *testing.T) primitiveScan {
	t.Helper()
	fset := token.NewFileSet()
	files, err := effectstest.FlatGlob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	structs := map[string]*ast.StructType{}
	applies := map[string]*ast.BlockStmt{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := effectstest.FlatReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		file, err := effectstest.FlatParseFile(fset, f, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if st, ok := ts.Type.(*ast.StructType); ok {
							structs[ts.Name.Name] = st
						}
					}
				}
			case *ast.FuncDecl:
				if d.Recv == nil || d.Name.Name != "Apply" || d.Body == nil {
					continue
				}
				recv := d.Recv.List[0].Type
				if star, ok := recv.(*ast.StarExpr); ok {
					recv = star.X
				}
				if id, ok := recv.(*ast.Ident); ok {
					applies[id.Name] = d.Body
				}
			}
		}
	}
	// untilEndOfTurn (until_end_of_turn.go) is ScopedEffectFor with an
	// until-end-of-turn duration, and so eotSnapshot underneath (#1497).
	// ApplyReports is BecomeCopy's Apply with the "if you do" report
	// (#3774): Apply delegates to it, and the guard lives there.
	markers := map[string]bool{"IsNewSourceObject": true, "IsNewSourceObjectAsThis": true, "WithoutNewSourceObject": true, "EotSnapshot": true, "ApplyFor": true, "UntilEndOfTurn": true, "ApplyReports": true}
	guarded := map[string]bool{}
	for changed := true; changed; {
		changed = false
		for name, body := range applies {
			if guarded[name] {
				continue
			}
			found := false
			ast.Inspect(body, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.SelectorExpr:
					if markers[x.Sel.Name] {
						found = true
					}
				case *ast.Ident:
					if markers[x.Name] {
						found = true
					}
				case *ast.CompositeLit:
					if id, ok := x.Type.(*ast.Ident); ok && guarded[id.Name] {
						found = true
					}
				}
				return !found
			})
			if found {
				guarded[name] = true
				changed = true
			}
		}
	}
	return primitiveScan{structs: structs, applies: applies, guarded: guarded}
}

// actsOnACard reports whether a primitive's struct names a card to act
// on: a `Target`, `Card` or `Targets` field of uuid type.
func actsOnACard(st *ast.StructType) bool {
	for _, f := range st.Fields.List {
		typ := effectstest.ExprString(f.Type)
		if typ != "uuid.UUID" && typ != "[]uuid.UUID" {
			continue
		}
		for _, n := range f.Names {
			if n.Name == "Target" || n.Name == "Card" || n.Name == "Targets" {
				return true
			}
		}
	}
	return false
}

// And the card files that skip the primitive: a game mutator called
// straight on `item.SourceCardID` / `ctx.Source()` must sit in a
// function that asks the #1432 question. Best effort by construction —
// a source ID copied into a variable first is not followed — but it
// catches the shape every one of the sites #1432 fixed had.
func TestNoCardActsOnItsOwnSourceThroughAGameMutatorWithoutAsking(t *testing.T) {
	mutators := map[string]bool{
		"AddCounterForEffect": true, "AddCounterThenForEffect": true,
		"TapTargetForEffect": true, "UntapTargetForEffect": true,
		"SacrificePermanentForEffect": true, "ExileCardForEffect": true, "ExileCardKeyedForEffect": true,
		"BounceToHandForEffect": true, "BounceToHandThenForEffect": true, "TuckToLibraryForEffect": true,
		"DestroyPermanentForEffect": true, "RegenerateForEffect": true,
		"TransformPermanentForEffect": true, "ExileAndReturnTransformedForEffect": true,
		"SetClassLevelForEffect": true, "SolveCaseForEffect": true, "HarnessForEffect": true,
		"BecomePreparedForEffect": true,
	}
	// SourcePermanent answers the same question: its Left is true for a
	// source that left and came back (#1418).
	// followUpThis (prevention_static.go) is SourcePermanent for a
	// prevention follow-up, refusing a source that has Left.
	asks := map[string]bool{"SourceIsNewObject": true, "IsNewSourceObject": true, "IsNewSourceObjectAsThis": true, "B09SourceStillOnBattlefield": true, "SourcePermanent": true, "FollowUpThis": true}
	// A SPELL moving itself — "Exile Teferi's Protection" — is not a
	// permanent's ability; a spell item is never judged.
	spells := map[string]bool{
		"teferis_protection.go": true, "teferis_reproach.go": true,
		"avatars_wrath.go": true, "blue_suns_zenith.go": true,
		// "Shuffle Beacon of Immortality into its owner's library." (#3245)
		"beacon_of_immortality.go": true,
		"temporal_mastery.go":      true, "temporal_trespass.go": true,
		// "Exile Bond of Insight." / "Exile Deliver Unto Evil." (S82 OLD batch 170, #5520)
		"s82_old_b170_helpers.go": true,
		// "Exile Draconic Intervention." (#3537)
		"draconic_intervention.go": true,
		"bound_determined.go":      true, // Bound's "Exile this card." (ADR 0106 PR 3)
		// Restorative Burst, the sorcery back face: "Exile Restorative Burst."
		"pestilent_cauldron.go": true,
		// "Exile Morningtide's Light." (ADR 0108 PR 7, #1904)
		"morningtides_light.go": true,
		// "Shuffle Beacon of Tomorrows into its owner's library." and
		// "Exile Rise of the Eldrazi." (S58 PR 5)
		"beacon_of_tomorrows.go": true, "rise_of_the_eldrazi.go": true,
		// "Exile Expropriate." (ADR 0146, #2143)
		"expropriate.go": true,
		// "Exile Travel Through Caradhras." (#2928)
		"travel_through_caradhras.go": true,
		// "Exile Yawgmoth's Vile Offering." (#2953)
		"yawgmoths_vile_offering.go": true,
		// "Exile Worst Fears." (#3857)
		"worst_fears.go": true,
		// "Exile Secret of Bloodbending." (#3999)
		"secret_of_bloodbending.go": true,
		// "Shuffle Beacon of Unrest into its owner's library." (#3209)
		"beacon_of_unrest.go": true,
		// "Exile Burning Wish." and "Exile Living Wish." (ADR 0191, #3506)
		"burning_wish.go": true, "living_wish.go": true,
		// "Exile Cunning Wish.", "Exile Glittering Wish.", "Exile Golden
		// Wish." and "Exile Death Wish." (#3525)
		"cunning_wish.go": true, "glittering_wish.go": true,
		"golden_wish.go": true, "death_wish.go": true,
		// "Exile Alchemist's Gambit." (ADR 0188, #3563)
		"alchemists_gambit.go": true,
		// "Put Approach of the Second Sun into its owner's library
		// seventh from the top." (#3654)
		"approach_of_the_second_sun.go": true,
		// "Exile The Needs of the Many." (#4973)
		"the_needs_of_the_many.go": true,
		// "return Haunting Imitation to its owner's hand." (S82 VOW b17)
		"haunting_imitation.go": true,
		// "Exile Revel in Silence." (S82 old b049, #5502)
		"flamescroll_celebrant_revel_in_silence.go": true,
		// "Exile Harness Infinity." (S82 STX b047, #5502)
		"harness_infinity.go": true,
		// "Exile Emergency Powers." (S82 old b179, #5521)
		"emergency_powers.go": true,
	}
	isSelf := func(e ast.Expr) bool {
		s := effectstest.ExprString(e)
		return s == "item.SourceCardID" || s == "it.SourceCardID" || s == "ctx.Source()" || s == "c.Source()"
	}
	files, err := effectstest.FlatGlob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var bad []string
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || spells[f] || f == "source_object_guard.go" {
			continue
		}
		file, err := effectstest.FlatParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		var stack []ast.Node
		ast.Inspect(file, func(n ast.Node) bool {
			if n == nil {
				stack = stack[:len(stack)-1]
				return true
			}
			stack = append(stack, n)
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !mutators[sel.Sel.Name] || !slices.ContainsFunc(call.Args, isSelf) {
				return true
			}
			// The innermost enclosing function must ask.
			for i := len(stack) - 1; i >= 0; i-- {
				var body *ast.BlockStmt
				switch fn := stack[i].(type) {
				case *ast.FuncLit:
					body = fn.Body
				case *ast.FuncDecl:
					body = fn.Body
				}
				if body == nil {
					continue
				}
				found := false
				ast.Inspect(body, func(m ast.Node) bool {
					if id, ok := m.(*ast.Ident); ok && asks[id.Name] {
						found = true
					}
					return !found
				})
				if !found {
					bad = append(bad, fset.Position(call.Pos()).String()+" "+sel.Sel.Name)
				}
				break
			}
			return true
		})
	}
	if len(bad) > 0 {
		t.Errorf("a card acts on its own source through a game mutator without asking whether that source is a new object (#1432):\n  %s\n"+
			"Use the primitive (it asks for you) or check SourceIsNewObject(g, item) first.", strings.Join(bad, "\n  "))
	}
}

// testHeardExile and testHeardExileTail let a test hear ExileTarget's keyed
// "then" (#4253 slice 9): the tail writes whether the card was exiled.
// testHeardThenTail is the same for any single-card door, under the tail's
// Word (#4253 retired SacrificePermanent's and BounceToHand's closure Then).
var (
	testHeardExile     map[string]bool
	testHeardExileTail = effects.CardTail("test/4253-heard-exile", func(_ *effects.Context, o game.TailOutcome) error {
		testHeardExile["exile"] = len(o.IDs) == 1
		return nil
	})
	testHeardThenTail = effects.CardTail("test/4253-heard-then", func(_ *effects.Context, o game.TailOutcome) error {
		testHeardExile[o.Tail.Word] = effects.OneWentThisWay(o)
		return nil
	})
)
