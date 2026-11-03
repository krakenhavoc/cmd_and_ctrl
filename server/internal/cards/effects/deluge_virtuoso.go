package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deluge Virtuoso — Creature — Human Wizard {2}{U}, 2/2 (Secrets of
// Strixhaven, #4456):
//
//	"When this creature enters, tap target creature an opponent controls
//	 and put a stun counter on it. (If a permanent with a stun counter
//	 would become untapped, remove one from it instead.)
//	 Opus — Whenever you cast an instant or sorcery spell, this creature
//	 gets +1/+1 until end of turn. If five or more mana was spent to cast
//	 that spell, this creature gets +2/+2 until end of turn instead."
//
// The enters trigger is the shared tap-and-stun body (tapTargetAndStunIt):
// a target that has left or stopped being a legal pick by resolution is
// left alone (CR 608.2b). Opus is an ability word (CR 207.2c); the ability
// is one trigger on your instant and sorcery casts. The mana spent is
// counted as the ability triggers (the cast path records it before the
// event, and a waived payment asks what was paid: ReadsYourCasts) and
// carried on the trigger, so the ruling holds: it resolves before the spell,
// and the count it was cast with is what the "instead" reads even if that
// spell was countered. The pump is a single +1/+1 or a single +2/+2, never
// both.
//
// No simplification.
func init() {
	const opusLabel = "Deluge Virtuoso — Opus: +1/+1 until end of turn, or +2/+2 if five or more mana was spent"
	opus := On(game.EventCast, YouCast(Or(Instant(), Sorcery())), opusLabel,
		func(g *game.Game, item *game.StackItem) error {
			if !sourceIsStillThisPermanent(g, item) {
				return nil
			}
			n := 1
			if item.Params.Amount >= 5 {
				n = 2
			}
			return BoostUntilEOT{
				Target:    item.SourceCardID,
				Power:     n,
				Toughness: n,
				Label:     "Deluge Virtuoso — Opus pump",
			}.Apply(NewContext(g, item))
		})
	opus.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, opusLabel)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:       "416c70ab-f136-43e0-b5d5-9af8d4211c01",
		Name:           "Deluge Virtuoso",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsYourCasts,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Deluge Virtuoso — tap target creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("target creature an opponent controls", OpponentControls())),
			opus,
		},
	})
}
