package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestNghcCardsAreRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.NghcForgottenCreation:  "Forgotten Creation",
		effectstest.NghcEndlessEvil:        "Endless Evil",
		effectstest.NghcNihilith:           "Nihilith",
		effectstest.NghcOverchargedAmalgam: "Overcharged Amalgam",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// nghcHandIDs lists the instance IDs in a seat's hand.
func nghcHandIDs(p *game.Player) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(p.Hand.Cards))
	for _, c := range p.Hand.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}

// nghcNamedCards seeds n uniquely named cards onto the TOP of a seat's
// library (so the next draws are these), returning their IDs in draw
// order.
func nghcNamedLibraryTop(p *game.Player, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		ids[i] = uuid.New()
		p.Library.PushTop(game.Card{
			InstanceID: ids[i], Name: names[i], TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID,
		})
	}
	return ids
}

// --- Forgotten Creation -------------------------------------------------

func nghcForgottenCreationAtUpkeep(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fc := effectstest.NghcPushCreature(g, me.ID, "Forgotten Creation", "Creature — Zombie Horror", effectstest.NghcForgottenCreation, 3, 3)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	return g, me, fc
}

func TestNghcForgottenCreationHasSkulk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	fc := effectstest.NghcPushCreature(g, g.Seats[0].ID, "Forgotten Creation", "Creature — Zombie Horror", effectstest.NghcForgottenCreation, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, fc, "skulk") {
		t.Error("Forgotten Creation printed skulk")
	}
}

func TestNghcForgottenCreationDiscardsTheHandAndDrawsThatMany(t *testing.T) {
	g, me, _ := nghcForgottenCreationAtUpkeep(t)
	old := nghcHandIDs(me)
	if len(old) == 0 {
		t.Fatal("the fixture hand is empty")
	}
	drawn := nghcNamedLibraryTop(me, "Drawn A", "Drawn B", "Drawn C", "Drawn D", "Drawn E", "Drawn F", "Drawn G", "Drawn H", "Drawn I", "Drawn J")
	libBefore := me.Library.Size()

	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range old {
		if me.Hand.Contains(id) {
			t.Errorf("card %s is still in hand", id)
		}
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %s was not discarded", id)
		}
	}
	if got := me.Hand.Size(); got != len(old) {
		t.Errorf("hand size = %d, want the %d cards discarded", got, len(old))
	}
	if got := me.Library.Size(); got != libBefore-len(old) {
		t.Errorf("library %d -> %d, want %d drawn", libBefore, got, len(old))
	}
	for i := 0; i < len(old); i++ {
		if !me.Hand.Contains(drawn[i]) {
			t.Errorf("the draw skipped the top card %d", i)
		}
	}
}

func TestNghcForgottenCreationMayDecline(t *testing.T) {
	g, me, _ := nghcForgottenCreationAtUpkeep(t)
	old := nghcHandIDs(me)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range old {
		if !me.Hand.Contains(id) {
			t.Errorf("declined, but card %s left the hand", id)
		}
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("declined, but the graveyard holds %d cards", me.Graveyard.Size())
	}
}

func TestNghcForgottenCreationWithAnEmptyHandDrawsNothing(t *testing.T) {
	g, me, _ := nghcForgottenCreationAtUpkeep(t)
	me.Hand.Cards = nil
	libBefore := me.Library.Size()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 || me.Library.Size() != libBefore {
		t.Errorf("an empty hand drew: hand %d, library %d -> %d", me.Hand.Size(), libBefore, me.Library.Size())
	}
}

func TestNghcForgottenCreationOnlyTriggersOnItsControllersUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.NghcPushCreature(g, me.ID, "Forgotten Creation", "Creature — Zombie Horror", effectstest.NghcForgottenCreation, 3, 3)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("asked on another player's upkeep")
	}
}

