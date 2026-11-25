package cardsth

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Celestial Toymaker — Legendary Creature — Rogue Artificer
// {W}{U}{B}, 2/4 (#6175):
//
//	"Whenever The Celestial Toymaker attacks, look at the top three
//	 cards of your library. Exile any number of them in a face-down
//	 pile and the rest in a face-up pile. Defending player chooses one
//	 of those piles. Put that pile into your hand.
//	 At the beginning of each end step, each opponent loses 2 life for
//	 each spell or ability that caused one or more players to guess or
//	 to group cards or permanents into a pile this turn."
//
// # The attack trigger
//
// The controller looks at the top three cards (so they are the one
// knower, LookAtTopOfPlayersLibraryForEffect) and picks the face-down
// pile out of them with a RevealPick over cards they own — Nael's
// shape. The picked cards are exiled face down and the rest face up, in
// that printed order, before anyone chooses: a pile the defending player
// takes goes to the controller's HAND and the pile they leave stays in
// exile, which is why the exile cannot be deferred to the end the way
// Fact or Fiction's split defers its move.
//
// The defending player then picks a pile. The prompt is an option pick
// whose two options carry the piles, so the wire's one redaction pass
// keeps the face-down pile's backs out of their view while its size
// rides the label (Abstract Performance's shape). The defending player
// is read off the triggering attack event each time it is needed rather
// than carried, since nothing between the prompts can change it.
//
// A creature attacking a planeswalker or a battle has a defending
// player too (the walker's controller, the battle's protector); one
// attacking nothing (CR 506.4c) has none, and then nobody chooses and
// both piles stay exiled, which is weaker than printed, never stronger.
//
// # The end step
//
// "Each spell or ability that caused one or more players to guess or to
// group cards or permanents into a pile this turn" is the engine's turn
// tally (#6175, game/pile_guess_tally.go, CR 700.3): every pile and
// guess effect in the catalog notes itself there, once per spell or
// ability, and the count is read as this trigger resolves. It counts
// everyone's effects, the Toymaker's own attack trigger among them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "388fe92a-0bc0-4c0a-a7c6-910c5e71e342",
		Name:         "The Celestial Toymaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("The Celestial Toymaker — look at the top three cards of your library, exile them in a face-down pile and a face-up pile, and the defending player chooses which pile goes to your hand",
				celestialToymakerAttack),
			effects.On(game.EventBeginEndStep, effects.AnyPlayer,
				"The Celestial Toymaker — each opponent loses 2 life for each spell or ability that made a player guess or group objects into a pile this turn",
				celestialToymakerDrain),
		},
	})
}

// celestialToymakerDefender is the attack's defending player, read off
// the triggering event (CR 506.2): uuid.Nil for an attacker that is
// attacking nothing.
func celestialToymakerDefender(ctx *effects.Context) uuid.UUID {
	return effects.B17DefendingPlayer(ctx.Game, ctx.Trigger().Event)
}

func celestialToymakerAttack(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	you := item.Controller
	// #6175: the controller groups these cards into two piles; a pile can
	// hold zero objects (CR 700.3d), so an empty library still counts.
	effects.NotePileOrGuess(ctx)
	looked := g.LookAtTopOfPlayersLibraryForEffect(you, you, 3)
	if len(looked) == 0 {
		return nil
	}
	return effects.RevealPick{
		Player:   you,
		Owner:    you,
		Question: "The Celestial Toymaker — choose any number of them for the face-down pile; the rest go in the face-up pile",
		Cards:    looked,
		Min:      0,
		Max:      len(looked),
		Keyed:    celestialToymakerSplit,
	}.Apply(ctx)
}

// celestialToymakerSplit exiles the chosen cards face down; the cards
// left ride its tail and are exiled face up next. An on-disk key: never
// renamed, never reused.
var celestialToymakerSplit = effects.RegisterRevealPick("choose-cards/celestial-toymaker-split",
	func(ctx *effects.Context, picked, left []uuid.UUID, _ []int) error {
		return ctx.Game.ExileCardsFaceDownChosenKeyedForEffect(ctx.Controller(), picked,
			effects.WithCarry(ctx.ItemTail(celestialToymakerFaceDownExiled), left...))
	})

// celestialToymakerFaceDownExiled exiles the rest face up, carrying the
// face-down pile that just landed. An on-disk key.
var celestialToymakerFaceDownExiled = effects.CardTail("then/celestial-toymaker-face-down-exiled",
	func(ctx *effects.Context, o game.TailOutcome) error {
		faceDown := append([]uuid.UUID(nil), o.IDs...)
		return ctx.Game.ExileCardsKeyedForEffect(o.Tail.Carry,
			effects.WithCarry(ctx.ItemTail(celestialToymakerPilesExiled), faceDown...))
	})

// celestialToymakerPilesExiled asks the defending player which pile
// goes to the controller's hand. An on-disk key.
var celestialToymakerPilesExiled = effects.CardTail("then/celestial-toymaker-piles-exiled",
	func(ctx *effects.Context, o game.TailOutcome) error {
		g := ctx.Game
		defender := celestialToymakerDefender(ctx)
		if defender == uuid.Nil || g.PlayerByIDForEffect(defender) == nil {
			// Nobody is defending: no pile is chosen and both stay exiled.
			return nil
		}
		faceDown := effects.AbstractPerformanceStillExiled(g, o.Tail.Carry)
		faceUp := effects.AbstractPerformanceStillExiled(g, o.IDs)
		if len(faceDown) == 0 && len(faceUp) == 0 {
			return nil
		}
		return effects.PickOption{
			Player:   defender,
			From:     ctx.Controller(),
			Question: "The Celestial Toymaker — choose a pile; it goes into its owner's hand",
			Options: []game.ChoiceOption{
				{Label: "The face-down pile (" + strconv.Itoa(len(faceDown)) + " cards)", Cards: faceDown},
				{Label: "The face-up pile (" + strconv.Itoa(len(faceUp)) + " cards)", Cards: faceUp},
			},
			ThenKey:  celestialToymakerPilePicked,
			Carry:    effects.PackIDLists(faceDown, faceUp),
			WithItem: true,
		}.Apply(ctx)
	})

// celestialToymakerPilePicked puts the chosen pile into the Toymaker's
// controller's hand; the other pile stays in exile. An on-disk key.
var celestialToymakerPilePicked = effects.OptionPickThen("option-pick/celestial-toymaker-pile",
	func(ctx *effects.Context, r game.OptionPicked) error {
		if r.Option == nil {
			return nil
		}
		g, you, source := ctx.Game, ctx.Controller(), ctx.Source()
		faceDown, faceUp := effects.UnpackIDPair(r.Carry)
		taken := faceDown
		if r.Index == 1 {
			taken = faceUp
		}
		for _, id := range effects.AbstractPerformanceStillExiled(g, taken) {
			if _, err := g.ReturnFromAnyZoneToHandForEffect(you, source, id); err != nil {
				return err
			}
		}
		return nil
	})

// celestialToymakerDrain is the end step: 2 life per spell or ability
// that made a player guess or group objects into a pile this turn, read
// as the trigger resolves (CR 608.2). None is no loss at all.
func celestialToymakerDrain(g *game.Game, item *game.StackItem) error {
	n := g.PilesOrGuessesThisTurnForEffect()
	if n <= 0 {
		return nil
	}
	return effects.EachOpponentLosesLifeNow(g, item, 2*n)
}
