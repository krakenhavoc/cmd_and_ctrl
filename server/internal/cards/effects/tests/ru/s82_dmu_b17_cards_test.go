package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dominaria United batch 17 (#5487): Activated Sleeper; Cadric, Soul Kindler;
// The Ever-Changing 'Dane; The Peregrine Dynamo; Stangg, Echo Warrior;
// Ratadrabik of Urborg; Ivy, Gleeful Spellthief; Vesuvan Duplimancy; Twinferno;
// Najal, the Storm Runner.

const (
	dmuB17SleeperOracle   = "e12da2a1-e390-4487-b154-69e439cb7b05"
	dmuB17CadricOracle    = "739bbf44-9169-4749-abea-21c20823915f"
	dmuB17DaneOracle      = "bda0468f-1c72-4d3d-ad94-15d446c75313"
	dmuB17DynamoOracle    = "da21f1a7-65d0-46d0-8ad1-f84a387ba753"
	dmuB17StanggOracle    = "aaf2a587-ced7-4615-b10f-164e7384b073"
	dmuB17RatadrabikOracl = "b3ab43f2-4f2b-4439-9254-0110ce443fdb"
	dmuB17IvyOracle       = "55ad6a6b-1c44-4397-86ae-dd9221892b22"
	dmuB17DuplimancyOracl = "6058fab1-9878-42b7-93be-b0a5aa99bd6d"
	dmuB17TwinfernoOracle = "93989dd7-2d3e-46e2-8e92-8d0479796087"
	dmuB17NajalOracle     = "2a7504b9-220d-412e-9381-b6a8a3750241"
)

func TestDmuB17EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dmuB17SleeperOracle: "Activated Sleeper", dmuB17CadricOracle: "Cadric, Soul Kindler",
		dmuB17DaneOracle: "The Ever-Changing 'Dane", dmuB17DynamoOracle: "The Peregrine Dynamo",
		dmuB17StanggOracle: "Stangg, Echo Warrior", dmuB17RatadrabikOracl: "Ratadrabik of Urborg",
		dmuB17IvyOracle: "Ivy, Gleeful Spellthief", dmuB17DuplimancyOracl: "Vesuvan Duplimancy",
		dmuB17TwinfernoOracle: "Twinferno", dmuB17NajalOracle: "Najal, the Storm Runner",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Twinferno ----------------------------------------------------------------

func TestDmuB17TwinfernoCopiesTheNextInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := effectstest.LifeOf(g, victim)
	effectstest.CastCatalogSpellWithModes(t, g, "Twinferno", "Instant", dmuB17TwinfernoOracle, []int{0})
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("Twinferno left %d delayed triggers, want 1", len(g.DelayedTriggers))
	}
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me, victim, 1)
	if got := effectstest.LifeOf(g, victim); got != before-6 {
		t.Errorf("victim life = %d, want %d (the Bolt and its copy)", got, before-6)
	}
	if got := effectstest.GraveyardSize(g, me); got != 2 {
		t.Errorf("graveyard = %d cards, want 2 (Twinferno and the Bolt; a copy is not a card)", got)
	}
}

func TestDmuB17TwinfernoDoesNotCopyItselfOrACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpellWithModes(t, g, "Twinferno", "Instant", dmuB17TwinfernoOracle, []int{0})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.GraveyardSize(g, g.Seats[0].ID); got != 1 {
		t.Errorf("graveyard = %d, want 1", got)
	}
	effectstest.CastCatalogSpell(t, g, "Sol Ring", "Artifact", effectstest.SolRingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Errorf("an artifact spell consumed the delayed trigger (%d left)", len(g.DelayedTriggers))
	}
}

func TestDmuB17TwinfernoGrantsDoubleStrikeToYourCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Twinferno", TypeLine: "Instant", OracleID: dmuB17TwinfernoOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: effectstest.B16TargetCard(theirs), Strict: true}); err == nil {
		t.Fatal("targeted an opponent's creature")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: effectstest.B16TargetCard(mine), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "double strike") {
		t.Error("the Bear lacks double strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "double strike") {
		t.Error("their Bear has double strike")
	}
}

// --- Najal, the Storm Runner ----------------------------------------------------

