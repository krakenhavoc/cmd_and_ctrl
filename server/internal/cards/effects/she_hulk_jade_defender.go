package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// She-Hulk, Jade Defender — {3}{G} Legendary Creature — Gamma Hero 4/4
// (Marvel Super Heroes, #4455):
//
//	"Reach, trample
//	 Power-up — {4}{G}{G}: Destroy up to one target artifact or
//	 enchantment. Put a +1/+1 counter on She-Hulk. (Activate each
//	 power-up ability only once. Reduce the cost by her mana cost if she
//	 entered this turn.)"
//
// Power-up is the engine's (CR 702.193). The target is "up to one", so the
// counter goes on whether or not a target was chosen, and a target that is
// gone by resolution is skipped (CR 608.2b) without stopping the counter.
//
// No simplification.
func init() {
	// Purpose set by hand until #4581 lands powerUpAnswering without its
	// answers parameter.
	ab := PowerUp("{4}{G}{G}: Destroy up to one target artifact or enchantment. Put a +1/+1 counter on She-Hulk.",
		ManaCost("{4}{G}{G}"), sheHulkJadeDefenderPowerUp)
	ab.Targets = TargetPermanent("up to one target artifact or enchantment", Or(Artifact(), Enchantment())).WithCount(0, 1)
	ab.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:        "bbd0a2a4-f81b-4a6b-bc4d-d7e53dc51333",
		Name:            "She-Hulk, Jade Defender",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Activated:       []ActivatedAbility{ab},
	})
}

func sheHulkJadeDefenderPowerUp(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		break
	}
	return plusOneCountersOnThis(1)(g, item)
}
