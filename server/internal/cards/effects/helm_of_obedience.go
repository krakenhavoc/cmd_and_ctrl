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
// #1159 made that clause read what ARRIVED, so the caveat this card
// carried from the day it shipped is gone: a milled commander whose
// owner takes the command zone was never put into the graveyard
// (CR 400.7), so it does not end the run and the Helm keeps milling
// until a creature card really is put there. What is left of the
// caveat is the OTHER half of the printed sentence — "or X cards have
// been put into their graveyard this way" — where X still bounds the
// cards the run takes off the library rather than the cards that
// arrive. Fixing that means making the bound itself landed-counted,
// which is a change to the mill AMOUNT (CR 701.13b counts cards
// moved, and it is the number Bruvac the Grandiloquent doubles) and
// not to the clause; it is filed rather than smuggled in here.
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
			"X counts the cards the run takes off the library rather than the cards that reach the graveyard, so a card a replacement diverts on the way (a commander whose owner takes the command zone, CR 903.9) still uses up one of the X. The creature-card half of the clause is exact.",
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
	return MillToZone{
		Player: victim,
		N:      ctx.X(),
		// #1159: answered against what reached the graveyard, so a
		// commander whose owner takes the command zone does not end
		// the run — the Helm keeps milling until a creature card
		// really is put there, which is what the card says.
		Until: UntilCard(func(c game.Card) bool { return c.IsCreature() }),
		// #893: the reanimation reads what was PUT INTO THE GRAVEYARD,
		// so it runs from the continuation. A milled commander stops to
		// answer CR 903.9 and the creature card to reanimate is not
		// knowable until it does — the Helm used to read the list with
		// that prompt still open, find nothing, and drop its own second
		// half on the floor.
		Then: func(ctx *Context, milled []uuid.UUID) error {
			var creature uuid.UUID
			for _, id := range milled {
				c, ok := ctx.Game.LookupCardForEffect(id)
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
		},
	}.Apply(ctx)
}