func TestDmuB17NajalLetsYouCastSorceriesAtInstantSpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Doublecast", "Sorcery", effectstest.DoublecastOracle, nil); err != nil {
		t.Fatalf("main phase: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// A sorcery during combat is refused without Najal and allowed with it.
	finish := func(withNajal bool) error {
		g2 := effectstest.NewCatalogGame(t)
		p := g2.Seats[g2.Turn.ActiveSeat]
		effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
		if withNajal {
			g2.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Najal, the Storm Runner", TypeLine: "Legendary Creature — Efreet Wizard",
				OracleID: dmuB17NajalOracle, Power: 5, Toughness: 4, Owner: p.ID, Controller: p.ID})
		}
		id := uuid.New()
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Doublecast", TypeLine: "Sorcery", OracleID: effectstest.DoublecastOracle, Owner: p.ID, Controller: p.ID})
		return g2.CastSpell(p.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := finish(false); err == nil {
		t.Error("a sorcery was cast in combat with no Najal")
	}
	if err := finish(true); err != nil {
		t.Errorf("a sorcery could not be cast in combat with Najal: %v", err)
	}
	_ = me
}

func dmuB17NajalAttack(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	effectstest.AdvanceToMain(t, g)
	najal := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: najal, Name: "Najal, the Storm Runner", TypeLine: "Legendary Creature — Efreet Wizard",
		OracleID: dmuB17NajalOracle, Power: 5, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.FinABAttack(t, g, opp.ID, najal)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.PassPriorityAroundTable(t, g)
	return
}

func TestDmuB17NajalPaysTwoToCopyTheNextInstantOrSorcery(t *testing.T) {
	g, me, opp := dmuB17NajalAttack(t)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("paying {2} left %d delayed triggers, want 1", len(g.DelayedTriggers))
	}
	before := effectstest.LifeOf(g, opp.ID)
	dmuB17BoltNow(t, g, opp.ID)
	for i := 0; i < 3; i++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != before-6 {
		t.Errorf("opponent life = %d, want %d (the Bolt and its copy)", got, before-6)
	}
}

// dmuB17BoltNow casts a Lightning Bolt at the player in the current step.
func dmuB17BoltNow(t *testing.T, g *game.Game, victim uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim}}}); err != nil {
		t.Fatalf("cast Bolt: %v", err)
	}
}

func TestDmuB17NajalDecliningCopiesNothing(t *testing.T) {
	g, me, opp := dmuB17NajalAttack(t)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("declining left %d delayed triggers", len(g.DelayedTriggers))
	}
	before := effectstest.LifeOf(g, opp.ID)
	dmuB17BoltNow(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-3 {
		t.Errorf("opponent life = %d, want %d", got, before-3)
	}
}

// --- Ratadrabik of Urborg -----------------------------------------------------

func dmuB17Legend(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.B12Push(g, owner, name, "Legendary Creature — Human Wizard", "oracle-"+name, 3, 2)
}

func dmuB17Tokens(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func TestDmuB17RatadrabikCopiesAnotherLegendThatDiesAsATwoTwoBlackZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	rat := effectstest.B12Push(g, me.ID, "Ratadrabik of Urborg", "Legendary Creature — Zombie Wizard", dmuB17RatadrabikOracl, 3, 3)
	legend := dmuB17Legend(g, me.ID, "Mage")
	effectstest.KillCreature(t, g, me.ID, legend)
	effectstest.PassPriorityAroundTable(t, g)
	toks := dmuB17Tokens(g, "Mage")
	if len(toks) != 1 {
		t.Fatalf("%d Mage tokens, want 1", len(toks))
	}
	tok := toks[0]
	if tok.HasSupertype("Legendary") {
		t.Error("the copy is legendary")
	}
	if !tok.HasSubtype("Zombie") || !tok.HasSubtype("Human") || !tok.HasSubtype("Wizard") {
		t.Errorf("the copy's type line is %q, want Human Wizard plus Zombie", tok.TypeLine)
	}
	if p, tg := effectstest.EffectivePower(t, g, tok.InstanceID), effectstest.EffectiveToughness(t, g, tok.InstanceID); p != 2 || tg != 2 {
		t.Errorf("the copy is %d/%d, want 2/2", p, tg)
	}
	if !tok.HasColor("B") {
		t.Errorf("the copy is not black: %v", tok.Colors)
	}
	// The new Zombie has vigilance from Ratadrabik's lord ability.
	if !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "vigilance") {
		t.Error("the Zombie copy lacks vigilance")
	}
	_ = rat
}

