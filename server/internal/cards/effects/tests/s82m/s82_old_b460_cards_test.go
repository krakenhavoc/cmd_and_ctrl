package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	b460SurgeNode   = "3b30f043-aa79-4262-b13f-6cd794b1920b"
	b460Chamber     = "46af64c9-2ab5-4e63-98db-cfb885a96109"
	b460Limitless   = "587461ee-d104-4eef-a4d8-49b1abd56393"
	b460Boundless   = "6047a38c-ec03-428d-9f4f-15d767f8176e"
	b460Piercing    = "9017da9a-2cd5-42fe-a82d-63c34e952344"
	b460Loyal       = "9696ad3d-2e17-4ab7-88ee-3c56a632b7a8"
	b460Etched      = "d0dd582a-ac35-4a46-beca-6304fdd2d358"
	b460Burning     = "e2535383-79f5-4900-8dce-227b0b3650c4"
	b460BoltOracle  = effectstest.LightningBoltOracle
	b460ShrineTypes = "Artifact"
)

func TestOldB460EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		b460SurgeNode: "Surge Node", b460Chamber: "Conversion Chamber", b460Limitless: "Shrine of Limitless Power",
		b460Boundless: "Shrine of Boundless Growth", b460Piercing: "Shrine of Piercing Vision",
		b460Loyal: "Shrine of Loyal Legions", b460Etched: "Etched Monstrosity", b460Burning: "Shrine of Burning Rage",
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

func b460Activate(t *testing.T, g *game.Game, who, source uuid.UUID, idx int, targets ...game.TargetRef) {
	t.Helper()
	if err := g.ActivateCatalogAbility(who, source, idx, game.ActivateAbilityParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("activate %d: %v", idx, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB460SurgeNodeEntersWithSixAndMovesCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Surge Node", "Artifact", b460SurgeNode, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var node uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == b460SurgeNode {
			node = c.InstanceID
		}
	}
	if node == uuid.Nil {
		t.Fatal("Surge Node did not enter")
	}
	if got := effectstest.CountersOn(g, node, game.CounterCharge); got != 6 {
		t.Fatalf("charge counters on entry = %d, want 6", got)
	}
	other := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Other", "Artifact", "")
	effectstest.AddPool(t, g, me.ID, "{C}")
	b460Activate(t, g, me.ID, node, 0, game.TargetRef{Kind: game.TargetCard, ID: other})
	if got := effectstest.CountersOn(g, node, game.CounterCharge); got != 5 {
		t.Errorf("Surge Node counters = %d, want 5", got)
	}
	if got := effectstest.CountersOn(g, other, game.CounterCharge); got != 1 {
		t.Errorf("target artifact counters = %d, want 1", got)
	}
	// A creature is not a legal target.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AddPool(t, g, me.ID, "{C}")
	g.Battlefield.Cards[b460Index(g, node)].Tapped = false
	if err := g.ActivateCatalogAbility(me.ID, node, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Error("a nonartifact creature was accepted as the target")
	}
}

func b460Index(g *game.Game, id uuid.UUID) int {
	for i, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return i
		}
	}
	return -1
}

func TestOldB460ConversionChamberExilesThenMakesAGolem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chamber := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Conversion Chamber", "Artifact", b460Chamber)
	art := effectstest.PushGraveyardCardTyped(opp, "Relic", "Artifact")
	creature := effectstest.PushGraveyardCardTyped(opp, "Bear", "Creature — Bear")
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, chamber, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}}}); err == nil {
		t.Error("a nonartifact card was accepted as the target")
	}
	b460Activate(t, g, me.ID, chamber, 0, game.TargetRef{Kind: game.TargetCard, ID: art})
	if effectstest.GraveyardHas(opp, art) {
		t.Error("the artifact card is still in the graveyard")
	}
	if got := effectstest.CountersOn(g, chamber, game.CounterCharge); got != 1 {
		t.Fatalf("charge counters = %d, want 1", got)
	}
	g.Battlefield.Cards[b460Index(g, chamber)].Tapped = false
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	b460Activate(t, g, me.ID, chamber, 1)
	if got := effectstest.CountersOn(g, chamber, game.CounterCharge); got != 0 {
		t.Errorf("charge counters after the Golem = %d, want 0", got)
	}
	n, _ := effectstest.CountTokensNamed(g, me.ID, "Phyrexian Golem")
	if n != 1 {
		t.Fatalf("Phyrexian Golems = %d, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Golem" && (c.Power != 3 || c.Toughness != 3 || !c.IsArtifact() || !c.IsCreature()) {
			t.Errorf("Golem %+v: want a 3/3 artifact creature", c)
		}
	}
	// With no counter left the second ability cannot be paid.
	g.Battlefield.Cards[b460Index(g, chamber)].Tapped = false
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, chamber, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the Golem ability was paid with no charge counter")
	}
}

