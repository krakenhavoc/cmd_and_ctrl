package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valley Floodcaller — Creature — Otter Wizard {2}{U}, 2/2:
//
//	"Flash
//	 You may cast noncreature spells as though they had flash.
//	 Whenever you cast a noncreature spell, Birds, Frogs, Otters, and
//	 Rats you control get +1/+1 until end of turn. Untap them."
//
// The flash grant is Gandalf's narrowed timing rule (noncreature
// instead of sorcery). The trigger snapshots the matching creatures as
// it resolves (CR 611.2c): they all get the pump, and the same set is
// untapped. A creature that arrives afterwards gets neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4879c8f0-8832-4290-bc71-9838940f75cd",
		Name:            "Valley Floodcaller",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{NoncreatureOnly: true},
				"You may cast noncreature spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Valley Floodcaller — Birds, Frogs, Otters, and Rats you control get +1/+1 and untap", valleyFloodcallerPumpAndUntap),
		},
	})
}

// valleyFloodcallerCreatures is "Birds, Frogs, Otters, and Rats you
// control".
func valleyFloodcallerCreatures() effects.CardPredicate {
	return effects.And(effects.YouControl(), effects.Or(effects.OfCreatureType("Bird"), effects.OfCreatureType("Frog"), effects.OfCreatureType("Otter"), effects.OfCreatureType("Rat")))
}

func valleyFloodcallerPumpAndUntap(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	match := valleyFloodcallerCreatures()
	var ids []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if match(g, item.Controller, c) {
			ids = append(ids, c)
		}
	}
	if err := (effects.BoostUntilEOT{Match: match, Power: 1, Toughness: 1, Label: "Valley Floodcaller — +1/+1"}).Apply(ctx); err != nil {
		return err
	}
	for _, c := range ids {
		if err := (effects.UntapTarget{Target: c.InstanceID}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