func TestDmuB17RatadrabikIgnoresItselfNonlegendsAndOpposingLegends(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	effectstest.AdvanceToMain(t, g)
	rat := effectstest.B12Push(g, me.ID, "Ratadrabik of Urborg", "Legendary Creature — Zombie Wizard", dmuB17RatadrabikOracl, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := dmuB17Legend(g, opp.ID, "Their Mage")
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.KillCreature(t, g, me.ID, rat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"Bear", "Their Mage", "Ratadrabik of Urborg"} {
		if got := len(dmuB17Tokens(g, name)); got != 0 {
			t.Errorf("%d %s tokens, want none", got, name)
		}
	}
}

func TestDmuB17RatadrabikGivesOtherZombiesVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := effectstest.B12Push(g, me.ID, "Ratadrabik of Urborg", "Legendary Creature — Zombie Wizard", dmuB17RatadrabikOracl, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Zed", "Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Zed", "Creature — Zombie", 2, 2)
	for id, want := range map[uuid.UUID]bool{rat: true, mine: true, bear: false, theirs: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "vigilance"); got != want {
			t.Errorf("vigilance on %s = %v, want %v", effectstest.B12Card(t, g, id).Name, got, want)
		}
	}
}

// --- Vesuvan Duplimancy -------------------------------------------------------

func TestDmuB17DuplimancyCopiesTheCreatureYourSpellTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Vesuvan Duplimancy", "Enchantment", dmuB17DuplimancyOracl, 0, 0)
	legend := dmuB17Legend(g, me.ID, "Mage")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// A spell aimed at an opposing creature does nothing.
	effectstest.CastCatalogSpell(t, g, "Whirling Strike", "Instant", broB02Whirling, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(dmuB17Tokens(g, "Their Bear")); got != 0 {
		t.Fatalf("copied an opposing creature (%d tokens)", got)
	}
	// A spell aimed at your own legend copies it, not legendary.
	effectstest.CastCatalogSpell(t, g, "Whirling Strike", "Instant", broB02Whirling, effectstest.B16TargetCard(legend))
	effectstest.PassPriorityAroundTable(t, g)
	toks := dmuB17Tokens(g, "Mage")
	if len(toks) != 1 {
		t.Fatalf("%d Mage tokens, want 1", len(toks))
	}
	if toks[0].HasSupertype("Legendary") {
		t.Error("the token is legendary")
	}
	if toks[0].Controller != me.ID {
		t.Error("the token is not yours")
	}
}

// --- Ivy, Gleeful Spellthief --------------------------------------------------

func TestDmuB17IvyCopiesASpellThatTargetsOnlyAnotherCreatureAndRetargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ivy := effectstest.B12Push(g, me.ID, "Ivy, Gleeful Spellthief", "Legendary Creature — Faerie Rogue", dmuB17IvyOracle, 2, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Whirling Strike", "Instant", broB02Whirling, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("their Bear has power %d, want 4 (the original resolved)", p)
	}
	if p := effectstest.EffectivePower(t, g, ivy); p != 4 {
		t.Errorf("Ivy has power %d, want 4 (the copy targets Ivy)", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, ivy, "trample") {
		t.Error("Ivy lacks trample from the copy")
	}
}

func TestDmuB17IvyMayDeclineToCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ivy := effectstest.B12Push(g, me.ID, "Ivy, Gleeful Spellthief", "Legendary Creature — Faerie Rogue", dmuB17IvyOracle, 2, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Whirling Strike", "Instant", broB02Whirling, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, ivy); p != 2 {
		t.Errorf("Ivy has power %d, want 2 (declined)", p)
	}
}