func TestOldB460ShrinesCountUpkeepAndTheirColourOfSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	red := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Shrine of Burning Rage", b460ShrineTypes, b460Burning)
	blue := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Shrine of Piercing Vision", b460ShrineTypes, b460Piercing)
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: b460BoltOracle,
		Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: effectstest.B16TargetPlayer(g.Seats[1].ID)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MonSettle(t, g)
	if got := effectstest.CountersOn(g, red, game.CounterCharge); got != 1 {
		t.Errorf("red Shrine after a red spell = %d, want 1", got)
	}
	if got := effectstest.CountersOn(g, blue, game.CounterCharge); got != 0 {
		t.Errorf("blue Shrine after a red spell = %d, want 0", got)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.MonSettle(t, g)
	for name, id := range map[string]uuid.UUID{"red": red, "blue": blue} {
		want := 1
		if name == "red" {
			want = 2
		}
		if got := effectstest.CountersOn(g, id, game.CounterCharge); got != want {
			t.Errorf("%s Shrine after my upkeep = %d, want %d", name, got, want)
		}
	}
}

func TestOldB460ShrineOfBurningRageDealsDamageEqualToCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shrine := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Shrine of Burning Rage", b460ShrineTypes, b460Burning)
	effectstest.SetCounters(g, shrine, map[string]int{game.CounterCharge: 4})
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}")
	before := effectstest.LifeOf(g, opp.ID)
	b460Activate(t, g, me.ID, shrine, 0, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if got := before - effectstest.LifeOf(g, opp.ID); got != 4 {
		t.Errorf("damage dealt = %d, want 4", got)
	}
	if g.Battlefield.Contains(shrine) {
		t.Error("the Shrine was not sacrificed")
	}
}

func TestOldB460ShrineOfLoyalLegionsMakesOneMyrPerCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shrine := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Shrine of Loyal Legions", b460ShrineTypes, b460Loyal)
	effectstest.SetCounters(g, shrine, map[string]int{game.CounterCharge: 3})
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}")
	b460Activate(t, g, me.ID, shrine, 0)
	n, _ := effectstest.CountTokensNamed(g, me.ID, "Phyrexian Myr")
	if n != 3 {
		t.Errorf("Phyrexian Myr tokens = %d, want 3", n)
	}
}

func TestOldB460ShrineOfBoundlessGrowthAddsOneColorlessPerCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shrine := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Shrine of Boundless Growth", b460ShrineTypes, b460Boundless)
	effectstest.SetCounters(g, shrine, map[string]int{game.CounterCharge: 3})
	effectstest.ActivateManaFor(t, g, me.ID, shrine, 0, game.ManaAbilityParams{})
	if got := effectstest.PoolCount(me, "C"); got != 3 {
		t.Errorf("colorless mana = %d, want 3", got)
	}
	if g.Battlefield.Contains(shrine) {
		t.Error("the Shrine was not sacrificed")
	}
}

func TestOldB460ShrineOfLimitlessPowerDiscardsPerCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shrine := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Shrine of Limitless Power", b460ShrineTypes, b460Limitless)
	effectstest.SetCounters(g, shrine, map[string]int{game.CounterCharge: 2})
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{C}")
	hand := opp.Hand.Size()
	b460Activate(t, g, me.ID, shrine, 0, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	p := effectstest.DiscardPromptFor(g, opp)
	if p == nil {
		t.Fatal("no discard prompt for the target player")
	}
	var picks []uuid.UUID
	for _, c := range opp.Hand.Cards[:2] {
		picks = append(picks, c.InstanceID)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, picks...)
	if got := opp.Hand.Size(); got != hand-2 {
		t.Errorf("hand %d → %d, want two discarded", hand, got)
	}
}

func TestOldB460ShrineOfPiercingVisionLooksAtXAndTakesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shrine := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Shrine of Piercing Vision", b460ShrineTypes, b460Piercing)
	effectstest.SetCounters(g, shrine, map[string]int{game.CounterCharge: 3})
	top := effectstest.LibraryTopIDs(me, 4)
	hand := me.Hand.Size()
	b460Activate(t, g, me.ID, shrine, 0)
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no choose-cards prompt")
	}
	if len(p.ChooseCards) != 3 {
		t.Errorf("looked at %d cards, want 3", len(p.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top[1])
	effectstest.MonSettle(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card taken", hand, me.Hand.Size())
	}
	if !effectstest.InZone(me.Hand, top[1]) {
		t.Error("the chosen card is not in hand")
	}
}

func TestOldB460EtchedMonstrosityEntersFivePointFiveAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mon := effectstest.CastCreatureCard(t, g, "Etched Monstrosity", "Artifact Creature — Phyrexian Golem", b460Etched, 10, 10)
	if effectstest.BattlefieldCardFor(g, mon) == nil {
		t.Fatal("Etched Monstrosity did not enter")
	}
	if got := effectstest.CountersOn(g, mon, game.CounterMinusOne); got != 5 {
		t.Fatalf("-1/-1 counters = %d, want 5", got)
	}
	if c, ok := effectstest.BattlefieldCard(g, mon); !ok || c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("size = %d/%d, want 5/5", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.AddPool(t, g, me.ID, "{W}{U}{B}{R}{G}")
	hand := me.Hand.Size()
	b460Activate(t, g, me.ID, mon, 0, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d → %d, want three cards drawn", hand, got)
	}
	if got := effectstest.CountersOn(g, mon, game.CounterMinusOne); got != 0 {
		t.Errorf("-1/-1 counters after the ability = %d, want 0", got)
	}
}
