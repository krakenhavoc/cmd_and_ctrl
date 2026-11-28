package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b810_helpers.go — Mirage creatures (#5633).

// S82OldB810BlockPairFoe is the other creature of a block pair that names the
// source, when it passes ok as it stands (CR 509.3b, 509.3d: one event for
// each blocker-attacker pair). The colour or type word is read through the
// source's own text (CR 612.2).
func S82OldB810BlockPairFoe(ev game.Event, source *game.Card, g *game.Game, ok func(foe game.Card) bool) bool {
	foe := SkewerSlingerFoe(ev, source.InstanceID)
	if foe == uuid.Nil {
		return false
	}
	c, found := g.LookupCardForEffect(foe)
	return found && c.IsCreature() && ok(game.ThroughTextOf(source, c))
}

// S82OldB810DestroyBlockPairFoeAtEndOfCombat is "Whenever this creature blocks
// or becomes blocked by a <foe> creature, destroy that creature at end of
// combat" (Rock Basilisk, Dread Specter): one trigger for each pair, which
// records the other creature as the object it was (CR 400.7) and, as it
// resolves, schedules the destruction at the beginning of the end of combat
// step (CR 511.2) of that object if it is still on the battlefield.
func S82OldB810DestroyBlockPairFoeAtEndOfCombat(label string, ok func(foe game.Card) bool) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBlock},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return S82OldB810BlockPairFoe(ev, source, g, ok)
		},
		Key: label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Object = DamagedCreatureRef(g, SkewerSlingerFoe(ev, source.InstanceID))
			return item
		},
		Effect: ScheduleDestroyRecordedObjectAtEndOfCombat(label),
	}
}

// S82OldB810PhaseOutWithBlockPairFoe is "Whenever this creature blocks or
// becomes blocked by a creature, this creature and that creature phase out"
// (Dream Fighter): both leave together, each skipped if it is no longer on
// the battlefield (CR 608.2b).
func S82OldB810PhaseOutWithBlockPairFoe(label string) game.TriggeredAbility {
	return On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return S82OldB810BlockPairFoe(ev, source, g, func(game.Card) bool { return true })
	}, label, func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		ids := []uuid.UUID{item.SourceCardID}
		if foe := SkewerSlingerFoe(item.Trigger.Event, item.SourceCardID); foe != uuid.Nil {
			ids = append(ids, foe)
		}
		return g.PhaseOutForEffect(item.SourceCardID, ids...)
	})
}

// S82OldB810DiesToTopOfLibraryInstead is "If this creature would die, put it
// on top of its owner's library instead" (Gravebane Zombie): a self
// replacement of the move from the battlefield to the graveyard (CR 614.6),
// redirected to the top of its owner's library.
func S82OldB810DiesToTopOfLibraryInstead(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:         []game.EventKind{game.EventZoneMove},
		SelfReplacement: true,
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove && ev.CardID == src.InstanceID &&
				ev.OldZone == game.ZoneBattlefield && ev.NewZone == game.ZoneGraveyard
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.NewZone = game.ZoneLibrary
			ev.NewZoneOwner = uuid.Nil
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// ScheduleDestroyRecordedObjectAtEndOfCombat is the body of "destroy that
// creature at end of combat": the object the trigger recorded on
// item.Params.Object gets a delayed trigger at the beginning of the end of
// combat step (CR 511.2) that destroys it if it is still that object on the
// battlefield (CR 400.7). Lowland Basilisk, Rock Basilisk and Dread Specter
// share it.
func ScheduleDestroyRecordedObjectAtEndOfCombat(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ref := item.Params.Object
		if ref.ID == uuid.Nil {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:     game.StepEndCombat,
			Label:  label,
			Cards:  []uuid.UUID{ref.ID},
			Body:   DestroyTheObjectBody,
			Params: game.EffectParams{Object: ref},
		}.Apply(NewContext(g, item))
	}
}
