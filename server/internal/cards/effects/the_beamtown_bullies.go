package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Beamtown Bullies — Legendary Creature — Ogre Devil Warrior
// {1}{B}{R}{G}, 4/4 (Streets of New Capenna Commander, #5491):
//
//	"Vigilance, haste
//	 {T}: Target opponent whose turn it is puts target nonlegendary
//	 creature card from your graveyard onto the battlefield under their
//	 control. It gains haste. Goad it. At the beginning of the next end
//	 step, exile it. (Until your next turn, that creature attacks each
//	 combat if able and attacks a player other than you if able.)"
//
// Two target clauses (Clauses, read by slot): an opponent whose turn it is,
// and a nonlegendary creature card in the Bullies' controller's own
// graveyard. Each is re-checked as the ability resolves and the whole
// ability does nothing unless it can still use them (CR 608.2b is per
// target, and with either gone the put cannot happen). The card enters
// under the chosen opponent's control (CR 110.2) and the rest of the
// sentence reads the permanent that entered: it gains haste for as long as
// it stays that object, it is goaded by the Bullies' controller until that
// player's next turn (CR 701.15), and a delayed trigger (CR 603.7) exiles it
// at the beginning of the next end step. If the entry is replaced or
// refused (a Clone, a creature that can't enter), nothing is goaded and
// nothing is scheduled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "62d394e1-6e7d-4703-8fd8-494f0f043876",
		Name:            "The Beamtown Bullies",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
		Activated: []ActivatedAbility{{
			Label: "{T}: Target opponent whose turn it is puts target nonlegendary creature card from your graveyard onto the battlefield under their control. It gains haste. Goad it. At the beginning of the next end step, exile it.",
			Cost:  TapCost(),
			Targets: Clauses(
				TargetPlayer("target opponent whose turn it is", trekB39OpponentWhoseTurnItIs()),
				TargetCardInGraveyard("target nonlegendary creature card from your graveyard", Creature(), Not(Legendary()), YouOwn()),
			),
			Effect: beamtownBulliesResolve,
		}},
	})
}

func beamtownBulliesResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	who, ok := ctx.ClauseTarget(0)
	if !ok || who.Kind != game.TargetPlayer {
		return nil
	}
	card, ok := ctx.ClauseTarget(1)
	if !ok || card.Kind != game.TargetCard {
		return nil
	}
	return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(card.ID, who.ID, false, ctx.ItemTail(beamtownBulliesRider))
}

// beamtownBulliesRider is "It gains haste. Goad it. At the beginning of the
// next end step, exile it." for the permanent that entered. An on-disk key:
// never renamed, never reused.
var beamtownBulliesRider = EnteredTail("beamtown-bullies/haste-goad-exile", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	if err := (ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "The Beamtown Bullies — it gains haste",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := goadAndScheduleClear(ctx, []uuid.UUID{entered}); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		Label: "The Beamtown Bullies — exile it",
		Body:  exileListedCardsBody,
		Cards: []uuid.UUID{entered},
	}.Apply(ctx)
})
