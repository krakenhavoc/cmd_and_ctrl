package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b007_a_test.go — Secret Lair Drop planeswalkers, batch 007 group A
// (#5462): Daretti, Ral, Sarkhan and Vraska, driven through the real engine.

const (
	s82OldB007Daretti = "add0f7e7-8990-4fd1-8ac7-8115f7802912"
	s82OldB007Ral     = "b7da26c9-78a6-41c5-a5ef-6555e8274d7a"
	s82OldB007Sarkhan = "c2c1e44c-2a19-466f-a6d5-4cb71a940b86"
	s82OldB007Vraska  = "cff8b4e9-c60c-42c1-ad2e-74ae9d7f3afb"
)

func TestS82OldB007ACardsAreFullWithNoCaveats(t *testing.T) {
	for name, oracle := range map[string]string{
		"Daretti, Ingenious Iconoclast": s82OldB007Daretti, "Ral, Storm Conduit": s82OldB007Ral,
		"Sarkhan, Dragonsoul": s82OldB007Sarkhan, "Vraska, Swarm's Eminence": s82OldB007Vraska,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v", name, spec, ok)
		}
	}
}

func s82OldB007Keyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) {
	t.Helper()
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if c := &g.Battlefield.Cards[i]; c.InstanceID == id {
				c.Keywords = append(c.Keywords, kw)
			}
		}
	})
}

func s82OldB007Named(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// --- Vraska, Swarm's Eminence ---------------------------------------------------

func TestS82OldB007AVraskaMinusTwoMakesADeathtouchAssassin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Swarm's Eminence", s82OldB007Vraska, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	toks := s82OldB007Named(g, "Assassin")
	if len(toks) != 1 {
		t.Fatalf("%d Assassins, want 1", len(toks))
	}
	a := toks[0]
	if a.Power != 1 || a.Toughness != 1 || !game.HasKeyword(&a, "deathtouch") || a.Controller != me.ID {
		t.Errorf("Assassin is %d/%d, keywords %v, controller %v", a.Power, a.Toughness, a.Keywords, a.Controller)
	}
	if got := effectstest.LoyaltyOf(g, vraska); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

func TestS82OldB007AVraskaPutsACounterOnADeathtouchCreatureThatHitsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.S82Walker(g, me.ID, "Vraska, Swarm's Eminence", s82OldB007Vraska, 5)
	snake := effectstest.B12Creature(g, me.ID, "Snake", "Creature — Snake", 1, 1)
	s82OldB007Keyword(t, g, snake, "deathtouch")
	effectstest.DealNoncombatDamage(g, me.ID, snake, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, snake, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters on the deathtouch creature: %d, want 1", n)
	}
	effectstest.DealCombatDamage(g, me.ID, snake, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, snake, game.CounterPlusOne); n != 2 {
		t.Errorf("after a second hit: %d counters, want 2", n)
	}
}

func TestS82OldB007AVraskaPutsACounterWhenADeathtouchCreatureHitsAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.S82Walker(g, me.ID, "Vraska, Swarm's Eminence", s82OldB007Vraska, 5)
	enemy := effectstest.S82Walker(g, opp.ID, "Test Walker", "", 9)
	snake := effectstest.B12Creature(g, me.ID, "Snake", "Creature — Snake", 1, 1)
	s82OldB007Keyword(t, g, snake, "deathtouch")
	effectstest.DealCombatDamage(g, me.ID, snake, enemy)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, snake, game.CounterPlusOne); n != 1 {
		t.Errorf("counters after hitting a planeswalker: %d, want 1", n)
	}
}

