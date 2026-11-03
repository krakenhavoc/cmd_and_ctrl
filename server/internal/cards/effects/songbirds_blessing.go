package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Songbirds' Blessing — Enchantment — Aura {3}{W} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Enchant creature
//	 Whenever enchanted creature attacks, reveal cards from the top of your
//	 library until you reveal an Aura card. You may put that card onto the
//	 battlefield. If you don't, put it into your hand. Put the rest on the
//	 bottom of your library in a random order."
//
// Aqueous Form's attack trigger; the Aura's controller reveals, whoever
// controls the creature. The reveal is revealUntil (a token is revealed but
// never the hit). The hit is offered as a "you may put" pick
// (PutFromLibraryOntoBattlefield, Optional) and an Aura put onto the
// battlefield without being cast chooses what it enchants as it enters
// (CR 303.4f), without targeting: the engine asks, and with no legal host
// the Aura stays in the library (CR 303.4g), exactly the ruling's "you
// won't be able to put it onto the battlefield". The tail then reads what
// happened (PutFromLibraryTailCarrying): an Aura that did not enter, because
// the player said no or it could not, goes to the hand, and the rest of the
// revealed cards go to the bottom in a random order. With no Aura in the
// library the whole run goes to the bottom.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2598294c-a9a7-4bef-a562-23f297a80536",
		Name:         "Songbirds' Blessing",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			WheneverEnchantedCreatureAttacks("Songbirds' Blessing — reveal until an Aura card; you may put it onto the battlefield, otherwise put it into your hand",
				songbirdsBlessingReveal),
		},
	})
}

func songbirdsBlessingReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	run, hit := revealUntil(ctx, player, func(c game.Card) bool { return c.IsAura() }, "Songbirds' Blessing")
	if hit == uuid.Nil {
		return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, cardsStillInALibrary(g, run))
	}
	return PutFromLibraryOntoBattlefield{
		Player:    player,
		Cards:     run,
		Match:     func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID == hit },
		Max:       1,
		Optional:  true,
		Label:     "Songbirds' Blessing — you may put the Aura card onto the battlefield",
		ThenCarry: []uuid.UUID{hit},
		ThenTail:  songbirdsBlessingTail,
	}.Apply(ctx)
}

// songbirdsBlessingTail is "If you don't, put it into your hand. Put the
// rest on the bottom of your library in a random order", told whether the
// Aura entered. Carry[0] is the revealed Aura card. An on-disk key: never
// renamed, never reused.
var songbirdsBlessingTail = PutFromLibraryTailCarrying("put-from-library/songbirds-blessing-hand-or-bottom", 1,
	func(ctx *Context, res PutFromLibraryResult) error {
		rest := res.Rest
		if len(res.Entered) == 0 && len(res.Carry) == 1 && slices.Contains(rest, res.Carry[0]) {
			ctx.Game.TakeFromLibraryToHandForEffect(res.Player, []uuid.UUID{res.Carry[0]})
			rest = slices.DeleteFunc(slices.Clone(rest), func(id uuid.UUID) bool { return id == res.Carry[0] })
		}
		return ctx.Game.PutOnBottomInRandomOrderForEffect(res.Player, game.ZoneLibrary, rest)
	})
