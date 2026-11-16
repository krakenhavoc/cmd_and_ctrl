package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Maeve, Insidious Singer — Legendary Creature — Siren {2}{U}{U}, 3/4
// (Game Night: Free-for-All, #5485):
//
//	"{2}{U}: Goad target creature. Whenever that creature attacks one of
//	 your opponents this turn, you draw a card. (A goaded creature attacks
//	 each combat if able and attacks a player other than you if able.)"
//
// The goad is GoadTarget's (CR 701.15a): the target is re-checked as the
// ability resolves (CR 608.2b), and a target gone by then gets neither the
// goad nor the watch. The engine ends the goad as Maeve's controller's next
// turn begins.
//
// The second sentence is a delayed triggered ability (CR 603.7) the
// resolution creates, Okoye, Mighty and Adored's shape: bound to the
// goaded creature's object, so a creature that left and came back never
// fires it (CR 400.7); repeating for the rest of the turn (CR 603.7b), so
// a creature that attacks in a second combat draws again; and controlled
// by the player who activated the ability (CR 603.7e), who is "you". It
// fires only when the creature is declared attacking one of that player's
// opponents itself (CR 508.3a): an attack on a planeswalker or a battle is
// not one. "This turn" means the watch is only worth anything when Maeve
// is activated on the turn the creature will attack, typically during its
// controller's turn, before combat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2494daad-c81d-4a80-ba5d-e7011af8de46",
		Name:         "Maeve, Insidious Singer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Goad target creature. Whenever that creature attacks one of your opponents this turn, you draw a card.",
			Cost:    ManaCost("{2}{U}"),
			Targets: TargetCreature("target creature"),
			Effect:  maeveGoadAndWatch,
		}},
	})
}

// maeveThatCreatureAttacksAnOpponentCondition is "whenever that creature
// attacks one of your opponents": the one creature CondParams.Object
// names, declared as an attacker of a player other than the delayed
// trigger's controller. An on-disk key: never renamed.
var maeveThatCreatureAttacksAnOpponentCondition = game.DelayedCondition(
	"attack/that-creature-attacks-an-opponent-of-the-controller", maeveThatCreatureAttacksAnOpponent)

func maeveThatCreatureAttacksAnOpponent(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
	if ev.Kind != game.EventAttack || p.Object.ID == uuid.Nil || ev.CardID != p.Object.ID {
		return false
	}
	c, ok := battlefieldCardForEffect(g, p.Object.ID)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return false
	}
	opp := attackedPlayer(g, ev)
	return opp != uuid.Nil && opp != dt.Controller
}

// maeveGoadAndWatch goads the target and schedules the draw watch for
// that same creature.
func maeveGoadAndWatch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := goadAndScheduleClear(ctx, []uuid.UUID{id}); err != nil {
		return err
	}
	c, ok := battlefieldCardForEffect(g, id)
	if !ok {
		return nil
	}
	return DelayedOnEvent{
		Label:      "Maeve, Insidious Singer — that creature attacked one of your opponents: draw a card",
		On:         []game.EventKind{game.EventAttack},
		Condition:  maeveThatCreatureAttacksAnOpponentCondition,
		CondParams: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
		Body:       drawACardForTheControllerBody,
		Repeats:    true,
	}.Apply(ctx)
}
