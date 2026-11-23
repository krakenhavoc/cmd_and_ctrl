package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sword of the Paruns — Artifact — Equipment {4} (Ravnica Remastered,
// #5948):
//
//	"As long as equipped creature is tapped, tapped creatures you control
//	 get +2/+0.
//	 As long as equipped creature is untapped, untapped creatures you
//	 control get +0/+2.
//	 {3}: You may tap or untap equipped creature.
//	 Equip {3}"
//
// Two layer 7c statics. Each reads the host's tapped state as the layers
// are rebuilt, so tapping or untapping the host moves the bonus in the same
// beat; an unattached Sword gives nothing. "Tapped creatures you control"
// includes the equipped creature itself.
//
// "You may tap or untap equipped creature" is one yes/no: a tapped host can
// only be untapped (tapping it again does nothing) and an untapped one can
// only be tapped, so the prompt names the one change that does something
// (swordOfTheParunsToggle).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc33444e-da8c-4af5-b3ed-552e91b0e656",
		Name:         "Sword of the Paruns",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			swordOfTheParunsStatic(true, "As long as equipped creature is tapped, tapped creatures you control get +2/+0.", 2, 0),
			swordOfTheParunsStatic(false, "As long as equipped creature is untapped, untapped creatures you control get +0/+2.", 0, 2),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}: You may tap or untap equipped creature.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{3}"),
				Effect:  swordOfTheParunsAsk,
			},
			effects.EquipAbility("{3}"),
		},
	})
}

// swordOfTheParunsStatic is one of the two statics: it applies to the
// creatures its controller has in the host's tapped state, and only while
// the host is in that state.
func swordOfTheParunsStatic(tapped bool, label string, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:    label,
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			if !target.IsCreature() || target.Controller != source.Controller || target.Tapped != tapped {
				return false
			}
			host := g.AttachedHostOf(source)
			return host != nil && host.IsCreature() && host.Tapped == tapped
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

func swordOfTheParunsAsk(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	host := effects.R35EquippedCreatureOf(g, item.SourceCardID)
	if host == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(host)
	if !ok {
		return nil
	}
	q, yes := "Sword of the Paruns — tap the equipped creature?", "Tap"
	if c.Tapped {
		q, yes = "Sword of the Paruns — untap the equipped creature?", "Untap"
	}
	return effects.MayChoice{Question: q, YesLabel: yes, NoLabel: "Leave it", Then: swordOfTheParunsToggle, Carry: []uuid.UUID{host}}.Apply(ctx)
}

// swordOfTheParunsToggle taps the equipped creature if it is untapped and
// untaps it if it is tapped, as it is now (CR 608.2b).
var swordOfTheParunsToggle = effects.MayYes("confirm/sword-of-the-paruns-tap-or-untap", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(r.Carry[0])
	if !ok {
		return nil
	}
	if c.Tapped {
		return effects.UntapTarget{Target: c.InstanceID}.Apply(ctx)
	}
	return effects.TapTarget{Target: c.InstanceID}.Apply(ctx)
})
