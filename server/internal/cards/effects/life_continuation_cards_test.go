package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_continuation_cards_test.go is #793 seen from the catalog: the
// drain family ("each opponent loses X life. You gain life equal to
// the life lost this way") under boards where the life lost this way is
// NOT X times the number of opponents.
//
// The engine-side suite is server/internal/game/life_continuation_test.go.
// What these tests add is the card sentence: Exsanguinate, Gray
// Merchant of Asphodel, Kokusho and Debt to the Deathless all share one
// body, and it now takes its gain from the continuation rather than
// from arithmetic over the opponent count.
//
// ALHAMMARRET'S ARCHIVE IS NOT IN THE CATALOG. The mixed CR 616 window
// #793 names — Rhox Faithmender plus a second, DIFFERENT life
// replacement, the one board that still prompts after #792 — cannot be
// built from catalog cards today: the Faithmender is the only card in
// the catalog with a RepEventLife replacement, and two Faithmenders
// collapse without asking. So the second effect is registered for the
// test. When the Archive ships, the stand-in below is what it replaces.

// oneMoreLifeLostFor is a life-loss replacement gated to one player:
// "if that player would lose life, they lose one more". Registered per
// test as the second, DISTINCT effect it takes to make a CR 616
// ordering prompt happen at all.
func oneMoreLifeLostFor(victim uuid.UUID) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
			return ev.Kind == game.RepEventLife && ev.LifeDelta < 0 && ev.LifePlayer == victim
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.LifeDelta--
			return nil
		},
		Label: "they lose one more life",
	}
}

// twiceTheLifeLostFor is the other half of the pair: "if that player
// would lose life, they lose twice that much". A different declared
// effect from the one above, which is what stops #792 collapsing the
// two and keeps the prompt.
func twiceTheLifeLostFor(victim uuid.UUID) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
			return ev.Kind == game.RepEventLife && ev.LifeDelta < 0 && ev.LifePlayer == victim
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.LifeDelta *= 2
			return nil
		},
		Label: "they lose twice that much life",
	}
}

// answerTheOrderingPrompt answers the single queued CR 616 prompt in
// the order it was offered.
func answerTheOrderingPrompt(t *testing.T, g *game.Game) {
	t.Helper()
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
			prompt = c
			break
		}
	}
	if prompt == nil {
		t.Fatalf("no CR 616 ordering prompt queued (have %d choices)", len(g.PendingChoices))
	}
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

// TestExsanguinateGainsWhatWasActuallyLostAcrossACR616Prompt is the
// issue's headline repro. One opponent's life loss has two different
// replacements on it, so it pauses for a CR 616 ordering prompt; the
// other two opponents lose the printed amount. The caster's gain is the
// TRUE total — and it is then doubled by their own Rhox Faithmender,
// which is the mixed window the issue names.
//
// Before #793 the card read each opponent's life total back on the line
// after changing it. The paused opponent's total had not moved yet, so
// they counted as having lost nothing and the gain was short by
// everything that opponent lost.
func TestExsanguinateGainsWhatWasActuallyLostAcrossACR616Prompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	victim, o2, o3 := g.Seats[1], g.Seats[2], g.Seats[3]
	for _, p := range []*game.Player{victim, o2, o3} {
		p.Life = 40
	}
	me.Life = 10

	b12Push(g, me.ID, "Rhox Faithmender", "Creature — Rhino Monk", b14RhoxFaithmenderOracle, 1, 5)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(oneMoreLifeLostFor(victim.ID))
		g.RegisterReplacementForTest(twiceTheLifeLostFor(victim.ID))
	})

	castXSpell(t, g, "Exsanguinate", "Sorcery", exsanguinateOracle, "{X}{B}{B}", 3, nil)
	passPriorityAroundTable(t, g)

	if victim.Life != 40 {
		t.Fatalf("the paused opponent's life moved to %d before the prompt was answered", victim.Life)
	}
	if me.Life != 10 {
		t.Fatalf("the caster gained %d before the drain finished", me.Life-10)
	}

	answerTheOrderingPrompt(t, g)

	// The prompt is offered in gather order: lose one more, then lose
	// twice that much. 3 → 4 → 8.
	if victim.Life != 32 {
		t.Errorf("ordered opponent life = %d, want 32 — (3+1)*2 lost", victim.Life)
	}
	if o2.Life != 37 || o3.Life != 37 {
		t.Errorf("other opponents' lives = %d/%d, want 37/37", o2.Life, o3.Life)
	}
	// 8 + 3 + 3 = 14 lost this way, doubled by the Faithmender.
	if me.Life != 10+28 {
		t.Errorf("caster life = %d, want %d — 14 lost this way, doubled to 28", me.Life, 10+28)
	}
}

// TestGrayMerchantGainsOnlyWhatWasActuallyLost pins the arithmetic the
// card comment used to admit was an approximation: the gain was
// computed as X × (live opponents), which is right only while nothing
// is replacing anybody's life loss. An opponent whose life total cannot
// change loses nothing, so the drain gains less.
func TestGrayMerchantGainsOnlyWhatWasActuallyLost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	immune := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	others := []*game.Player{
		g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)],
		g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)],
	}
	me.Life = 10

	// Devotion to black comes from permanents the caster controls; the
	// Merchant's own {2}{B}{B} is two, and nothing else on this board
	// is black, so X = 2.
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventLife && ev.LifePlayer == immune.ID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Label: "that player's life total can't change",
		})
	})

	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Gray Merchant of Asphodel",
		TypeLine: "Creature — Zombie", OracleID: b02bGrayMerchantOracle,
		ManaCost: "{3}{B}{B}", Power: 2, Toughness: 4,
		Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)

	if immune.Life != game.StartingLife {
		t.Errorf("immune opponent life = %d, want %d — their loss was replaced away",
			immune.Life, game.StartingLife)
	}
	for i, p := range others {
		if p.Life != game.StartingLife-2 {
			t.Errorf("opponent %d life = %d, want %d", i+1, p.Life, game.StartingLife-2)
		}
	}
	if me.Life != 14 {
		t.Errorf("caster life = %d, want 14 — 4 lost this way, not 2 × 3 opponents", me.Life)
	}
}
