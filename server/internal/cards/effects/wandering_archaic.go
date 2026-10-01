package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wandering Archaic // Explore the Vastlands — modal double-faced card
// (slice 296-m; the back face #1743):
//
//	Wandering Archaic — Creature — Avatar {5}, 4/4
//	 "Whenever an opponent casts an instant or sorcery spell, they may
//	  pay {2}. If they don't, you may copy that spell. You may choose
//	  new targets for the copy."
//	Explore the Vastlands — Sorcery {3}
//	 "Each player looks at the top five cards of their library and may
//	  reveal a land card and/or an instant or sorcery card from among
//	  them. Each player puts the cards they revealed this way into
//	  their hand and the rest on the bottom of their library in a
//	  random order. Each player gains 3 life."
//
// # Two keys, one card
//
// One oracle_id per CARD, so game.CatalogKey makes it composite: the
// creature keeps the bare oracle ID and the sorcery takes
// "<oracle_id>#1", the keyspace every MDFC back face lives in.
//
// # The front face
//
// Its trigger is Rhystic Study's own pay-unless shape one level
// deeper: PayUnless asks the CASTER (ev.Actor, not this controller) to
// pay {2}; declining opens a second, independent "you may copy that
// spell" question to Wandering Archaic's controller (MayChoice, the
// same "you may [do X] mid-resolution" primitive Rhystic Study's draw
// uses), and a "yes" copies the spell with CR 707.10c's "choose new
// targets" prompt.
//
// # The back face
//
// Until #1743 it was left uncatalogued: every "look at the top N"
// shape was one player choosing at most one card against one
// predicate, and this is every player choosing up to two cards against
// two predicates from one look. It is now
// EachPlayerTakesFromLibrary with two TakeSlots:
//
//   - Each player looks at their own five (only they become a knower)
//     and is asked one choose_cards question over the cards that are a
//     land OR an instant or sorcery, at most one per slot. A card that
//     is somehow both fills one slot, not two. A player with neither
//     in their five is not asked.
//   - The questions go up together, APNAP, and are answered privately
//     (CR 101.4a); nothing is revealed until all of them are answered,
//     and then every player's picks are revealed, put into their hand,
//     and the rest of each look goes to the bottom of its library in a
//     random order — that player's own random-order stream.
//   - Then each player still in the game gains 3 life.
//
// No simplifications on either face.
func init() {
	Register(Spec{
		OracleID:     wanderingArchaicOracleID,
		Name:         "Wandering Archaic",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (c.IsInstant() || c.IsSorcery())
			}, "Wandering Archaic — they may pay {2} or you may copy that spell", wanderingArchaicPayOrCopy),
		},
	})
	Register(Spec{
		OracleID:     wanderingArchaicOracleID + "#1",
		Name:         "Explore the Vastlands",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			source := ctx.Source()
			return EachPlayerTakesFromLibrary{
				N: 5,
				Take: TakeFromLibraryToHand{
					Slots: []TakeSlot{
						{Label: "a land card", Match: Land(), Max: 1},
						{Label: "an instant or sorcery card", Match: Or(Instant(), Sorcery()), Max: 1},
					},
					Optional: true,
					Reveal:   true,
					Label:    "Explore the Vastlands — you may reveal a land card and/or an instant or sorcery card",
					Then:     TakeRestOnBottomInRandomOrder,
				},
				Then: func(g *game.Game, _ []TakeFromLibraryResult) error {
					for _, player := range apnapPlayers(g) {
						if err := g.ChangePlayerLifeForEffect(source, player, 3); err != nil {
							return err
						}
					}
					return nil
				},
			}.Apply(ctx)
		},
	})
}

const wanderingArchaicOracleID = "6556c4c0-b10d-4208-821b-0c0a49abd188"

func wanderingArchaicPayOrCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spellID := ctx.Trigger().Event.CardID
	caster := ctx.Trigger().Event.Actor
	controller := item.Controller
	return PayUnless{
		Chooser:  caster,
		Cost:     "{2}",
		Question: "Wandering Archaic — pay {2}?",
		OnDecline: func(ctx *Context) error {
			return MayChoice{
				Player:   controller,
				Question: "Wandering Archaic — copy that spell?",
				OnYes: func(ctx *Context) error {
					return CopySpell{
						StackID:          spellID,
						Controller:       controller,
						ChooseNewTargets: true,
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	}.Apply(ctx)
}
