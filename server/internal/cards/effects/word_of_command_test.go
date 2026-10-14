package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// ADR 0164, #2164: Word of Command and the control of a player it
// gives (CR 723.2).

const wordOfCommandOracle = "e8ad3a77-b293-4d69-b080-27ca9f95d443"

// wocSetup casts Word of Command from seat 0 at seat 1 and answers
// nothing. The victim's hand is `hand`; it controls one untapped basic
// land of each type named in `lands`.
func wocSetup(t *testing.T, lands []string, hand ...game.Card) (g *game.Game, caster, victim *game.Player, ids []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	caster, victim = g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { ids = revealHand(victim, hand...) })
	for _, land := range lands {
		pushCatalogPermanent(g, victim.ID, land, "Basic Land — "+land, "", false)
	}
	castAt(t, g, "Word of Command", "Instant", wordOfCommandOracle, pbPlayer(victim.ID))
	return g, caster, victim, ids
}

func wocBolt() game.Card {
	c := rhBolt()
	c.OracleID = lightningBoltOracle
	return c
}

// openControlledPlay returns the one controlled_play prompt, or nil.
func openControlledPlay(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceControlledPlay {
			return c
		}
	}
	return nil
}

// The look is the caster's alone (CR 701.20e), any card may be chosen,
// and the chosen card stays in the hand.
func TestWordOfCommandLooksAtTheHandAndKeepsTheCard(t *testing.T) {
	g, caster, victim, ids := wocSetup(t, []string{"Mountain"}, rhForest(), wocBolt())
	pick := openVariant(t, g)
	if !sameIDs(pick.DiscardOptions, ids) || pick.PickDestination != game.PickKeep || pick.Chooser != caster.ID {
		t.Fatalf("pick = options %v destination %q chooser %v; want the whole hand, keep, the caster",
			pick.DiscardOptions, pick.PickDestination, pick.Chooser)
	}
	third := g.Seats[2].ID
	for _, c := range victim.Hand.Cards {
		if !c.IsKnownTo(caster.ID) {
			t.Errorf("%s: the caster did not look at it", c.Name)
		}
		if c.IsKnownTo(third) {
			t.Errorf("%s: a third player saw it; a look is not a reveal", c.Name)
		}
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids[1:]); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if !victim.Hand.Contains(ids[1]) || rvDiscards(g, ids[1]) != 0 {
		t.Fatal("the chosen card left the hand")
	}
}

// The forced play: the caster controls the victim, casts the Bolt as
// them with their Mountain, and nobody gets priority in between.
func TestWordOfCommandForcesTheCastWithTheirLands(t *testing.T) {
	g, caster, victim, ids := wocSetup(t, []string{"Mountain"}, wocBolt())
	bolt := ids[0]
	if err := g.ResolvePendingChoice(openVariant(t, g).ID, caster.ID, ids); err != nil {
		t.Fatalf("pick: %v", err)
	}
	play := openControlledPlay(g)
	if play == nil {
		t.Fatal("no controlled_play prompt")
	}
	if play.Chooser != caster.ID || play.FromPlayer != victim.ID {
		t.Fatalf("prompt chooser %v from %v; want the caster, the victim", play.Chooser, play.FromPlayer)
	}
	if got := g.ControllerOfPlayer(victim.ID); got != caster.ID {
		t.Fatalf("victim is controlled by %v, want the caster", got)
	}
	if !g.ResolutionPaused() {
		t.Fatal("Word of Command's resolution is not waiting on the forced play")
	}
	// "It can't be played" is refused while it can be.
	if err := g.ResolvePendingChoice(play.ID, caster.ID, nil); !errors.Is(err, game.ErrControlledPlayPossible) {
		t.Fatalf("declining: err = %v, want ErrControlledPlayPossible", err)
	}
	life := victim.Life
	params := game.CastSpellParams{
		Strict:  true,
		AutoTap: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
	}
	if err := g.ControlledPlayCast(caster.ID, play.ID, bolt, params); err != nil {
		t.Fatalf("forced cast: %v", err)
	}
	if !g.Stack.Contains(bolt) {
		t.Fatal("the Bolt is not on the stack")
	}
	if c, _ := g.LookupCardForEffect(bolt); c.Controller != victim.ID {
		t.Errorf("the Bolt is controlled by %v; CR 723.3 keeps it the victim's", c.Controller)
	}
	if openControlledPlay(g) != nil {
		t.Fatal("the prompt is still open")
	}
	if got := g.ControllerOfPlayer(victim.ID); got != victim.ID {
		t.Fatalf("victim is still controlled by %v while the Bolt waits", got)
	}
	passPriorityAroundTable(t, g)
	if victim.Life != life-3 {
		t.Errorf("victim life = %d, want %d", victim.Life, life-3)
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("a control record outlived the spell: %+v", g.PlayerControls)
	}
}