// Information safety: the discarded cards are public in the graveyard,
// the cards drawn afterwards stay hidden from every other seat.
func TestNghcForgottenCreationRedactsTheDrawnCardsPerSeat(t *testing.T) {
	g, me, _ := nghcForgottenCreationAtUpkeep(t)
	opp := g.Seats[1]
	nghcNamedLibraryTop(me, "Secret A", "Secret B", "Secret C", "Secret D", "Secret E", "Secret F", "Secret G", "Secret H", "Secret I", "Secret J")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	mineView := protocol.ViewOfGameFor(g, me.ID.String())
	theirs := protocol.ViewOfGameFor(g, opp.ID.String())
	var mineHand, theirHand []string
	for _, c := range mineView.Seats[0].Hand.Cards {
		mineHand = append(mineHand, c.Name)
	}
	for _, c := range theirs.Seats[0].Hand.Cards {
		theirHand = append(theirHand, c.Name)
	}
	if len(mineHand) == 0 || mineHand[0] == "" {
		t.Fatalf("the drawing seat sees no names in its own hand: %v", mineHand)
	}
	for _, n := range theirHand {
		if n != "" {
			t.Errorf("an opponent's view of the hand leaks %q", n)
		}
	}
	if theirs.Seats[0].Hand.Count != me.Hand.Size() {
		t.Errorf("the opponent's hand count = %d, want %d", theirs.Seats[0].Hand.Count, me.Hand.Size())
	}
	if len(theirs.Seats[0].Graveyard.Cards) == 0 {
		t.Error("the discarded cards are not visible in the graveyard")
	}
}

// --- Endless Evil -------------------------------------------------------

func nghcHorror(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.NghcPushCreature(g, owner, "Nihilith", "Creature — Horror", effectstest.NghcNihilith, 4, 4)
}

func nghcCountNamed(g *game.Game, controller uuid.UUID, name string, token bool) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == name && c.IsToken() == token {
				n++
			}
		}
	})
	return n
}

func TestNghcEndlessEvilMakesAOneOneTokenCopyEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	horror := nghcHorror(g, me.ID)
	effectstest.AuraCast(t, g, "Endless Evil", effectstest.NghcEndlessEvil, horror)

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if got := nghcCountNamed(g, me.ID, "Nihilith", true); got != 1 {
		t.Fatalf("%d token copies, want 1", got)
	}
	var tok uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.IsToken() && c.Name == "Nihilith" {
				tok = c.InstanceID
			}
		}
	})
	if p, tt := effectstest.EffectivePower(t, g, tok), effectstest.EffectiveToughness(t, g, tok); p != 1 || tt != 1 {
		t.Errorf("the token is %d/%d, want 1/1", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, tok, "fear") {
		t.Error("the copy lost the copied creature's fear")
	}
	if p := effectstest.EffectivePower(t, g, horror); p != 4 {
		t.Errorf("the original shrank to %d power", p)
	}
	if !g.Battlefield.Contains(tok) {
		t.Error("the token is not on the battlefield")
	}
}

func TestNghcEndlessEvilIgnoresOtherPlayersUpkeeps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	horror := nghcHorror(g, me.ID)
	effectstest.AuraCast(t, g, "Endless Evil", effectstest.NghcEndlessEvil, horror)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := nghcCountNamed(g, me.ID, "Nihilith", true); got != 0 {
		t.Errorf("%d tokens on another player's upkeep, want 0", got)
	}
}

func TestNghcEndlessEvilEnchantsOnlyCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := nghcHorror(g, opp.ID)
	err := effectstest.CastCatalogSpellErr(t, g, "Endless Evil", effectstest.AuraTypeLine, effectstest.NghcEndlessEvil,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	if err == nil {
		t.Error("Endless Evil was cast on an opponent's creature")
	}
}

func TestNghcEndlessEvilReturnsToHandWhenAHorrorDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	horror := nghcHorror(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Endless Evil", effectstest.NghcEndlessEvil, horror)
	effectstest.KillCreature(t, g, me.ID, horror)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) || me.Graveyard.Contains(aura) {
		t.Error("the Aura did not return to its owner's hand")
	}
}

func TestNghcEndlessEvilStaysInTheGraveyardWhenANonHorrorDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.AuraCast(t, g, "Endless Evil", effectstest.NghcEndlessEvil, bear)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(aura) || !me.Graveyard.Contains(aura) {
		t.Error("the Aura returned although the creature was not a Horror")
	}
}

