package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// putAttackingYourChoice asks what `card` attacks — any player,
// planeswalker or battle the controller could attack — and puts it onto
// the battlefield under the controller's control tapped and attacking
// that (CR 506.3c, CR 508.4). With one choice there is no question.
//
// `from` is the zone the card must still be in when it moves, and it is
// checked twice: the controller's answer to "what does it attack?" is a
// pause in which the card can leave (CR 400.7 — a card that has changed
// zones is a new object, and the instruction names the old one). The
// helper began life in the Grim Captain's file hard-wired to exile
// (#2709); Interceptor, Shadow's Hound returns itself from the graveyard
// (#3771), and the question and the entry are the same.
//
// The card was never declared as an attacker, so its own "whenever this
// attacks" does not trigger and neither does anyone else's "whenever a
// creature attacks".
func putAttackingYourChoice(ctx *Context, card uuid.UUID, from game.ZoneKind, name string) error {
	me := ctx.Controller()
	refs := ctx.Game.AttackTargetsForEffect(me)
	if len(refs) == 0 {
		return nil
	}
	put := func(g *game.Game, target uuid.UUID) error {
		if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != from {
			return nil
		}
		return g.PutOntoBattlefieldTogetherKeyedForEffect(
			[]game.BatchEntry{{CardID: card, From: from}},
			game.ZoneEntryOptions{Controller: me, Tapped: true, Attacking: target}, game.KeyedTail{})
	}
	if len(refs) == 1 {
		return put(ctx.Game, refs[0].ID)
	}
	options := make([]game.ChoiceOption, len(refs))
	for i, r := range refs {
		options[i] = game.ChoiceOption{Label: "Attack " + attackTargetName(ctx.Game, r.ID)}
		if r.Kind != game.AttackTargetPlayer {
			options[i].Cards = []uuid.UUID{r.ID}
		}
	}
	return PickOption{
		Question: name + " — what does it attack?",
		Options:  options,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(refs) {
				index = 0
			}
			return put(ctx.Game, refs[index].ID)
		},
	}.Apply(ctx)
}
