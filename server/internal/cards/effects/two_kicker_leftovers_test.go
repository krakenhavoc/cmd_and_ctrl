package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// two_kicker_leftovers_test.go — #2353: the "Kicker [A] and/or [B]"
// cards #2333 left over, and the populate cards #2318 held back. One
// test per card; the two-kicker cards run every combination (neither,
// each alone, both).

const (
	thunderscapeBattlemageOracle = "9aa66e7d-d5e3-4c48-9c68-b61ff2af0b1a"
	anaBattlemageOracle          = "c37537ae-616a-41e4-890c-f4b9475dbf31"
	stormscapeBattlemageOracle   = "38ee748d-adcd-41df-9b23-d2a34829784c"
	nightscapeBattlemageOracle   = "e96e68b4-cb32-4b75-a885-e1781f4b74ab"
	wastescapeBattlemageOracle   = "3cc6475d-aae9-4592-a26b-0fbb4269759d"
	strongholdArenaOracle        = "7dabd64c-8d0b-4e56-b6be-5651c9e985c5"
	illuminateOracle             = "f944c245-08ff-41a8-bdfc-7d21a6a63eae"
	vodalianMindsingerOracle     = "3823afcf-6b15-46a9-ba49-3a9a39f8114e"
	temporalFirestormOracle      = "a0f25ffa-0e53-4e38-8001-d595fd2c2045"
	arborealAllianceOracle       = "2f497352-c812-48c4-a922-cc311bcdcdc6"
	musterTheDepartedOracle      = "f0ba9cca-b946-465c-acae-0b465d7e1622"
)

// lftKickerCombos are the four ways to pay a two-kicker card, as
// positions in its OptionalCosts: neither, the first, the second, both.
var lftKickerCombos = []struct {
	name     string
	optional []int
	first    bool
	second   bool
}{
	{"unkicked", nil, false, false},
	{"first kicker", []int{0}, true, false},
	{"second kicker", []int{1}, false, true},
	{"both kickers", []int{0, 1}, true, true},
}

