package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mana Vault —
//
// "This artifact doesn't untap during your untap step. At the beginning of
// your upkeep, you may pay {4}. If you do, untap this artifact. At the
// beginning of your draw step, if this artifact is tapped, it deals 1 damage
// to you. {T}: Add {C}{C}{C}."
//
// The draw-step trigger checks its condition both when it triggers and when it
// resolves (CR 603.4). The resolution check reads "this artifact" through
// Context.SourcePermanent: the Vault as it is now while it is still the object
// the trigger came from, and its last-known tapped status once it has left
// (#1396, CR 608.2h). The damage comes from that same object.
//
// "This artifact" is the OBJECT, not the card (CR 400.7). A Vault that leaves
// and comes back before the trigger resolves is a new object, so the trigger
// judges the Vault that left — tapped, so it deals its 1 damage — and never
// the new, normally untapped one. That was the caveat until the trigger item
// learned its source object (#1418). The upkeep trigger's untap is the same
// read: a Vault that is a new object since the trigger was put on the stack
// is not the one it untaps.
func init() {
	Register(Spec{
		OracleID:              "736892cb-a34b-4bb9-b56c-e26e3db207a2",
		Name:                  "Mana Vault",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		ManaAbilities:         []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}{C}{C}", Label: "Add {C}{C}{C}"}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Mana Vault — pay {4} to untap", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{4}",
					Question: "Mana Vault — pay {4} to untap it?",
					OnPay:    manaVaultUntapThis,
				}.Apply(NewContext(g, item))
			}),
			On(game.EventBeginDrawStep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && ev.Kind == game.EventBeginDrawStep && source.Tapped
			}, "Mana Vault — deal 1 damage", manaVaultDrawStepDamage),
		},
	})
}

// manaVaultUntapThis is the upkeep trigger's "if you do, untap this
// artifact": only while the Vault is still the object the trigger came from
// (#1418). Last-known information is read, never untapped.
func manaVaultUntapThis(ctx *Context) error {
	vault, ok := ctx.SourcePermanent()
	if !ok || vault.Left {
		return nil
	}
	return UntapTarget{Target: ctx.Source()}.Apply(ctx)
}

// manaVaultDrawStepDamage is the draw-step trigger's resolution: "if this
// artifact is tapped, it deals 1 damage to you", with the intervening if
// re-checked now (CR 603.4) against the Vault the trigger came from, as it is
// or as it last existed once it has left (#1396, #1418).
func manaVaultDrawStepDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	vault, ok := g.PermanentForEffect(ref)
	if !ok || !vault.Tapped {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: item.Controller, Amount: 1}.Apply(ctx)
}
