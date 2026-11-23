package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// King of the Oathbreakers — Legendary Creature — Spirit Noble {2}{W}{B}, 3/3
// (The Lord of the Rings: Tales of Middle-earth):
//
//	"Flying
//	 Whenever King of the Oathbreakers or another Spirit you control
//	 becomes the target of a spell, it phases out.
//	 Whenever King of the Oathbreakers or another Spirit you control
//	 phases in, create a tapped 1/1 white Spirit creature token with
//	 flying."
//
// The first trigger is Bonecrusher Giant's becomes-the-target-of-a-spell
// (EventBecomesTarget, read against the stack so abilities don't count),
// widened to this Spirit or any other Spirit its controller controls, and
// phases the targeted permanent out (PhaseOut, CR 702.26): until its
// controller's next untap step, and the spell then has an illegal target
// (CR 608.2b). The second watches EventPhaseIn for the same set, one trigger
// for each permanent that phases in (a permanent's own phase-in trigger is
// harvested once it is back, CR 702.26b). The token is created tapped.
// Creature type words are read through the source's text (CR 612.2).
//
// No simplification.
func init() {
	phasesOut := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBecomesTarget},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if ev.Kind != game.EventBecomesTarget || !kingOfTheOathbreakersSpirit(ev.CardID, source, g) {
				return false
			}
			return effects.B40TargetedByASpell(ev, g)
		},
		Key: "King of the Oathbreakers — it phases out",
		Effect: func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
				return nil
			}
			return effects.PhaseOut{Targets: []uuid.UUID{item.Trigger.Event.CardID}}.Apply(effects.NewContext(g, item))
		},
	}
	phasesIn := effects.On(game.EventPhaseIn, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Kind == game.EventPhaseIn && kingOfTheOathbreakersSpirit(ev.CardID, source, g)
	}, "King of the Oathbreakers — create a tapped 1/1 white Spirit creature token with flying",
		func(g *game.Game, item *game.StackItem) error {
			return effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("1/1 white Spirit with flying")).EntersTapped(), N: 1}.Apply(effects.NewContext(g, item))
		})
	effects.Register(effects.Spec{
		OracleID:        "3ee451e7-2b2f-4ce7-af96-540bba555c60",
		Name:            "King of the Oathbreakers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{phasesOut, phasesIn},
	})
}

// kingOfTheOathbreakersSpirit is "King of the Oathbreakers or another Spirit
// you control": the permanent `id` is the source, or a Spirit the source's
// controller controls.
func kingOfTheOathbreakersSpirit(id uuid.UUID, source *game.Card, g *game.Game) bool {
	if id == source.InstanceID {
		return true
	}
	c, ok := g.LookupCardForEffect(id)
	return ok && c.Controller == source.Controller && c.HasSubtype(source.TextSubtype("Spirit"))
}