// Only their lands pay. A Sol Ring is no help, so a Bolt with nothing
// but a Sol Ring can't be played and Word of Command does nothing.
func TestWordOfCommandUsesOnlyLandMana(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	g.WithWriteLock(func() { ids = revealHand(victim, wocBolt()) })
	pushCatalogPermanent(g, victim.ID, "Sol Ring", "Artifact", solRingOracle, false)
	castAt(t, g, "Word of Command", "Instant", wordOfCommandOracle, pbPlayer(victim.ID))
	if err := g.ResolvePendingChoice(openVariant(t, g).ID, caster.ID, ids); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if openControlledPlay(g) != nil {
		t.Fatal("a forced play was asked for a card only a Sol Ring could pay for")
	}
	if len(g.PlayerControls) != 0 || g.ControllerOfPlayer(victim.ID) != victim.ID {
		t.Fatal("the victim is still controlled")
	}
	if !victim.Hand.Contains(ids[0]) {
		t.Fatal("the Bolt left the hand")
	}
}

// While the forced spell resolves, the caster makes its controller's
// choices: Demonic Tutor's search is answered by the caster, through
// the ordinary resolve_choice, for the victim.
func TestWordOfCommandControlsThePlayerWhileTheSpellResolves(t *testing.T) {
	tutor := game.Card{Name: "Demonic Tutor", TypeLine: "Sorcery", ManaCost: "{1}{B}", OracleID: demonicTutorOracle}
	g, caster, victim, ids := wocSetup(t, []string{"Swamp", "Swamp"}, tutor)
	if err := g.ResolvePendingChoice(openVariant(t, g).ID, caster.ID, ids); err != nil {
		t.Fatalf("pick: %v", err)
	}
	play := openControlledPlay(g)
	if play == nil {
		t.Fatal("no controlled_play prompt")
	}
	// The caster's cast_spell names the victim and the prompt, and runs
	// as the victim (actions.routeControlledPlayer).
	raw, _ := json.Marshal(map[string]any{
		"instance_id":     ids[0].String(),
		"strict":          true,
		"auto_tap":        true,
		"controlled_play": play.ID.String(),
	})
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeCastSpell, Player: victim.ID, Caller: caster.ID, Params: raw}); err != nil {
		t.Fatalf("forced cast through the dispatcher: %v", err)
	}
	// The Tutor waits on the stack, the victim answers for themselves
	// again, and the table passes as usual.
	passUntilChoice(t, g, game.PendingChoiceSearchLibrary)
	var search *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSearchLibrary {
			search = c
		}
	}
	if search.Chooser != victim.ID {
		t.Fatalf("search chooser %v, want the victim", search.Chooser)
	}
	if got := g.ControllerOfPlayer(victim.ID); got != caster.ID {
		t.Fatalf("while the Tutor resolves the victim is controlled by %v, want the caster", got)
	}
	// The victim's own answer is refused (CR 723.5)…
	pick := victim.Library.Cards[0].InstanceID
	ans, _ := json.Marshal(map[string]any{"choice_id": search.ID.String(), "card_ids": []string{pick.String()}})
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeResolveChoice, Player: victim.ID, Caller: victim.ID, Params: ans}); !errors.Is(err, game.ErrPlayerControlled) {
		t.Fatalf("victim answering: err = %v, want ErrPlayerControlled", err)
	}
	// …and the caster's, sent as themselves, is taken for the victim.
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeResolveChoice, Player: caster.ID, Caller: caster.ID, Params: ans}); err != nil {
		t.Fatalf("caster answering for the victim: %v", err)
	}
	if !victim.Hand.Contains(pick) {
		t.Fatal("the searched card is not in the victim's hand")
	}
	passPriorityAroundTable(t, g)
	if len(g.PlayerControls) != 0 || g.ControllerOfPlayer(victim.ID) != victim.ID {
		t.Fatalf("control outlived the Tutor: %+v", g.PlayerControls)
	}
}

