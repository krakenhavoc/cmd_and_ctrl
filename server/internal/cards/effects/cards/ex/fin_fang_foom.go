package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fin Fang Foom — {2}{R}{R} Legendary Creature — Alien Dragon Villain 3/5:
//
//	"Flying
//	 Whenever you cast an instant or sorcery spell that targets an artifact
//	 or land, copy that spell. You may choose new targets for the copy. Put
//	 two +1/+1 counters on Fin Fang Foom."
//
// The trigger reads the announced targets off the spell on the stack (the
// way Arcee, Sharpshooter does). A target counts if it is an artifact or a
// land permanent now. The copy is the controller's (CR 707.10) and the
// counters follow it.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "52ec472e-a745-4e25-b9a7-8e1eb8873594",
		Name:            "Fin Fang Foom",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, finFangFoomCast,
				"Fin Fang Foom — copy that spell and put two +1/+1 counters on him", finFangFoomCopy),
		},
	})
}

func finFangFoomCast(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !(spell.IsInstant() || spell.IsSorcery()) {
		return false
	}
	return finFangFoomTargetsArtifactOrLand(g, ev.CardID)
}

func finFangFoomTargetsArtifactOrLand(g *game.Game, spellID uuid.UUID) bool {
	item := g.StackItemForEffect(spellID)
	if item == nil || item.Kind != game.StackItemSpell {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || !effects.OnBattlefield(g, t.ID) {
			continue
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok && (c.IsArtifact() || c.IsLand()) {
			return true
		}
	}
	return false
}

func finFangFoomCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.CopySpell{StackID: ctx.Trigger().Event.CardID, ChooseNewTargets: true}).Apply(ctx); err != nil {
		return err
	}
	return effects.PlusOneCountersOnThis(2)(g, item)
}
