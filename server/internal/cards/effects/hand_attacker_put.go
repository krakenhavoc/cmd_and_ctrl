package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hand_attacker_put.go — "you may put a <creature> card from your hand onto
// the battlefield tapped and attacking" (Ultra Magnus, Tactician; Nick Fury,
// Spymaster).
//
// The card enters tapped and attacking a player, planeswalker or battle its
// controller chooses (CR 508.4, 506.3c), the triggering attack's own
// defender first. It was never declared as an attacker, so it triggers no
// "whenever attacks" ability. With nothing to attack it enters tapped, not
// attacking (CR 508.4a).
//
// The put is keyed (#4253 slice 8): the defender the source is attacking
// rides the card prompt as the first of its carry, the card and the
// defenders ride the "what does it attack?" prompt, and the item rides both
// as data. The two continuations are registered once per card under that
// card's key; the card supplies what follows an entry (tail), which is
// told what entered, so "if you do" is an entered != uuid.Nil.

// HandAttackerPutter is one card's registered put.
type HandAttackerPutter struct {
	name    string
	picked  game.ChooseCardsThen
	attacks game.OptionPickThen
	tail    game.TailRef
}

// NewHandAttackerPutter registers the two continuations of a card's put under
// "choose-cards/<key>-put" and "option-pick/<key>-attacks" (on-disk keys,
// never renamed) and keeps `tail`, run once the card entered. Call it from a
// package-level var. `name` heads the prompts.
func NewHandAttackerPutter(key, name string, tail game.TailRef) *HandAttackerPutter {
	p := &HandAttackerPutter{name: name, tail: tail}
	p.picked = game.RegisterChooseCardsThen("choose-cards/"+key+"-put",
		func(g *game.Game, r game.CardsChosen) error {
			if !r.Answered || len(r.Picked) == 0 || r.Asked == nil {
				return nil
			}
			var prefer uuid.UUID
			if len(r.Carry) > 0 {
				prefer = r.Carry[0]
			}
			return p.chooseDefender(g, r.Asked, r.Picked[0], prefer)
		})
	p.attacks = OptionPickThen("option-pick/"+key+"-attacks",
		func(ctx *Context, r game.OptionPicked) error {
			if len(r.Carry) < 2 {
				return nil
			}
			card, targets := r.Carry[0], r.Carry[1:]
			index := r.Index
			if index < 0 || index >= len(targets) {
				index = 0
			}
			return p.put(ctx.Game, ctx.Item, card, targets[index])
		})
	return p
}

// Offer asks the item's controller, who may decline, to pick one of
// `candidates` (cards in their hand) to put onto the battlefield tapped and
// attacking. The defender the item's trigger names is offered first.
func (p *HandAttackerPutter) Offer(g *game.Game, item *game.StackItem, candidates []uuid.UUID, question string) {
	if len(candidates) == 0 {
		return
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Destination: game.ChooseOntoBattlefieldTapped,
		Chooser:     item.Controller,
		Source:      item.SourceCardID,
		Question:    question,
		Cards:       candidates,
		Min:         0,
		Max:         1,
		ThenKey:     p.picked,
		LiveZone:    game.ZoneHand,
		Carry:       []uuid.UUID{triggerAttackTarget(item)},
		Asked:       game.NewAskedItem(item),
	})
}

// chooseDefender asks what the entering creature attacks when there is more
// than one choice (CR 508.4), `prefer` first.
func (p *HandAttackerPutter) chooseDefender(g *game.Game, item *game.StackItem, card, prefer uuid.UUID) error {
	var targets []uuid.UUID
	for _, ref := range g.AttackTargetsForEffect(item.Controller) {
		targets = append(targets, ref.ID)
	}
	if i := slices.Index(targets, prefer); i > 0 {
		targets = append([]uuid.UUID{prefer}, slices.Delete(targets, i, i+1)...)
	}
	switch len(targets) {
	case 0:
		// Nothing to attack: it enters tapped, not attacking (CR 508.4a).
		return p.put(g, item, card, uuid.Nil)
	case 1:
		return p.put(g, item, card, targets[0])
	}
	options := make([]game.ChoiceOption, len(targets))
	for i, id := range targets {
		options[i] = game.ChoiceOption{Label: "Attack " + attackTargetName(g, id)}
	}
	return PickOption{
		Question: p.name + " — what does it attack?",
		Options:  options,
		ThenKey:  p.attacks,
		WithItem: true,
		Carry:    append([]uuid.UUID{card}, targets...),
	}.Apply(NewContext(g, item))
}

// put moves the card from the hand, tapped and attacking `attacking`
// (uuid.Nil: not attacking), then runs the card's tail.
func (p *HandAttackerPutter) put(g *game.Game, item *game.StackItem, card, attacking uuid.UUID) error {
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneHand {
		return nil
	}
	return g.PutFromHandOntoBattlefieldKeyedForEffect(card, game.HandEntryOptions{
		Controller: item.Controller,
		Tapped:     true,
		Attacking:  attacking,
	}, NewContext(g, item).ItemTail(p.tail))
}
