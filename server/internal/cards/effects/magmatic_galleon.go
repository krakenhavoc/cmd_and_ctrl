package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magmatic Galleon — Artifact — Vehicle, {3}{R}{R}:
//
//	"When this Vehicle enters, it deals 5 damage to target creature an
//	 opponent controls.
//	 Whenever one or more creatures your opponents control are dealt
//	 excess noncombat damage, create a Treasure token.
//	 Crew 2"
//
// The ETB is the ordinary Alchemax Slayer-Bots shape: a mandatory
// structured target (TargetCreature + OpponentControls), 5 damage read
// off ctx.LegalTargets() at resolution.
//
// The second ability watches EventDealDamage for a hit that pushed a
// creature past what it needed to die — "excess" isn't a field the
// event carries, so magmaticGalleonExcessNoncombatDamage reconstructs
// it: DamageMarked already includes this event's Amount by the time
// the trigger harvester sees it (applyDamageToPermanentLocked runs
// before emitDealDamageLocked), so subtracting the event's own Amount
// back out recovers the BEFORE state, and the predicate fires only
// when this specific event is what pushed the total past toughness —
// not merely because the creature was already over from an earlier,
// unrelated hit (an indestructible creature that already has lethal
// damage marked and then takes one more noncombat point should not
// mint a second Treasure off the old excess). OncePerBatch (#587)
// collapses "one or more" the same way every other batched wording
// does: three opponent creatures overkilled by one sweeper is one
// Treasure, not three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "59322432-591a-4e8d-aff5-12ca1feb1028",
		Name:         "Magmatic Galleon",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Magmatic Galleon — 5 damage to target creature an opponent controls",
				magmaticGalleonETBDamage),
				TargetCreature("target creature an opponent controls", OpponentControls())),
			OncePerBatch(On(game.EventDealDamage, magmaticGalleonExcessNoncombatDamage,
				"Magmatic Galleon — create a Treasure", Do(CreateToken{Template: TreasureToken(), N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:  "Crew 2",
			Cost:   CrewCost(2),
			Effect: CrewEffect("Magmatic Galleon"),
		}},
	})
}

// magmaticGalleonETBDamage reads the chosen target off ctx.LegalTargets()
// (CR 608.2b re-checks legality; an illegal-by-resolution target is
// simply not there any more) and deals 5 damage.
func magmaticGalleonETBDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: ts[0].ID, Amount: 5}.Apply(ctx)
}

// magmaticGalleonExcessNoncombatDamage reports whether ev is a
// noncombat EventDealDamage that dealt an opponent's creature more
// damage than it needed to die, and that THIS event is what did it
// (see the file comment for why the before/after subtraction
// matters).
func magmaticGalleonExcessNoncombatDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Combat || ev.Amount <= 0 {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !target.IsCreature() || target.Controller == source.Controller {
		return false
	}
	toughness := target.CurrentToughness()
	after := target.DamageMarked - toughness
	if after <= 0 {
		return false
	}
	before := (target.DamageMarked - ev.Amount) - toughness
	if before < 0 {
		before = 0
	}
	return after > before
}
