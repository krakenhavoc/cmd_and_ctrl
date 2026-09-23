package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// waterbend_test.go — #1310 / #1311's enumerator half. A waterbend
// cost may be paid partly by tapping artifacts and creatures
// (CR 701.67a), so a bot with no mana but a board of creatures CAN
// pay — and an enumerator that asked only "can the pool pay" would
// never offer Aang's transform or the Unagi's ward payment at all.
// Every move offered here is dispatched against a clone (#544).

func waterbendClause(extra string) *game.TapPermanentsCost {
	return &game.TapPermanentsCost{
		Key:   "waterbend",
		Label: "Waterbend " + extra,
		Extra: extra,
		Spec: &game.TargetSpec{
			Mode: "permanent", Label: "an untapped artifact or creature you control",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.IsArtifact() || c.IsCreature()
			},
		},
	}
}

func waterbendSource(g *game.Game, p *game.Player, cost game.AbilityCost) uuid.UUID {
	return battlefieldCard(g, p, game.Card{
		Name: "Waterbender", TypeLine: "Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "Waterbend: mark",
			Cost:   cost,
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
}

type waterbendMove struct {
	SourceCardID string   `json:"source_card_id"`
	WaterbendIDs []string `json:"waterbend_ids"`
	XValue       int      `json:"x_value"`
}

func waterbendMovesFor(t *testing.T, moves []legal.Move, src uuid.UUID) []waterbendMove {
	t.Helper()
	var out []waterbendMove
	for _, m := range moves {
		if m.Type != legal.TypeActivateAbility {
			continue
		}
		var p waterbendMove
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if p.SourceCardID == src.String() {
			out = append(out, p)
		}
	}
	return out
}

// No mana at all, three creatures and the artifact source itself:
// "Waterbend {3}" is offered tapping three, and the engine accepts it.
func TestWaterbendAbilityIsOfferedPaidByTapping(t *testing.T) {
	g := newTable(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	src := waterbendSource(g, me, game.AbilityCost{Mana: "{3}", Waterbend: waterbendClause("{3}")})
	for i := 0; i < 3; i++ {
		battlefieldCard(g, me, creature("Helper", "{1}", 1, 1))
	}
	moves := legal.EnumerateFor(g, me.ID)
	got := waterbendMovesFor(t, moves, src)
	if len(got) != 1 || len(got[0].WaterbendIDs) != 3 {
		t.Fatalf("waterbend moves = %+v, want one tapping all three", got)
	}
	dispatchAll(t, g, me.ID, moves)
}

// The source and one creature cannot pay {3} with no mana: not offered
// (#544). (The source is an artifact whose cost prints no {T}, so it
// is one of the two.)
func TestWaterbendAbilityShortOfHelpersIsNotOffered(t *testing.T) {
	g := newTable(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	src := waterbendSource(g, me, game.AbilityCost{Mana: "{3}", Waterbend: waterbendClause("{3}")})
	battlefieldCard(g, me, creature("Helper", "{1}", 1, 1))
	if got := waterbendMovesFor(t, legal.EnumerateFor(g, me.ID), src); len(got) != 0 {
		t.Fatalf("offered %+v, which the engine would refuse", got)
	}
}

// Waterbend {X} with "X can't be 0": the largest X the taps (and no
// mana) reach, never X=0.
func TestWaterbendXAbilityIsOfferedAtTheLargestReachableX(t *testing.T) {
	g := newTable(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	src := waterbendSource(g, me, game.AbilityCost{Mana: "{X}", MinX: 1, Waterbend: waterbendClause("{X}")})
	for i := 0; i < 2; i++ {
		battlefieldCard(g, me, creature("Helper", "{1}", 1, 1))
	}
	moves := legal.EnumerateFor(g, me.ID)
	got := waterbendMovesFor(t, moves, src)
	// Three waterbenders: the source is an artifact whose cost prints
	// no {T}, so it may help pay for itself.
	if len(got) != 1 || got[0].XValue != 3 || len(got[0].WaterbendIDs) != 3 {
		t.Fatalf("waterbend moves = %+v, want X=3 tapping three", got)
	}
	dispatchAll(t, g, me.ID, moves)
}

// The pay-unless half: a "Ward—Waterbend {2}" payer with no mana and
// two creatures is offered the payment, the move carries the taps, and
// answering it keeps the spell alive.
func TestWaterbendWardPaymentIsOfferedByTapping(t *testing.T) {
	g := newTable(t)
	payer := g.Seats[1]
	a := battlefieldCard(g, payer, creature("Helper A", "{1}", 1, 1))
	b := battlefieldCard(g, payer, creature("Helper B", "{1}", 1, 1))
	spell := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{InstanceID: spell, Name: "Doom Blade", TypeLine: "Instant",
			Owner: payer.ID, Controller: payer.ID})
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*game.StackItem)
		}
		g.StackMeta[spell] = &game.StackItem{ID: spell, Kind: game.StackItemSpell,
			Controller: payer.ID, Owner: payer.ID, SourceCardID: spell}
		if err := g.QueueCounterUnlessPaidForEffect(game.CounterUnlessPaidPrompt{
			StackItem: spell, Source: uuid.New(), Cost: "{2}",
			Question: "Ward — waterbend {2}", Waterbend: waterbendClause("{2}"),
		}); err != nil {
			t.Fatalf("queue: %v", err)
		}
	})

	moves := legal.EnumerateFor(g, payer.ID)
	var pay *legal.Move
	for i := range moves {
		var p struct {
			Apply  *bool    `json:"apply"`
			TapIDs []string `json:"tap_ids"`
		}
		_ = json.Unmarshal(moves[i].Params, &p)
		if p.Apply != nil && *p.Apply {
			if len(p.TapIDs) != 2 {
				t.Fatalf("pay move taps %v, want both creatures", p.TapIDs)
			}
			pay = &moves[i]
		}
	}
	if pay == nil {
		t.Fatalf("no pay move offered — the taps can pay the whole ward (moves %v)", labels(moves))
	}
	dispatchAll(t, g, payer.ID, moves)
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(pay.Type), Player: pay.Player, Caller: payer.ID, Params: pay.Params,
	}); err != nil {
		t.Fatalf("dispatch pay: %v", err)
	}
	if g.StackItemForEffect(spell) == nil {
		t.Error("a paid waterbend ward countered the spell")
	}
	for _, id := range []uuid.UUID{a, b} {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id && !c.Tapped {
				t.Errorf("%v was not tapped to pay", id)
			}
		}
	}
}
