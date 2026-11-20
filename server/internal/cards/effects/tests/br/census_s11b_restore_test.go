package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// census_s11b_restore_test.go — #4361 S11b: the engine routes whose
// continuation was a closure (zoneRoute.then) now ride the keyed route
// batch with a registered tail. Each test moves a card named Needle (or,
// for the meld, a Urza named Needle), so two order probes ask for their
// order (CR 616.1); the table is restored STRICTLY (the boot path) while
// that prompt is open, the prompt is answered on the restored table, and
// the rest of the effect is checked to have run off the disk:
//
//   - meld (engine/meld-exiled): the two exiled cards enter melded;
//   - a revealed-hand exile, unlinked (Appetite for Brains) and linked
//     (Null Summoner): the card lands in exile, linked or not;
//   - a revealed-hand pick's "then" after a discard (Tourach's Canticle,
//     engine/revealed-pick-then): the random discard follows;
//   - explore (engine/explore-done): a land to hand, and a nonland to the
//     graveyard on a yes, each followed by "has explored";
//   - discover (engine/discovered-to-hand): the lapse to hand announces
//     the discover;
//   - paradigm (engine/paradigm-exiled): the delayed trigger is created;
//   - rebound (engine/rebound-exiled): the upkeep cast is scheduled;
//   - Goliath Daydreamer's exile with a counter (engine/exiled-with-counter);
//   - the Adventure exile (engine/adventure-exiled): the creature is
//     castable from exile.

// s11bStrictAnswer restores g strictly with the frame `frame` open,
// answers the one ordering prompt on the restored table and returns it.
func s11bStrictAnswer(t *testing.T, g *game.Game, frame string) *game.Game {
	t.Helper()
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("pending choices %+v, want the move's ordering prompt", g.PendingChoices)
	}
	g = effectstest.StrictRestoreWithFrame(t, g, frame)
	effectstest.AnswerExileOrder(t, g)
	return g
}

// s11bDelayedLabelled is whether a delayed trigger with this label is
// scheduled.
func s11bDelayedLabelled(g *game.Game, label string) bool {
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Label == label {
			return true
		}
	}
	return false
}

// --- meld ----------------------------------------------------------------

func TestAMeldExileOrderIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, stone := effectstest.PushUrzaPair(g, me.ID)
	effectstest.FindBattlefieldCardForTest(g, urza).Name = "Needle"
	effectstest.SeedExileOrderProbes(g, me.ID)
	if err := effectstest.ActivateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Battlefield.Contains(urza) {
		t.Fatal("Urza left the battlefield before the ordering prompt was answered")
	}

	g = s11bStrictAnswer(t, g, game.ReplacementResumeMoveBatch)
	m := effectstest.MeldedUrza(t, g)
	if !m.IsMelded() || len(m.MeldedFrom) != 2 {
		t.Fatalf("the permanent is not two cards after the restored answer: %+v", m.MeldedFrom)
	}
	for _, id := range []uuid.UUID{urza, stone} {
		if g.Exile.Contains(id) {
			t.Errorf("card %v was left in exile outside the melded permanent", id)
		}
	}
}

// --- a revealed-hand pick ----------------------------------------------------

func s11bNeedleBig() game.Card {
	return game.Card{Name: "Needle", TypeLine: "Sorcery", ManaCost: "{4}{G}{G}{G}", Colors: []string{"G"}}
}

func TestAnAppetiteForBrainsExileOrderIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(victim, s11bNeedleBig())
	effectstest.SeedExileOrderProbes(g, caster.ID)
	effectstest.CastAt(t, g, "Appetite for Brains", "Sorcery", effectstest.AppetiteForBrainsOracle, effectstest.PbPlayer(victim.ID))
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	if g.Exile.Contains(ids[0]) {
		t.Fatal("the card reached exile before the ordering prompt was answered")
	}

	g = s11bStrictAnswer(t, g, game.ReplacementResumeMoveBatch)
	if got := effectstest.CardInExile(t, g, ids[0]).ExiledWith; got.ID != uuid.Nil {
		t.Errorf("an unlinked exile is linked to %+v", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts open after the exile finished", len(g.PendingChoices))
	}
}

func TestANullSummonerExileOrderIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := effectstest.RevealHand(opp, game.Card{Name: "Needle", TypeLine: "Creature — Bear", ManaCost: "{3}{R}{R}", Colors: []string{"R"}, Power: 3, Toughness: 3})
	effectstest.SeedExileOrderProbes(g, me.ID)
	summoner := effectstest.CastCatalogSpell(t, g, "Null Summoner", "Creature — Human Warlock", effectstest.NullSummonerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, ids); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}

	g = s11bStrictAnswer(t, g, game.ReplacementResumeMoveBatch)
	if got := effectstest.CardInExile(t, g, ids[0]).ExiledWith.ID; got != summoner {
		t.Errorf("the restored exile linked the card to %s, want the Summoner %s", got, summoner)
	}
}

func TestATourachsCanticleDiscardOrderIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), game.Card{Name: "Needle", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}})
	effectstest.SeedOrderProbes(g, victim.ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	effectstest.CastAt(t, g, "Tourach's Canticle", "Sorcery", effectstest.TourachsCanticleOracle, effectstest.PbPlayer(victim.ID))
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids[1:]); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	if victim.Hand.Size() != 2 {
		t.Fatalf("hand = %d before the ordering prompt was answered, want both cards", victim.Hand.Size())
	}

	g = s11bStrictAnswer(t, g, game.ReplacementResumeDiscard)
	victim = effectstest.SeatByID(g, victim.ID)
	if !victim.Graveyard.Contains(ids[1]) {
		t.Error("the chosen card was not discarded after the restored answer")
	}
	if victim.Hand.Size() != 0 || !victim.Graveyard.Contains(ids[0]) {
		t.Errorf("hand = %d: the keyed \"then discards a card at random\" did not run off the disk", victim.Hand.Size())
	}
}

// --- explore --------------------------------------------------------------

// s11bMapExplore has a Scout explore with a Map over a top card named
// Needle of the given type, and returns the game, the controller and the
// Scout.
func s11bMapExplore(t *testing.T, typeLine string, probes []string, probeName string) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	scout := effectstest.PushVanillaCreature(g, me.ID, "Scout", 1, 1)
	m := effectstest.PushToken(g, me.ID, effects.MapToken())
	top := effectstest.PutOnTopForTest(me, "Needle", typeLine)
	effectstest.SeedOrderProbes(g, me.ID, probes, probeName)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: scout}},
	}); err != nil {
		t.Fatalf("activate the Map: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, scout, top
}

func s11bExplored(g *game.Game, scout uuid.UUID) int {
	n := 0
	for _, ev := range effectstest.EventsOfKind(g, game.EventExplored) {
		if ev.CardID == scout {
			n++
		}
	}
	return n
}

func TestAnExploredLandsTripToHandIsARestorePoint(t *testing.T) {
	g, me, scout, top := s11bMapExplore(t, "Land", effectstest.LibraryHandOrderProbeOracles, "Library Hand Order Probe")
	if s11bExplored(g, scout) != 0 {
		t.Fatal("the Scout explored before the land reached the hand")
	}

	g = s11bStrictAnswer(t, g, game.ReplacementResumeMoveBatch)
	if !effectstest.SeatByID(g, me.ID).Hand.Contains(top) {
		t.Error("the land did not reach the hand after the restored answer")
	}
	if got := s11bExplored(g, scout); got != 1 {
		t.Errorf("the Scout explored %d times, want once off the disk", got)
	}
}

func TestAnExploredCardsTripToTheGraveyardIsARestorePoint(t *testing.T) {
	g, me, scout, top := s11bMapExplore(t, "Sorcery", effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	effectstest.AnswerExploreQuestion(t, g, me.ID, true)
	if s11bExplored(g, scout) != 0 {
		t.Fatal("the Scout explored before the card reached the graveyard")
	}

	g = s11bStrictAnswer(t, g, game.ReplacementResumeMoveBatch)
	if !effectstest.SeatByID(g, me.ID).Graveyard.Contains(top) {
		t.Error("the revealed card did not reach the graveyard after the restored answer")
	}
	if got := s11bExplored(g, scout); got != 1 {
		t.Errorf("the Scout explored %d times, want once off the disk", got)
	}
}

// --- discover -------------------------------------------------------------

func TestADiscoverLapseToHandIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.PriorityHolder]
	me.Library.Cards = nil
	hit := effectstest.PushLibraryCardForTest(me, effectstest.DiscoverLibraryCard(me.ID, "Needle", "Instant", "{U}"))
	effectstest.SeedOrderProbes(g, me.ID, effectstest.LibraryHandOrderProbeOracles, "Library Hand Order Probe")
	source := uuid.New()
	g.WithWriteLock(func() {
		if err := (effects.Discover{N: 2}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: source})); err != nil {
			t.Fatalf("Discover: %v", err)
		}
	})
	if !g.Exile.Contains(hit) {
		t.Fatal("the discovered card is not in exile")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if len(effectstest.EventsOfKind(g, game.EventDiscover)) != 0 {
		t.Fatal("the discover was announced before the card reached the hand")
	}

	g = s11bStrictAnswer(t, g, game.ReplacementResumeMoveBatch)
	if !effectstest.SeatByID(g, me.ID).Hand.Contains(hit) {
		t.Error("the discovered card did not reach the hand after the restored answer")
	}
	evs := effectstest.EventsOfKind(g, game.EventDiscover)
	if len(evs) != 1 || evs[0].CardID != hit || evs[0].Amount != 2 || evs[0].Source != source {
		t.Errorf("EventDiscover = %+v, want one naming the card, 2 and the source, off the disk", evs)
	}
}