// lftSettle resolves everything on the stack, answering each of
// `me`'s pick_target prompts with `pick`. It looks for a prompt before
// each pass, because a cast trigger asks as the spell is cast. Returns
// how many prompts were answered.
func lftSettle(t *testing.T, g *game.Game, me uuid.UUID, pick func(p *game.PendingChoice) []game.TargetRef) int {
	t.Helper()
	asked := 0
	for i := 0; i < 24; i++ {
		if c := openTriggerOrder(g, me); c != nil {
			if err := g.ResolveTriggerOrder(c.ID, me, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := latestPickTarget(g, me); p != nil {
			if err := g.ResolvePickTargets(p.ID, me, pick(p)); err != nil {
				t.Fatalf("ResolvePickTargets(%q): %v", p.Reason, err)
			}
			asked++
			continue
		}
		if stackFullyEmpty(g) {
			return asked
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("the stack did not settle")
	return asked
}

// pickPlayer answers every prompt with `id` (a player).
func lftPickPlayer(id uuid.UUID) func(*game.PendingChoice) []game.TargetRef {
	return func(*game.PendingChoice) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
	}
}

// lftPickCards answers every prompt with the offered cards among `want`.
func lftPickCards(want ...uuid.UUID) func(*game.PendingChoice) []game.TargetRef {
	return func(p *game.PendingChoice) []game.TargetRef {
		var out []game.TargetRef
		for _, id := range want {
			if hasID(p.PickTargetCards, id) {
				out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
			}
		}
		return out
	}
}

func lftSeed(g *game.Game, controller uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Power: power, Toughness: toughness,
		Owner: controller, Controller: controller,
	})
}

func lftHandCards(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

// lftCastKicked is castXSpell with the optional costs claimed.
func lftCastKicked(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, x int, targets []game.TargetRef, optional []int) uuid.UUID {
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
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets, XValue: x, OptionalCosts: optional}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func TestTwoKickerLeftoversDeclareTwoKickersAsPrinted(t *testing.T) {
	for oracle, want := range map[string][2]string{
		thunderscapeBattlemageOracle: {"{1}{B}", "{G}"},
		anaBattlemageOracle:          {"{2}{U}", "{1}{B}"},
		stormscapeBattlemageOracle:   {"{W}", "{2}{B}"},
		nightscapeBattlemageOracle:   {"{2}{U}", "{2}{R}"},
		wastescapeBattlemageOracle:   {"{G}", "{1}{U}"},
		strongholdArenaOracle:        {"{G}", "{W}"},
		illuminateOracle:             {"{2}{R}", "{3}{U}"},
		vodalianMindsingerOracle:     {"{1}{R}", "{1}{G}"},
		temporalFirestormOracle:      {"{1}{W}", "{1}{U}"},
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: Completeness = %s, want full", spec.Name, spec.Completeness)
		}
		costs := game.OptionalCostsFor(oracle)
		if len(costs) != 2 || costs[0].ManaCost != want[0] || costs[1].ManaCost != want[1] {
			t.Errorf("%s: optional costs = %+v, want kickers %v in printed order", spec.Name, costs, want)
		}
	}
}

func TestThunderscapeBattlemageEachKickerTriggersItsOwnAbility(t *testing.T) {
	for _, tc := range lftKickerCombos {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			aura := lftSeed(g, victim.ID, "Pacifism", "Enchantment — Aura", 0, 0)
			lftHandCards(victim, 3)
			if _, err := castWithOptionalCosts(t, g, "Thunderscape Battlemage", "Creature — Human Wizard",
				thunderscapeBattlemageOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			asked := settleAnsweringPicks(t, g, me.ID, victim.ID)
			if want := len(tc.optional); asked != want {
				t.Errorf("answered %d target prompts, want %d — one per kicker paid", asked, want)
			}
			owed := 0
			if tc.first {
				owed = 2
			}
			if got := discardOwed(g, victim.ID); got != owed {
				t.Errorf("victim owes %d discards, want %d", got, owed)
			}
			if gone := findBattlefieldCardByID(g, aura) == nil; gone != tc.second {
				t.Errorf("enchantment destroyed = %v, want %v", gone, tc.second)
			}
		})
	}
}

func TestAnaBattlemageEachKickerTriggersItsOwnAbility(t *testing.T) {
	for _, tc := range lftKickerCombos {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			ogre := lftSeed(g, victim.ID, "Ogre", "Creature — Ogre", 3, 3)
			lftHandCards(victim, 4)
			life := victim.Life
			if _, err := castWithOptionalCosts(t, g, "Ana Battlemage", "Creature — Human Wizard",
				anaBattlemageOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			pick := func(p *game.PendingChoice) []game.TargetRef {
				if len(p.PickTargetPlayers) > 0 {
					return lftPickPlayer(victim.ID)(p)
				}
				return lftPickCards(ogre)(p)
			}
			if asked := lftSettle(t, g, me.ID, pick); asked != len(tc.optional) {
				t.Errorf("answered %d target prompts, want %d", asked, len(tc.optional))
			}
			owed := 0
			if tc.first {
				owed = 3
			}
			if got := discardOwed(g, victim.ID); got != owed {
				t.Errorf("victim owes %d discards, want %d", got, owed)
			}
			tapped := findBattlefieldCardByID(g, ogre).Tapped
			if tapped != tc.second {
				t.Errorf("ogre tapped = %v, want %v", tapped, tc.second)
			}
			wantLoss := 0
			if tc.second {
				wantLoss = 3
			}
			if got := life - victim.Life; got != wantLoss {
				t.Errorf("victim lost %d life, want %d — the ogre deals its power to its controller", got, wantLoss)
			}
		})
	}
}

func TestAnaBattlemageTapsOnlyAnUntappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	tapped := lftSeed(g, victim.ID, "Sleepy Ogre", "Creature — Ogre", 3, 3)
	g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Tapped = true
	if _, err := castWithOptionalCosts(t, g, "Ana Battlemage", "Creature — Human Wizard",
		anaBattlemageOracle, nil, []int{1}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt: Ana herself is an untapped creature")
	}
	if hasID(p.PickTargetCards, tapped) {
		t.Errorf("a tapped creature is offered as \"target untapped creature\"")
	}
}

