package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salvation Colossus — Artifact Creature — Construct {6}{W}{W}, 9/9:
//
//	"Flying, vigilance, trample
//	 Whenever you attack, other creatures you control get +2/+2 and
//	 gain indestructible until end of turn.
//	 Unearth—Pay eight {E}. (Pay eight energy counters: Return this card
//	 from your graveyard to the battlefield. It gains haste. Exile it at
//	 the beginning of the next end step or if it would leave the
//	 battlefield. Unearth only as a sorcery.)"
//
// "Whenever you attack" is one trigger per attack declaration
// (OncePerBatch over EventAttack), whether or not the Colossus attacks.
// The pump and the grant lock their set as the trigger resolves (CR
// 611.2c). Unearth (CR 702.84a) is an activated ability from the
// graveyard, so its eight energy is an ordinary energy component (ADR
// 0129 §5), checked before anything is paid and never waived.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8b5c6754-36bd-44ac-b0c0-5a3cd2d9f2c3",
		Name:            "Salvation Colossus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Salvation Colossus — other creatures get +2/+2 and indestructible", salvationColossusAttack)),
		},
		Activated: []effects.ActivatedAbility{effects.UnearthPaying("Unearth—Pay eight {E}", effects.PayEnergy(8))},
	})
}

func salvationColossusAttack(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	others := effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID))
	if err := (effects.BoostUntilEOT{Match: others, Power: 2, Toughness: 2, Label: "Salvation Colossus — +2/+2"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Match: others, Keywords: []string{"indestructible"}, Label: "Salvation Colossus — indestructible"}.Apply(ctx)
}