func TestS82OldB007AVraskaIgnoresOtherDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.S82Walker(g, me.ID, "Vraska, Swarm's Eminence", s82OldB007Vraska, 5)
	plain := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	snake := effectstest.B12Creature(g, me.ID, "Snake", "Creature — Snake", 1, 1)
	s82OldB007Keyword(t, g, snake, "deathtouch")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Snake", "Creature — Snake", 1, 1)
	s82OldB007Keyword(t, g, theirs, "deathtouch")
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	effectstest.DealCombatDamage(g, me.ID, plain, opp.ID)  // no deathtouch
	effectstest.DealCombatDamage(g, opp.ID, theirs, me.ID) // not yours
	effectstest.DealCombatDamage(g, me.ID, snake, victim)  // a creature, not a player
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"plain": plain, "snake": snake, "theirs": theirs} {
		if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != 0 {
			t.Errorf("%s got %d counters, want none", name, n)
		}
	}
}

func TestS82OldB007AAssassinTokenDestroysAPlaneswalkerItDamages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Swarm's Eminence", s82OldB007Vraska, 5)
	enemy := effectstest.S82Walker(g, opp.ID, "Test Walker", "", 9)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	toks := s82OldB007Named(g, "Assassin")
	if len(toks) != 1 {
		t.Fatalf("%d Assassins", len(toks))
	}
	effectstest.DealCombatDamage(g, me.ID, toks[0].InstanceID, enemy)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, enemy); ok {
		t.Error("the damaged planeswalker is still on the battlefield")
	}
	// The token's hit also earns Vraska's counter, since it has deathtouch.
	if n := effectstest.CounterCount(g, toks[0].InstanceID, game.CounterPlusOne); n != 1 {
		t.Errorf("Assassin counters %d, want 1", n)
	}
}

func TestS82OldB007AAssassinTokenLeavesAPlayerOrCreatureHitAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Swarm's Eminence", s82OldB007Vraska, 5)
	enemy := effectstest.S82Walker(g, opp.ID, "Test Walker", "", 9)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	tok := s82OldB007Named(g, "Assassin")[0].InstanceID
	effectstest.DealCombatDamage(g, me.ID, tok, opp.ID)
	effectstest.DealCombatDamage(g, me.ID, tok, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, enemy); !ok {
		t.Error("an unrelated planeswalker was destroyed")
	}
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the creature was destroyed by the token's trigger")
	}
}

// --- Daretti, Ingenious Iconoclast ------------------------------------------------

func TestS82OldB007ADarettiPlusOneMakesADefenderConstruct(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	daretti := effectstest.S82Walker(g, me.ID, "Daretti, Ingenious Iconoclast", s82OldB007Daretti, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, daretti, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	toks := s82OldB007Named(g, "Construct")
	if len(toks) != 1 {
		t.Fatalf("%d Constructs, want 1", len(toks))
	}
	c := toks[0]
	if c.Power != 1 || c.Toughness != 1 || !game.HasKeyword(&c, "defender") || len(c.Colors) != 0 || !c.IsArtifact() || !c.IsCreature() {
		t.Errorf("Construct is %d/%d %q colours %v keywords %v", c.Power, c.Toughness, c.TypeLine, c.Colors, c.Keywords)
	}
	if got := effectstest.LoyaltyOf(g, daretti); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

func TestS82OldB007ADarettiMinusOneSacrificesAnArtifactToDestroyACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	daretti := effectstest.S82Walker(g, me.ID, "Daretti, Ingenious Iconoclast", s82OldB007Daretti, 3)
	trinket := effectstest.PushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, daretti, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, trinket)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, trinket); ok {
		t.Error("the artifact was not sacrificed")
	}
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("the target creature survived")
	}
	if got := effectstest.LoyaltyOf(g, daretti); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestS82OldB007ADarettiMinusOneCanDestroyAnArtifactToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	daretti := effectstest.S82Walker(g, me.ID, "Daretti, Ingenious Iconoclast", s82OldB007Daretti, 3)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Relic", "Artifact", "", false)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, daretti, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("the target artifact survived")
	}
}

