package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avatar Roku, Firebender — Legendary Creature — Human Avatar
// {3}{R}{R}{R}, 6/6 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Whenever a player attacks, add six {R}. Until end of combat, you
//	 don't lose this mana as steps end.
//	 {R}{R}{R}: Target creature gets +3/+0 until end of turn."
//
// "Whenever a player attacks" triggers once for each declaration of one or
// more attackers by any player (CR 508.3d), opponents included, so it is
// one trigger per declaration (OncePerBatch). The mana goes to Roku's
// controller and is firebending's mana: six red that the step-boundary
// sweep keeps through every step of combat and drops at the first
// boundary outside it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9eb9e4c1-fabf-45ff-a584-fbe4b94fb41c",
		Name:         "Avatar Roku, Firebender",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, AnyPlayer,
				"Avatar Roku, Firebender — a player attacked: add six {R} until end of combat", Do(redUntilEndOfCombat(6)))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{R}{R}{R}: Target creature gets +3/+0 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 3, 0)),
			Cost:    ManaCost("{R}{R}{R}"),
			Targets: TargetCreature("target creature"),
			Effect:  boostTargetUntilEOT("Avatar Roku, Firebender — +3/+0", 3, 0),
		}},
	})
}
