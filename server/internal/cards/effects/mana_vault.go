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
// resolves (CR 603.4). If Mana Vault has left the battlefield by then, the
// resolution check reads its last-known tapped status and the damage comes
// from the departed Vault (#1396, CR 608.2h).
//
// Declared simplification: the trigger does not remember WHICH Mana Vault
// object it came from — a triggered item carries its source's card, not its
// object epoch — so the resolution reads the Vault the card most recently
// was. A Vault that leaves and comes back before the trigger resolves is
// therefore judged as the new one, which normally enters untapped and so
// deals no damage. Tracked as #1418.
func init() {
	Register(Spec{
		OracleID:              "736892cb-a34b-4bb9-b56c-e26e3db207a2",
		Name:                  "Mana Vault",
		Completeness:          CompletenessCaveats,
		Caveats:               []string{"If Mana Vault leaves the battlefield and comes back before its draw-step ability resolves, the ability checks the new Mana Vault, so it usually deals no damage."},
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		ManaAbilities:         []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}{C}{C}", Label: "Add {C}{C}{C}"}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Mana Vault — pay {4} to untap", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{4}",
					Question: "Mana Vault — pay {4} to untap it?",
					OnPay: func(ctx *Context) error {
						return UntapTarget{Target: item.SourceCardID}.Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
			On(game.EventBeginDrawStep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && ev.Kind == game.EventBeginDrawStep && source.Tapped
			}, "Mana Vault — deal 1 damage", manaVaultDrawStepDamage),
		},
	})
}

// manaVaultDrawStepDamage is the draw-step trigger's resolution: "if this
// artifact is tapped, it deals 1 damage to you", with the intervening if
// re-checked now (CR 603.4) against the Vault as it is, or as it last
// existed once it has left (#1396).
func manaVaultDrawStepDamage(g *game.Game, item *game.StackItem) error {
	ref, ok := g.PermanentRefForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	vault, ok := g.PermanentForEffect(ref)
	if !ok || !vault.Tapped {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: item.Controller, Amount: 1}.Apply(NewContext(g, item))
}