func TestStormscapeBattlemageEachKickerTriggersItsOwnAbility(t *testing.T) {
	for _, tc := range lftKickerCombos {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			bear := lftSeed(g, victim.ID, "Bear", "Creature — Bear", 2, 2)
			g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Colors = []string{"G"}
			shade := lftSeed(g, victim.ID, "Shade", "Creature — Shade", 1, 1)
			g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Colors = []string{"B"}
			life := me.Life
			if _, err := castWithOptionalCosts(t, g, "Stormscape Battlemage", "Creature — Metathran Wizard",
				stormscapeBattlemageOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			lftSettle(t, g, me.ID, func(p *game.PendingChoice) []game.TargetRef {
				if hasID(p.PickTargetCards, shade) {
					t.Errorf("a black creature is offered as \"target nonblack creature\"")
				}
				return lftPickCards(bear)(p)
			})
			wantLife := 0
			if tc.first {
				wantLife = 3
			}
			if got := me.Life - life; got != wantLife {
				t.Errorf("caster gained %d life, want %d", got, wantLife)
			}
			if gone := findBattlefieldCardByID(g, bear) == nil; gone != tc.second {
				t.Errorf("nonblack creature destroyed = %v, want %v", gone, tc.second)
			}
			if findBattlefieldCardByID(g, shade) == nil {
				t.Error("the black creature was destroyed")
			}
		})
	}
}

func TestNightscapeBattlemageEachKickerTriggersItsOwnAbility(t *testing.T) {
	for _, tc := range lftKickerCombos {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			a := lftSeed(g, victim.ID, "Bear A", "Creature — Bear", 2, 2)
			b := lftSeed(g, victim.ID, "Bear B", "Creature — Bear", 2, 2)
			forest := lftSeed(g, victim.ID, "Forest", "Basic Land — Forest", 0, 0)
			if _, err := castWithOptionalCosts(t, g, "Nightscape Battlemage", "Creature — Zombie Wizard",
				nightscapeBattlemageOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			lftSettle(t, g, me.ID, lftPickCards(a, b, forest))
			// "Up to two" with two chosen: both come home. The land is the
			// {2}{R} clause's target, so a prompt that offers it takes it.
			bounced := findBattlefieldCardByID(g, a) == nil && findBattlefieldCardByID(g, b) == nil
			if bounced != tc.first {
				t.Errorf("both creatures returned to hand = %v, want %v", bounced, tc.first)
			}
			if gone := findBattlefieldCardByID(g, forest) == nil; gone != tc.second {
				t.Errorf("land destroyed = %v, want %v", gone, tc.second)
			}
			if tc.first && victim.Hand.Size() < 2 {
				t.Errorf("victim's hand has %d cards, want the two bounced creatures", victim.Hand.Size())
			}
		})
	}
}

func TestWastescapeBattlemageCastTriggersAreLinkedToTheirKickers(t *testing.T) {
	for _, tc := range lftKickerCombos {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			relic := lftSeed(g, victim.ID, "Relic", "Artifact", 0, 0)
			ogre := lftSeed(g, victim.ID, "Ogre", "Creature — Ogre", 3, 3)
			mine := lftSeed(g, me.ID, "My Relic", "Artifact", 0, 0)
			if _, err := castWithOptionalCosts(t, g, "Wastescape Battlemage", "Creature — Eldrazi Wizard",
				wastescapeBattlemageOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			// Cast triggers ask as the spell is cast, with the Battlemage
			// still on the stack.
			if tc.first || tc.second {
				if p := latestPickTarget(g, me.ID); p == nil {
					t.Fatal("no target prompt for a kicked cast trigger")
				} else if hasID(p.PickTargetCards, mine) {
					t.Error("\"an opponent controls\" offers the caster's own artifact")
				}
			}
			asked := lftSettle(t, g, me.ID, lftPickCards(relic, ogre))
			if want := len(tc.optional); asked != want {
				t.Errorf("answered %d target prompts, want %d", asked, want)
			}
			if exiled := findBattlefieldCardByID(g, relic) == nil; exiled != tc.first {
				t.Errorf("artifact exiled = %v, want %v", exiled, tc.first)
			}
			if bounced := findBattlefieldCardByID(g, ogre) == nil; bounced != tc.second {
				t.Errorf("creature bounced = %v, want %v", bounced, tc.second)
			}
			if findBattlefieldCardByID(g, mine) == nil {
				t.Error("the caster's own artifact was exiled")
			}
		})
	}
}