func TestNghcEndlessEvilDoesNothingWhenTheCreatureIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	horror := nghcHorror(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Endless Evil", effectstest.NghcEndlessEvil, horror)
	effectstest.ExileCreature(t, g, horror)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(aura) {
		t.Error("the Aura returned although the creature did not die")
	}
}

// The Aura returns only if it is still the card that fell off: one
// exiled in response stays exiled (CR 400.7).
func TestNghcEndlessEvilDoesNotFollowAnAuraExiledInResponse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	horror := nghcHorror(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Endless Evil", effectstest.NghcEndlessEvil, horror)
	effectstest.KillCreature(t, g, me.ID, horror)
	if effectstest.StackFullyEmpty(g) {
		t.Fatal("the dies trigger never reached the stack")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, aura); err != nil {
		t.Fatalf("exile the Aura in response: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(aura) {
		t.Error("the Aura came back from exile")
	}
}

// "This card" is the object that fell off: an Aura that left the
// graveyard and came back to it while the trigger waited is a new object
// the ability cannot find (CR 400.7).
func TestNghcEndlessEvilDoesNotFollowAnAuraThatCameBackAsANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	horror := nghcHorror(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Endless Evil", effectstest.NghcEndlessEvil, horror)
	effectstest.KillCreature(t, g, me.ID, horror)
	if effectstest.StackFullyEmpty(g) {
		t.Fatal("the dies trigger never reached the stack")
	}
	gy := game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}
	if err := g.MoveCardByID(gy, game.ZoneRef{Kind: game.ZoneExile}, aura); err != nil {
		t.Fatalf("exile the Aura: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, gy, aura); err != nil {
		t.Fatalf("return it to the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(aura) {
		t.Error("the Aura returned although it is no longer the object that fell off")
	}
}

// The Horror check reads the creature as it last existed on the
// battlefield: a creature that was a Horror only through changeling
// (every creature type, CR 702.73a) counts, though it prints no type.
func TestNghcEndlessEvilJudgesTheHorrorTypeAsItLastExisted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shifter", TypeLine: "Creature — Shapeshifter",
		Keywords: []string{game.KeywordChangeling},
		Power:    2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	aura := effectstest.AuraCast(t, g, "Endless Evil", effectstest.NghcEndlessEvil, bear)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("a creature that was a Horror when it died did not bring the Aura back")
	}
}

// --- Nihilith -----------------------------------------------------------

func nghcTimeCounters(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			n = c.Counters[game.CounterTime]
		}
	})
	return n
}

// nghcSuspendNihilith pays {1}{B} and suspends a Nihilith from the
// active seat's hand.
func nghcSuspendNihilith(t *testing.T, g *game.Game) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.HandCardForTest(me, "Nihilith", "Creature — Horror", effectstest.NghcNihilith)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend Nihilith: %v", err)
	}
	return me, id
}

// nghcOppDiscards empties the seat's hand and discards exactly one
// fresh card from it, returning that card.
func nghcOppDiscards(t *testing.T, g *game.Game, opp *game.Player) uuid.UUID {
	t.Helper()
	opp.Hand.Cards = nil
	id := effectstest.HandCard(opp, "Discarded", "Sorcery")
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(opp.ID, 1); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	return id
}

func TestNghcNihilithDeclaresFearAndSuspendSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := nghcHorror(g, g.Seats[0].ID)
	if !effectstest.HasEffectiveKeyword(t, g, id, "fear") {
		t.Error("Nihilith printed fear")
	}
	sa := effectstest.SuspendOf(effectstest.NghcNihilith)
	if sa == nil || sa.Counters != 7 || sa.Cost != "{1}{B}" {
		t.Errorf("suspend = %+v, want 7 — {1}{B}", sa)
	}
}

func TestNghcNihilithRemovesATimeCounterWhenACardHitsAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := nghcSuspendNihilith(t, g)
	opp := g.Seats[1]
	if got := nghcTimeCounters(g, id); got != 7 {
		t.Fatalf("suspended with %d counters, want 7", got)
	}
	nghcOppDiscards(t, g, opp)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := nghcTimeCounters(g, id); got != 6 {
		t.Errorf("after one card: %d counters, want 6", got)
	}
}

func TestNghcNihilithMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := nghcSuspendNihilith(t, g)
	nghcOppDiscards(t, g, g.Seats[1])
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := nghcTimeCounters(g, id); got != 7 {
		t.Errorf("declined, but counters = %d", got)
	}
}

func TestNghcNihilithAsksOncePerCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := nghcSuspendNihilith(t, g)
	opp := g.Seats[1]
	opp.Hand.Cards = nil
	effectstest.HandCard(opp, "A", "Sorcery")
	effectstest.HandCard(opp, "B", "Sorcery")
	effectstest.HandCard(opp, "C", "Sorcery")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 3) })
	// One trigger per card: three prompts, answered one after another.
	yes := 0
	for i := 0; i < 8; i++ {
		if effectstest.LatestTriggerPrompt(g, me.ID) == nil {
			effectstest.PassPriorityAroundTable(t, g)
		}
		if effectstest.LatestTriggerPrompt(g, me.ID) == nil {
			break
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		yes++
	}
	effectstest.PassPriorityAroundTable(t, g)
	if yes != 3 {
		t.Errorf("%d prompts for three cards, want one each", yes)
	}
	if got := nghcTimeCounters(g, id); got != 7-yes {
		t.Errorf("counters = %d, want %d", got, 7-yes)
	}
}

func nghcTriggerPrompts(g *game.Game, chooser uuid.UUID) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			n++
		}
	}
	return n
}

func TestNghcNihilithIgnoresYourOwnGraveyardTokensAndAlreadyBuriedCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := nghcSuspendNihilith(t, g)
	opp := g.Seats[1]

	// Your own graveyard.
	effectstest.HandCard(me, "Mine", "Sorcery")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if nghcTriggerPrompts(g, me.ID) != 0 {
		t.Error("a card put into your own graveyard asked")
	}

	// A token dying is not a card.
	tok := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tok) })
	effectstest.PassPriorityAroundTable(t, g)
	if nghcTriggerPrompts(g, me.ID) != 0 {
		t.Error("a token dying asked (a token is not a card)")
	}

	// A card moving from one graveyard to another place is not "put into".
	exiled := nghcOppDiscards(t, g, opp)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneExile}, exiled); err != nil {
		t.Fatalf("exile from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if nghcTriggerPrompts(g, me.ID) != 0 {
		t.Error("a card leaving a graveyard asked")
	}
}

func TestNghcNihilithNeedsToBeSuspended(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// A Nihilith exiled by something else has no time counters.
	card := game.Card{InstanceID: uuid.New(), Name: "Nihilith", TypeLine: "Creature — Horror", OracleID: effectstest.NghcNihilith, Owner: me.ID, Controller: me.ID}
	g.Exile.PushTop(card)
	nghcOppDiscards(t, g, g.Seats[1])
	effectstest.PassPriorityAroundTable(t, g)
	if nghcTriggerPrompts(g, me.ID) != 0 {
		t.Error("a Nihilith with no time counters asked (it is not suspended)")
	}
}

func TestNghcNihilithLastCounterOffersTheFreeCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := nghcSuspendNihilith(t, g)
	opp := g.Seats[1]
	for i := 0; i < 7; i++ {
		nghcOppDiscards(t, g, opp)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := nghcTimeCounters(g, id); got != 0 {
		t.Fatalf("after seven cards: %d counters, want 0", got)
	}
	offer := effectstest.LatestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatal("removing the last time counter offered no free cast")
	}
}

