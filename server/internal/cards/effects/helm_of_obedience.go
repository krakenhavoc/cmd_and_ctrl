package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helm of Obedience — Artifact {4}:
//
//	"{X}, {T}: Target opponent mills a card, then repeats this process
//	 until a creature card or X cards have been put into their
//	 graveyard this way, whichever comes first. If one or more
//	 creature cards were put into that graveyard this way, sacrifice
//	 this artifact and put one of them onto the battlefield under your
//	 control. X can't be 0."
//
// The card that named the gap on #74, and the reason AbilityCost
// grew a MinX rather than just an X. "X can't be 0" is a printed
// floor on the announcement, not a suggestion: without it Helm is a
// free repeatable tap that mills nothing, and an enumerator that
// offered X=0 would be handing bots a move worth exactly nothing
// forever. With it, a controller who cannot pay {1} has no
// activation at all — which is the right shape, because the
// alternative is an ability that is always "legal" and never does
// anything.
//
// The mill is MillToZone's Until clause verbatim: "until a creature
// card ... whichever comes first" is a run that stops after the
// first card it accepts, and the card that ends it still moves.
// Diffing the graveyard afterwards would be wrong the moment
// anything else put a card there in the same resolution.
//
// "One of them" needs no prompt. The run stops AT the first creature
// card, so there is never more than one to choose from — the plural
// in the oracle text is there for the rules, not for the player.
//
// The sacrifice is at RESOLUTION, not part of the cost: the Helm is
// still on the battlefield while the mill happens, and an opponent
// who removes it in response to the activation does not stop the
// ability (CR 608.2). It is also conditional — a mill that finds no
// creature leaves the Helm in play to try again.
//
// The reanimation says "under YOUR control", so it routes through
// ReturnFromGraveyardUnderControlForEffect with the activator named
// explicitly. The default there is "under its owner's control",
// which would hand the opponent's creature straight back to the
// opponent — the whole point of the card, silently inverted.
func init() {
	Register(Spec{
		OracleID:     "16cadebf-c484-41f8-9e38-5c2c528f5b54",
		Name:         "Helm of Obedience",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"If the creature card milled is a commander, the Helm stops milling but puts nothing onto the battlefield — its owner is still being asked whether the commander goes to the command zone instead.",
		},
		Activated: []ActivatedAbility{{
			Label:   "{X}, {T}: Target opponent mills until a creature card or X cards are in their graveyard; reanimate it.",
			Cost:    Plus(ManaCost("{X}"), TapCost(), MinX(1)),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect:  helmOfObedienceMill,
		}},
	})
}

func helmOfObedienceMill(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
		return nil
	}
	victim := item.Targets[0].ID
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	ctx := NewContext(g, item)
	// ctx.X() is the value announced at CR 602.2b and locked onto
	// the stack item — the same X the cost charged. MinX(1) above is
	// what guarantees it is at least 1 here, so the "X can't be 0"
	// clause needs no second check in the effect.
	var milled []uuid.UUID
	if err := (MillToZone{
		Player: victim,
		N:      ctx.X(),
		Until:  func(c game.Card) bool { return c.IsCreature() },
		Milled: &milled,
	}).Apply(ctx); err != nil {
		return err
	}
	var creature uuid.UUID
	for _, id := range milled {
		c, ok := g.LookupCardForEffect(id)
		if ok && c.IsCreature() {
			creature = id
			break
		}
	}
	if creature == uuid.Nil {
		return nil
	}
	// "sacrifice this artifact AND put one of them onto the
	// battlefield" — in that order, so an aristocrats payoff
	// watching the Helm die sees it before the creature arrives.
	if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return ReturnFromGraveyard{
		Target:     creature,
		Dest:       game.ZoneBattlefield,
		Controller: item.Controller,
	}.Apply(ctx)
}
