package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_15_cards_test.go — Marvel Super Heroes Commander, fifteenth batch
// (#4455). Each test plays the real catalog entry and asserts what its text
// says.

const (
	msc15LoveOracle     = "645dbd86-fe30-4897-a3ca-633f07ec5b54"
	msc15TorchOracle    = "d55cb52d-b35c-4602-a518-cd51f155a2d1"
	msc15KangOracle     = "4e420f43-a4a3-44bb-9a4e-a943f93a3534"
	msc15HawkeyeOracle  = "e90d6903-5415-4ddf-88c0-15ddd8614bde"
	msc15HerbieOracle   = "064a2f4a-ff3f-4622-8727-24b22965416d"
	msc15ExtremisOracle = "b058ee2e-ddb9-4f49-bbe0-16b72d54d4da"
	msc15GammaOracle    = "b92b4894-0786-49a7-8003-d99393f080d4"
	msc15ReturnOracle   = "4b7a8c05-15cd-452e-b9fc-83800e81793d"
	msc15DoombotOracle  = "37042a3e-c377-4282-80f9-068e4955f8a1"
	msc15HerbOracle     = "20418984-bc6d-42e5-ada6-30b5281925c0"
)

func TestMsc15EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc15LoveOracle:     "Love on the Battlefield",
		msc15TorchOracle:    "Human Torch",
		msc15KangOracle:     "Kang Dynasty",
		msc15HawkeyeOracle:  "Hawkeye, Avenging Archer",
		msc15HerbieOracle:   "H.E.R.B.I.E., Lovable Robot",
		msc15ExtremisOracle: "Extremis Elite",
		msc15GammaOracle:    "Gamma Grotesque",
		msc15ReturnOracle:   "Heroic Return",
		msc15DoombotOracle:  "Doombot Harbinger",
		msc15HerbOracle:     "Heart-Shaped Herb",
	}
	for oracle, name := range want {
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

// msc15PendingDelayed counts the delayed triggers carrying a label.
func msc15PendingDelayed(g *game.Game, label string) int {
	n := 0
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Label == label {
			n++
		}
	}
	return n
}

// --- Love on the Battlefield -----------------------------------------------------------

func msc15LoveTable(t *testing.T) (g *game.Game, me, opp *game.Player, love, a, b uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	love = effectstest.B27Push(g, me.ID, "Love on the Battlefield", "Enchantment", msc15LoveOracle, "{1}{U}{R}", 0, 0)
	a = effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b = effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 3, 3)
	return
}

func TestMsc15LoveTwoAttackersGetFirstStrikeDrawAndGrowWhenTheyConnect(t *testing.T) {
	g, me, opp, _, a, b := msc15LoveTable(t)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, a, "first strike")
	effectstest.AssertKeywords(t, g, b, "first strike")
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("cards drawn = %d, want 1", got)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Bear A": a, "Bear B": b} {
		if got := effectstest.PlusOnes(g, id); got != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", name, got)
		}
	}
	if opp.Life != 40-5 {
		t.Errorf("opponent life = %d, want 35", opp.Life)
	}
}

func TestMsc15LoveABlockedAttackerDoesNotGrowAndTheWatchEndsWithCombat(t *testing.T) {
	g, me, opp, _, a, b := msc15LoveTable(t)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, a); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, a); got != 0 {
		t.Errorf("the blocked attacker has %d counters, want 0", got)
	}
	if got := effectstest.PlusOnes(g, b); got != 1 {
		t.Errorf("the unblocked attacker has %d counters, want 1", got)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if n := msc15PendingDelayed(g, "Love on the Battlefield — put a +1/+1 counter on it"); n != 0 {
		t.Errorf("%d watches survive the combat, want 0", n)
	}
	_ = me
}

func TestMsc15LoveOnlyExactlyTwoAttackersTrigger(t *testing.T) {
	g, me, opp, _, a, b := msc15LoveTable(t)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 1, 1)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a, b, c)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("three attackers drew %d cards, want 0", me.Hand.Size()-hand)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, a) + effectstest.PlusOnes(g, b) + effectstest.PlusOnes(g, c); got != 0 {
		t.Errorf("counters = %d, want 0 with three attackers", got)
	}

	g, me, opp, _, a, _ = msc15LoveTable(t)
	hand = me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("one attacker drew %d cards, want 0", me.Hand.Size()-hand)
	}
}

