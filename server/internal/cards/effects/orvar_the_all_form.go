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
// on the item as its payload as it triggers, because the spell's targets
// are read then and the spell may be gone by resolution. "One of those
// permanents" is your pick when there are several. A permanent that has
// left the battlefield by the time the ability resolves is not copied.
//
// The discard ability is the family's "this card" trigger (ADR 0153 §4),
// watching from the GRAVEYARD on a discard an opponent's spell or ability
// caused (the stack item's controller, CR 113.8). It has its own target
// clause, any permanent, chosen when it goes on the stack.
//
// Caveats, declared: a permanent that left the battlefield before the
// cast trigger resolves is not copied, where the printed card would copy
// it as it last existed; and if madness or Rest in Peace sends the
// discarded card to exile instead of the graveyard, its discard ability
// does not trigger from exile (owner decision 6).
func init() {
	Register(Spec{
		OracleID:     "276c4643-13ad-43bf-840b-08c025ebc065",
		Name:         "Orvar, the All-Form",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"If the permanent your spell targeted has left the battlefield before Orvar's copy ability resolves, no token is created.",
			discardTriggerFromExileCaveat,
		},
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
			Targeting(InGraveyard(On(game.EventDiscardCard, DiscardedByAnOpponent(true),
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
		item.Payload = append(item.Payload, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return item
}

// orvarCopyOneOfThem copies the only recorded permanent still on the
// battlefield, or asks its controller to pick among several.
func orvarCopyOneOfThem(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var left []uuid.UUID
	for _, id := range ctx.PayloadCards() {
		if g.Battlefield.Contains(id) {
			left = append(left, id)
		}
	}
	switch len(left) {
	case 0:
		return nil
	case 1:
		return CreateTokenCopy{Copy: left[0], N: 1}.Apply(ctx)
	}
	return ChoosePermanents{
		Question: "Orvar, the All-Form — copy one of those permanents",
		Candidates: func(_ *game.Game, _ uuid.UUID) ([]uuid.UUID, int, int) {
			return left, 1, 1
		},
		Then: func(ctx *Context, picked game.PromptedPicks) error {
			for _, id := range picked.Cards() {
				return CreateTokenCopy{Copy: id, N: 1}.Apply(ctx)
			}
			return nil
		},
	}.Apply(ctx)
}
