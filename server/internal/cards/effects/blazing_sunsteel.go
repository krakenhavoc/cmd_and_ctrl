package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blazing Sunsteel — Artifact — Equipment {1}{R} (Commander Legends;
// deck request #4724):
//
//	"Equipped creature gets +1/+0 for each opponent you have.
//	 Whenever equipped creature is dealt damage, it deals that much
//	 damage to any target.
//	 Equip {4}"
//
// The pump is Blackblade Reforged's counted bonus (PumpAttachedPer),
// counting the opponents of the EQUIPMENT's controller who are still in
// the game. It is read on every layer pass, so it shrinks the moment an
// opponent loses or concedes, and a stolen carrier still counts the
// Sunsteel controller's opponents.
//
// The reflect is Pain for All's watch on the attached creature with Ill-
// Tempered Loner's target: the trigger is the Equipment's, so its
// controller picks any target as it goes on the stack, and the damaged
// creature deals the damage on resolution, as printed ("it deals"). The
// engine opens one damage event per source, so the trigger waits for the
// damage to finish (AtBatchEnd, ThatMuchDamage) and reflects the TOTAL
// the creature was dealt at the same time with one target pick: a
// creature blocked by two creatures triggers once (CR 120.3, CR 510.2).
// A creature that died to the damage still deals it, from its last-known
// information, and the trigger still resolves if the Sunsteel was moved
// or left in response: it names the creature that was dealt the damage.
// "Any target" includes that creature itself.
//
// Equip is the ordinary sorcery-speed activation.
//
// No simplification.
func init() {
	const label = "Blazing Sunsteel — equipped creature deals that much damage to any target"
	Register(Spec{
		OracleID:     "46c77100-ab2e-43eb-b1cb-c22ed114213e",
		Name:         "Blazing Sunsteel",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 0, blazingSunsteelOpponentsYouHave),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == ev.Target
			},
			Targets:    TargetAny(),
			Key:        label,
			AtBatchEnd: ThatMuchDamage(),
			Effect:     blazingSunsteelReflect,
		}},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}

// blazingSunsteelOpponentsYouHave counts the Equipment controller's
// opponents still in the game. It walks the seats only, so it is safe
// inside a layer pass.
func blazingSunsteelOpponentsYouHave(g *game.Game, source *game.Card) int {
	n := 0
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.ID != source.Controller {
			n++
		}
	}
	return n
}

// blazingSunsteelReflect is the trigger's body: the creature the
// triggering event damaged deals the settled total to the target.
func blazingSunsteelReflect(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	if ev.Amount <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: ev.Target, Target: t.ID, Amount: ev.Amount}.Apply(ctx)
	}
	return nil
}