func TestS82OldB007ADarettiMinusOneDeclinedDestroysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	daretti := effectstest.S82Walker(g, me.ID, "Daretti, Ingenious Iconoclast", s82OldB007Daretti, 3)
	trinket := effectstest.PushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, daretti, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID) // sacrifice nothing
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the creature died although no artifact was sacrificed")
	}
	if _, ok := effectstest.BattlefieldCard(g, trinket); !ok {
		t.Error("the artifact was sacrificed although the player declined")
	}
}

func TestS82OldB007ADarettiMinusOneWithoutAnArtifactDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	daretti := effectstest.S82Walker(g, me.ID, "Daretti, Ingenious Iconoclast", s82OldB007Daretti, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, daretti, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the creature was destroyed with no artifact to sacrifice")
	}
}

func TestS82OldB007ADarettiMinusOneRefusesAnIllegalTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	daretti := effectstest.S82Walker(g, me.ID, "Daretti, Ingenious Iconoclast", s82OldB007Daretti, 3)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.AdvanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, daretti, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(land)})
	if err == nil {
		t.Error("a land is neither an artifact nor a creature, but the activation was accepted")
	}
}

func TestS82OldB007ADarettiMinusSixCopiesAnArtifactOnTheBattlefieldThreeTimes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	daretti := effectstest.S82Walker(g, me.ID, "Daretti, Ingenious Iconoclast", s82OldB007Daretti, 6)
	relic := effectstest.PushCatalogPermanent(g, opp.ID, "Their Relic", "Artifact", "", false)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, daretti, 2, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(relic)})
	effectstest.PassPriorityAroundTable(t, g)
	copies := 0
	for _, c := range s82OldB007Named(g, "Their Relic") {
		if c.InstanceID != relic && c.Controller == me.ID && c.IsToken() {
			copies++
		}
	}
	if copies != 3 {
		t.Errorf("%d token copies under my control, want 3", copies)
	}
}

func TestS82OldB007ADarettiMinusSixCopiesAnArtifactCardFromAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	daretti := effectstest.S82Walker(g, me.ID, "Daretti, Ingenious Iconoclast", s82OldB007Daretti, 6)
	card := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Buried Relic", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, daretti, 2, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(card)})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(s82OldB007Named(g, "Buried Relic")); n != 3 {
		t.Errorf("%d copies of the graveyard artifact on the battlefield, want 3", n)
	}
	if !opp.Graveyard.Contains(card) {
		t.Error("the card left the graveyard; only copies are made")
	}
}

func TestS82OldB007ADarettiMinusSixRefusesANonArtifactCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	daretti := effectstest.S82Walker(g, me.ID, "Daretti, Ingenious Iconoclast", s82OldB007Daretti, 6)
	card := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, daretti, 2, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(card)})
	if err == nil {
		t.Error("a creature card is not an artifact, but the activation was accepted")
	}
}

// --- Ral, Storm Conduit ---------------------------------------------------------

func TestS82OldB007ARalPlusTwoScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ral := effectstest.S82Walker(g, me.ID, "Ral, Storm Conduit", s82OldB007Ral, 4)
	effectstest.SeedLibraryTop(me, "Top", "Sorcery", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ral, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if got := effectstest.LoyaltyOf(g, ral); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestS82OldB007ARalPingsAnOpponentWhenYouCastAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.S82Walker(g, me.ID, "Ral, Storm Conduit", s82OldB007Ral, 4)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.B02dCastInstantAs(t, g, me, "Idle Instant", nil)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
}

func TestS82OldB007ARalIgnoresACreatureSpellAndAnOpponentsInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.S82Walker(g, me.ID, "Ral, Storm Conduit", s82OldB007Ral, 4)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.CastCreatureCard(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("a creature spell pinged: life %d, want %d", opp.Life, life)
	}
}

