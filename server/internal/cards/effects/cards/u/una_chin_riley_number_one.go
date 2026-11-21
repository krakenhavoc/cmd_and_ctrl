package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Una Chin-Riley, Number One — Legendary Creature — Illyrian Officer
// {4}{W}, 4/5 (Star Trek):
//
//	Whenever you attack while creatures you control have total power 8
//	or greater, target attacking creature gets +2/+0 and gains first
//	strike until end of turn.
//
// "Whenever you attack" is once per declaration (OncePerBatch, The Chief
// Warg's shape). "While" is part of the trigger condition, not an
// intervening "if": the total power — current power, attackers included —
// is read as attackers are declared and is not checked again on
// resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0529c6fa-a35f-4d33-b274-93446ce68419",
		Name:         "Una Chin-Riley, Number One",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller) && effects.CreaturesYouControlTotalPowerAtLeast(8)(g, source.Controller)
				}, "Una Chin-Riley — target attacking creature gets +2/+0 and gains first strike until end of turn",
					unaChinRileyPump),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()))),
		},
	})
}

func unaChinRileyPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Una Chin-Riley — +2/+0 until end of turn"}).Apply(ctx); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike"}, Label: "Una Chin-Riley — first strike until end of turn"}.Apply(ctx)
	}
	return nil
}
