package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orvar, the All-Form — Legendary Creature — Shapeshifter {3}{U}, 3/3:
//
//	"Changeling
//	 Whenever you cast an instant or sorcery spell, if it targets one or
//	 more other permanents you control, create a token that's a copy of
//	 one of those permanents.
//	 When a spell or ability an opponent controls causes you to discard
//	 this card, create a token that's a copy of target permanent."
//
// Three abilities, as printed.
//
// The cast trigger has an intervening "if" (CR 603.4): when the spell is
// cast it must be an instant or sorcery that targets at least one
// permanent you control other than Orvar. The permanents are recorded
// on the item as its payload as it triggers, each as the OBJECT it was
// (a card and its CR 400.7 epoch), because the spell's targets are read
// then and the spell may be gone by resolution. "One of those
// permanents" is your pick when there are several. A permanent that has
// left the battlefield by the time the ability resolves is copied as it
// last existed there (CR 608.2h, CR 707.2: its last-known copiable
// values), not skipped, and a card that has come back as a new object is
// not mistaken for the one that was targeted.
//
// The discard ability is the family's "this card" trigger (ADR 0153 §4),
// watching from wherever the card landed (WhereverDiscarded, #3859) on a
// discard an opponent's spell or ability caused (the stack item's
// controller, CR 113.8). It has its own target clause, any permanent,
// chosen when it goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "276c4643-13ad-43bf-840b-08c025ebc065",
		Name:              "Orvar, the All-Form",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"changeling"},
		DiscardByOpponent: game.DiscardByOpponentTrigger,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventCast},
				AppliesTo: orvarSpellTargetsYourOtherPermanents,
				Key:       orvarCopyLabel,
				Build:     orvarRecordTargets,
				Effect:    orvarCopyOneOfThem,
			},
			Targeting(WhereverDiscarded(On(game.EventDiscardCard, DiscardedByAnOpponent(true),
				"Orvar, the All-Form — discarded: create a token that's a copy of target permanent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					return CreateTokenCopy{Copy: ts[0].ID, N: 1}.Apply(ctx)
				})), TargetPermanent("target permanent")),
		},
	})
}

const orvarCopyLabel = "Orvar, the All-Form — create a token that's a copy of one of those permanents"

// orvarTargetedPermanents lists the permanents the cast spell targets
// that are other permanents its caster controls, in target order and
// without repeats. nil when the spell is not an instant or sorcery.
func orvarTargetedPermanents(ev game.Event, source *game.Card, g *game.Game) []uuid.UUID {
	if ev.Actor != source.Controller {
		return nil
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil || item.Kind != game.StackItemSpell {
		return nil
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !(spell.HasCardType("instant") || spell.HasCardType("sorcery")) {
		return nil
	}
	var out []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || t.ID == source.InstanceID || seen[t.ID] {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || c.Controller != source.Controller || !g.Battlefield.Contains(t.ID) {
			continue
		}
		seen[t.ID] = true
		out = append(out, t.ID)
	}
	return out
}

func orvarSpellTargetsYourOtherPermanents(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return len(orvarTargetedPermanents(ev, source, g)) > 0
}

// orvarRecordTargets fills the item's payload with the permanents the
// spell targeted as it was cast.
func orvarRecordTargets(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
	ids := orvarTargetedPermanents(ev, source, g)
	if len(ids) == 0 {
		return nil
	}
	item := game.NewTriggeredItem(source, orvarCopyLabel)
	for _, id := range ids {
		ref, ok := g.PermanentRefForEffect(id)
		if !ok {
			continue
		}
		item.Payload = append(item.Payload, game.PayloadObjectRef(ref))
	}
	return item
}

// orvarCopyOneOfThem copies the only recorded permanent, or asks its
// controller to pick among several. Each is read as the object it was
// when the spell was cast (CR 608.2h): still on the battlefield, as it is
// now; gone from it, as it last existed there.
func orvarCopyOneOfThem(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var objects []game.ObjectRef
	for _, ref := range ctx.Payload() {
		if ref.Kind != game.TargetCard {
			continue
		}
		if _, ok := g.TokenCopyTemplateForObjectForEffect(ref.Object()); ok {
			objects = append(objects, ref.Object())
		}
	}
	switch len(objects) {
	case 0:
		return nil
	case 1:
		return orvarCopyObject(ctx, objects[0])
	}
	options := make([]game.ChoiceOption, len(objects))
	carry := make([]uuid.UUID, len(objects))
	epochs := make([]int, len(objects))
	for i, o := range objects {
		options[i] = game.ChoiceOption{Label: orvarOptionLabel(g, o)}
		carry[i], epochs[i] = o.ID, o.Epoch
	}
	return PickOption{
		Question: "Orvar, the All-Form — copy one of those permanents",
		Options:  options,
		ThenKey:  orvarCopyPicked,
		WithItem: true,
		Carry:    carry,
		Nums:     epochs,
	}.Apply(ctx)
}

// orvarCopyObject is the token copy of one recorded permanent, if its copiable
// values can still be read.
func orvarCopyObject(ctx *Context, o game.ObjectRef) error {
	tmpl, ok := ctx.Game.TokenCopyTemplateForObjectForEffect(o)
	if !ok {
		return nil
	}
	return createTokenCopies(ctx, ctx.Controller(), tmpl, 1)
}

// orvarCopyPicked is the recorded permanent chosen to be copied, keyed
// (#4253): the permanents ride the prompt as their IDs (the carry) and
// their epochs (the numbers), in the order they were listed. An on-disk key:
// never renamed, never reused.
var orvarCopyPicked = OptionPickThen("option-pick/orvar-the-all-form-copy", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(r.Carry) || r.Index >= len(r.Nums) {
		return nil
	}
	return orvarCopyObject(ctx, game.ObjectRef{ID: r.Carry[r.Index], Epoch: r.Nums[r.Index]})
})

// orvarOptionLabel names a recorded permanent for the pick: its name as
// it last existed on the battlefield.
func orvarOptionLabel(g *game.Game, o game.ObjectRef) string {
	if info, ok := g.PermanentForEffect(o); ok && info.Copiable != nil {
		return info.Copiable.Name
	}
	return "a permanent"
}
