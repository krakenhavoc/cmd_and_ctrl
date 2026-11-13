package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyrogoyf — Creature — Lhurgoyf {3}{R}, */1+* (Modern Horizons 3
// Commander, #5468):
//
//	"Pyrogoyf's power is equal to the number of card types among cards
//	 in all graveyards and its toughness is equal to that number plus 1.
//	 Whenever this creature or another Lhurgoyf creature you control
//	 enters, that creature deals damage equal to its power to any
//	 target."
//
// The power and toughness are Tarmogoyf's characteristic-defining
// ability (layer 7a), recomputed whenever a card enters or leaves a
// graveyard.
//
// The trigger fires for Pyrogoyf itself, and for another creature with
// the Lhurgoyf creature type (read through Pyrogoyf's text, CR 612.2, in
// the layered view, so a changeling counts) entering under its
// controller's control; a Lhurgoyf that is not a creature does not. The
// target is chosen as the trigger goes on the stack. The entering
// creature is the damage source, and its power is read as the trigger
// resolves, or as it last existed on the battlefield if it has left
// (Flayer of the Hatebound's shape, CR 608.2h); a power below zero deals
// no damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fd7457a-388d-4cca-a7cf-86b4ea922037",
		Name:         "Pyrogoyf",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Power is equal to the number of card types among cards in all graveyards, and its toughness is that number plus 1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := game.DistinctCardTypesInAllGraveyards(g)
				c.Power = n
				c.Toughness = n + 1
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, pyrogoyfLhurgoyfEntered,
				"Pyrogoyf — that creature deals damage equal to its power to any target",
				pyrogoyfDamage), TargetAny()),
		},
	})
}

// pyrogoyfLhurgoyfEntered is "whenever this creature or another Lhurgoyf
// creature you control enters".
func pyrogoyfLhurgoyfEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventETB && ev.CardID == source.InstanceID {
		return true
	}
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Lhurgoyf"))
}

// pyrogoyfDamage has the entering creature deal damage equal to its power,
// read now or as it last existed, to the trigger's target.
func pyrogoyfDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return DealDamage{
		Source: item.Trigger.Event.CardID,
		Target: targets[0].ID,
		Amount: info.Power,
	}.Apply(ctx)
}
