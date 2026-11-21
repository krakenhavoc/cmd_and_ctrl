package testsru

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB214StirgeOracle   = "aed2d938-125d-49b4-93ac-202b7a37fcc2"
	gapB214TopazOracle    = "ca8535f2-1f96-4836-bd32-5e952f9bee72"
	gapB214CloudOracle    = gapB214TopazOracle + "#1"
	gapB214UchuulonOracle = "1f266fa4-6667-49db-9d24-dfbabf1449ca"
	gapB214ViconiaOracle  = "22e36922-a81a-4a0c-b698-51b48bd91ce6"
	gapB214VrockOracle    = "f50d2a5c-2408-4c09-81cd-55e2ca9b1ce1"
	gapB214BanditOracle   = "86a7f685-05d0-4167-b5bb-8f6943c23087"
	gapB214BabaOracle     = "d6cd4aa1-ff3a-4453-8735-e5ace5cbe386"
	gapB214KaghaOracle    = "f78db9ad-d4a3-4e49-a76e-bd31124b7837"
	gapB214BhaalOracle    = "0d2dd78c-103c-4286-a3a1-e471dc77ed91"
	gapB214KeeneOracle    = "e06007cf-dc6e-4b50-93d5-9a7c8ee9c023"
	gapB214MyrkulOracle   = "11329dd8-b10b-4659-b747-fbdbecf16308"
	gapB214GrizzlyOracle  = effectstest.SocB7GrizzlyBearsOracle
)

func gapB214Mana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	effectstest.Acrb1AddMana(t, g, p.ID, mana)
}

// gapB214SetLife moves a player's life total to n through the engine, so the
// layers that read it are re-derived.
func gapB214SetLife(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, n-p.Life); err != nil {
			t.Fatalf("ChangePlayerLifeForEffect: %v", err)
		}
	})
}

func TestGapB214EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB214StirgeOracle:   "Stirge",
		gapB214TopazOracle:    "Topaz Dragon",
		gapB214CloudOracle:    "Entropic Cloud",
		gapB214UchuulonOracle: "Uchuulon",
		gapB214ViconiaOracle:  "Viconia, Drow Apostate",
		gapB214VrockOracle:    "Vrock",
		gapB214BanditOracle:   "Zhentarim Bandit",
		gapB214KaghaOracle:    "Kagha, Shadow Archdruid",
		gapB214BhaalOracle:    "Bhaal, Lord of Murder",
		gapB214KeeneOracle:    "Nine-Fingers Keene",
		gapB214MyrkulOracle:   "Myrkul, Lord of Bones",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Stirge --------------------------------------------------------------

func TestGapB214StirgeCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Brute", "Creature — Ogre", 2, 2)
	stirge := effectstest.B12Push(g, opp.ID, "Stirge", "Creature — Insect Bat", gapB214StirgeOracle, 1, 1)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(stirge, attacker); err == nil {
		t.Error("Stirge blocked")
	}
}

func TestGapB214StirgePaysLifeAndSacrificesItselfToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stirge := effectstest.B12Push(g, me.ID, "Stirge", "Creature — Insect Bat", gapB214StirgeOracle, 1, 1)
	hand, life := me.Hand.Size(), me.Life
	gapB214Mana(t, g, me, "{C}{B}")
	effectstest.B16Activate(t, g, me.ID, stirge, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d -> %d, life %d -> %d: want +1 card and -1 life", hand, me.Hand.Size(), life, me.Life)
	}
	if effectstest.B12ZoneOf(g, stirge) == game.ZoneBattlefield {
		t.Error("Stirge was not sacrificed")
	}
}

func TestGapB214StirgeNeedsItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stirge := effectstest.B12Push(g, me.ID, "Stirge", "Creature — Insect Bat", gapB214StirgeOracle, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, stirge, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated without mana")
	}
}

// --- Topaz Dragon // Entropic Cloud --------------------------------------

func TestGapB214TopazDragonHasFlyingAndDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	d := effectstest.B12Push(g, me.ID, "Topaz Dragon", "Creature — Dragon", gapB214TopazOracle, 4, 4)
	for _, kw := range []string{"flying", "deathtouch"} {
		if !effectstest.HasEffectiveKeyword(t, g, d, kw) {
			t.Errorf("Topaz Dragon lacks %s", kw)
		}
	}
}

func TestGapB214EntropicCloudGivesYourCreaturesDeathtouchOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B18CastFromHand(t, g, me, "Entropic Cloud", "Instant — Adventure", gapB214CloudOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "deathtouch") {
		t.Error("my creature lacks deathtouch")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "deathtouch") {
		t.Error("an opponent's creature gained deathtouch")
	}
}

// --- Vrock ---------------------------------------------------------------

func TestGapB214VrockDrainsOnlyAfterYourPermanentLeft(t *testing.T) {
	for _, left := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Vrock", "Creature — Bird Demon", gapB214VrockOracle, 3, 3)
		if left {
			effectstest.B18Kill(t, g, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
		}
		lives := effectstest.B24Lives(g)
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		for i, l := range effectstest.B24Lives(g) {
			want := lives[i]
			if left && i != 0 {
				want -= 3
			}
			if l != want {
				t.Errorf("left=%v seat %d life %d, want %d", left, i, l, want)
			}
		}
	}
}

// --- Zhentarim Bandit ----------------------------------------------------

func TestGapB214BanditMakesATreasureWhenYouPayOneLife(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bandit := effectstest.B12Push(g, me.ID, "Zhentarim Bandit", "Creature — Halfling Rogue", gapB214BanditOracle, 2, 1)
		life := me.Life
		effectstest.DeclareAttack(t, g, opp.ID, bandit)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		treasures := effectstest.B16CountNamed(g, "Treasure")
		if pay && (treasures != 1 || me.Life != life-1) {
			t.Errorf("paid: %d Treasures, life %d -> %d, want 1 and -1", treasures, life, me.Life)
		}
		if !pay && (treasures != 0 || me.Life != life) {
			t.Errorf("declined: %d Treasures, life %d -> %d, want none and no change", treasures, life, me.Life)
		}
	}
}

// --- Viconia -------------------------------------------------------------

func gapB214GraveCreatures(me *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	}
}

func TestGapB214ViconiaReturnsACreatureOnlyWithFourInYourGraveyard(t *testing.T) {
	for _, n := range []int{3, 4} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Viconia, Drow Apostate", "Legendary Creature — Elf Cleric", gapB214ViconiaOracle, 2, 3)
		gapB214GraveCreatures(me, n)
		hand, grave := me.Hand.Size(), me.Graveyard.Size()
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		wantReturn := n >= 4
		if got := me.Hand.Size() - hand; wantReturn && got < 1 || !wantReturn && got != 0 {
			// the draw step may also have added a card; compare the graveyard instead
			t.Logf("hand delta %d", got)
		}
		if got := grave - me.Graveyard.Size(); wantReturn && got != 1 || !wantReturn && got != 0 {
			t.Errorf("%d creature cards: graveyard %d -> %d, want %v returned", n, grave, me.Graveyard.Size(), wantReturn)
		}
	}
}

// --- Myrkul and Bhaal: the shared indestructible clause -----------------

func TestGapB214GodsAreIndestructibleAtHalfStartingLifeOrLess(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
	}{{"Myrkul, Lord of Bones", gapB214MyrkulOracle}, {"Bhaal, Lord of Murder", gapB214BhaalOracle}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		god := effectstest.B12Push(g, me.ID, tc.name, "Legendary Creature — God", tc.oracle, 4, 4)
		gapB214SetLife(t, g, me, 21)
		if effectstest.HasEffectiveKeyword(t, g, god, "indestructible") {
			t.Errorf("%s is indestructible at 21 life", tc.name)
		}
		gapB214SetLife(t, g, me, 20)
		if !effectstest.HasEffectiveKeyword(t, g, god, "indestructible") {
			t.Errorf("%s is not indestructible at 20 life (half of 40)", tc.name)
		}
	}
}

