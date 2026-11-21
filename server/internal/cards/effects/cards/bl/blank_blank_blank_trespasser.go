package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// _____ _____ _____ Trespasser — Creature — Human Rogue Guest {1}{U}, 2/1:
//
//	"When this creature enters, you may put a name sticker on it.
//	 {3}{U}: This creature gets +1/+0 until end of turn for each name
//	 sticker on it. It can't be blocked this turn."
//
// The count is read as the ability resolves, from the creature while it
// stands. Both halves are the one ability's effect: the pump and "can't be
// blocked" are given together, so an ability that finds the creature gone
// does neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75854e14-5f15-443d-b342-b27fc3ab4b6e",
		Name:         "_____ _____ _____ Trespasser",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("_____ _____ _____ Trespasser — you may put a name sticker on it",
				effects.Do(effects.PutASticker{On: effects.StickerOnThis, Kinds: []game.StickerKind{game.StickerName}})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: This creature gets +1/+0 until end of turn for each name sticker on it. It can't be blocked this turn.",
			Cost:    effects.ManaCost("{3}{U}"),
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Effect:  blankTrespasserPump,
		}},
	})
}

func blankTrespasserPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if ctx.IsNewSourceObject(item.SourceCardID) || !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	if err := (effects.BoostUntilEOT{Target: c.InstanceID, Power: effects.NameStickerCount(c, nil), Label: "_____ _____ _____ Trespasser"}).Apply(ctx); err != nil {
		return err
	}
	return effects.RestrictUntilEOT{Target: c.InstanceID, Restrictions: game.CantBeBlocked, Label: "_____ _____ _____ Trespasser — can't be blocked"}.Apply(ctx)
}