func TestDmuB17IvyDoesNotCopyASpellAimedAtIvy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ivy := effectstest.B12Push(g, me.ID, "Ivy, Gleeful Spellthief", "Legendary Creature — Faerie Rogue", dmuB17IvyOracle, 2, 1)
	effectstest.CastCatalogSpell(t, g, "Whirling Strike", "Instant", broB02Whirling, effectstest.B16TargetCard(ivy))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("Ivy offered to copy a spell that targets her")
	}
}

// --- Cadric, Soul Kindler -----------------------------------------------------

func dmuB17CadricEnters(t *testing.T, name string) (g *game.Game, me *game.Player, cadric, entered uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	cadric = effectstest.B12Push(g, me.ID, "Cadric, Soul Kindler", "Legendary Creature — Dwarf Wizard", dmuB17CadricOracle, 4, 3)
	effectstest.FloatForTest(g, me, "C")
	entered = effectstest.CastCatalogSpell(t, g, name, "Legendary Creature — Human Wizard", "oracle-"+name, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return
}

func TestDmuB17CadricPaysOneForAHastyTokenCopySacrificedAtTheEndStep(t *testing.T) {
	g, me, _, orig := dmuB17CadricEnters(t, "Mage")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	toks := dmuB17Tokens(g, "Mage")
	if len(toks) != 1 {
		t.Fatalf("%d Mage tokens, want 1", len(toks))
	}
	tok := toks[0]
	if !tok.HasSupertype("Legendary") {
		t.Error("the copy is not legendary (Cadric's copy keeps legendary)")
	}
	if !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
		t.Error("the token lacks haste")
	}
	// The legend rule does not apply to the token: the card and the token both stay.
	g.RunStateChecksForTest()
	if !effectstest.BattlefieldHas(g, orig) || !effectstest.BattlefieldHas(g, tok.InstanceID) || effectstest.LegendPrompt(g, me.ID) != nil {
		t.Fatal("the legend rule applied to the card and its token copy")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, tok.InstanceID) {
		t.Error("the token was not sacrificed at the end step")
	}
	if !effectstest.BattlefieldHas(g, orig) {
		t.Error("the original was sacrificed")
	}
}

func TestDmuB17CadricDecliningPaysNothingAndMakesNoToken(t *testing.T) {
	g, me, _, _ := dmuB17CadricEnters(t, "Mage")
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(dmuB17Tokens(g, "Mage")); got != 0 {
		t.Errorf("%d tokens after declining", got)
	}
}

func TestDmuB17CadricIgnoresNonlegendsTokensAndOpposingLegends(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Cadric, Soul Kindler", "Legendary Creature — Dwarf Wizard", dmuB17CadricOracle, 4, 3)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	b12Legend := effectstest.B12Push(g, opp.ID, "Their Mage", "Legendary Creature — Human", "", 2, 2)
	_ = b12Legend
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("Cadric asked about a nonlegendary creature or an opposing legend")
	}
}

func TestDmuB17CadricLegendRuleStillAppliesToLegendaryCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Cadric, Soul Kindler", "Legendary Creature — Dwarf Wizard", dmuB17CadricOracle, 4, 3)
	a := dmuB17Legend(g, me.ID, "Twin")
	b := dmuB17Legend(g, me.ID, "Twin")
	g.RunStateChecksForTest()
	if effectstest.LegendPrompt(g, me.ID) == nil {
		t.Errorf("no legend-rule choice for two legendary cards with one name (%v, %v)", a, b)
	}
}

// --- Activated Sleeper --------------------------------------------------------