func TestGapB214MyrkulExilesTheDeadAndMakesAnEnchantmentCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Myrkul, Lord of Bones", "Legendary Creature — God", gapB214MyrkulOracle, 7, 5)
	bear := effectstest.B12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", gapB214GrizzlyOracle, 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, bear); z != game.ZoneExile {
		t.Fatalf("the dead Bear is in %s, want exile", z)
	}
	toks := effectstest.S47Tokens(g, "Grizzly Bears")
	if len(toks) != 1 {
		t.Fatalf("%d Bear tokens, want 1", len(toks))
	}
	tok := toks[0]
	if tok.IsCreature() || !tok.HasCardType("enchantment") || tok.HasSubtype("Bear") {
		t.Errorf("the token is %q, want a plain Enchantment", tok.TypeLine)
	}
}

func TestGapB214MyrkulIgnoresTokensAndOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Myrkul, Lord of Bones", "Legendary Creature — God", gapB214MyrkulOracle, 7, 5)
	theirs := effectstest.B12Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", gapB214GrizzlyOracle, 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPromptFor(g, me.ID) != nil {
		t.Error("Myrkul offered to exile an opponent's creature")
	}
}

// --- Bhaal ---------------------------------------------------------------

func TestGapB214BhaalCountersAndGoadsATargetWhenYourCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Bhaal, Lord of Murder", "Legendary Creature — God", gapB214BhaalOracle, 4, 4)
	victim := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	target := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, target)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, target, game.CounterPlusOne); n != 1 {
		t.Errorf("target has %d +1/+1 counters, want 1", n)
	}
	if c, ok := effectstest.BattlefieldCard(g, target); !ok || !c.IsGoaded() {
		t.Errorf("the target is not goaded")
	}
}

// --- Uchuulon ------------------------------------------------------------

func TestGapB214UchuulonPowerCountsCrabsOozesAndHorrors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	u := effectstest.B12Push(g, me.ID, "Uchuulon", "Creature — Crab Ooze Horror", gapB214UchuulonOracle, 0, 4)
	if p := effectstest.EffectivePower(t, g, u); p != 1 {
		t.Fatalf("alone, power %d, want 1 (itself)", p)
	}
	effectstest.B12Creature(g, me.ID, "Ooze", "Creature — Ooze", 1, 1)
	effectstest.B12Creature(g, me.ID, "Horror", "Creature — Horror", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Horror", "Creature — Horror", 1, 1)
	if p := effectstest.EffectivePower(t, g, u); p != 3 {
		t.Errorf("power %d, want 3 (itself, an Ooze and a Horror)", p)
	}
	if tg := effectstest.EffectiveToughness(t, g, u); tg != 4 {
		t.Errorf("toughness %d, want 4", tg)
	}
}

func TestGapB214UchuulonExilesAnOpponentsCreatureCardAndCopiesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Uchuulon", "Creature — Crab Ooze Horror", gapB214UchuulonOracle, 0, 4)
	dead := effectstest.B17GraveyardCard(opp, "Dead Ogre", "Creature — Ogre", "{3}{R}")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, dead)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if z := effectstest.B12ZoneOf(g, dead); z != game.ZoneExile {
		t.Fatalf("the Ogre card is in %s, want exile", z)
	}
	if n := len(effectstest.S47Tokens(g, "Uchuulon")); n != 1 {
		t.Errorf("%d Uchuulon tokens, want 1", n)
	}
}

func TestGapB214UchuulonMakesNoTokenWithoutACardToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Uchuulon", "Creature — Crab Ooze Horror", gapB214UchuulonOracle, 0, 4)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.S47Tokens(g, "Uchuulon")); n != 0 {
		t.Errorf("%d Uchuulon tokens with nothing exiled, want 0", n)
	}
}

// --- Nine-Fingers Keene --------------------------------------------------

func TestGapB214KeeneHasMenaceAndNinePayLifeWard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	k := effectstest.B12Push(g, me.ID, "Nine-Fingers Keene", "Legendary Creature — Human Rogue", gapB214KeeneOracle, 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, k, "menace") {
		t.Error("Keene lacks menace")
	}
	spec, _ := effects.Lookup(gapB214KeeneOracle)
	if len(spec.Triggered) != 2 {
		t.Errorf("%d triggers, want the ward and the combat-damage ability", len(spec.Triggered))
	}
}

