package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mask of Memory — Artifact — Equipment for {2} (EDHREC rank 990):
//
//	"Whenever equipped creature deals combat damage to a player, you
//	 may draw two cards. If you do, discard a card.
//	 Equip {1}"
//
// The third user of attachedCreatureDealtCombatDamageToPlayer, after
// the two Swords — the shape that makes an Equipment a card-advantage
// engine rather than a stat stick, and the reason evasive
// one-drops are worth equipping.
//
// "You may draw two cards. If you do, discard a card" is a single
// linked clause: the discard is the price of the draw, not a separate
// instruction, so declining the draw declines the discard. The
// engine has no yes/no prompt at trigger resolution, so the "may" is
// resolved as YES — which is the choice a player makes essentially
// every time (two-for-one into a pitch of the worst card in hand)
// and, critically, is the choice that keeps the linked discard
// attached to its draw. Taking the draw and skipping the discard
// would be stronger than printed; skipping both would be weaker.
// Declared as a caveat because an empty-library or hellbent corner
// exists where a player would genuinely decline.
//
// The discard used to be random, with a caveat saying so, because the
// catalog had no prompt for a chosen discard that anything waited on.
// #651 gave it one, so the controller picks their own pitch as
// printed (CR 701.8a), from the hand the two cards have already been
// drawn into.
func init() {
	Register(Spec{
		OracleID:     "d6b2c998-a226-426c-a40d-6e6007041bfe",
		Name:         "Mask of Memory",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Mask of Memory — draw two, discard one", maskOfMemoryMayDrawTwo),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}

// maskOfMemoryMayDrawTwo is the printed body: "you MAY draw two
// cards. If you do, discard a card."
//
// The question is asked at RESOLUTION (#796), which is where the card
// prints it — until the free yes/no existed the draw was simply taken,
// and the card said so in a caveat. The difference is real in two
// corners a Commander game reaches: a player who would deck
// themselves, and a player holding a card they would rather not be
// made to pitch.
//
// The discard is LINKED to the draw (CR 607.2) — it happens only on
// the branch where the cards were actually drawn, which is why it
// lives inside OnYes rather than after the prompt.
//
// Caller holds g.mu.
func maskOfMemoryMayDrawTwo(g *game.Game, item *game.StackItem) error {
	return MayChoice{
		Question: "Mask of Memory — draw two cards, then discard a card?",
		YesLabel: "Draw two",
		NoLabel:  "Decline",
		OnYes:    maskOfMemoryDrawThenDiscard,
	}.Apply(NewContext(g, item))
}

// maskOfMemoryDrawThenDiscard is the "if you do" branch.
//
// Caller holds g.mu.
func maskOfMemoryDrawThenDiscard(ctx *Context) error {
	if err := (DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
		return err
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player: ctx.Controller(),
		Source: ctx.Source(),
		N:      1,
	})
	return nil
}