// A cast trigger is not an enters trigger: with the Battlemage
// countered the kicked ability still resolves.
func TestWastescapeBattlemageKickedAbilityResolvesEvenIfCountered(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ogre := lftSeed(g, victim.ID, "Ogre", "Creature — Ogre", 3, 3)
	spell, err := castWithOptionalCosts(t, g, "Wastescape Battlemage", "Creature — Eldrazi Wizard",
		wastescapeBattlemageOracle, nil, []int{1}, nil)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the kicked cast trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: ogre}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	counterID := uuid.New()
	victim.Hand.PushTop(game.Card{
		InstanceID: counterID, Name: "Counterspell", TypeLine: "Instant",
		OracleID: "cc187110-1148-4090-bbb8-e205694a39f5", Owner: victim.ID, Controller: victim.ID,
	})
	if err := g.CastSpell(victim.ID, counterID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("CastSpell Counterspell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(spell) {
		t.Error("the Battlemage was not countered")
	}
	if findBattlefieldCardByID(g, ogre) != nil {
		t.Error("the kicked cast trigger did not resolve after the Battlemage was countered")
	}
}

func TestStrongholdArenaGainsThreePerKick(t *testing.T) {
	for _, tc := range lftKickerCombos {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			life := me.Life
			if _, err := castWithOptionalCosts(t, g, "Stronghold Arena", "Enchantment",
				strongholdArenaOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got, want := me.Life-life, 3*len(tc.optional); got != want {
				t.Errorf("gained %d life, want %d — 3 per time kicked", got, want)
			}
		})
	}
}

func TestStrongholdArenaFlipsAndLosesLifeEqualToManaValue(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		b12Push(g, me.ID, "Stronghold Arena", "Enchantment", strongholdArenaOracle, 0, 0)
		attacker := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Expensive", TypeLine: "Sorcery", ManaCost: "{3}{R}",
			Owner: me.ID, Controller: me.ID,
		})
		hand, life := me.Hand.Size(), me.Life

		b784AttackEach(t, g, [2]uuid.UUID{attacker, g.Seats[1].ID})
		answerLatestTriggerPrompt(t, g, me.ID, accept)
		passPriorityAroundTable(t, g)

		wantHand, wantLoss := 0, 0
		if accept {
			wantHand, wantLoss = 1, 4
		}
		if got := me.Hand.Size() - hand; got != wantHand {
			t.Errorf("accept=%v: hand grew by %d, want %d", accept, got, wantHand)
		}
		if got := life - me.Life; got != wantLoss {
			t.Errorf("accept=%v: lost %d life, want %d (the card's mana value)", accept, got, wantLoss)
		}
	}
}

func TestIlluminateKickersAreLinkedToTheirClauses(t *testing.T) {
	for _, tc := range lftKickerCombos {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			ogre := lftSeed(g, victim.ID, "Ogre", "Creature — Ogre", 5, 5)
			life := victim.Life
			hand := me.Hand.Size()
			lftCastKicked(t, g, "Illuminate", "Sorcery", illuminateOracle, "{X}{R}", 3,
				[]game.TargetRef{{Kind: game.TargetCard, ID: ogre}}, tc.optional)
			passPriorityAroundTable(t, g)
			if got := findBattlefieldCardByID(g, ogre); got == nil || got.DamageMarked != 3 {
				t.Errorf("target creature = %+v, want it on the battlefield with 3 damage", got)
			}
			wantLoss := 0
			if tc.first {
				wantLoss = 3
			}
			if got := life - victim.Life; got != wantLoss {
				t.Errorf("controller lost %d life, want %d (the {2}{R} kicker)", got, wantLoss)
			}
			// The spell leaves the hand and its card is on the stack or in the
			// graveyard, so the net change is X cards drawn minus the cast.
			wantDraw := 0
			if tc.second {
				wantDraw = 3
			}
			if got := me.Hand.Size() - hand; got != wantDraw {
				t.Errorf("hand grew by %d, want %d (the {3}{U} kicker draws X)", got, wantDraw)
			}
		})
	}
}

