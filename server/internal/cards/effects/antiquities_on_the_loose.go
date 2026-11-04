package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Antiquities on the Loose — Sorcery {1}{W}{W} (Secrets of Strixhaven,
// #4456):
//
//	"Create two 2/2 red and white Spirit creature tokens. Then if this
//	 spell was cast from anywhere other than your hand, put a +1/+1
//	 counter on each Spirit you control.
//	 Flashback {4}{W}{W} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// "Cast from anywhere other than your hand" reads the spell's announce
// record (CastFromZone), as Increasing Vengeance does: the flashback cast
// counts, and so does any other cast out of a graveyard, exile or the
// command zone. The counters go on every Spirit you control, the two new
// ones and any other, after the tokens exist (a second step, so a
// creation that waits on a question still gets them).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "9c45682f-48b5-4e58-9277-ff4211f389ae",
		Name:             "Antiquities on the Loose",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 2},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{W}{W}")},
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("2/2 red and white Spirit"), N: 2},
			Instruction(antiquitiesCounters),
		),
	})
}

// antiquitiesCounters is the "then if cast from anywhere other than your
// hand" half.
func antiquitiesCounters(ctx *Context) error {
	from := ctx.Item.CastFromZone
	if from == "" || from == game.ZoneHand {
		return nil
	}
	self, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok {
		return nil
	}
	spirit := self.TextSubtype("Spirit")
	var spirits []game.Card
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && c.HasSubtype(spirit) {
			spirits = append(spirits, c)
		}
	}
	for _, c := range spirits {
		if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