func TestS82OldB007ARalMinusTwoCopiesTheNextInstantAndBothPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ral := effectstest.S82Walker(g, me.ID, "Ral, Storm Conduit", s82OldB007Ral, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ral, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.B02dCastInstantAs(t, g, me, "Idle Instant", nil)
	for i := 0; i < 4 && len(g.PendingChoices) > 0; i++ {
		effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.PassPriorityAroundTable(t, g)
	copied := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventSpellCopied {
			copied++
		}
	}
	if copied != 1 {
		t.Errorf("%d spell copies, want 1", copied)
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d (one ping for the cast, one for the copy)", opp.Life, life-2)
	}
	// Only the NEXT instant or sorcery is copied.
	effectstest.B02dCastInstantAs(t, g, me, "Second Instant", nil)
	for i := 0; i < 4 && len(g.PendingChoices) > 0; i++ {
		effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.PassPriorityAroundTable(t, g)
	copied = 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventSpellCopied {
			copied++
		}
	}
	if copied != 1 {
		t.Errorf("%d spell copies after a second instant, want still 1", copied)
	}
}

// --- Sarkhan, Dragonsoul --------------------------------------------------------

func TestS82OldB007ASarkhanPlusTwoHitsOnlyOpponentsAndTheirCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	sarkhan := effectstest.S82Walker(g, me.ID, "Sarkhan, Dragonsoul", s82OldB007Sarkhan, 5)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, sarkhan, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 1
		if s.ID == me.ID {
			want = lives[s.ID]
		}
		if s.Life != want {
			t.Errorf("seat %v life %d, want %d", s.ID, s.Life, want)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.DamageMarked != 1 {
		t.Errorf("their creature has %d damage, want 1", c.DamageMarked)
	}
	if c, _ := effectstest.BattlefieldCard(g, mine); c.DamageMarked != 0 {
		t.Errorf("my creature has %d damage, want 0", c.DamageMarked)
	}
	if got := effectstest.LoyaltyOf(g, sarkhan); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestS82OldB007ASarkhanMinusThreeDealsFourToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	sarkhan := effectstest.S82Walker(g, me.ID, "Sarkhan, Dragonsoul", s82OldB007Sarkhan, 5)
	life := opp.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, sarkhan, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d, want %d", opp.Life, life-4)
	}
	if got := effectstest.LoyaltyOf(g, sarkhan); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestS82OldB007ASarkhanMinusThreeDealsFourToAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	sarkhan := effectstest.S82Walker(g, me.ID, "Sarkhan, Dragonsoul", s82OldB007Sarkhan, 5)
	enemy := effectstest.S82Walker(g, opp.ID, "Test Walker", "", 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, sarkhan, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(enemy)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, enemy); got != 2 {
		t.Errorf("planeswalker loyalty %d, want 2", got)
	}
}

func TestS82OldB007ASarkhanMinusThreeRefusesACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	sarkhan := effectstest.S82Walker(g, me.ID, "Sarkhan, Dragonsoul", s82OldB007Sarkhan, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sarkhan, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(bear)}); err == nil {
		t.Error("a creature is not a player or planeswalker, but the activation was accepted")
	}
}

func TestS82OldB007ASarkhanUltimateFetchesEveryDragonCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sarkhan := effectstest.S82Walker(g, me.ID, "Sarkhan, Dragonsoul", s82OldB007Sarkhan, 9)
	d1 := effectstest.SeedLibraryTop(me, "Red Dragon", "Creature — Dragon", "{4}{R}")
	d2 := effectstest.SeedLibraryTop(me, "Blue Dragon", "Creature — Dragon", "{4}{U}")
	wurm := effectstest.SeedLibraryTop(me, "Wurm", "Creature — Wurm", "{4}")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, sarkhan, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, d1, d2)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{d1, d2} {
		if _, ok := effectstest.BattlefieldCard(g, id); !ok {
			t.Errorf("Dragon %v did not enter the battlefield", id)
		}
	}
	if _, ok := effectstest.BattlefieldCard(g, wurm); ok {
		t.Error("a non-Dragon entered the battlefield")
	}
}
