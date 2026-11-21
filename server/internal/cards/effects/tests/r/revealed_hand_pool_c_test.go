package testsr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func rhBear() game.Card {
	return game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2}
}

func rhNoPick(t *testing.T, g *game.Game) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDiscardFromHand {
			t.Fatalf("a revealed-hand prompt is open: %+v", c)
		}
	}
}

// A triggered pick: Grief enters, its controller targets an opponent,
// and the trigger's controller (CR 113.8) chooses a nonland card from
// that opponent's revealed hand. The prompt names Grief as its source.
func TestGriefEntersAndItsControllerChoosesANonlandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), rhBear())

	grief := effectstest.CastAndResolveCreature(t, g, "Grief", "Creature — Elemental Incarnation", effectstest.GriefOracle)
	effectstest.PickPlayer(t, g, me.ID, victim.ID)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.OpenPick(t, g)
	if pick.Chooser != me.ID || pick.FromPlayer != victim.ID || pick.Source != grief {
		t.Fatalf("prompt = chooser %v from %v source %v; want %v from %v source %v",
			pick.Chooser, pick.FromPlayer, pick.Source, me.ID, victim.ID, grief)
	}
	if !effectstest.SameIDs(pick.DiscardOptions, ids[1:]) {
		t.Errorf("options = %v, want the Bolt and the Bears %v", pick.DiscardOptions, ids[1:])
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("taking the Bears: %v", err)
	}
	if !victim.Graveyard.Contains(ids[2]) {
		t.Error("the Bears were not discarded")
	}
}

// Grief's trigger targets an opponent, never its controller.
func TestGriefCannotTargetItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastAndResolveCreature(t, g, "Grief", "Creature — Elemental Incarnation", effectstest.GriefOracle)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Grief's trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Error("Grief's trigger targeted its own controller")
	}
}

// An activated pick whose cost sacrifices its source: the Imp is in the
// graveyard by resolution, and the activator still chooses (CR 113.8).
// "Activate only as a sorcery" refuses it outside a main phase
// (CR 602.5d).
func TestPilferingImpSacrificesItselfAndItsActivatorChooses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt())
	imp := effectstest.PushCatalogPermanent(g, me.ID, "Pilfering Imp", "Creature — Imp", effectstest.PilferingImpOracle, false)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}}

	if g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain {
		t.Fatalf("the fixture starts in a main phase (%v); the sorcery-timing check needs another step", g.Turn.Step)
	}
	if err := g.ActivateCatalogAbility(me.ID, imp, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Fatal("Pilfering Imp activated outside a main phase")
	}
	if !g.Battlefield.Contains(imp) {
		t.Fatal("a refused activation sacrificed the Imp")
	}

	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, imp, 0, game.ActivateAbilityParams{Targets: target}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(imp) || !me.Graveyard.Contains(imp) {
		t.Fatal("the Imp was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.OpenPick(t, g)
	if pick.Chooser != me.ID || pick.FromPlayer != victim.ID || pick.Source != imp {
		t.Fatalf("prompt = chooser %v from %v source %v", pick.Chooser, pick.FromPlayer, pick.Source)
	}
	if !effectstest.SameIDs(pick.DiscardOptions, ids[1:]) {
		t.Errorf("options = %v, want only the Bolt", pick.DiscardOptions)
	}
}

// He Who Hungers: "Sacrifice a Spirit" takes only a Spirit, the pick
// has no filter (a land is a legal choice), and sacrificing itself
// triggers soulshift 4 (CR 702.46a), which returns a Spirit card of
// mana value 4 or less and offers nothing else.
func TestHeWhoHungersSacrificesASpiritAndSoulshifts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt())
	effectstest.AdvanceToMain(t, g)
	hungers := effectstest.PushCatalogPermanent(g, me.ID, "He Who Hungers", "Legendary Creature — Spirit", effectstest.HeWhoHungersOracle, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	smallSpirit := uuid.New()
	bigSpirit := uuid.New()
	notASpirit := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: smallSpirit, Name: "Small Spirit", TypeLine: "Creature — Spirit", ManaCost: "{3}{B}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: bigSpirit, Name: "Big Spirit", TypeLine: "Creature — Spirit", ManaCost: "{4}{B}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: notASpirit, Name: "Bear Card", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}}

	if err := g.ActivateCatalogAbility(me.ID, hungers, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear}, Targets: target,
	}); err == nil {
		t.Fatal("He Who Hungers sacrificed a creature that isn't a Spirit")
	}
	if err := g.ActivateCatalogAbility(me.ID, hungers, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{hungers}, Targets: target,
	}); err != nil {
		t.Fatalf("sacrificing itself: %v", err)
	}

	// Soulshift: "you may", then the target.
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no soulshift target prompt")
	}
	if !effectstest.HasID(p.PickTargetCards, smallSpirit) || effectstest.HasID(p.PickTargetCards, bigSpirit) || effectstest.HasID(p.PickTargetCards, notASpirit) {
		t.Fatalf("soulshift offered %v; want only the mana value 4 Spirit %v", p.PickTargetCards, smallSpirit)
	}
	effectstest.PickCard(t, g, me.ID, smallSpirit)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(smallSpirit) {
		t.Error("soulshift did not return the Spirit card")
	}

	pick := effectstest.OpenPick(t, g)
	if pick.Chooser != me.ID || !effectstest.SameIDs(pick.DiscardOptions, ids) {
		t.Errorf("prompt = chooser %v options %v; want %v choosing from %v", pick.Chooser, pick.DiscardOptions, me.ID, ids)
	}
}

