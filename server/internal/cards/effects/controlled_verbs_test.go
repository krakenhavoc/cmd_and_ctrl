package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4025, ADR 0166 amendment of 2026-10-19: what the controller of a
// player may send as that player, stage by stage (CR 723.5b). Each test
// reaches a stage through the real cards, then sends a table of verbs
// through actions.Dispatch as the controller naming the controlled
// player. The Emrakul turn stage is tabled in
// internal/actions/controlled_turn_test.go, and the verb sets per stage
// in internal/actions/controlled_stage_test.go.

// controlledSandboxVerbs are manual table verbs. None is a choice the
// rules or an object call for, so no stage lets the controller send one
// as the controlled player.
var controlledSandboxVerbs = []actions.Type{
	actions.TypeDrawCard, actions.TypePlayCard, actions.TypeMoveCard,
	actions.TypeTap, actions.TypeUntap, actions.TypeUntapAll,
	actions.TypePassTurn, actions.TypeShuffleLibrary, actions.TypeChangeLife,
	actions.TypeAddCounter, actions.TypeAddPlayerCounter, actions.TypeSetPoison,
	actions.TypeSetMonarch, actions.TypeAdvanceStep, actions.TypeRevealCard,
	actions.TypeCounterSpell, actions.TypeSacrificePermanent, actions.TypeMarkDamage,
}

// controlledRulesVerbs are rules verbs that a whole-turn controller may
// send (ADR 0167) and that a controller of one resolution or one search
// may not: nobody holds priority during either.
var controlledRulesVerbs = []actions.Type{
	actions.TypePassPriority, actions.TypeEndTurn, actions.TypeActivateAbility,
	actions.TypeActivateLoyalty, actions.TypeSpecialAction, actions.TypeOutsidePlay,
	actions.TypeDeclareAttacker, actions.TypeDeclareAttackers, actions.TypeDiscardSelection,
}

// assertControlledVerbs sends each verb as `controller` naming
// `controlled`. A refused verb must come back ErrPlayerControlled. An
// allowed one carries no parameters, so it may fail for any other reason,
// but never because the player is controlled.
func assertControlledVerbs(t *testing.T, g *game.Game, controller, controlled uuid.UUID, refused, allowed []actions.Type) {
	t.Helper()
	for _, ty := range refused {
		err := actions.Dispatch(g, actions.Action{Type: ty, Player: controlled, Caller: controller})
		if !errors.Is(err, game.ErrPlayerControlled) {
			t.Errorf("%s as the controlled player: err = %v, want ErrPlayerControlled", ty, err)
		}
	}
	for _, ty := range allowed {
		err := actions.Dispatch(g, actions.Action{Type: ty, Player: controlled, Caller: controller})
		if errors.Is(err, game.ErrPlayerControlled) {
			t.Errorf("%s as the controlled player was refused as controlled: %v", ty, err)
		}
	}
}

// firstPermanentOf is the first battlefield card p controls.
func firstPermanentOf(g *game.Game, p uuid.UUID) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func joinVerbs(lists ...[]actions.Type) []actions.Type {
	var out []actions.Type
	for _, l := range lists {
		out = append(out, l...)
	}
	return out
}

// assertControllerCannotTouchTheirCards sends, as the CONTROLLER
// themselves (not as the controlled player), every verb that names a
// card, against the controlled player's battlefield card and hand card.
// The card verbs gate on the card's controller (requireCardController),
// so each must come back ErrCardCallerMismatch while a control is live,
// and the cards must not move: no path lets the controller edit the
// controlled player's cards by hand (#4025 follow-up).
func assertControllerCannotTouchTheirCards(t *testing.T, g *game.Game, controller, controlled *game.Player, onBattlefield, inHand uuid.UUID) {
	t.Helper()
	zone := func(kind string, owner uuid.UUID) map[string]string {
		return map[string]string{"kind": kind, "owner": owner.String()}
	}
	bf, hand := onBattlefield.String(), inHand.String()
	table := []struct {
		name   string
		ty     actions.Type
		params map[string]any
	}{
		{"move_card, battlefield to graveyard", actions.TypeMoveCard, map[string]any{"instance_id": bf, "src": zone("battlefield", controlled.ID), "dst": zone("graveyard", controlled.ID)}},
		{"move_card, hand to battlefield", actions.TypeMoveCard, map[string]any{"instance_id": hand, "src": zone("hand", controlled.ID), "dst": zone("battlefield", controlled.ID)}},
		{"move_card, battlefield to the controller's hand", actions.TypeMoveCard, map[string]any{"instance_id": bf, "src": zone("battlefield", controlled.ID), "dst": zone("hand", controller.ID)}},
		{"tap", actions.TypeTap, map[string]any{"instance_id": bf}},
		{"untap", actions.TypeUntap, map[string]any{"instance_id": bf}},
		{"add_counter", actions.TypeAddCounter, map[string]any{"instance_id": bf, "name": "+1/+1", "delta": 1}},
		{"set_battlefield_position", actions.TypeSetBattlefieldPosition, map[string]any{"instance_id": bf, "x": 0.5, "y": 0.5}},
		{"sacrifice_permanent", actions.TypeSacrificePermanent, map[string]any{"instance_id": bf}},
		{"mark_damage", actions.TypeMarkDamage, map[string]any{"instance_id": bf, "delta": 1}},
		{"remove_from_combat", actions.TypeRemoveFromCombat, map[string]any{"instance_id": bf}},
		{"reveal_card, from the hand", actions.TypeRevealCard, map[string]any{"card_id": hand}},
	}
	handSize, bfSize := controlled.Hand.Size(), g.Battlefield.Size()
	for _, row := range table {
		raw, _ := json.Marshal(row.params)
		err := actions.Dispatch(g, actions.Action{Type: row.ty, Player: controller.ID, Caller: controller.ID, Params: raw})
		if !errors.Is(err, game.ErrCardCallerMismatch) {
			t.Errorf("%s as the controller: err = %v, want ErrCardCallerMismatch", row.name, err)
		}
	}
	// A player verb that names the controlled player is the stage
	// table's business, not the card gate's: refused too.
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeDrawCard, Player: controlled.ID, Caller: controller.ID}); !errors.Is(err, game.ErrPlayerControlled) {
		t.Errorf("draw_card naming the controlled player: err = %v, want ErrPlayerControlled", err)
	}
	if controlled.Hand.Size() != handSize || g.Battlefield.Size() != bfSize {
		t.Errorf("cards moved: hand %d -> %d, battlefield %d -> %d",
			handSize, controlled.Hand.Size(), bfSize, g.Battlefield.Size())
	}
}

