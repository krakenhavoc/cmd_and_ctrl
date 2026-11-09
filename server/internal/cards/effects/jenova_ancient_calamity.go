package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jenova, Ancient Calamity — Legendary Creature — Alien {2}{B}{G}, 1/5
// (Final Fantasy, #4462):
//
//	"At the beginning of combat on your turn, put a number of +1/+1
//	 counters equal to Jenova's power on up to one other target creature.
//	 That creature becomes a Mutant in addition to its other types.
//	 Whenever a Mutant you control dies during your turn, you draw cards
//	 equal to its power."
//
// The counters count Jenova's power as it last existed (the ruling, CR
// 608.2h): current power while she is on the battlefield, else the
// departure record's. "Becomes a Mutant" is Sorin's layer-4 subtype add,
// pinned to the object for good (CR 611.2a). The death trigger reads the
// creature's subtypes and controller as it last existed (CR 603.10a) and
// its power the same way; it triggers for each Mutant dying together with
// Jenova, Jenova included if she is a Mutant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9370f6bf-d6a0-4c5b-828f-7229f563cc07",
		Name:         "Jenova, Ancient Calamity",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat(jenovaCountersLabel, jenovaCounters),
				Another(UpToOneTargetCreature("up to one other target creature"))),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || !isActivePlayer(g, source.Controller) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					if !ok || !leftAsType(ev, c, "creature") || leftUnderControlOf(ev, c) != source.Controller {
						return false
					}
					mutant, _ := ev.WasSubtype("Mutant")
					return mutant
				},
				Key: jenovaDrawLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, jenovaDrawLabel)
					epoch := -1
					if c, ok := g.LookupCardForEffect(ev.CardID); ok {
						epoch = c.ObjectEpoch
					}
					item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: epoch}
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{N: b29PowerAsItLastStood(g, item.Params.Object.ID)}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

const (
	jenovaCountersLabel = "Jenova, Ancient Calamity — put +1/+1 counters equal to her power on up to one other target creature; it becomes a Mutant"
	jenovaDrawLabel     = "Jenova, Ancient Calamity — draw cards equal to the Mutant's power"
)

// jenovaCounters is the combat trigger's body.
func jenovaCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if n := b29PowerAsItLastStood(g, item.SourceCardID); n > 0 {
		if err := (AddCounter{Target: target, Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return jenovaMutant(ctx, g, target)
}

func jenovaMutant(ctx *Context, g *game.Game, target uuid.UUID) error {
	return ScopedEffectFor{
		Target:   target,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Mutant")...)},
		Duration: g.PinnedTo(game.IndefiniteDuration(), target),
		Label:    "Jenova, Ancient Calamity — it becomes a Mutant in addition to its other types",
	}.Apply(ctx)
}
