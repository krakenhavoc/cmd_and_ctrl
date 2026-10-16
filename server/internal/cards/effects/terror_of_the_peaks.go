package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terror of the Peaks — Creature — Dragon {3}{R}{R}, 5/4 (EDHREC
// rank 519):
//
//	"Flying
//	 Spells your opponents cast that target this creature cost an
//	 additional 3 life to cast.
//	 Whenever another creature you control enters, this creature
//	 deals damage equal to that creature's power to any target."
//
// The Dragon that turns every creature you cast into a Lightning
// Bolt or bigger. The trigger is targeted ("any target"), so the
// controller picks when it fires; the damage equals the entering
// creature's power read when the trigger RESOLVES (a pump in
// response counts, as printed), falling back to its power at entry
// if it has already left the battlefield — CR 608.2h's last known
// information, approximated by the value captured in Build. The
// Dragon itself is the damage source, so a damage doubler applies.
//
// "Spells your opponents cast that target this creature cost an
// additional 3 life to cast" is Spec.TargetTaxes (#3860, ADR 0048's
// addendum on target taxes): CR 601.2f's additional cost, charged by a
// permanent other than the spell and read once the spell's targets are
// announced. The cost-modifier engine (game.CostModifier, S28) prices
// GENERIC MANA only, so this is the AttackTax family's sibling
// (game/target_tax.go) rather than a modifier: the cast gate refuses a
// caster who can't pay the life (CR 119.4) and charges it with the other
// additional costs, the spell already on the stack; the legal-move
// enumerator prices each target set and drops the ones the caster can't
// afford, so neither the picker nor a bot is offered a cast that would
// be refused; and the view names the tax beside the Dragon in the
// picker. It is charged once however many of the spell's targets name
// the Dragon, does not apply to the Dragon's own controller or to an
// ability that targets it (that is ward's business), and goes away with
// the Dragon's abilities (CR 613.1f).
func init() {
	Register(Spec{
		OracleID:        "f8e17f4f-080d-4bba-bd05-ca27e94ccecc",
		Name:            "Terror of the Peaks",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TargetTaxes: []game.TargetTax{
			OpponentsSpellsTargetingThisCostLife(3,
				"Spells your opponents cast that target this creature cost an additional 3 life to cast."),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature()
			},
			Targets: TargetAny(),
			Key:     "Terror of the Peaks — damage equal to that creature's power to any target",
			// The power at entry is a board read made at trigger time
			// (ADR 0041 P9's fill-in Build) and goes onto Params.Amount
			// rather than being closed over; the entering creature's
			// identity is on item.Trigger.Event.CardID already.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Terror of the Peaks — damage equal to that creature's power to any target")
				if c, ok := g.LookupCardForEffect(ev.CardID); ok {
					item.Params.Amount = c.CurrentPower()
				}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				entered := item.Trigger.Event.CardID
				amount := item.Params.Amount
				if z := g.FindCardZoneForEffect(entered); z != nil && z.Kind == game.ZoneBattlefield {
					if c, ok := g.LookupCardForEffect(entered); ok {
						amount = c.CurrentPower()
					}
				}
				return DealDamage{
					Source: item.SourceCardID,
					Target: item.Targets[0].ID,
					Amount: amount,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
