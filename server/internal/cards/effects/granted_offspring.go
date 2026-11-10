package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// granted_offspring.go — #5428: offspring a permanent gives the spells
// its controller casts. Zinnia, Valley's Voice: "Creature spells you cast
// gain offspring {2} as you cast them."
//
// CR 702.175a: "Offspring represents two abilities. "Offspring [cost]"
// means "You may pay an additional [cost] as you cast this spell" and
// "When this permanent enters, if its offspring cost was paid, create a
// token that's a copy of it, except it's 1/1.""
//
// The cost is a granted optional cost (game/granted_optional_cost.go)
// under a key of its own, so a creature that prints offspring keeps its
// own cost beside the granted one and each pays off by itself (CR
// 702.175b: "If a spell has multiple instances of offspring, each is paid
// separately and triggers based on the payments made for it, not any
// other instances of offspring."). For the same rule the grant is
// EachSource: two permanents granting offspring give two instances.
//
// The trigger is the permanent's own, derived from its cast record
// (game/granted_entry_trigger.go): the spell gained offspring as it was
// cast (CR 610.5), so the token is made whether or not the permanent
// that granted it is still on the battlefield.

// GrantedOffspringKey is the key of a granted offspring cost, as
// recorded on PaidCost.GrantedOptional and CastProvenance.GrantedOptional.
// An on-disk identity: never renamed, never reused.
const GrantedOffspringKey = "granted-offspring"

// OffspringForSpellsYouCast is "<spells> you cast gain offspring {cost}
// as you cast them" (Zinnia, Valley's Voice), declared in
// Spec.GrantedOptionalCosts. `label` is the static's own text; `spells`
// is read against the card being cast.
func OffspringForSpellsYouCast(label string, spells CardPredicate, mana string) game.GrantedOptionalCost {
	gr := SpellsYouCastHave(label, spells, game.AdditionalCost{
		Optional: true,
		Key:      GrantedOffspringKey,
		ManaCost: mana,
		Label:    "Offspring " + mana,
	})
	gr.EachSource = true
	return gr
}

// grantedOffspringTokenBody is the granted keyword's "create a token
// that's a copy of it, except it's 1/1", OffspringToken's body under a
// key of its own because the permanent has no catalog row for it. The
// copy is of the permanent as it is, or as it last existed if it has left
// (CR 608.2h). An on-disk key.
var grantedOffspringTokenBody = game.SimpleDelayedBody("offspring/granted-token", func(g *game.Game, item *game.StackItem) error {
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       item.SourceCardID,
		N:          1,
		Except:     offspringOneOne,
	}.Apply(NewContext(g, item))
})

func init() {
	game.RegisterGrantedOptionalEntryTrigger(GrantedOffspringKey, game.TriggeredAbility{
		Keyword: OffspringKey,
		Watches: []game.EventKind{game.EventETB},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID
		},
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return game.NewKeyedTriggeredItem(source, source.Name+" — offspring: create a 1/1 token copy of it",
				grantedOffspringTokenBody, game.EffectParams{})
		},
	})
}
