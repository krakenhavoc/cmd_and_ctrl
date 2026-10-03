package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gruulBeastmasterLabel is the attack trigger's stack label.
const gruulBeastmasterLabel = "Gruul Beastmaster — another target creature you control gets +X/+0"

// Gruul Beastmaster — Creature — Human Shaman {3}{G}, 2/2:
//
//	"Riot (This creature enters with your choice of a +1/+1 counter or
//	 haste.)
//	 Whenever this creature attacks, another target creature you
//	 control gets +X/+0 until end of turn, where X is this creature's
//	 power."
//
// Riot is the engine's keyword (ADR 0109 §10). X is read as the trigger
// resolves (CR 608.2h), from the Beastmaster as it is then, or as it
// last existed if it has left the battlefield (CR 113.7a) — so a riot
// counter and any pump in response count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "924771b2-8566-4bdf-b089-85c3257b9900",
		Name:            "Gruul Beastmaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks(gruulBeastmasterLabel, gruulBeastmasterPump),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}

// gruulBeastmasterPump gives the target +X/+0, X the Beastmaster's power.
func gruulBeastmasterPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	return vsPumpTheTarget(ctx, info.Power, 0, gruulBeastmasterLabel)
}
