package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "75854e14-5f15-443d-b342-b27fc3ab4b6e",
		Name:         "_____ _____ _____ Trespasser",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("_____ _____ _____ Trespasser — you may put a name sticker on it",
				Do(PutASticker{On: StickerOnThis, Kinds: []game.StickerKind{game.StickerName}})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: This creature gets +1/+0 until end of turn for each name sticker on it. It can't be blocked this turn.",
			Cost:    ManaCost("{3}{U}"),
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Effect:  blankTrespasserPump,
		}},
	})
}

func blankTrespasserPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if ctx.isNewSourceObject(item.SourceCardID) || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	if err := (BoostUntilEOT{Target: c.InstanceID, Power: nameStickerCount(c, nil), Label: "_____ _____ _____ Trespasser"}).Apply(ctx); err != nil {
		return err
	}
	return RestrictUntilEOT{Target: c.InstanceID, Restrictions: game.CantBeBlocked, Label: "_____ _____ _____ Trespasser — can't be blocked"}.Apply(ctx)
}
