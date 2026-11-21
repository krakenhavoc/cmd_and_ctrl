package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Borg Queen — Legendary Artifact Creature — Borg Noble
// {4}{W}{U}{B}, 4/4 (Star Trek Commander, #4974):
//
//	"Artifact creatures you control get +2/+2.
//	 Whenever The Borg Queen enters or attacks, target opponent exiles
//	 cards from the top of their library until they exile a creature
//	 card. Assimilate that card. They put the rest on the bottom in a
//	 random order. (To assimilate a card, put it onto the battlefield
//	 under your control with a +1/+1 counter. It's a Borg artifact
//	 creature and loses all other creature types.)"
//
// The anthem is Tempered Steel's and includes the Queen herself.
//
// The trigger is one ability watching two events (Sun Titan). The
// opponent's library is exiled one card at a time until a creature card
// lands (Helm of Obedience's stop clause with no count bound); that card
// is assimilated from exile under the Queen's controller (the
// assimilate entry of Death Is Irrelevant: a +1/+1 counter, Artifact
// added, its creature types replaced by Borg, CR 205.1b), and once it
// has landed the other exiled cards go to the bottom of their owner's
// library in a random order. A library with no creature card is exiled
// whole and all of it goes to the bottom. An opponent no longer a legal
// target makes the trigger do nothing (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b03d6bc2-d9f1-4d03-b1b3-77993bec6547",
		Name:         "The Borg Queen",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16Anthem(effects.B36ArtifactCreatureYouControl, 2, 2)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEntersOrAttacks("The Borg Queen — target opponent exiles until a creature card; assimilate it", theBorgQueenExile),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

func theBorgQueenExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
		}
	}
	if opp == uuid.Nil {
		return nil
	}
	return effects.MillToZone{
		Player:   opp,
		To:       game.ZoneExile,
		Until:    effects.UntilCreatureOrCount(0),
		ThenTail: effects.WithCarry(ctx.ItemTail(theBorgQueenExiled), opp),
	}.Apply(ctx)
}

// theBorgQueenExiled runs once the exile has stopped: the last card
// exiled is the creature card (when there is one), assimilated; the rest
// wait in the next tail's Carry, the opponent first. An on-disk key.
var theBorgQueenExiled = effects.CardTail("trek-b31/borg-queen-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	opp := effects.TailCarry0(o)
	creature := uuid.Nil
	if n := len(o.IDs); n > 0 {
		if c, ok := ctx.Game.LookupCardForEffect(o.IDs[n-1]); ok && c.IsCreature() {
			creature = o.IDs[n-1]
		}
	}
	rest := []uuid.UUID{opp}
	for _, id := range o.IDs {
		if id != creature {
			rest = append(rest, id)
		}
	}
	then := effects.WithCarry(ctx.ItemTail(theBorgQueenRestOnBottom), rest...)
	if creature == uuid.Nil {
		return then.Run(ctx.Game, nil)
	}
	return effects.TrekB31AssimilateFromExileThen(ctx, creature, "The Borg Queen — assimilated: a Borg artifact creature", then)
})

// theBorgQueenRestOnBottom puts the other exiled cards on the bottom of
// their owner's library in a random order. Carry[0] is the opponent. An
// on-disk key.
var theBorgQueenRestOnBottom = effects.CardTail("trek-b31/borg-queen-rest-on-bottom", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) < 2 {
		return nil
	}
	return ctx.Game.PutOnBottomInRandomOrderForEffect(o.Tail.Carry[0], game.ZoneExile, o.Tail.Carry[1:])
})