func TestIlluminateDamageToControllerIsReadBeforeTheCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	bear := lftSeed(g, victim.ID, "Bear", "Creature — Bear", 2, 2)
	life := victim.Life
	lftCastKicked(t, g, "Illuminate", "Sorcery", illuminateOracle, "{X}{R}", 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}, []int{0})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, bear) != nil {
		t.Error("the 2/2 survived 2 damage")
	}
	if got := life - victim.Life; got != 2 {
		t.Errorf("controller lost %d life, want 2 even though the creature died", got)
	}
}

func TestVodalianMindsingerCountersAndTheft(t *testing.T) {
	for _, tc := range lftKickerCombos {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			// Power 3 is stealable only once the Mindsinger has grown past it.
			giant := lftSeed(g, victim.ID, "Giant", "Creature — Giant", 3, 3)
			lftCastSizedCreature(t, g, "Vodalian Mindsinger", "Creature — Merfolk Wizard", vodalianMindsingerOracle, 2, tc.optional)
			lftSettle(t, g, me.ID, lftPickCards(giant))
			mind := lftFindNamed(g, "Vodalian Mindsinger")
			if mind == nil {
				t.Fatal("the Mindsinger did not enter")
			}
			wantPower := 2 + 2*len(tc.optional)
			if got := mind.CurrentPower(); got != wantPower {
				t.Errorf("power = %d, want %d — two counters per time kicked", got, wantPower)
			}
			stolen := findBattlefieldCardByID(g, giant).Controller == me.ID
			if want := wantPower > 3; stolen != want {
				t.Errorf("3-power creature stolen = %v, want %v at power %d", stolen, want, wantPower)
			}
		})
	}
}

func TestVodalianMindsingerTheftEndsWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := lftSeed(g, victim.ID, "Bear", "Creature — Bear", 1, 1)
	lftCastSizedCreature(t, g, "Vodalian Mindsinger", "Creature — Merfolk Wizard", vodalianMindsingerOracle, 2, nil)
	lftSettle(t, g, me.ID, lftPickCards(bear))
	if findBattlefieldCardByID(g, bear).Controller != me.ID {
		t.Fatal("the 1-power creature was not stolen")
	}
	mind := lftFindNamed(g, "Vodalian Mindsinger")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mind.InstanceID, game.DestroyOptions{}) })
	if got := layeredCard(t, g, bear).Controller; got != victim.ID {
		t.Errorf("the creature stayed with %s after the Mindsinger died", got)
	}
}