// A controlled player decides nothing (CR 723.5) but may concede (CR
// 723.6); the controller may not concede for them.
func TestWordOfCommandControlledPlayerMayOnlyConcede(t *testing.T) {
	g, caster, victim, ids := wocSetup(t, []string{"Mountain"}, wocBolt())
	if err := g.ResolvePendingChoice(openVariant(t, g).ID, caster.ID, ids); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if openControlledPlay(g) == nil {
		t.Fatal("no controlled_play prompt")
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypePassPriority, Player: victim.ID, Caller: victim.ID}); !errors.Is(err, game.ErrPlayerControlled) {
		t.Fatalf("victim passing: err = %v, want ErrPlayerControlled", err)
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeConcede, Player: victim.ID, Caller: caster.ID}); err == nil {
		t.Fatal("the caster conceded for the victim")
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeConcede, Player: victim.ID, Caller: victim.ID}); err != nil {
		t.Fatalf("victim conceding: %v", err)
	}
	g.SettleResolution()
	if openControlledPlay(g) != nil || len(g.PlayerControls) != 0 {
		t.Fatal("the forced play outlived the controlled player")
	}
}

// The bot side: the caster's seat is offered the forced plays, each
// naming the prompt, and the victim's seat is offered nothing.
func TestWordOfCommandMovesGoToTheController(t *testing.T) {
	g, caster, victim, ids := wocSetup(t, []string{"Mountain"}, wocBolt())
	if err := g.ResolvePendingChoice(openVariant(t, g).ID, caster.ID, ids); err != nil {
		t.Fatalf("pick: %v", err)
	}
	play := openControlledPlay(g)
	if play == nil {
		t.Fatal("no controlled_play prompt")
	}
	if moves := legal.EnumerateFor(g, victim.ID); len(moves) != 0 {
		t.Fatalf("the controlled seat was offered %d moves", len(moves))
	}
	moves := legal.EnumerateFor(g, caster.ID)
	if len(moves) == 0 {
		t.Fatal("the controller was offered nothing")
	}
	for _, m := range moves {
		if m.ControlledPlay != play.ID.String() || m.Type != legal.TypeCastSpell || m.Player != victim.ID {
			t.Fatalf("move %+v; want only the victim's casts of the Bolt, naming the prompt", m)
		}
	}
	// Each one is accepted as sent.
	m := moves[0]
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: caster.ID, Params: m.Params}); err != nil {
		t.Fatalf("dispatching %s: %v", m.Label, err)
	}
	if !g.Stack.Contains(ids[0]) {
		t.Fatal("the Bolt is not on the stack")
	}
	// The move names no payment, and the forced play is paid anyway:
	// their Mountain is tapped.
	for _, c := range g.Battlefield.Cards {
		if c.Controller == victim.ID && c.Name == "Mountain" && !c.Tapped {
			t.Error("the Mountain is untapped: the forced cast was not paid for")
		}
	}
}

// "You do get to choose which lands get tapped": the controller locks
// the lands they don't want tapped.
func TestWordOfCommandControllerChoosesTheLands(t *testing.T) {
	g, caster, victim, ids := wocSetup(t, []string{"Mountain", "Mountain"}, wocBolt())
	if err := g.ResolvePendingChoice(openVariant(t, g).ID, caster.ID, ids); err != nil {
		t.Fatalf("pick: %v", err)
	}
	play := openControlledPlay(g)
	if play == nil {
		t.Fatal("no controlled_play prompt")
	}
	var mountains []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == victim.ID {
			mountains = append(mountains, c.InstanceID)
		}
	}
	params := game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
		LockedSources: mountains[:1],
	}
	if err := g.ControlledPlayCast(caster.ID, play.ID, ids[0], params); err != nil {
		t.Fatalf("forced cast: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		switch c.InstanceID {
		case mountains[0]:
			if c.Tapped {
				t.Error("the locked Mountain was tapped")
			}
		case mountains[1]:
			if !c.Tapped {
				t.Error("the chosen Mountain was not tapped")
			}
		}
	}
}

// A land is played only on its owner's turn (CR 305.3): on the
// caster's turn there is nothing the victim can be made to play, so
// Word of Command ends with the land still in hand.
func TestWordOfCommandCantMakeThemPlayALandOnYourTurn(t *testing.T) {
	g, caster, victim, ids := wocSetup(t, []string{"Mountain"}, rhForest())
	if err := g.ResolvePendingChoice(openVariant(t, g).ID, caster.ID, ids); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if openControlledPlay(g) != nil {
		t.Fatal("a forced land play was asked on another player's turn")
	}
	if !victim.Hand.Contains(ids[0]) || len(g.PlayerControls) != 0 {
		t.Fatal("the land moved, or the victim is still controlled")
	}
}

// passUntilChoice passes priority until a prompt of that kind is open.
func passUntilChoice(t *testing.T, g *game.Game, kind game.PendingChoiceKind) {
	t.Helper()
	for range 16 {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == kind {
				return
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("no %s prompt opened", kind)
}