// Word of Command, the forced play (ControlPlaying): the play, the
// lands' mana and the prompts it raises; nothing else.
func TestControlledVerbsWhileWordOfCommandForcesThePlay(t *testing.T) {
	g, caster, victim, ids := wocSetup(t, []string{"Mountain"}, wocBolt())
	if err := g.ResolvePendingChoice(openVariant(t, g).ID, caster.ID, ids); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if openControlledPlay(g) == nil {
		t.Fatal("no controlled_play prompt")
	}
	if got := g.ControlStageOfPlayer(victim.ID); got != game.ControlPlaying {
		t.Fatalf("control stage = %q, want %q", got, game.ControlPlaying)
	}
	assertControlledVerbs(t, g, caster.ID, victim.ID,
		joinVerbs(controlledSandboxVerbs, controlledRulesVerbs),
		[]actions.Type{actions.TypeCastSpell, actions.TypeActivateManaAbility, actions.TypeResolveChoice})
	assertControllerCannotTouchTheirCards(t, g, caster, victim, firstPermanentOf(g, victim.ID), ids[0])
}

// Word of Command, the forced spell resolving (ControlSpellResolving):
// only the prompts the resolution raises. This is the stage #4025 found
// open to every verb. Demonic Tutor's search is the prompt.
func TestControlledVerbsWhileTheForcedSpellResolves(t *testing.T) {
	tutor := game.Card{Name: "Demonic Tutor", TypeLine: "Sorcery", ManaCost: "{1}{B}", OracleID: demonicTutorOracle}
	g, caster, victim, ids := wocSetup(t, []string{"Swamp", "Swamp"}, tutor, wocBolt())
	spare := ids[1]
	ids = ids[:1]
	if err := g.ResolvePendingChoice(openVariant(t, g).ID, caster.ID, ids); err != nil {
		t.Fatalf("pick: %v", err)
	}
	play := openControlledPlay(g)
	if play == nil {
		t.Fatal("no controlled_play prompt")
	}
	raw, _ := json.Marshal(map[string]any{
		"instance_id":     ids[0].String(),
		"strict":          true,
		"auto_tap":        true,
		"controlled_play": play.ID.String(),
	})
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeCastSpell, Player: victim.ID, Caller: caster.ID, Params: raw}); err != nil {
		t.Fatalf("forced cast: %v", err)
	}
	passUntilChoice(t, g, game.PendingChoiceSearchLibrary)
	if got := g.ControlStageOfPlayer(victim.ID); got != game.ControlSpellResolving {
		t.Fatalf("control stage = %q, want %q", got, game.ControlSpellResolving)
	}
	hand, life := victim.Hand.Size(), victim.Life
	assertControlledVerbs(t, g, caster.ID, victim.ID,
		joinVerbs(controlledSandboxVerbs, controlledRulesVerbs,
			[]actions.Type{actions.TypeCastSpell, actions.TypeActivateManaAbility}),
		[]actions.Type{actions.TypeResolveChoice})
	assertControllerCannotTouchTheirCards(t, g, caster, victim, firstPermanentOf(g, victim.ID), spare)
	if victim.Hand.Size() != hand || victim.Life != life {
		t.Errorf("a refused verb still changed the victim: hand %d -> %d, life %d -> %d",
			hand, victim.Hand.Size(), life, victim.Life)
	}
}

// Opposition Agent, the search (ControlSearching): only the search
// prompt (#3998).
func TestControlledVerbsWhileAnOpponentSearches(t *testing.T) {
	g, searcher, ctl, _, _ := oaTable(t)
	pushLibraryCardForTest(searcher, game.Card{
		Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: lightningBoltOracle,
	})
	castCatalogSpell(t, g, "Demonic Tutor", "Sorcery", demonicTutorOracle, nil)
	passPriorityAroundTable(t, g)
	oaOpenSearch(t, g, searcher, ctl)
	if got := g.ControlStageOfPlayer(searcher.ID); got != game.ControlSearching {
		t.Fatalf("control stage = %q, want %q", got, game.ControlSearching)
	}
	assertControlledVerbs(t, g, ctl.ID, searcher.ID,
		joinVerbs(controlledSandboxVerbs, controlledRulesVerbs,
			[]actions.Type{actions.TypeCastSpell, actions.TypeActivateManaAbility}),
		[]actions.Type{actions.TypeResolveChoice})
	perm := pushCatalogPermanent(g, searcher.ID, "Grizzly Bears", "Creature — Bear", "", false)
	var handCard uuid.UUID
	g.WithWriteLock(func() {
		handCard = revealHand(searcher, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})[0]
	})
	assertControllerCannotTouchTheirCards(t, g, ctl, searcher, perm, handCard)
}