// --- Human Torch -----------------------------------------------------------------------

func msc15Torch(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Human Torch", "Legendary Creature — Human Hero", msc15TorchOracle, 3, 2)
}

func TestMsc15TorchGainsFlyingDoubleStrikeAndHasteAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	torch := msc15Torch(g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveAbilities(t, g, torch); len(got) != 0 {
		t.Errorf("abilities without a noncreature spell: %v, want none", got)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	torch = msc15Torch(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, torch, "flying", "double strike", "haste")
}

func TestMsc15TorchPaidAttackSharesCombatDamageWithTheOtherOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	torch := msc15Torch(g, me.ID)
	effectstest.DeclareAttack(t, g, o1.ID, torch)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment prompt after Human Torch attacked")
	}
	effectstest.FloatMana(t, g, me, "{R}{G}{W}{U}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	for name, p := range map[string]*game.Player{"first": o1, "second": o2, "third": o3} {
		if p.Life != 40-3 {
			t.Errorf("the %s opponent has %d life, want 37 (3 damage each)", name, p.Life)
		}
	}
	if me.Life != 40 {
		t.Errorf("my life = %d, want 40", me.Life)
	}
}

func TestMsc15TorchDecliningOrShortOfManaSharesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	torch := msc15Torch(g, me.ID)
	effectstest.DeclareAttack(t, g, o1.ID, torch)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if o1.Life != 37 || o2.Life != 40 {
		t.Errorf("lives %d and %d, want 37 and 40", o1.Life, o2.Life)
	}

	g = effectstest.NewCatalogGame(t)
	me, o1, o2 = g.Seats[0], g.Seats[1], g.Seats[2]
	torch = msc15Torch(g, me.ID)
	effectstest.DeclareAttack(t, g, o1.ID, torch)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{R}{G}{W}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if o1.Life != 37 || o2.Life != 40 {
		t.Errorf("underpaid: lives %d and %d, want 37 and 40", o1.Life, o2.Life)
	}
}

// --- Kang Dynasty ----------------------------------------------------------------------

func TestMsc15KangDynastyTapsGoadsAndDrawsWhenAGoadedCreatureConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	c1 := effectstest.B12Creature(g, o1.ID, "Their Bear", "Creature — Bear", 2, 2)
	c2 := effectstest.B12Creature(g, o2.ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Kang Dynasty", "Enchantment — Saga", msc15KangOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID, c1, c2)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"first": c1, "second": c2} {
		if !effectstest.IsTapped(g, id) {
			t.Errorf("the %s opponent's creature is untapped", name)
		}
		if c := effectstest.B12Card(t, g, id); len(c.Goads) == 0 {
			t.Errorf("the %s opponent's creature is not goaded", name)
		}
	}
	if n := msc15PendingDelayed(g, "Kang Dynasty — draw a card"); n != 2 {
		t.Fatalf("draw watches = %d, want 2", n)
	}
	// The goaded creature untaps on its controller's turn and connects with
	// another player: that is "a player", not only me.
	hand := me.Hand.Size()
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(c1, o3(g).ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("cards drawn when the goaded creature connected = %d, want 1", got)
	}
}

func o3(g *game.Game) *game.Player { return g.Seats[3] }

func TestMsc15KangDynastyWatchEndsAsMyNextTurnBegins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	c1 := effectstest.B12Creature(g, o1.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Kang Dynasty", "Enchantment — Saga", msc15KangOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID, c1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := msc15PendingDelayed(g, "Kang Dynasty — draw a card"); n != 1 {
		t.Fatalf("draw watches = %d, want 1", n)
	}
	for _, dt := range g.DelayedTriggers {
		if dt == nil || dt.Label != "Kang Dynasty — draw a card" {
			continue
		}
		if dt.Duration == nil || dt.Duration.Kind != game.UntilYourNextTurn || dt.Duration.Player != me.ID {
			t.Errorf("the watch lasts %+v, want until my next turn", dt.Duration)
		}
	}
}

func TestMsc15KangDynastyChapterThreePumpsByHandSizeAndCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, o1.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.CastCatalogSpell(t, g, "Kang Dynasty", "Enchantment — Saga", msc15KangOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.HobA9NextChapter(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.HobA9NextChapter(t, g)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if got := effectstest.EffectivePowerWithCounters(t, g, mine); got != 2+hand {
		t.Errorf("power = %d, want %d (2 plus a card in hand each)", got, 2+hand)
	}
	effectstest.DeclareAttack(t, g, o1.ID, mine)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, mine); err == nil {
		t.Error("a creature blocked Kang Dynasty's target")
	}
}

