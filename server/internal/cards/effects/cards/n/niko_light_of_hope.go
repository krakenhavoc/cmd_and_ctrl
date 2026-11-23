package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Niko, Light of Hope — Legendary Creature — Human Wizard {2}{W}{U}, 3/4
// (Duskmourn: House of Horror, #4466):
//
//	"When Niko enters, create two Shard tokens. (They're enchantments with
//	 "{2}, Sacrifice this token: Scry 1, then draw a card.")
//	 {2}, {T}: Exile target nonlegendary creature you control. Shards you
//	 control become copies of it until the next end step. Return it to the
//	 battlefield under its owner's control at the beginning of the next end
//	 step."
//
// The tokens are the predefined Shard (CR 111.10e). The ability exiles the
// target, then makes each Shard you control at that moment a copy of the
// creature as it last existed on the battlefield (PermanentInfo.Copiable,
// read by the creature's object, so a Clone the target was copying is
// copied as what it copied, CR 707.2), for a duration of its own: the new
// game.UntilNextEndStep, which ends as the next end step of any turn begins
// (BecomeCopy.Duration CopyUntilNextEndStep). The return is the delayed
// trigger every blink uses, firing at the same step. If the target was
// made illegal, none of it happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f153e369-1d5b-4d8b-a2f6-d93b849cabb5",
		Name:         "Niko, Light of Hope",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Niko, Light of Hope — create two Shard tokens", effects.Do(effects.CreateToken{Template: effects.ShardToken(), N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, {T}: Exile target nonlegendary creature you control. Shards you control become copies of it until the next end step. " +
				"Return it to the battlefield under its owner's control at the beginning of the next end step.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target nonlegendary creature you control", effects.Not(effects.Legendary()), effects.YouControl()),
			Effect:  effects.StdB26NikoExile,
		}},
	})
}
