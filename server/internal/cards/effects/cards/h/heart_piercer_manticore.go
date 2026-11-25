package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heart-Piercer Manticore — Creature — Manticore {2}{R}{R}, 4/3
// (Commander 2019):
//
//	"When this creature enters, you may sacrifice another creature. When you
//	 do, this creature deals damage equal to that creature's power to any
//	 target.
//	 Embalm {5}{R} ({5}{R}, Exile this card from your graveyard: Create a
//	 token that's a copy of it, except it's a white Zombie Manticore with no
//	 mana cost. Embalm only as a sorcery.)"
//
// Ruthless Lawbringer's sacrifice and Minsc and Boo's payload: the
// sacrifice is asked as the enters trigger resolves (another creature, a
// floor of zero for the "may"), and once one has really gone the "when you
// do" reflexive trigger is created, carrying the sacrificed creature's
// last-known power (CR 608.2h), and chooses any target as it goes on the
// stack (CR 603.12). Embalm is the shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "253bec96-108c-4880-9c04-8bc99ceafa64",
		Name:         "Heart-Piercer Manticore",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Heart-Piercer Manticore — you may sacrifice another creature; when you do, it deals damage equal to that creature's power to any target", heartPiercerManticoreSacrifice),
		},
		Activated: []effects.ActivatedAbility{effects.Embalm("{5}{R}")},
	})
}

func heartPiercerManticoreSacrifice(g *game.Game, item *game.StackItem) error {
	return effects.MaySacrificeOne(g, item, "Heart-Piercer Manticore — sacrifice another creature to deal damage equal to its power to any target?",
		effects.Creature(), true, heartPiercerManticoreSacrificed)
}

// heartPiercerManticoreSacrificed is the "when you do". An on-disk key:
// never renamed, never reused.
var heartPiercerManticoreSacrificed = effects.RegisterSacrificedOne("prompt-run/heart-piercer-manticore-sacrifice", func(ctx *effects.Context, sacrificed uuid.UUID) error {
	return effects.ReflexiveTrigger{
		Label:  "Heart-Piercer Manticore — this creature deals damage equal to the sacrificed creature's power to any target",
		Body:   heartPiercerManticoreDamageBody,
		Params: game.EffectParams{Amount: effects.DepartedCreaturePower(ctx.Game, sacrificed)},
	}.Apply(ctx)
})

// heartPiercerManticoreDamageBody is the reflexive trigger: Params.Amount
// damage from this creature to the chosen target. An on-disk key: never
// renamed, never reused.
var heartPiercerManticoreDamageBody = game.ReflexiveBody("heart-piercer-manticore/damage", effects.BreechesBlastEffect,
	effects.ConstTargets(func() *game.TargetSpec { return effects.TargetAny() }))