// --- Hawkeye, Avenging Archer ----------------------------------------------------------

func TestMsc15HawkeyePingsAndDrawsWhenItKillsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hawkeye := effectstest.B12Push(g, me.ID, "Hawkeye, Avenging Archer", "Legendary Creature — Human Archer Hero", msc15HawkeyeOracle, 3, 4)
	effectstest.AssertKeywords(t, g, hawkeye, "reach")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 1)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, hawkeye, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the 2/1 survived 1 damage")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("cards drawn = %d, want 1", got)
	}
}

func TestMsc15HawkeyeDrawsNothingForOtherDamageOrMyOwnCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Hawkeye, Avenging Archer", "Legendary Creature — Human Archer Hero", msc15HawkeyeOracle, 3, 4)
	killer := effectstest.B12Creature(g, me.ID, "Killer", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 1)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 1)
	hand := me.Hand.Size()
	effectstest.B27Damage(g, killer, theirs, 1)
	effectstest.B27Damage(g, killer, mine, 1)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(mine) {
		t.Fatal("the bears survived")
	}
	if me.Hand.Size() != hand {
		t.Errorf("drew %d cards though Hawkeye dealt no damage to them", me.Hand.Size()-hand)
	}
}

func TestMsc15HawkeyeOwnCreatureKilledByHawkeyeDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hawkeye := effectstest.B12Push(g, me.ID, "Hawkeye, Avenging Archer", "Legendary Creature — Human Archer Hero", msc15HawkeyeOracle, 3, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 1)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, hawkeye, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine)}); err != nil {
		t.Fatalf("ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("drew %d cards for my own creature dying", me.Hand.Size()-hand)
	}
}

func TestMsc15HawkeyePingsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hawkeye := effectstest.B12Push(g, me.ID, "Hawkeye, Avenging Archer", "Legendary Creature — Human Archer Hero", msc15HawkeyeOracle, 3, 4)
	if err := g.ActivateCatalogAbility(me.ID, hawkeye, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("opponent life = %d, want 39", opp.Life)
	}
	if !effectstest.IsTapped(g, hawkeye) {
		t.Error("the ping did not tap Hawkeye")
	}
}

// --- H.E.R.B.I.E., Lovable Robot -------------------------------------------------------

func TestMsc15HerbieMakesColorlessAndFiltersForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herbie := effectstest.B12Push(g, me.ID, "H.E.R.B.I.E., Lovable Robot", "Legendary Artifact Creature — Robot Scout", msc15HerbieOracle, 1, 1)
	effectstest.AssertKeywords(t, g, herbie, "flying")
	if err := g.ActivateManaAbility(me.ID, herbie, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if got := me.ManaPool.Units(); len(got) != 1 {
		t.Errorf("mana in the pool = %d units, want 1", len(got))
	}
}

func TestMsc15HerbieSurveilsAtCombatOnlyAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "H.E.R.B.I.E., Lovable Robot", "Legendary Artifact Creature — Robot Scout", msc15HerbieOracle, 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Fatal("surveilled with no noncreature spell cast")
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.B12Push(g, me.ID, "H.E.R.B.I.E., Lovable Robot", "Legendary Artifact Creature — Robot Scout", msc15HerbieOracle, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Error("no surveil prompt after a noncreature spell")
	}
}

// --- Extremis Elite --------------------------------------------------------------------

func TestMsc15ExtremisElitePowerUpGrowsAndPings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	elite := effectstest.CastPowerUpCreature(t, g, me, "Extremis Elite", "Creature — Human Mercenary Villain", "{1}{R}", "{C}{R}", msc15ExtremisOracle, 2)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, elite); got != "{3}" {
		t.Fatalf("the turn it entered: %q, want {3} ({4}{R} less {1}{R})", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, elite, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, elite); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if opp.Life != 39 {
		t.Errorf("opponent life = %d, want 39", opp.Life)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, elite, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("a second power-up activation was allowed")
	}
}

// --- Gamma Grotesque -------------------------------------------------------------------

