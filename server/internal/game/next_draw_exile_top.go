package game

import "github.com/google/uuid"

// next_draw_exile_top.go — a one-shot draw replacement a resolving
// ability sets up on a player (#5875): Urabrask, Heretic Praetor's "the
// next time they would draw a card this turn, instead they exile the top
// card of their library. They may play it this turn."
//
// # The shape
//
// A replacement created by a resolving ability is a ScopedEffect record
// (ADR 0041 P8, scoped_replacements.go), here ModNextDrawExileTopPlay
// naming the player whose draw it replaces, lasting until end of turn
// (CR 611.2a). It is Squid Fire Knight's dice shield's shape on the draw
// window: the record's adapted ReplacementEffect declares a DrawInstead
// (draw_instead.go), so applying it cancels that one draw (CR 614.6:
// nothing is drawn and no "whenever a player draws" trigger sees it) and,
// once the window has settled, runs nextDrawExileTopInstead. Applying it
// also spends the record ("the next time"), so a second draw that turn is
// an ordinary draw. A record no draw meets is swept at cleanup with the
// rest of the turn's effects: it expires unused.
//
// The replaced draw's player is the one who exiles and the one who may
// play the card, whoever controls the ability that made the record (the
// opponent, for Urabrask). The exile goes through the impulse-exile
// primitive (exile_play.go), so a commander on top is offered the command
// zone (CR 903.9) and gets no permission if it goes there, and an empty
// library exiles nothing: the replacement still applies, and nobody loses
// for drawing from an empty library, because nothing was drawn (CR
// 121.6a). The rest of a "draw two" resumes once the exile has settled
// (CR 121.6b, 614.11a).

// ModNextDrawExileTopPlay is "the next time <Player> would draw a card
// this turn, instead they exile the top card of their library. They may
// play it this turn." One-use: the record is dropped as it applies. An
// on-disk identity.
const ModNextDrawExileTopPlay ModKind = "nextDrawExileTopPlay"

var (
	// nextDrawExileTopInstead is the record's DrawInstead body.
	nextDrawExileTopInstead DrawInsteadRef
	// nextDrawExileTopDone pays the replaced draw's tail once the exile
	// has settled; its Carry[0] is the parked draw.
	nextDrawExileTopDone TailRef
)

// The keys are on-disk identities: never renamed, never reused.
func init() {
	nextDrawExileTopInstead = RegisterDrawInstead("engine/next-draw-exile-top-play", nextDrawExileTopRun)
	nextDrawExileTopDone = RegisterTail("engine/next-draw-exile-top-done", nextDrawExileTopPayDraw)
}

// nextDrawExileTopPayDraw runs the parked draw's tail.
//
// Caller must hold g.mu (write).
func nextDrawExileTopPayDraw(g *Game, o TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return DrawDoneByID(o.Tail.Carry[0]).Run(g)
}

// NextDrawExileTopPlayableForEffect registers the replacement on
// `player` until end of turn: the next time they would draw a card this
// turn, instead they exile the top card of their library and may play it
// this turn. `source` is the card whose ability made it and `controller`
// that ability's controller. Reports whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) NextDrawExileTopPlayableForEffect(source, controller, player uuid.UUID, label string) bool {
	if player == uuid.Nil || controller == uuid.Nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(source, ScopeGame, controller,
		[]Mod{{Kind: ModNextDrawExileTopPlay, Player: player}}, g.UntilEndOfTurnDuration(), label)
}

// nextDrawExileTopAppliesLocked is the record's AppliesTo: a draw by the
// named player.
func nextDrawExileTopAppliesLocked(m Mod, ev *ReplacementEvent) bool {
	return ev.Kind == RepEventDraw && ev.DrawCount > 0 && m.Player != uuid.Nil && ev.DrawPlayer == m.Player
}

// applyNextDrawExileTopLocked runs as the record's replacement applies,
// after the draw was cancelled and the body recorded: it names the
// record's source as the body's source and spends the record.
//
// Caller must hold g.mu (write).
func (g *Game) applyNextDrawExileTopLocked(e ScopedEffect, ev *ReplacementEvent) {
	ev.drawInstead.source = e.Source.ID
	seq := e.Seq
	g.dropScopedEffectsLocked(func(_ int, x ScopedEffect) bool { return x.Seq == seq })
}

// nextDrawExileTopRun is the body: the drawer exiles the top card of
// their library and may play it this turn, then the draw's instruction
// goes on.
//
// Caller must hold g.mu (write).
func nextDrawExileTopRun(g *Game, drawer, _ uuid.UUID, done DrawDone) error {
	then := KeyedTail{Ref: nextDrawExileTopDone, Player: drawer, Carry: []uuid.UUID{done.Park(g)}}
	return g.ExileTopWithPermissionKeyedForEffect(drawer, drawer, 1, CastPermission{}, then)
}
