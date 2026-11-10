package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f153e369-1d5b-4d8b-a2f6-d93b849cabb5",
		Name:         "Niko, Light of Hope",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Niko, Light of Hope — create two Shard tokens", Do(CreateToken{Template: ShardToken(), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Exile target nonlegendary creature you control. Shards you control become copies of it until the next end step. " +
				"Return it to the battlefield under its owner's control at the beginning of the next end step.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCreature("target nonlegendary creature you control", Not(Legendary()), YouControl()),
			Effect:  stdB26NikoExile,
		}},
	})
}
