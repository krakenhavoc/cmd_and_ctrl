package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stalking Yeti — Snow Creature — Yeti {2}{R}{R}, 3/3 (Coldsnap):
//
//	"When this creature enters, if it's on the battlefield, it deals damage
//	 equal to its power to target creature an opponent controls and that
//	 creature deals damage equal to its power to this creature.
//	 {2}{S}: Return this creature to its owner's hand. Activate only as a
//	 sorcery."
//
// Not a fight (CR 701.14): nothing here needs the two to be creatures at the
// time, but both are. The two damage amounts are read as the trigger resolves
// and dealt together as one damage event (CR 120.3, CR 608.2h). The Yeti must
// still be on the battlefield as the ability resolves; if it is not, it does
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e79e67db-3f2e-4002-9e6c-8f06ca9d8af4",
		Name:         "Stalking Yeti",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Stalking Yeti — it deals damage equal to its power to target creature an opponent controls and that creature deals damage equal to its power to it",
				stalkingYetiBite),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{S}: Return this creature to its owner's hand. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{2}{S}"),
			SorcerySpeed: true,
			Effect:       effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}

func stalkingYetiBite(g *game.Game, item *game.StackItem) error {
	if !effects.B09SourceStillOnBattlefield(g, item) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	yeti, ok1 := g.LookupCardForEffect(item.SourceCardID)
	other, ok2 := g.LookupCardForEffect(target)
	if !ok1 || !ok2 {
		return nil
	}
	fromYeti, fromOther := max(yeti.CurrentPower(), 0), max(other.CurrentPower(), 0)
	return g.DamageInstanceForEffect(func() error {
		if err := (effects.DealDamage{Source: item.SourceCardID, Target: target, Amount: fromYeti}).Apply(ctx); err != nil {
			return err
		}
		return effects.DealDamage{Source: target, Target: item.SourceCardID, Amount: fromOther}.Apply(ctx)
	})
}
