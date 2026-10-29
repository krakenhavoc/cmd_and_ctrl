package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bard the Bowman — Legendary Creature — Human Archer {1}{W}{U}, 1/3:
//
//	"Reach
//	 Whenever you draw your second card each turn, put a +1/+1
//	 counter on target creature. It gains lifelink until end of turn."
//
// "Second card" is Knights of Dol Amroth's condition. The target is any
// creature, not only yours. Both halves act on the same creature; if
// it left or became illegal in response, neither happens (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ec076f5b-b0e3-4b6f-9293-a8fc42f20bd8",
		Name:            "Bard the Bowman",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Bard the Bowman — put a +1/+1 counter on target creature; it gains lifelink until end of turn",
				Do(Instruction(bardCounterAndLifelink))),
				TargetCreature("target creature")),
		},
	})
}

// bardCounterAndLifelink is the whole body: a +1/+1 counter on the
// target, which then gains lifelink until end of turn.
func bardCounterAndLifelink(ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: "Bard the Bowman — lifelink"}.Apply(ctx)
}
