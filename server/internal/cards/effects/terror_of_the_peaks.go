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
// Sandbox simplification, WEAKER than printed: the life tax on
// spells that target it is not implemented. That is a
// cost-modification effect (#93, the roadmap's largest missing
// mechanic) with no seam in the cast path; the card is played for
// the trigger, which is complete.
func init() {
	Register(Spec{
		OracleID:        "f8e17f4f-080d-4bba-bd05-ca27e94ccecc",
		Name:            "Terror of the Peaks",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Opponents' spells that target it do not cost an additional 3 life."},
		PrintedKeywords: []string{"flying"},
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
				item := game.NewTriggeredItem(source, "Terror of the Peaks — damage equal to that creature's power to any target", nil)
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
