package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whirlwind, Killer Cyclone — Legendary Creature — Mutant Villain {2}{R},
// 2/3 (Marvel Super Heroes Commander, #4455):
//
//	"Haste
//	 Whenever one or more creatures you control that entered this turn
//	 attack a player, target creature that player controls can't block
//	 this turn."
//
// Attuma's "attack a player" batch (OncePerBatchPerPlayer, CR 603.2c): one
// declaration is one trigger per player attacked however many new creatures
// go at that player, and a second player attacked is a second trigger with
// its own target. A creature "entered this turn" if it has come onto the
// battlefield since the turn began (Game.EnteredThisTurn, not summoning
// sickness), so a creature that was already there when the turn started
// does not set it off. The target is a creature controlled by the player
// that attack was aimed at, chosen as the trigger goes on the stack
// (TargetsFrom reads the triggering event, CR 603.3d), and Clamor Shaman's
// "can't block this turn" is put on it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "72eabfcc-61b6-4a07-9a9c-b40a9c05cf66",
		Name:            "Whirlwind, Killer Cyclone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			withTargetsFrom(
				OncePerBatchPerPlayer(On(game.EventAttack, whirlwindNewCreatureAttacksAPlayer, whirlwindLabel, whirlwindCantBlock)),
				whirlwindTarget),
		},
	})
}

const whirlwindLabel = "Whirlwind, Killer Cyclone — target creature that player controls can't block this turn"

// whirlwindNewCreatureAttacksAPlayer is the When: a creature the source's
// controller controls that entered this turn was declared attacking a player.
func whirlwindNewCreatureAttacksAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return b16YouAttackedAPlayer(ev, source, g) && b06EnteredThisTurn(g, ev.CardID)
}

// whirlwindTarget is "target creature that player controls": the player the
// declared attack was aimed at.
func whirlwindTarget(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	defender := b17DefendingPlayer(g, tc.Event)
	if defender == uuid.Nil {
		return nil
	}
	return TargetCreature("target creature that player controls", ControlledBy(defender))
}

func whirlwindCantBlock(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: whirlwindLabel}.Apply(ctx)
}