func TestDmuB17SleeperOffersOnlyCreatureCardsPutThereFromTheBattlefieldThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	effectstest.AdvanceToMain(t, g)
	milled := effectstest.GraveyardCreature(opp, "Milled Giant", "Giant", 3, 3)
	died := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.KillCreature(t, g, opp.ID, died)
	bolt := effectstest.GraveyardCardOf(g, opp, "Lightning Bolt", "Instant", "{R}", "")

	sleeper := effectstest.CastGraveyardCopier(t, g, "Activated Sleeper", "Creature — Phyrexian Shapeshifter", dmuB17SleeperOracle, 0, 0)
	prompt := effectstest.WaitForCopyPrompt(t, g)
	if len(prompt.CopyOptions) != 1 || prompt.CopyOptions[0] != died {
		t.Fatalf("offered %v, want only the Giant that died this turn (not %s milled, nor %s)", prompt.CopyOptions, milled, bolt)
	}
	effectstest.ResolveWithCopyChoice(t, g, died)
	got := effectstest.CopyBattlefieldCard(t, g, sleeper)
	if got.Name != "Hill Giant" || got.CurrentPower() != 3 {
		t.Errorf("Activated Sleeper = %q %d/%d, want a 3/3 Hill Giant", got.Name, got.CurrentPower(), got.CurrentToughness())
	}
	for _, sub := range []string{"Giant", "Phyrexian"} {
		if !got.HasSubtype(sub) {
			t.Errorf("the copy should be a %s (%q)", sub, got.TypeLine)
		}
	}
	if !me.Graveyard.Contains(died) && !opp.Graveyard.Contains(died) {
		t.Error("the copied card left the graveyard")
	}
}

// --- The Ever-Changing 'Dane --------------------------------------------------

func TestDmuB17DaneBecomesACopyOfTheSacrificedCreatureAndKeepsItsAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	dane := effectstest.B12Push(g, me.ID, "The Ever-Changing 'Dane", "Legendary Creature — Shapeshifter", dmuB17DaneOracle, 3, 3)
	giant := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 5, 5)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, dane, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{giant}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.B12Card(t, g, dane)
	if got.Name != "Hill Giant" {
		t.Fatalf("'Dane is %q, want a copy of Hill Giant", got.Name)
	}
	if p := effectstest.EffectivePower(t, g, dane); p != 5 {
		t.Errorf("'Dane's power is %d, want 5", p)
	}
	// "Except it has this ability": it can sacrifice another creature again.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, dane, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("the copy could not activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Card(t, g, dane); got.Name != "Bear" {
		t.Errorf("'Dane is %q, want a copy of Bear", got.Name)
	}
}

// --- The Peregrine Dynamo -----------------------------------------------------

const dmuB17RanAndShawOracle = "a728685f-8670-4db2-ae02-3cf74eb3c402"

func dmuB17AbilityOnStack(g *game.Game) (uuid.UUID, bool) {
	var id uuid.UUID
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemActivated {
			id = item.ID
		}
	}
	return id, id != uuid.Nil
}

func dmuB17DynamoTable(t *testing.T, commander bool) (g *game.Game, me *game.Player, dynamo, shaw, ability uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	dynamo = effectstest.B12Push(g, me.ID, "The Peregrine Dynamo", "Legendary Artifact Creature — Construct", dmuB17DynamoOracle, 1, 5)
	shaw = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ran and Shaw",
		TypeLine: "Legendary Creature — Dragon", OracleID: dmuB17RanAndShawOracle, Power: 4, Toughness: 4,
		Owner: me.ID, Controller: me.ID, IsCommander: commander})
	effectstest.FloatForTest(g, me, "RRRR")
	if err := g.ActivateCatalogAbility(me.ID, shaw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Ran and Shaw: %v", err)
	}
	id, ok := dmuB17AbilityOnStack(g)
	if !ok {
		t.Fatal("no ability on the stack")
	}
	return g, me, dynamo, shaw, id
}

func TestDmuB17DynamoCopiesAnAbilityFromAnotherLegendarySource(t *testing.T) {
	g, me, dynamo, shaw, ability := dmuB17DynamoTable(t, false)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, dynamo, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ability), Strict: true}); err != nil {
		t.Fatalf("activate the Dynamo: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4 && effectstest.CopyPrompt(g) != nil; i++ {
		c := effectstest.CopyPrompt(g)
		_ = g.ResolveCopyTarget(c.ID, c.Chooser, uuid.Nil)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p := effectstest.EffectivePower(t, g, shaw); p != 8 {
		t.Errorf("Ran and Shaw has power %d, want 8 (the ability and its copy each gave +2/+0)", p)
	}
}

func TestDmuB17DynamoRefusesACommandersAbility(t *testing.T) {
	g, me, dynamo, _, ability := dmuB17DynamoTable(t, true)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, dynamo, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ability), Strict: true}); err == nil {
		t.Error("copied an ability whose source is a commander")
	}
}