// lftCastSizedCreature casts a creature whose printed P/T is a square
// `size`, so a power read off it means something.
func lftCastSizedCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, size int, optional []int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: size, Toughness: size, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{OptionalCosts: optional}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func lftFindNamed(g *game.Game, name string) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == name {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func TestTemporalFirestormPhasesOutWhatTheKickersChoose(t *testing.T) {
	for _, tc := range lftKickerCombos {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			mine1 := lftSeed(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			mine2 := lftSeed(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
			theirs := lftSeed(g, victim.ID, "Their Bear", "Creature — Bear", 2, 2)
			walker := lftSeed(g, victim.ID, "Their Walker", "Planeswalker — Test", 0, 0)
			lftCastKicked(t, g, "Temporal Firestorm", "Sorcery", temporalFirestormOracle, "{3}{R}{R}", 0, nil, tc.optional)
			passPriorityAroundTable(t, g)
			x := len(tc.optional)
			if x > 0 {
				c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
				if c == nil {
					t.Fatalf("no choice offered for up to %d permanents", x)
				}
				pick := []uuid.UUID{mine1, mine2}[:x]
				if err := g.ResolveOwnPermanents(c.ID, me.ID, pick); err != nil {
					t.Fatalf("ResolveOwnPermanents: %v", err)
				}
				passPriorityAroundTable(t, g)
				for _, id := range pick {
					if !phasedOutInCatalogGame(g, id) {
						t.Errorf("a chosen permanent did not phase out")
					}
				}
			}
			survivors := 0
			for _, id := range []uuid.UUID{mine1, mine2} {
				if phasedOutInCatalogGame(g, id) {
					survivors++
				} else if findBattlefieldCardByID(g, id) != nil {
					t.Errorf("an unchosen creature took 5 damage and lived")
				}
			}
			if survivors != x {
				t.Errorf("%d of my creatures phased out, want %d (X = times kicked)", survivors, x)
			}
			if findBattlefieldCardByID(g, theirs) != nil {
				t.Error("an unchosen opposing creature took 5 and lived")
			}
			if findBattlefieldCardByID(g, walker) != nil {
				t.Error("a planeswalker with no loyalty counters survived the 5 damage")
			}
		})
	}
}

func TestArborealAllianceMakesAnXByXTreefolk(t *testing.T) {
	for _, x := range []int{0, 3} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		lftCastKicked(t, g, "Arboreal Alliance", "Enchantment", arborealAllianceOracle, "{X}{G}{G}", x, nil, nil)
		passPriorityAroundTable(t, g)
		trees := populateTokens(g, me.ID, "Treefolk")
		if x == 0 {
			if len(trees) != 0 {
				t.Errorf("X=0 made %d tokens, want none", len(trees))
			}
			continue
		}
		if len(trees) != 1 {
			t.Fatalf("X=%d made %d Treefolk, want 1", x, len(trees))
		}
		if c := findBattlefieldCardByID(g, trees[0]); c.CurrentPower() != x || c.CurrentToughness() != x {
			t.Errorf("Treefolk is %d/%d, want %d/%d", c.CurrentPower(), c.CurrentToughness(), x, x)
		}
	}
}

func TestArborealAlliancePopulatesOnlyWhenAnElfAttacks(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		want     int
	}{
		{"an Elf attacks", "Creature — Elf Warrior", 2},
		{"a Bear attacks", "Creature — Bear", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			b12Push(g, me.ID, "Arboreal Alliance", "Enchantment", arborealAllianceOracle, 0, 0)
			pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
			attacker := b12Creature(g, me.ID, "Attacker", tc.typeLine, 2, 2)

			b784AttackEach(t, g, [2]uuid.UUID{attacker, g.Seats[1].ID})
			passPriorityAroundTable(t, g)
			if got := len(populateTokens(g, me.ID, "Soldier")); got != tc.want {
				t.Errorf("Soldiers = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestArborealAllianceOneTriggerHoweverManyElvesAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Arboreal Alliance", "Enchantment", arborealAllianceOracle, 0, 0)
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	a := b12Creature(g, me.ID, "Elf A", "Creature — Elf", 1, 1)
	b := b12Creature(g, me.ID, "Elf B", "Creature — Elf", 1, 1)

	b784AttackEach(t, g, [2]uuid.UUID{a, g.Seats[1].ID}, [2]uuid.UUID{b, g.Seats[1].ID})
	passPriorityAroundTable(t, g)
	if got := len(populateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Errorf("Soldiers = %d, want 2 — \"one or more Elves\" populates once", got)
	}
}

func TestMusterTheDepartedMakesASpiritAndMorbidPopulates(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		castCatalogSpell(t, g, "Muster the Departed", "Enchantment", musterTheDepartedOracle, nil)
		passPriorityAroundTable(t, g)
		if got := len(populateTokens(g, me.ID, "Spirit")); got != 1 {
			t.Fatalf("Spirits after entering = %d, want 1", got)
		}
		if died {
			g.TurnTally.CreaturesDied = 1
		}
		advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
		passPriorityAroundTable(t, g)
		want := 1
		if died {
			want = 2
		}
		if got := len(populateTokens(g, me.ID, "Spirit")); got != want {
			t.Errorf("creature died this turn = %v: Spirits = %d, want %d", died, got, want)
		}
	}
}