// The ability re-checks that the card is still suspended as it resolves
// (CR 603.4): a last counter removed in response leaves nothing to take.
func TestNghcNihilithResolutionRechecksTheSuspendedState(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := nghcSuspendNihilith(t, g)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterTime, -6); err != nil {
			t.Fatalf("set up one counter: %v", err)
		}
	})
	nghcOppDiscards(t, g, g.Seats[1])
	g.RunStateChecksForTest()
	if effectstest.StackFullyEmpty(g) {
		t.Fatal("the ability never reached the stack")
	}
	// In response the last counter comes off (the free cast is offered
	// and declined), so the card is no longer suspended when it resolves.
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterTime, -1); err != nil {
			t.Fatalf("remove the last counter in response: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	offer := effectstest.LatestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatal("the last counter coming off offered no free cast")
	}
	if err := g.ResolvePendingChoice(offer.ID, me.ID, nil); err != nil {
		t.Fatalf("decline the free cast: %v", err)
	}
	// The ability still on the stack finds the card no longer suspended
	// (CR 603.4): if it asks at all, the answer changes nothing.
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestTriggerPrompt(g, me.ID); c != nil {
		if err := g.ResolveTriggerPrompt(c.ID, me.ID, true); err != nil {
			t.Fatalf("answer: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := nghcTimeCounters(g, id); got != 0 {
		t.Errorf("counters = %d, want 0 (the ability must do nothing to an unsuspended card)", got)
	}
}

// The ability finds the OBJECT it triggered from: a Nihilith that left
// exile and came back suspended again while the ability waited is a new
// object (CR 400.7), and loses nothing to it.
func TestNghcNihilithDoesNotFindANewlySuspendedObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := nghcSuspendNihilith(t, g)
	nghcOppDiscards(t, g, g.Seats[1])
	g.RunStateChecksForTest()
	if effectstest.StackFullyEmpty(g) {
		t.Fatal("the ability never reached the stack")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, id); err != nil {
		t.Fatalf("take it back to hand: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatalf("exile it again: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterTime, 7); err != nil {
			t.Fatalf("suspend the new object: %v", err)
		}
	})
	if c := effectstest.LatestTriggerPrompt(g, me.ID); c != nil {
		t.Fatal("the \"you may\" is asked as the ability resolves, not before")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := nghcTimeCounters(g, id); got != 7 {
		t.Errorf("counters = %d, want 7: the ability took a counter from a different object", got)
	}
}

// Information safety: a suspended card is exiled face up, so every seat
// sees it and its time counters.
func TestNghcNihilithSuspendedCardIsPublicToEverySeat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, id := nghcSuspendNihilith(t, g)
	for _, seat := range g.Seats {
		view := protocol.ViewOfGameFor(g, seat.ID.String())
		found := false
		for _, c := range view.Exile.Cards {
			if c.InstanceID == id.String() {
				found = true
				if c.Name != "Nihilith" || c.Counters[game.CounterTime] != 7 {
					t.Errorf("seat %s sees %q with %d time counters", seat.ID, c.Name, c.Counters[game.CounterTime])
				}
			}
		}
		if !found {
			t.Errorf("seat %s cannot see the suspended card in exile", seat.ID)
		}
	}
}

// --- Overcharged Amalgam / exploit --------------------------------------

func nghcAmalgamCard(owner *game.Player) uuid.UUID {
	id := uuid.New()
	owner.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Overcharged Amalgam", TypeLine: "Creature — Zombie Horror",
		OracleID: effectstest.NghcOverchargedAmalgam, Power: 3, Toughness: 3, Owner: owner.ID, Controller: owner.ID,
	})
	return id
}

func nghcCastAmalgam(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := nghcAmalgamCard(me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Overcharged Amalgam: %v", err)
	}
	return id
}

func TestNghcOverchargedAmalgamHasFlashAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.NghcPushCreature(g, g.Seats[0].ID, "Overcharged Amalgam", "Creature — Zombie Horror", effectstest.NghcOverchargedAmalgam, 3, 3)
	for _, kw := range []string{"flash", "flying"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Overcharged Amalgam printed %s", kw)
		}
	}
}

// nghcBoltOnStack has the active seat cast a Lightning Bolt at an
// opponent and keep priority, so the spell sits on the stack.
func nghcBoltOnStack(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	bolt := effectstest.HandCardForTest(me, "Lightning Bolt", "Instant", effectstest.NghcLightningBolt)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}},
	}); err != nil {
		t.Fatalf("cast Lightning Bolt: %v", err)
	}
	return bolt
}