func TestMsc15GammaGrotesquePowerUpDrawsForEachCreatureWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	marked := effectstest.B12Creature(g, me.ID, "Marked Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{marked, theirs} {
			_ = g.AddCounterForEffect(id, game.CounterPlusOne, 1)
		}
	})
	grot := effectstest.CastPowerUpCreature(t, g, me, "Gamma Grotesque", "Creature — Gamma Horror Villain", "{2}{G}", "{C}{C}{G}", msc15GammaOracle, 3)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, grot); got != "{2}{G}" {
		t.Fatalf("the turn it entered: %q, want {2}{G} ({4}{G}{G} less {2}{G})", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{G}")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, grot, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, grot); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("cards drawn = %d, want 2 (the Grotesque and the marked Bear; not the plain Bear, not theirs)", got)
	}
}

// --- Heroic Return ---------------------------------------------------------------------

func msc15ReturnPrice(t *testing.T, g *game.Game, me *game.Player) int {
	t.Helper()
	card := game.Card{InstanceID: uuid.New(), Name: "Heroic Return", TypeLine: "Instant", ManaCost: "{5}{W}",
		OracleID: msc15ReturnOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	price, err := g.PriceCast(me.ID, card, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	return price.Total.ManaValue()
}

func TestMsc15HeroicReturnCostsTwoLessOnlyWhileACreatureAttacksYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if got := msc15ReturnPrice(t, g, me); got != 6 {
		t.Fatalf("price with nothing attacking = %d, want 6", got)
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Msc15AttackedBy(t, g, bear, g.Seats[2].ID)
	if got := msc15ReturnPrice(t, g, me); got != 6 {
		t.Errorf("price while a creature attacks someone else = %d, want 6", got)
	}
	effectstest.Msc15AttackedBy(t, g, bear, me.ID)
	if got := msc15ReturnPrice(t, g, me); got != 4 {
		t.Errorf("price while a creature attacks me = %d, want 4", got)
	}
}

func msc15CastHeroicReturn(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Heroic Return", "Instant", msc15ReturnOracle, effectstest.CardRefs(target))
	effectstest.PassPriorityAroundTable(t, g)
}

func TestMsc15HeroicReturnGivesAHeroTwoCountersAndOthersNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hero := effectstest.PushGraveyardPermanent(me, "Fallen Hero", "Creature — Human Hero", "{2}")
	msc15CastHeroicReturn(t, g, me, hero)
	if !g.Battlefield.Contains(hero) {
		t.Fatal("the Hero did not return")
	}
	if got := effectstest.PlusOnes(g, hero); got != 2 {
		t.Errorf("the Hero entered with %d +1/+1 counters, want 2", got)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	bear := effectstest.PushGraveyardPermanent(me, "Fallen Bear", "Creature — Bear", "{2}")
	msc15CastHeroicReturn(t, g, me, bear)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear did not return")
	}
	if got := effectstest.PlusOnes(g, bear); got != 0 {
		t.Errorf("a non-Hero entered with %d +1/+1 counters, want 0", got)
	}
}

func TestMsc15HeroicReturnRefusesANoncreatureCardAndALowPayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.PushGraveyardPermanent(me, "Old Rock", "Artifact", "{2}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Heroic Return", TypeLine: "Instant", ManaCost: "{5}{W}", OracleID: msc15ReturnOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(rock)}); err == nil {
		t.Error("Heroic Return targeted a noncreature card")
	}
	hero := effectstest.PushGraveyardPermanent(me, "Fallen Hero", "Creature — Human Hero", "{2}")
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{W}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(hero), Strict: true}); err == nil {
		t.Error("cast for four mana with nothing attacking me")
	}
}

// --- Doombot Harbinger -----------------------------------------------------------------

