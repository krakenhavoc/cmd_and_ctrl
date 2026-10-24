package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// guile_resolution_cast_4286_test.go — #4286, CR 608.2g. Guile's "you may
// play that card without paying its mana cost" is made while the
// countering spell or ability resolves: the replaced counter asks the
// resolution_cast prompt over the exiled card, inside that resolution,
// for every counter the engine has. One table, four checks per counter:
//
//   - cast: the card is cast through the prompt and the cast is logged
//     before the counterspell finishes (EventCast, then the held
//     resolution's EventResolutionFinished), and nothing was countered;
//   - decline: the empty answer leaves the card in exile with no
//     permission, and lets the counterspell finish;
//   - restore: a STRICT restore while the prompt is open keeps the offer,
//     and the cast is made on the restored table;
//   - (aiseat/guile_resolution_cast_bot_test.go) the bot answers it.

// guileCounter is a table brought to Guile's open offer: who casts, the
// card on offer and the cast's parameters, and the countering spell (nil
// for a ward's triggered ability) that must still be on the stack, not
// finished, while the prompt is open.
type guileCounter struct {
	g       *game.Game
	chooser uuid.UUID
	card    uuid.UUID
	params  game.CastSpellParams
	counter uuid.UUID
	owner   *game.Player // the countering spell's owner, whose graveyard it ends in
}

type guileCounterCase struct {
	name  string
	setup func(t *testing.T) guileCounter
}

func guileCounterCases() []guileCounterCase {
	exileFree := game.CastSpellParams{FromZone: "exile", Strict: true}
	shockAtThem := func(opp uuid.UUID) game.CastSpellParams {
		return game.CastSpellParams{FromZone: "exile", Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp}}}
	}
	return []guileCounterCase{
		{"a hard counter (Counterspell)", func(t *testing.T) guileCounter {
			g, _, opp, bear, cs := counterABear(t, 1)
			passPriorityAroundTable(t, g)
			return guileCounter{g, opp.ID, bear, exileFree, cs, opp}
		}},
		{"a declined counter-unless-pays (Spell Pierce)", func(t *testing.T) guileCounter {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			toMainForCost(t, g)
			pushGuile(g, opp.ID)
			victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
			pierce := castFromSeat(t, g, opp, "Spell Pierce", "Instant", s293SpellPierceOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
			passUntilTaxed(t, g, me.ID)
			answerPayUnless(t, g, me.ID, false)
			return guileCounter{g, opp.ID, victim, exileFree, pierce, opp}
		}},
		{"a ward (Rimeshield Frost Giant)", func(t *testing.T) guileCounter {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			giant := pushCatalogPermanent(g, me.ID, "Rimeshield Frost Giant",
				"Creature — Giant Warrior", rimeshieldOracle, false)
			pushGuile(g, me.ID)
			theirs := pushCatalogPermanent(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", false)
			toMainForCost(t, g)
			blade := castAtWardedCreature(t, g, opp, giant)
			passUntilTaxed(t, g, opp.ID)
			answerPayUnless(t, g, opp.ID, false)
			return guileCounter{g, me.ID, blade, game.CastSpellParams{FromZone: "exile", Strict: true,
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}, uuid.Nil, nil}
		}},
		{"a counter to the top of the library (Hinder)", func(t *testing.T) guileCounter {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pushGuile(g, me.ID)
			seedLibrary(opp, "Their Top")
			shock := castShockAt(t, g, opp, me.ID)
			hinder := castCounterAt(t, g, me, "Hinder", hinderOracle, shock)
			passPriorityAroundTable(t, g)
			ask := putInLibraryChoiceFor(g, me.ID)
			if ask == nil {
				t.Fatal("Hinder's top-or-bottom choice was not asked")
			}
			if err := g.ResolvePutInLibrary(ask.ID, me.ID, nil, []uuid.UUID{shock}); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			return guileCounter{g, me.ID, shock, shockAtThem(opp.ID), hinder, me}
		}},
		{"a counter to the bottom of the library (Hinder)", func(t *testing.T) guileCounter {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pushGuile(g, me.ID)
			seedLibrary(opp, "Their Top")
			shock := castShockAt(t, g, opp, me.ID)
			hinder := castCounterAt(t, g, me, "Hinder", hinderOracle, shock)
			passPriorityAroundTable(t, g)
			ask := putInLibraryChoiceFor(g, me.ID)
			if ask == nil {
				t.Fatal("Hinder's top-or-bottom choice was not asked")
			}
			if err := g.ResolvePutInLibrary(ask.ID, me.ID, []uuid.UUID{shock}, nil); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			return guileCounter{g, me.ID, shock, shockAtThem(opp.ID), hinder, me}
		}},
		{"a counter that exiles the spell (Devious Cover-Up)", func(t *testing.T) guileCounter {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pushGuile(g, me.ID)
			shock := castShockAt(t, g, opp, me.ID)
			cover := castCounterAt(t, g, me, "Devious Cover-Up", deviousCoverUpOracle, shock)
			passPriorityAroundTable(t, g)
			return guileCounter{g, me.ID, shock, shockAtThem(opp.ID), cover, me}
		}},
	}
}

// castEventsOf counts the casts of card in the log.
func castEventsOf(g *game.Game, card uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventCast && ev.CardID == card {
			n++
		}
	}
	return n
}

