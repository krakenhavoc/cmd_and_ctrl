package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Charismatic Conqueror — Creature — Vampire Soldier {1}{W}, 2/2:
//
//	"Vigilance
//	 Whenever an artifact or creature an opponent controls enters
//	 untapped, they may tap that permanent. If they don't, you create
//	 a 1/1 white Vampire creature token with lifelink."
//
// A tax on every opposing board development, and the decision is the
// opponent's — #568's yes/no half, asked of whoever controls the
// permanent that entered rather than of a target. Until MayChoice
// gained Player addressing (#796) there was no prompt that could ask
// it: pay_unless speaks about mana, and a trigger's CR 603.5 "you may"
// belongs to the trigger's controller, who is the wrong player here
// and would have been handed their opponent's decision.
//
// Two things the card needs that the engine does not do for it:
//
//   - "enters UNTAPPED" is read when the trigger FIRES, not at
//     resolution. A permanent that entered tapped never triggers,
//     however it is untapped afterwards; one that entered untapped and
//     was tapped in response still asks the question, and the answer
//     "tap it" then taps nothing and still makes no token.
//   - the permanent can LEAVE between the trigger and the answer.
//     "Tap that permanent" then happens to nothing — which is still an
//     answer the player gave, so it must not fall through to the
//     token.
//
// The permanent is captured in Build as a scalar instance ID, the way
// Amulet of Vigor captures it: a pointer into a zone slice would not
// survive the clone an undo restores from.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3dcc35f3-74dc-46a8-8aa9-3411189f0547",
		Name:            "Charismatic Conqueror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: charismaticConquerorWatch,
				Key:       charismaticConquerorLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					entered := ev.CardID
					if entered == uuid.Nil {
						return nil
					}
					return game.NewTriggeredItem(source, charismaticConquerorLabel,
						func(g *game.Game, item *game.StackItem) error {
							return charismaticConquerorAsk(g, item, entered)
						})
				},
			},
		},
	})
}

// charismaticConquerorLabel is the stack-overlay copy.
const charismaticConquerorLabel = "Charismatic Conqueror — they may tap it, or you get a Vampire"

// charismaticConquerorWatch is the trigger condition: an artifact or
// creature an OPPONENT controls entered UNTAPPED.
func charismaticConquerorWatch(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller == source.Controller || c.Tapped {
		return false
	}
	return c.IsArtifact() || c.IsCreature()
}

// charismaticConquerorAsk is the trigger's body: ask the permanent's
// controller.
//
// Caller holds g.mu.
func charismaticConquerorAsk(g *game.Game, item *game.StackItem, entered uuid.UUID) error {
	ctx := NewContext(g, item)
	c, ok := g.LookupCardForEffect(entered)
	if !ok || c.Controller == item.Controller {
		// It has left, or it changed hands to the Conqueror's own
		// controller. There is nobody to ask, so nothing happens —
		// not the token, which is the "if they don't" branch of a
		// question that was never put.
		return nil
	}
	name := c.Name
	if name == "" {
		name = "that permanent"
	}
	return MayChoice{
		Player:   c.Controller,
		Question: "Charismatic Conqueror — tap " + name + "?",
		YesLabel: "Tap it",
		NoLabel:  "They get a Vampire",
		OnYes:    charismaticConquerorTap(entered),
		OnNo:     charismaticConquerorToken,
	}.Apply(ctx)
}

// charismaticConquerorTap is the "they tap it" branch. A permanent
// that has left, or is already tapped, taps nothing — and still makes
// no token: the player answered, and the answer was yes.
func charismaticConquerorTap(entered uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		if z := ctx.Game.FindCardZoneForEffect(entered); z == nil || z.Kind != game.ZoneBattlefield {
			return nil
		}
		return TapTarget{Target: entered}.Apply(ctx)
	}
}

// charismaticConquerorToken is the "if they don't" branch.
//
// Caller holds g.mu.
func charismaticConquerorToken(ctx *Context) error {
	return CreateToken{
		Controller: ctx.Controller(),
		Template:   TokenCard("1/1 white Vampire with lifelink"),
		N:          1,
	}.Apply(ctx)
}
