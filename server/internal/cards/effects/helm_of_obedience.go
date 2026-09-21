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
// # Both halves of the sentence count ARRIVALS (#1159, #1161)
//
// "Until a creature card OR X CARDS have been put into their
// graveyard this way, whichever comes first" is one clause with two
// stop conditions and the same verb governing both — "put into their
// graveyard this way", CR 400.7's arrived object. #1159 fixed the
// creature half; this is the other half, and it is written as what it
// is: UntilAny(UntilCard(creature), UntilCount(X)) over the cards
// that landed.
//
// So the Helm asks for an UNBOUNDED mill and stops itself. X is not
// the mill's amount, and that is the whole of the fix:
//
//   - a card the CR 614 window diverts costs the run nothing. A
//     milled commander whose owner takes the command zone was never
//     put into that graveyard, so it neither ends the run nor uses up
//     one of the X, and the Helm mills another card in its place.
//   - under Rest in Peace or Leyline of the Void NOTHING is ever put
//     into that graveyard, so the run never reaches X and the Helm
//     mills the victim's whole library. That is the famous combo, and
//     it falls out of the reading rather than being special-cased.
//   - nothing the run takes off the library is a mill AMOUNT any
//     more. CR 701.13b's number is the one an instruction names, and
//     an "until" run names none, so a mill-amount replacement has
//     nothing to double (game/mill.go, millAmountIsReplaceable):
//     Bruvac the Grandiloquent doubles mills, not bounds, and X=2 is
//     two cards with him on the battlefield exactly as it is without.
//     Paper doubles each one-card repetition instead, which can
//     overshoot the bound by a card; the engine models the run as one
//     instruction and stops on the number the card prints.
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
		Completeness: CompletenessFull,
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
		// No N: the bound is the other half of the CLAUSE, not the
		// mill's amount (#1161). N <= 0 with an Until is "no limit but
		// the library", and the library is exactly how far this runs
		// when a replacement keeps every card out of the graveyard.
		//
		// #1159 / #1161: both conditions are answered against what
		// reached the graveyard, so a commander whose owner takes the
		// command zone neither ends the run nor spends one of the X —
		// the Helm mills another card in its place, which is what the
		// card says.
		Until: UntilAny(
			UntilCard(func(c game.Card) bool { return c.IsCreature() }),
			UntilCount(ctx.X()),
		),
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