// Entomber Exarch's second bullet: a noncreature card, so a land is a
// legal pick and the creature is not.
func TestEntomberExarchChoosesANoncreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), rhBear(), effectstest.RhBolt())

	effectstest.CastAndResolveCreature(t, g, "Entomber Exarch", "Creature — Phyrexian Cleric", effectstest.EntomberExarchOracle)
	mode := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt from the enters trigger: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("choosing the discard bullet: %v", err)
	}
	effectstest.PickPlayer(t, g, me.ID, victim.ID)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.OpenPick(t, g)
	if want := []uuid.UUID{ids[0], ids[2]}; !effectstest.SameIDs(pick.DiscardOptions, want) {
		t.Errorf("options = %v, want the Forest and the Bolt %v", pick.DiscardOptions, want)
	}
	if pick.DiscardLabel != "noncreature card" || pick.Chooser != me.ID {
		t.Errorf("label %q chooser %v", pick.DiscardLabel, pick.Chooser)
	}
}

// Entomber Exarch's first bullet returns a creature card from your
// graveyard.
func TestEntomberExarchReturnsACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := effectstest.PushGraveyardCardForTest(me, "Dead Bear")

	effectstest.CastAndResolveCreature(t, g, "Entomber Exarch", "Creature — Phyrexian Cleric", effectstest.EntomberExarchOracle)
	mode := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatal("no mode prompt from the enters trigger")
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("choosing the return bullet: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
}

// Deceit's two intervening ifs (CR 603.4) read the mana that paid for
// it: {U}{U} bounces, {B}{B} takes a card, one of each does neither.
func TestDeceitTriggersOnTheManaSpentToCastIt(t *testing.T) {
	for _, tc := range []struct {
		name          string
		pool          string
		bounce, takes bool
	}{
		{"blue and blue", "UUCCCC", true, false},
		{"black and black", "BBCCCC", false, true},
		{"one of each", "UBCCCC", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt())
			rock := effectstest.PushTypedCard(g, victim.ID, "Rock", "Artifact", "{1}")
			effectstest.AdvanceToMain(t, g)
			effectstest.FloatForTest(g, me, tc.pool)

			deceit := effectstest.CastFromHandForTest(t, g, me, "Deceit", "Creature — Elemental Incarnation",
				"{4}{U/B}{U/B}", effectstest.DeceitOracle, game.CastSpellParams{Strict: true})
			for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
				if err := g.PassPriority(); err != nil {
					if errors.Is(err, game.ErrChoicePending) {
						break
					}
					t.Fatalf("PassPriority: %v", err)
				}
			}
			if !g.Battlefield.Contains(deceit) {
				t.Fatal("Deceit did not resolve")
			}

			var bounced, took bool
			for i := 0; i < 4; i++ {
				p := effectstest.LatestPickTarget(g, me.ID)
				if p == nil {
					break
				}
				if effectstest.HasID(p.PickTargetCards, rock) {
					bounced = true
					effectstest.PickCard(t, g, me.ID, rock)
				} else {
					took = true
					effectstest.PickPlayer(t, g, me.ID, victim.ID)
				}
			}
			effectstest.PassPriorityAroundTable(t, g)
			if bounced != tc.bounce || took != tc.takes {
				t.Fatalf("bounce trigger %v, pick trigger %v; want %v, %v", bounced, took, tc.bounce, tc.takes)
			}
			if tc.bounce && g.Battlefield.Contains(rock) {
				t.Error("the {U}{U} trigger did not return the artifact")
			}
			if tc.takes {
				if pick := effectstest.OpenPick(t, g); pick.Chooser != me.ID || pick.FromPlayer != victim.ID {
					t.Errorf("prompt = chooser %v from %v", pick.Chooser, pick.FromPlayer)
				}
			} else {
				rhNoPick(t, g)
			}
		})
	}
}

// Elspeth's Nightmare through the Saga lifecycle: chapter I offers
// only an opponent's creature with power 2 or less, and chapter II, a
// turn later, takes a noncreature, nonland card. The hand is set after
// the turn cycle, since the chapter reads it as it resolves.
func TestElspethsNightmareChaptersOneAndTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	victim := g.Seats[(seat+1)%len(g.Seats)]
	small := effectstest.PushVanillaCreature(g, victim.ID, "Small", 2, 2)
	big := effectstest.PushVanillaCreature(g, victim.ID, "Big", 3, 3)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)

	effectstest.CastCatalogSpell(t, g, "Elspeth's Nightmare", "Enchantment — Saga", effectstest.ElspethsNightmareOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no chapter I target prompt")
	}
	if !effectstest.HasID(p.PickTargetCards, small) || effectstest.HasID(p.PickTargetCards, big) || effectstest.HasID(p.PickTargetCards, mine) {
		t.Fatalf("chapter I offered %v; want only the power-2 creature an opponent controls", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Fatal("chapter I did not destroy the creature")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), rhBear(), effectstest.RhBolt())
	effectstest.PickPlayer(t, g, me.ID, victim.ID)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenPick(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, ids[2:]) || pick.DiscardLabel != "noncreature, nonland card" || pick.Chooser != me.ID {
		t.Errorf("options %v label %q chooser %v; want only the Bolt", pick.DiscardOptions, pick.DiscardLabel, pick.Chooser)
	}
}
