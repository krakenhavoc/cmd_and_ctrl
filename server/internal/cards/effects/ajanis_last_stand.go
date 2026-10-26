package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani's Last Stand — Enchantment {2}{W}{W}:
//
//	"Whenever a creature or planeswalker you control dies, you may
//	 sacrifice this enchantment. If you do, create a 4/4 white Avatar
//	 creature token with flying.
//	 When a spell or ability an opponent controls causes you to discard
//	 this card, if you control a Plains, create a 4/4 white Avatar
//	 creature token with flying."
//
// The first ability is a "you may sacrifice" with an "if you do": the
// token appears only when the enchantment really left the battlefield.
//
// The second is the family's "this card" trigger (ADR 0153 §4): it
// watches from wherever the card landed (WhereverDiscarded, #3859),
// on a discard an opponent's spell or ability caused (the stack item's
// controller, CR 113.8). It carries an intervening "if" (CR 603.4), so
// the Plains is checked when the discard happens and again as the
// ability resolves; "you" is the card's owner wherever it landed
// (CR 108.4).
//
// No simplification: the discard ability is watched from wherever the
// discard left the card (WhereverDiscarded, #3859), so madness or Rest in
// Peace sending it to exile instead of the graveyard does not stop it.
func init() {
	Register(Spec{
		OracleID:          "17c4807c-158e-4a86-90f1-d8fb2ebe892a",
		Name:              "Ajani's Last Stand",
		Completeness:      CompletenessFull,
		DiscardByOpponent: game.DiscardByOpponentTrigger,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, creatureOrPlaneswalkerYouControlDied,
				"Ajani's Last Stand — you may sacrifice it to create a 4/4 white Avatar with flying",
				ajanisLastStandSacrifice),
			WhereverDiscarded(On(game.EventDiscardCard, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return DiscardedByAnOpponent(true)(ev, source, ch, g) && controlsPlains(g, source.Controller)
			}, "Ajani's Last Stand — discarded: create a 4/4 white Avatar with flying", func(g *game.Game, item *game.StackItem) error {
				if !controlsPlains(g, item.Controller) {
					return nil
				}
				return CreateToken{Template: TokenCard("4/4 white Avatar with flying"), N: 1}.Apply(NewContext(g, item))
			})),
		},
	})
}

// creatureOrPlaneswalkerYouControlDied is "whenever a creature or
// planeswalker you control dies", read from the permanent's last-known
// types and controller (CR 603.10a).
func creatureOrPlaneswalkerYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	dead, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || leftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	return leftAsType(ev, dead, "creature") || leftAsType(ev, dead, "planeswalker")
}

// controlsPlains is the intervening "if you control a Plains".
func controlsPlains(g *game.Game, controller uuid.UUID) bool {
	return controlsAnyCard(g, controller, Subtype("Plains"))
}

// ajanisLastStandSacrifice is the dies trigger: sacrifice this
// enchantment if its controller says so, and only then the token.
func ajanisLastStandSacrifice(g *game.Game, item *game.StackItem) error {
	return MayChoice{
		Question: "Ajani's Last Stand — sacrifice it to create a 4/4 white Avatar with flying?",
		Then:     ajanisLastStandAnswer,
	}.Apply(NewContext(g, item))
}

// ajanisLastStandAnswer is the "yes": sacrifice the enchantment, and only
// then the token. Keyed (#4291).
var ajanisLastStandAnswer = MayYes("confirm/ajanis-last-stand-sacrifice", func(ctx *Context, _ game.ConfirmAnswer) error {
	return ajanisLastStandSacrificeThen(ctx.Game, ctx.Item)
})

// ajanisLastStandSacrificeThen is the sacrifice and, if it went, the
// token (#4253 slice 10a, keyed).
var ajanisLastStandSacrificeThen = SacrificeThisThen("then/ajanis-last-stand-avatar", func(ctx *Context) error {
	return CreateToken{Template: TokenCard("4/4 white Avatar with flying"), N: 1}.Apply(ctx)
})
