package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jurati, Cordial Queen — Legendary Artifact Creature — Borg Scientist
// {1}{W}{U}, 3/3 (Star Trek Commander, #4974):
//
//	"Whenever one or more artifact creatures you control deal combat damage
//	 to a player, create a token that's a copy of target artifact token you
//	 control."
//
// One trigger per damaged player per combat damage step, however many of
// your artifact creatures hit that player (CR 603.2c; Stilt-Man's batch).
// The dealer's types are read as it dealt the damage. The target is an
// artifact token you control, chosen as the trigger goes on the stack;
// with none the trigger is removed (CR 603.3d). The copy is of the token's
// copiable values (CR 707.2), and is yours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4dc4c8a3-17d6-4c2d-bcbf-9cebf261eda6",
		Name:         "Jurati, Cordial Queen",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				OncePerBatchPerPlayer(On(game.EventDealDamage, trekB39ArtifactCreaturesHitAPlayer,
					"Jurati, Cordial Queen — create a token that's a copy of target artifact token you control",
					trekB39TokenCopyOfTheTarget)),
				TargetPermanent("target artifact token you control", Artifact(), IsTokenPredicate(), YouControl())),
		},
	})
}