func TestMsc15DoombotHarbingerMillsFourOnlyIfYouMay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	library := me.Library.Size()
	effectstest.CastHoldCreature(t, g, "Doombot Harbinger", "Artifact Creature — Robot Villain", msc15DoombotOracle, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 4 {
		t.Errorf("cards milled = %d, want 4", got)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	library = me.Library.Size()
	effectstest.CastHoldCreature(t, g, "Doombot Harbinger", "Artifact Creature — Robot Villain", msc15DoombotOracle, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != library {
		t.Errorf("milled %d cards after declining", library-me.Library.Size())
	}
}

func TestMsc15DoombotHarbingerDiesExilesItselfToRebuyACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bot := effectstest.B12Push(g, me.ID, "Doombot Harbinger", "Artifact Creature — Robot Villain", msc15DoombotOracle, 2, 1)
	target := effectstest.PushGraveyardPermanent(me, "Fallen Bear", "Creature — Bear", "{2}")
	rock := effectstest.PushGraveyardPermanent(me, "Old Rock", "Artifact", "{2}")
	effectstest.B27Kill(g, bot)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.MiInHand(me, target) {
		t.Error("the creature card did not return to hand")
	}
	if effectstest.MiInHand(me, rock) {
		t.Error("a noncreature card returned")
	}
	if z := g.FindCardZoneForEffect(bot); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the Harbinger is in %v, want exile", z)
	}
}

func TestMsc15DoombotHarbingerDecliningKeepsItInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bot := effectstest.B12Push(g, me.ID, "Doombot Harbinger", "Artifact Creature — Robot Villain", msc15DoombotOracle, 2, 1)
	target := effectstest.PushGraveyardPermanent(me, "Fallen Bear", "Creature — Bear", "{2}")
	effectstest.B27Kill(g, bot)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.MiInHand(me, target) {
		t.Error("a creature card returned after declining")
	}
	if z := g.FindCardZoneForEffect(bot); z == nil || z.Kind != game.ZoneGraveyard {
		t.Errorf("the Harbinger is in %v, want the graveyard", z)
	}
}

// --- Heart-Shaped Herb -----------------------------------------------------------------

func msc15Herb(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B27Push(g, owner, "Heart-Shaped Herb", "Artifact", msc15HerbOracle, "{4}", 0, 0)
}

func msc15DamageToPlayer(t *testing.T, g *game.Game, source, target uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, target, n); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
}

func TestMsc15HerbPreventsOneDamageFromOpponentSourcesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc15Herb(g, me.ID)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	msc15DamageToPlayer(t, g, theirs, me.ID, 3)
	if me.Life != 38 {
		t.Errorf("after 3 from an opponent's creature: life %d, want 38", me.Life)
	}
	msc15DamageToPlayer(t, g, theirs, me.ID, 1)
	if me.Life != 38 {
		t.Errorf("after 1 from an opponent's creature: life %d, want 38 (all prevented)", me.Life)
	}
	msc15DamageToPlayer(t, g, mine, me.ID, 3)
	if me.Life != 35 {
		t.Errorf("after 3 from my own creature: life %d, want 35 (not prevented)", me.Life)
	}
	msc15DamageToPlayer(t, g, theirs, opp.ID, 3)
	if opp.Life != 37 {
		t.Errorf("damage to an opponent: life %d, want 37 (the Herb protects only me)", opp.Life)
	}
}

func TestMsc15HerbSacrificeReturnsTheCreatureWithThreeCountersAndMakesYouMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herb := msc15Herb(g, me.ID)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, herb, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(herb) {
		t.Error("the Herb was not sacrificed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear did not return")
	}
	if got := effectstest.PlusOnes(g, bear); got != 3 {
		t.Errorf("the Bear has %d +1/+1 counters, want 3", got)
	}
	if g.Monarch != me.ID {
		t.Error("I did not become the monarch")
	}
}

func TestMsc15HerbDecliningTheSacrificeChangesNothingButTheHerb(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herb := msc15Herb(g, me.ID)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, herb, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(herb) {
		t.Error("the Herb stayed: sacrificing it is a cost")
	}
	if !g.Battlefield.Contains(bear) || effectstest.PlusOnes(g, bear) != 0 {
		t.Error("the Bear was touched after declining")
	}
	if g.Monarch == me.ID {
		t.Error("I became the monarch without sacrificing a creature")
	}
}

func TestMsc15HerbSacrificedTokenStillMakesYouMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herb := msc15Herb(g, me.ID)
	token := effectstest.B12Creature(g, me.ID, "Bear", "Token Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, herb, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the token returned")
	}
	if g.Monarch != me.ID {
		t.Error("sacrificing a token did not make me the monarch")
	}
}

func TestMsc15HerbActivationNeedsItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herb := msc15Herb(g, me.ID)
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, herb, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated for one of the two mana")
	}
	if !g.Battlefield.Contains(herb) {
		t.Error("a refused activation sacrificed the Herb")
	}
}