// --- the resolved spell's exits -----------------------------------------------

func TestAParadigmExileOrderIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	effectstest.SeedExileOrderProbes(g, me.ID)
	id := effectstest.CastCatalogSpell(t, g, "Needle", "Sorcery — Lesson", effectstest.DecorumDissertationOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)

	g = s11bStrictAnswer(t, g, game.ReplacementResumeMoveBatch)
	if !effectstest.InExile(g, id) {
		t.Fatal("the paradigm spell is not in exile after the restored answer")
	}
	if !s11bDelayedLabelled(g, "Paradigm — cast a copy of Needle") {
		t.Errorf("no paradigm trigger scheduled off the disk: %+v", g.DelayedTriggers)
	}
}

const s11bStaggershockOracle = "056c3b7d-b603-40b8-8404-18c2eb7e7129"

func TestAReboundExileOrderIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.SeedExileOrderProbes(g, g.Seats[g.Turn.ActiveSeat].ID)
	_, id := effectstest.CastFromHandAtMain(t, g, "Needle", "Instant", s11bStaggershockOracle, effectstest.PlayerTarget(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)

	g = s11bStrictAnswer(t, g, game.ReplacementResumeMoveBatch)
	if !effectstest.InExile(g, id) {
		t.Fatal("the rebound spell is not in exile after the restored answer")
	}
	if !s11bDelayedLabelled(g, "Rebound — cast Needle from exile") {
		t.Errorf("no rebound trigger scheduled off the disk: %+v", g.DelayedTriggers)
	}
}

func TestAGoliathDaydreamerExileOrderIsARestorePoint(t *testing.T) {
	g, _ := effectstest.GoliathGame(t)
	effectstest.SeedExileOrderProbes(g, g.Seats[0].ID)
	id := effectstest.CastCatalogSpell(t, g, "Needle", "Sorcery", effectstest.SpellControlDivinationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	g = s11bStrictAnswer(t, g, game.ReplacementResumeMoveBatch)
	c := effectstest.ExiledCard(g, id)
	if c == nil {
		t.Fatal("the resolved spell is not in exile after the restored answer")
	}
	if c.Counters[game.CounterDream] != 1 {
		t.Errorf("counters = %v, want the dream counter put on off the disk", c.Counters)
	}
}

func TestAnAdventureExileOrderIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := effectstest.FoulmireKnight(me.ID)
	c.Faces[1].Name = "Needle"
	c.SetFace(0)
	me.Hand.PushTop(c)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.SeedExileOrderProbes(g, me.ID)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	g = s11bStrictAnswer(t, g, game.ReplacementResumeMoveBatch)
	if !g.Exile.Contains(c.InstanceID) {
		t.Fatal("the Adventure card is not in exile after the restored answer")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(c.InstanceID); !perm.Granted() || perm.Player != me.ID {
		t.Errorf("CR 715.4 grant = %+v, want the caster's, granted off the disk", perm)
	}
}