// guileCounterDone fails unless the countering spell has finished and gone
// to its owner's graveyard (a ward's trigger has no card to check).
func guileCounterDone(t *testing.T, o guileCounter) {
	t.Helper()
	if o.counter == uuid.Nil {
		return
	}
	if o.g.Stack.Contains(o.counter) {
		t.Error("the countering spell is still on the stack")
	}
	if !o.owner.Graveyard.Contains(o.counter) {
		t.Error("the countering spell did not finish into its owner's graveyard")
	}
}

func TestGuileCastsTheCounteredSpellDuringTheCounterspellsResolution(t *testing.T) {
	for _, tc := range guileCounterCases() {
		t.Run(tc.name+"/cast", func(t *testing.T) {
			o := tc.setup(t)
			mustResolutionCastOffer(t, o.g, o.chooser, game.MayCastKeywordFreeCast, o.card)
			if !o.g.Exile.Contains(o.card) || countEvents(o.g, game.EventCounterSpell) != 0 {
				t.Fatal("the offered spell must be in exile and not countered")
			}
			if o.counter != uuid.Nil && !o.g.Stack.Contains(o.counter) {
				t.Fatal("the countering spell finished before the cast was offered")
			}
			if err := castThroughResolutionCast(t, o.g, o.chooser, o.card, o.params); err != nil {
				t.Fatalf("the cast through the offer: %v", err)
			}
			if !o.g.Stack.Contains(o.card) {
				t.Fatal("the card is not on the stack")
			}
			if resolutionCastOfferFor(o.g, o.chooser) != nil {
				t.Error("the offer stayed open after the one cast it was for")
			}
			castInsideTheResolution(t, o.g, o.card)
			if countEvents(o.g, game.EventCounterSpell) != 0 {
				t.Error("an exiled-instead spell was reported as countered")
			}
			passPriorityAroundTable(t, o.g)
			guileCounterDone(t, o)
		})
		t.Run(tc.name+"/decline", func(t *testing.T) {
			o := tc.setup(t)
			mustResolutionCastOffer(t, o.g, o.chooser, game.MayCastKeywordFreeCast, o.card)
			castsBefore := castEventsOf(o.g, o.card)
			declineResolutionCast(t, o.g, o.chooser)
			if !o.g.Exile.Contains(o.card) || o.g.Stack.Contains(o.card) {
				t.Fatal("the declined card left exile")
			}
			if perm := guilePermission(o.g, o.card); perm != nil {
				t.Errorf("the decline left a permission: %+v", perm)
			}
			if o.g.ResolutionPaused() {
				t.Error("the decline did not let the resolution finish")
			}
			if castEventsOf(o.g, o.card) != castsBefore {
				t.Error("the declined card was cast")
			}
			if countEvents(o.g, game.EventCounterSpell) != 0 {
				t.Error("an exiled-instead spell was reported as countered")
			}
			passPriorityAroundTable(t, o.g)
			guileCounterDone(t, o)
			if !o.g.Exile.Contains(o.card) {
				t.Error("the declined card did not stay in exile")
			}
		})
		t.Run(tc.name+"/restore", func(t *testing.T) {
			o := tc.setup(t)
			r := restoreRoundTrip(t, o.g, true)
			mustResolutionCastOffer(t, r, o.chooser, game.MayCastKeywordFreeCast, o.card)
			if err := castThroughResolutionCast(t, r, o.chooser, o.card, o.params); err != nil {
				t.Fatalf("the cast on the restored table: %v", err)
			}
			if !r.Stack.Contains(o.card) {
				t.Fatal("the restored cast did not reach the stack")
			}
			castInsideTheResolution(t, r, o.card)
		})
	}
}

// Guile on the wrong side, and a spell that can't be countered, never
// reach the prompt.
func TestGuileAsksNoPromptForACounterItDoesNotReplace(t *testing.T) {
	g, _, _, _, _ := counterABear(t, 0)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast {
			t.Fatal("a Guile on the countered spell's side asked a cast")
		}
	}
}
