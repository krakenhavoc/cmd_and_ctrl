package effects

import (
	"fmt"

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
	return putAttackingForm{from: from}.choose(ctx, card, name)
}

// putAttackingWithFinality is putAttackingYourChoice for "... tapped and
// attacking with a finality counter on it" (Grim Reaper, Lethal Legionnaire):
// the card enters with the counter (CR 614.1c), and the engine's CR 122.1h
// replacement does the rest.
func putAttackingWithFinality(ctx *Context, card uuid.UUID, from game.ZoneKind, name string) error {
	return putAttackingForm{from: from, finality: true}.choose(ctx, card, name)
}

// putAttackingForm is one way a card is put onto the battlefield tapped and
// attacking: the zone it must still be in, and whether it enters with a
// finality counter. Each form has its own keyed answer to "what does it
// attack?", because the answer carries the card and the targets as data and
// nothing else.
type putAttackingForm struct {
	from     game.ZoneKind
	finality bool
}

// enters is the counters the card enters with.
func (f putAttackingForm) enters() map[string]int {
	if f.finality {
		return map[string]int{game.CounterFinality: 1}
	}
	return nil
}

func (f putAttackingForm) choose(ctx *Context, card uuid.UUID, name string) error {
	then, ok := putAttackingThens[f]
	if !ok {
		return fmt.Errorf("%w: putAttackingYourChoice from %q", game.ErrInvalidParam, f.from)
	}
	me := ctx.Controller()
	refs := ctx.Game.AttackTargetsForEffect(me)
	if len(refs) == 0 {
		return nil
	}
	if len(refs) == 1 {
		return putAttacking(ctx.Game, me, card, f, refs[0].ID)
	}
	options := make([]game.ChoiceOption, len(refs))
	carry := []uuid.UUID{card}
	for i, r := range refs {
		options[i] = game.ChoiceOption{Label: "Attack " + attackTargetName(ctx.Game, r.ID)}
		if r.Kind != game.AttackTargetPlayer {
			options[i].Cards = []uuid.UUID{r.ID}
		}
		carry = append(carry, r.ID)
	}
	return PickOption{
		Question: name + " — what does it attack?",
		Options:  options,
		ThenKey:  then,
		Carry:    carry,
	}.Apply(ctx)
}

// putAttacking is the entry itself: the card, if it is still where the
// instruction found it, tapped and attacking `target`.
func putAttacking(g *game.Game, me, card uuid.UUID, f putAttackingForm, target uuid.UUID) error {
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != f.from {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: card, From: f.from}},
		game.ZoneEntryOptions{Controller: me, Tapped: true, Attacking: target, EntersWithCounters: f.enters()}, game.KeyedTail{})
}

// putAttackingThens are the keyed answers to "what does it attack?", one
// per form (#4253 slice 8). The prompt's Carry is the card and then the
// attack targets in the order the options list them, so the answer's index
// names one; the zone and the entry counters are the registration's.
var putAttackingThens = map[putAttackingForm]game.OptionPickThen{
	{from: game.ZoneExile}:                     putAttackingFromExile,
	{from: game.ZoneGraveyard}:                 putAttackingFromGraveyard,
	{from: game.ZoneGraveyard, finality: true}: putAttackingFromGraveyardFinality,
}

var (
	putAttackingFromExile             = OptionPickThen("option-pick/put-attacking-from-exile", putAttackingAnswer(putAttackingForm{from: game.ZoneExile}))
	putAttackingFromGraveyard         = OptionPickThen("option-pick/put-attacking-from-graveyard", putAttackingAnswer(putAttackingForm{from: game.ZoneGraveyard}))
	putAttackingFromGraveyardFinality = OptionPickThen("option-pick/put-attacking-from-graveyard-finality", putAttackingAnswer(putAttackingForm{from: game.ZoneGraveyard, finality: true}))
)

func putAttackingAnswer(form putAttackingForm) func(ctx *Context, r game.OptionPicked) error {
	return func(ctx *Context, r game.OptionPicked) error {
		if len(r.Carry) < 2 {
			return nil
		}
		card, targets := r.Carry[0], r.Carry[1:]
		index := r.Index
		if index < 0 || index >= len(targets) {
			index = 0
		}
		return putAttacking(ctx.Game, ctx.Controller(), card, form, targets[index])
	}
}