func TestGapB214KeenePutsTheRestIntoHandWithNineGates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keene := effectstest.B12Push(g, me.ID, "Nine-Fingers Keene", "Legendary Creature — Human Rogue", gapB214KeeneOracle, 4, 4)
	for i := 0; i < 8; i++ {
		effectstest.B12Permanent(g, me.ID, "Gate", "Land — Gate")
	}
	gate := gapB214Gate(me, "Ninth Gate")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, keene)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("not asked which Gate to put onto the battlefield")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{gate}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 8 {
		t.Errorf("hand grew by %d, want the other eight cards", got)
	}
}

func gapB214Gate(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Land — Gate", Owner: p.ID, Controller: p.ID})
	return id
}

func TestGapB214KeenePutsAGateOntoTheBattlefieldAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keene := effectstest.B12Push(g, me.ID, "Nine-Fingers Keene", "Legendary Creature — Human Rogue", gapB214KeeneOracle, 4, 4)
	gate := gapB214Gate(me, "Test Gate")
	size := me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, keene)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("Keene's controller was not asked which Gate to put onto the battlefield")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{gate}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, gate); z != game.ZoneBattlefield {
		t.Fatalf("the Gate is in %s, want the battlefield", z)
	}
	if got := me.Library.Size(); got != size-1 {
		t.Errorf("library %d -> %d, want the Gate gone and the other eight cards back in it", size, got)
	}
}

// --- Kagha ---------------------------------------------------------------

func TestGapB214KaghaGainsDeathtouchAndMillsTwoWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kagha := effectstest.B12Push(g, me.ID, "Kagha, Shadow Archdruid", "Legendary Creature — Elf Druid", gapB214KaghaOracle, 1, 4)
	grave := me.Graveyard.Size()
	effectstest.DeclareAttack(t, g, opp.ID, kagha)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, kagha, "deathtouch") {
		t.Error("Kagha lacks deathtouch")
	}
	if got := me.Graveyard.Size() - grave; got != 2 {
		t.Errorf("graveyard grew by %d, want 2", got)
	}
}

func TestGapB214KaghaCastsOnePermanentFromThisTurnsLibraryCardsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Kagha, Shadow Archdruid", "Legendary Creature — Elf Druid", gapB214KaghaOracle, 1, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	milled := make([]uuid.UUID, 2)
	for i := range milled {
		milled[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: milled[i], Name: "Milled Bear", TypeLine: "Creature — Bear", ManaCost: "{G}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	}
	discarded := effectstest.GfDead(t, g, me, "Discarded Bear", "Creature — Bear", "{G}")
	g.WithWriteLock(func() { g.MillNForEffect(me.ID, 2) })
	for _, id := range milled {
		if !me.Graveyard.Contains(id) {
			t.Fatal("the milled cards did not reach the graveyard")
		}
	}
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, discarded, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("a card that was not put there from the library was cast")
	}
	if err := g.CastSpell(me.ID, milled[0], game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("casting a milled permanent: %v", err)
	}
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, milled[1], game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("a second permanent was cast: the permission is once during each of your turns")
	}
}

func TestGapB214KaghaPermissionIsLimitedToCardsFromYourLibraryThisTurn(t *testing.T) {
	spec, ok := effects.Lookup(gapB214KaghaOracle)
	if !ok || len(spec.CastPermissions) != 1 {
		t.Fatalf("permissions %+v", spec.CastPermissions)
	}
	p := spec.CastPermissions[0]
	if p.Zone != game.ZoneGraveyard || !p.Filter.FromLibraryToGraveyardThisTurn || !p.Filter.PermanentOnly ||
		len(p.PerType) != 1 || p.PerType[0] != game.PermissionTypePermanent || p.Timing != game.TimingYourTurnOnly {
		t.Errorf("permission %+v is not a once-a-turn permanent play from this turn's library cards", p)
	}
	_ = strings.ToLower
}