func TestNghcOverchargedAmalgamExploitingACreatureCountersTheSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	bolt := nghcBoltOnStack(t, g, me)
	life := opp.Life

	amalgam := nghcCastAmalgam(t, g, me)
	effectstest.NghcPassOnce(t, g) // the Amalgam resolves; its exploit ability goes on the stack
	effectstest.NghcPassOnce(t, g) // the exploit ability resolves and asks

	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil || p.ChooseMin != 0 {
		t.Fatal("exploit asks \"you may sacrifice a creature\" with a floor of zero")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range p.ChooseCards {
		offered[id] = true
	}
	if !offered[fodder] || !offered[amalgam] {
		t.Errorf("the candidates should be every creature you control, itself included: %v", p.ChooseCards)
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature is sacrificed before the counter trigger asks for its target")
	}
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != life {
		t.Errorf("the Bolt resolved: life %d -> %d", life, opp.Life)
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
	if !g.Battlefield.Contains(amalgam) {
		t.Error("the Amalgam should still be in play")
	}
}

func TestNghcOverchargedAmalgamSacrificingItselfStillCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := nghcBoltOnStack(t, g, me)
	life := opp.Life
	amalgam := nghcCastAmalgam(t, g, me)
	effectstest.NghcPassOnce(t, g)
	effectstest.NghcPassOnce(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, amalgam)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(amalgam) {
		t.Error("the Amalgam should have been sacrificed")
	}
	if opp.Life != life || !me.Graveyard.Contains(bolt) {
		t.Errorf("the Bolt was not countered (life %d -> %d)", life, opp.Life)
	}
}

func TestNghcOverchargedAmalgamDecliningCountersNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nghcBoltOnStack(t, g, me)
	life := opp.Life
	amalgam := nghcCastAmalgam(t, g, me)
	effectstest.NghcPassOnce(t, g)
	effectstest.NghcPassOnce(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID) // sacrifice nothing
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a counter trigger asked for a target without an exploited creature")
	}
	if opp.Life != life-3 {
		t.Errorf("the Bolt should have resolved: life %d -> %d", life, opp.Life)
	}
	if !g.Battlefield.Contains(amalgam) {
		t.Error("the Amalgam should still be in play")
	}
}

// The counter can answer an ability as well as a spell (Disallow's
// three-way clause): here the Amalgam's own exploit-less twin's trigger.
func TestNghcOverchargedAmalgamCountersATriggeredAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)

	first := nghcCastAmalgam(t, g, me)
	effectstest.NghcPassOnce(t, g) // the first Amalgam resolves; its exploit ability is on the stack
	firstTrigger := effectstest.TriggerOnStack(g, first)
	if firstTrigger == nil {
		t.Fatal("the first exploit ability is not on the stack")
	}
	second := nghcCastAmalgam(t, g, me) // flash, in response
	effectstest.NghcPassOnce(t, g)      // the second resolves
	effectstest.NghcPassOnce(t, g)      // the second's exploit ability resolves and asks
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	var found bool
	for _, id := range p.PickTargetCards {
		if id == firstTrigger.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("the first exploit ability is not a legal target: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, firstTrigger.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// The first exploit ability was countered, so it never asked.
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("the countered exploit ability still asked what to sacrifice")
	}
	_ = second
}

// An exploiter that left before the ability resolved can still make its
// controller sacrifice another creature, but "exploits a creature" does
// not trigger (it is no longer there to trigger).
func TestNghcExploitWithoutTheExploiterOnTheBattlefieldMakesNoPayoff(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	nghcBoltOnStack(t, g, me)
	amalgam := nghcCastAmalgam(t, g, me)
	effectstest.NghcPassOnce(t, g) // the Amalgam resolves; exploit is on the stack
	effectstest.ExileCreature(t, g, amalgam)
	effectstest.NghcPassOnce(t, g) // the exploit ability resolves
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice should still happen")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the Amalgam left the battlefield, yet its counter trigger asked for a target")
	}
}