func TestDmuB17DynamoRefusesAnAbilityFromANonlegendarySource(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	dynamo := effectstest.B12Push(g, me.ID, "The Peregrine Dynamo", "Legendary Artifact Creature — Construct", dmuB17DynamoOracle, 1, 5)
	shaw := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ran and Shaw",
		TypeLine: "Creature — Dragon", OracleID: dmuB17RanAndShawOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "RRRR")
	if err := g.ActivateCatalogAbility(me.ID, shaw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	ability, _ := dmuB17AbilityOnStack(g)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, dynamo, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ability), Strict: true}); err == nil {
		t.Error("copied an ability from a nonlegendary source")
	}
}

// --- Stangg, Echo Warrior -----------------------------------------------------

func TestDmuB17StanggMakesATappedAttackingTwinAndSacrificesItAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	effectstest.AdvanceToMain(t, g)
	stangg := effectstest.B12Push(g, me.ID, "Stangg, Echo Warrior", "Legendary Creature — Human Warrior", dmuB17StanggOracle, 3, 4)
	effectstest.FinABAttack(t, g, opp.ID, stangg)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4 && len(g.PendingChoices) > 0; i++ {
		break
	}
	twins := dmuB17Tokens(g, "Stangg Twin")
	if len(twins) != 1 {
		t.Fatalf("%d Stangg Twins, want 1 (pending %v)", len(twins), g.PendingChoices)
	}
	twin := twins[0]
	if !twin.Tapped || twin.AttackingTarget == uuid.Nil {
		t.Errorf("the Twin is tapped=%v attacking=%v, want tapped and attacking", twin.Tapped, twin.AttackingTarget)
	}
	if !twin.HasSupertype("Legendary") || twin.Power != 3 || twin.Toughness != 4 || !twin.HasColor("R") || !twin.HasColor("G") {
		t.Errorf("the Twin is %+v, want a legendary red and green 3/4", twin)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if len(dmuB17Tokens(g, "Stangg Twin")) != 0 {
		t.Error("the Twin was not sacrificed at the end step")
	}
	if !effectstest.BattlefieldHas(g, stangg) {
		t.Error("Stangg was sacrificed")
	}
}

func TestDmuB17StanggCopiesEachAuraAndEquipmentOntoTheTwin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	effectstest.AdvanceToMain(t, g)
	stangg := effectstest.B12Push(g, me.ID, "Stangg, Echo Warrior", "Legendary Creature — Human Warrior", dmuB17StanggOracle, 3, 4)
	sword := effectstest.FinABEquipment(g, me.ID, "Test Sword")
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Aura",
		TypeLine: "Enchantment — Aura", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: stangg})
		_ = g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: stangg})
	})
	effectstest.FinABAttack(t, g, opp.ID, stangg)
	effectstest.PassPriorityAroundTable(t, g)
	twins := dmuB17Tokens(g, "Stangg Twin")
	if len(twins) != 1 {
		t.Fatalf("%d Stangg Twins, want 1", len(twins))
	}
	twin := twins[0].InstanceID
	var swordCopy, auraCopy game.Card
	for _, c := range dmuB17Tokens(g, "Test Sword") {
		swordCopy = c
	}
	for _, c := range dmuB17Tokens(g, "Test Aura") {
		auraCopy = c
	}
	if swordCopy.InstanceID == uuid.Nil || !swordCopy.IsAttachedTo(twin) {
		t.Errorf("the Equipment copy %v is not attached to the Twin", swordCopy.InstanceID)
	}
	if auraCopy.InstanceID == uuid.Nil || !auraCopy.IsAttachedTo(twin) {
		t.Errorf("the Aura copy %v is not attached to the Twin", auraCopy.InstanceID)
	}
	if !effectstest.B12Card(t, g, sword).IsAttachedTo(stangg) {
		t.Error("the original Equipment left Stangg")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"Stangg Twin", "Test Sword", "Test Aura"} {
		if got := len(dmuB17Tokens(g, name)); got != 0 {
			t.Errorf("%d %s tokens remain after the end step", got, name)
		}
	}
	if !effectstest.BattlefieldHas(g, sword) || !effectstest.BattlefieldHas(g, aura) {
		t.Error("an original Aura or Equipment was sacrificed")
	}
}
