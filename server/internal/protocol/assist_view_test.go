package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// assist_view_test.go — ADR 0150 §5 (#1857): a spell with assist offers
// its caster `assist: {max, players}`, and the helper's pay_amount
// carries `assist` on every seat's view.

func pushAssistCard(g *game.Game, owner *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		owner.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Fan Thing", TypeLine: "Creature — Human Rogue",
			ManaCost: "{3}{B}", Owner: owner.ID, Controller: owner.ID,
			Keywords: []string{game.KeywordAssist},
			KnownBy:  map[uuid.UUID]bool{owner.ID: true},
		})
	})
	return id
}

func TestAssistOfferIsOnTheCastersOwnCard(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushAssistCard(g, me)

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if c.Assist == nil || c.Assist.Max != 3 || len(c.Assist.Players) != 1 || c.Assist.Players[0] != opp.ID.String() {
		t.Fatalf("assist = %+v, want max 3 and the one other player", c.Assist)
	}
	raw, err := json.Marshal(c.CastSurfaceView)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"assist":{"max":3,"players":[`) {
		t.Errorf("wire form: %s", raw)
	}
	// A card without assist offers nothing.
	plain := pushGiftCard(g, me)
	if c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), plain.String()); c.Assist != nil {
		t.Errorf("a card without assist offers %+v", c.Assist)
	}
}

func TestAssistPromptIsOnEverySeatsView(t *testing.T) {
	g := buildActiveGame(t)
	me, helper := g.Seats[0], g.Seats[1]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	g.WithWriteLock(func() {
		for range 2 {
			g.Battlefield.PushTop(game.Card{
				InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
				Owner: helper.ID, Controller: helper.ID,
			})
		}
	})
	id := pushAssistCard(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true, AssistPlayer: helper.ID}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for _, viewer := range g.Seats {
		v := ViewOfGameFor(g, viewer.ID.String())
		if len(v.PendingChoices) != 1 {
			t.Fatalf("viewer %s sees %d prompts", viewer.Name, len(v.PendingChoices))
		}
		p := v.PendingChoices[0]
		if p.Kind != string(game.PendingChoicePayAmount) || p.Chooser != helper.ID.String() {
			t.Errorf("viewer %s: %s to %s", viewer.Name, p.Kind, p.Chooser)
		}
		a := p.Assist
		if a == nil || a.Caster != me.ID.String() || a.Spell != id.String() || a.SpellName != "Fan Thing" || a.ManaCost != "{3}{B}" || a.Generic != 3 {
			t.Errorf("viewer %s: assist = %+v", viewer.Name, a)
		}
		if pa := p.PayAmount; pa == nil || pa.Resource != game.PayResourceMana || pa.Unit != game.PayAmountAssist || pa.Max != 2 {
			t.Errorf("viewer %s: pay_amount = %+v, want mana / assist / max 2", viewer.Name, pa)
		}
	}
	raw, err := json.Marshal(ViewOfGameFor(g, me.ID.String()).PendingChoices[0])
	if err != nil {
		t.Fatal(err)
	}
	var back PendingChoiceView
	if err := json.Unmarshal(raw, &back); err != nil || back.Assist == nil || back.Assist.Generic != 3 {
		t.Errorf("round trip lost the assist data: %+v (%v) from %s", back.Assist, err, raw)
	}
}
