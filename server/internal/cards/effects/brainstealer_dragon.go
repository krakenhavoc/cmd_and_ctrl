package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brainstealer Dragon — Creature — Dragon Horror {5}{B}{B}, 6/6 (#4092):
//
//	"Flying
//	 At the beginning of your end step, exile the top card of each
//	 opponent's library. You may play those cards for as long as they
//	 remain exiled. If you cast a spell this way, you may spend mana as
//	 though it were mana of any color to cast it.
//	 Whenever a nonland permanent an opponent owns enters the
//	 battlefield under your control, they lose life equal to its mana
//	 value."
//
// The end step is Court of Locthwain's grant once per opponent:
// ExileTopWithPermission face up, WhileExiled ("for as long as they
// remain exiled"), not CastOnly (a land may be played), and AnyColor
// ("as though it were mana of any color", which is colour only, not
// type). The permission is the player's, so it outlives the Dragon.
//
// The second ability fires on the entry of any nonland permanent whose
// owner is someone else and whose controller is the Dragon's: a spell
// cast from that exile, a reanimated opponent's creature, the Dragon
// itself if it was an opponent's card. A gain-control effect moves no
// permanent onto the battlefield and does not trigger it. "They" is
// the owner and "its mana value" is the permanent's as it entered, both
// read off the trigger's object snapshot (CR 603.10a), so a permanent
// that leaves in response still costs its owner the life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "13cb958f-1e66-4219-901a-ff58275c8475",
		Name:            "Brainstealer Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Brainstealer Dragon — exile the top card of each opponent's library; you may play those cards", brainstealerDragonExile),
			On(game.EventETB, brainstealerDragonOpponentsPermanentEnters,
				"Brainstealer Dragon — its owner loses life equal to its mana value", brainstealerDragonDrain),
		},
	})
}

func brainstealerDragonExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	// One opponent's library at a time in APNAP order (#4266).
	return exileTopEachPlayerInTurn(ctx, ctx.Opponents(),
		ExileTopWithPermission{GrantTo: item.Controller, N: 1, AnyColor: true, WhileExiled: true}, game.KeyedTail{})
}

func brainstealerDragonOpponentsPermanentEnters(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok || c.IsLand() {
		return false
	}
	return c.Owner != uuid.Nil && c.Owner != source.Controller
}

func brainstealerDragonDrain(g *game.Game, item *game.StackItem) error {
	obj := NewContext(g, item).Trigger().Object
	if obj == nil || obj.Owner == uuid.Nil || obj.ManaValue <= 0 {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, obj.Owner, -obj.ManaValue)
}
